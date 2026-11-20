# Set APP to the name of the application (binary name)
APP:=microservice-template

# Set APP_ENTRY_POINT to the main Go file for the application
APP_ENTRY_POINT:=cmd/microservice-template.go

# Set BUILD_OUT_DIR to the directory where the built executable should be placed
BUILD_OUT_DIR:=./

# Go module path, read from go.mod so it never drifts after `make rename`
GO_MODULE:=$(shell awk '/^module /{print $$2; exit}' go.mod)

# path to version package
GITVER_PKG:=$(GO_MODULE)/pkg/version

# Set GOOS and GOARCH to the current system values using the go env command
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# set git related vars for versioning (tolerate repos without tags or commits)
TAG 		:= $(shell git describe --abbrev=0 --tags 2>/dev/null || true)
COMMIT		:= $(shell git rev-parse HEAD 2>/dev/null || true)
BRANCH		?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
REMOTE		:= $(shell git config --get remote.origin.url 2>/dev/null || true)
BUILD_DATE	:= $(shell date +'%Y-%m-%dT%H:%M:%SZ%Z')

# Set RELEASE to either the current TAG or COMMIT
RELEASE :=
ifeq ($(TAG),)
	RELEASE := $(COMMIT)
else
	RELEASE := $(TAG)
endif

# append versioner vars to ldflags
LDFLAGS += -X $(GITVER_PKG).ServiceName=$(APP)
LDFLAGS += -X $(GITVER_PKG).CommitTag=$(TAG)
LDFLAGS += -X $(GITVER_PKG).CommitSHA=$(COMMIT)
LDFLAGS += -X $(GITVER_PKG).CommitBranch=$(BRANCH)
LDFLAGS += -X $(GITVER_PKG).OriginURL=$(REMOTE)
LDFLAGS += -X $(GITVER_PKG).BuildDate=$(BUILD_DATE)
LDFLAGS += -X $(GITVER_PKG).Release=$(RELEASE)

# -----------------------------------------------------------------------------
# Tooling: every tool is pinned and installed into ./bin (git-ignored), and
# ./bin is put first on PATH for all recipes, so the pinned versions are used
# even if other versions are installed globally.
# -----------------------------------------------------------------------------
BIN_DIR := $(CURDIR)/bin
export PATH := $(BIN_DIR):$(PATH)

SWAGGER_VERSION            := v0.36.8
BUF_VERSION                := v1.75.0
PROTOC_GEN_GO_VERSION      := v1.36.12
PROTOC_GEN_GO_GRPC_VERSION := v1.6.2
MIGRATE_VERSION            := v4.20.1
GOLANGCI_LINT_VERSION      := v2.15.0
GOVULNCHECK_VERSION        := v1.10.0

# Tool binaries (invoked by full path: make does not use the exported PATH to
# look up commands it executes directly)
SWAGGER       := $(BIN_DIR)/swagger
BUF           := $(BIN_DIR)/buf
MIGRATE       := $(BIN_DIR)/migrate
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOVULNCHECK   := $(BIN_DIR)/govulncheck

# Stamp files record the installed version; changing a version reinstalls it.
$(BIN_DIR)/.swagger-$(SWAGGER_VERSION):
	GOBIN=$(BIN_DIR) go install github.com/go-swagger/go-swagger/cmd/swagger@$(SWAGGER_VERSION)
	@touch $@

$(BIN_DIR)/.buf-$(BUF_VERSION):
	GOBIN=$(BIN_DIR) go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@touch $@

$(BIN_DIR)/.protoc-gen-go-$(PROTOC_GEN_GO_VERSION):
	GOBIN=$(BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@touch $@

$(BIN_DIR)/.protoc-gen-go-grpc-$(PROTOC_GEN_GO_GRPC_VERSION):
	GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@touch $@

$(BIN_DIR)/.migrate-$(MIGRATE_VERSION):
	GOBIN=$(BIN_DIR) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	@touch $@

$(BIN_DIR)/.golangci-lint-$(GOLANGCI_LINT_VERSION):
	GOBIN=$(BIN_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@touch $@

$(BIN_DIR)/.govulncheck-$(GOVULNCHECK_VERSION):
	GOBIN=$(BIN_DIR) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@touch $@

# Code generation tools (needed for build/test/docker)
.PHONY: tools
tools: $(BIN_DIR)/.swagger-$(SWAGGER_VERSION) \
	$(BIN_DIR)/.buf-$(BUF_VERSION) \
	$(BIN_DIR)/.protoc-gen-go-$(PROTOC_GEN_GO_VERSION) \
	$(BIN_DIR)/.protoc-gen-go-grpc-$(PROTOC_GEN_GO_GRPC_VERSION)

# Backwards-compatible aliases
.PHONY: proto-install buf-install swagger-install
proto-install buf-install swagger-install: tools

.PHONY: migrate-install
migrate-install: $(BIN_DIR)/.migrate-$(MIGRATE_VERSION)

.PHONY: lint-install
lint-install: $(BIN_DIR)/.golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: vuln-install
vuln-install: $(BIN_DIR)/.govulncheck-$(GOVULNCHECK_VERSION)

# The all target generates code, then runs the tidy, build, and test targets
all: generate tidy build test

# The tidy target runs go mod tidy
tidy:
	go mod tidy

# The update target runs go get -u
update:
	go get -u ./...

# Migration configuration
MIGRATIONS_DIR := ./db/migrations

# Default database connection values (override via env vars)
DATABASE_HOST ?= localhost
DATABASE_PORT ?= 5432
DATABASE_USER ?= dev
DATABASE_PASSWORD ?= dev
DATABASE_NAME ?= microservice_template_dev
DATABASE_SSL_MODE ?= disable

# Construct DATABASE_URL from parts
DATABASE_URL := postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)

.PHONY: migrate-create
migrate-create: migrate-install
ifndef NAME
	@echo "Error: NAME parameter is required"
	@echo "Usage: make migrate-create NAME=add_user_roles"
	@exit 1
endif
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "Created migration files in $(MIGRATIONS_DIR)/"

.PHONY: migrate-up
migrate-up: migrate-install
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "Migrations applied successfully"

.PHONY: migrate-down
migrate-down: migrate-install
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "Last migration rolled back"

.PHONY: migrate-force
migrate-force: migrate-install
ifndef VERSION
	@echo "Error: VERSION parameter is required"
	@echo "Usage: make migrate-force VERSION=2"
	@exit 1
endif
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "Migration version forced to $(VERSION)"

.PHONY: migrate-version
migrate-version: migrate-install
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: migrate-drop
migrate-drop: migrate-install
	@echo "WARNING: This will drop all tables! Press Ctrl+C to cancel, Enter to continue..."
	@read _
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" drop -f
	@echo "All migrations dropped"

# The run target runs the application with race detection enabled
run:
	go run -race $(APP_ENTRY_POINT) serve

# The build target builds the application for the current system
build:
	env CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s ${LDFLAGS}" -o $(BUILD_OUT_DIR)/$(APP) $(APP_ENTRY_POINT)

# The test target runs go test with the race detector
test:
	go test -race ./...

.PHONY: test-with-gen
test-with-gen: generate test

# gRPC integration tests (runs only gRPC package tests, including integration)
.PHONY: test-grpc
test-grpc:
	go test ./internal/grpc -count=1

# The clean target deletes the build output file
clean:
	rm -f $(BUILD_OUT_DIR)/$(APP)

# Docker Compose helpers
.PHONY: compose-up
compose-up:
	docker compose up -d

.PHONY: compose-down
compose-down:
	docker compose down

.PHONY: compose-restart
compose-restart:
	docker compose down
	docker compose up -d

# The test-coverage target runs go test with coverage enabled and generates a coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# The lint target runs the pinned golangci-lint v2
.PHONY: lint
lint: lint-install
	$(GOLANGCI_LINT) run ./...

# Run the pinned golangci-lint in Docker (useful when the local Go toolchain is
# newer than the one golangci-lint supports). Generate code first.
.PHONY: lint-docker
lint-docker:
	docker run --rm -v "$(CURDIR)":/src -w /src golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run ./...

.PHONY: lint-with-gen
lint-with-gen: generate lint

# Vulnerability scan (reachable vulnerabilities only)
.PHONY: vuln
vuln: vuln-install
	$(GOVULNCHECK) ./...

# -----------------------------------------------------------------------------
# Code generation. Generated code is not committed; `make generate` is the only
# entry point and is used locally, in CI, in the Dockerfile and by `make rename`.
# -----------------------------------------------------------------------------
PROTO_DIR := ./protocols
EXAMPLE_PROTO_DIR := ./api/proto
PROTO_REPO ?= https://github.com/andskur/protocols-template.git
PROTO_BRANCH ?= main

.PHONY: generate
generate: generate-proto generate-api

.PHONY: generate-all
generate-all: generate

# Generates the bundled example API (api/proto) and, when present, the shared
# protocols subtree (protocols/, see buf.gen.yaml for the go_package override).
.PHONY: generate-proto
generate-proto: tools proto-clean
	@echo "Generating Go code from $(EXAMPLE_PROTO_DIR)..."
	@cd $(EXAMPLE_PROTO_DIR) && $(BUF) generate
	@if [ -d $(PROTO_DIR) ]; then \
		echo "Generating Go code from $(PROTO_DIR) (go_package prefix $(GO_MODULE)/protocols)..."; \
		$(BUF) generate $(PROTO_DIR) --template buf.gen.yaml; \
	fi
	@echo "Proto generation complete"

# Backwards-compatible aliases
.PHONY: proto-generate-all buf-generate-all
proto-generate-all buf-generate-all: generate-proto

# Generate a single package of the protocols subtree, e.g. PROTO_PACKAGE=user
.PHONY: proto-generate buf-generate
proto-generate buf-generate: tools
ifndef PROTO_PACKAGE
	@echo "Error: PROTO_PACKAGE parameter is required"
	@echo "Usage: make proto-generate PROTO_PACKAGE=user"
	@exit 1
endif
	@test -d $(PROTO_DIR)/$(PROTO_PACKAGE) || (echo "Error: $(PROTO_DIR)/$(PROTO_PACKAGE) directory not found" && exit 1)
	$(BUF) generate $(PROTO_DIR) --template buf.gen.yaml --path $(PROTO_DIR)/$(PROTO_PACKAGE)

.PHONY: proto-setup
proto-setup:
	@if [ -e $(PROTO_DIR) ]; then \
		echo "Error: $(PROTO_DIR) already exists; use 'make proto-update' to pull changes"; exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: working tree has uncommitted changes; commit or stash them first (required by git subtree)"; exit 1; \
	fi
	@echo "Adding protobuf subtree from $(PROTO_REPO) ($(PROTO_BRANCH)) at $(PROTO_DIR)..."
	git subtree add --prefix=$(PROTO_DIR:./%=%) $(PROTO_REPO) $(PROTO_BRANCH) --squash
	@echo "Subtree added. Run 'make generate' to generate Go code under $(GO_MODULE)/protocols/..."

.PHONY: proto-update
proto-update:
	@echo "Updating protobuf subtree from $(PROTO_REPO)..."
	git subtree pull --prefix=$(PROTO_DIR:./%=%) $(PROTO_REPO) $(PROTO_BRANCH) --squash
	@echo "Subtree updated successfully"

.PHONY: buf-lint
buf-lint: tools
	cd $(EXAMPLE_PROTO_DIR) && $(BUF) lint
	@if [ -d $(PROTO_DIR) ]; then $(BUF) lint $(PROTO_DIR); fi

.PHONY: buf-breaking
buf-breaking: tools
	@test -d $(PROTO_DIR) || (echo "Error: $(PROTO_DIR) directory not found" && exit 1)
	$(BUF) breaking $(PROTO_DIR) --against '.git#branch=main,subdir=protocols'

.PHONY: proto-clean
proto-clean:
	@echo "Cleaning generated proto files..."
	@find $(EXAMPLE_PROTO_DIR) -name "*.pb.go" -type f -delete
	@if [ -d $(PROTO_DIR) ]; then find $(PROTO_DIR) -name "*.pb.go" -type f -delete; fi

# Swagger/HTTP API targets
.PHONY: swagger-validate
swagger-validate: tools
	@echo "Validating swagger specification..."
	@$(SWAGGER) validate api/swagger.yaml
	@echo "Swagger spec is valid"

# The server package is regenerated from scratch every time (it contains an
# APP-specific configure_*.go file), then internal/http/module.go is synced to
# the API type name go-swagger actually generated.
.PHONY: generate-api
generate-api: tools swagger-clean
	@echo "Generating API server from swagger spec..."
	@$(SWAGGER) generate server --quiet \
		-A $(APP)-api \
		-P models.User \
		--server-package server \
		-f ./api/swagger.yaml \
		--exclude-main \
		--keep-spec-order \
		--flag-strategy pflag \
		--target ./internal/http \
		--spec ./api/swagger.yaml
	@bash scripts/sync-swagger-api.sh
	@echo "API generation complete"

.PHONY: swagger-clean
swagger-clean:
	@rm -rf internal/http/server internal/http/models

.PHONY: clean-generated
clean-generated: proto-clean swagger-clean

.PHONY: test-http
test-http:
	@echo "Running HTTP module tests..."
	@go test -v -race -count=1 ./internal/http/...

# Template synchronization
TEMPLATE_REMOTE_NAME := template
TEMPLATE_REMOTE_URL ?= https://github.com/andskur/go-microservice-template.git
TEMPLATE_BRANCH ?= main

.PHONY: template-setup
template-setup:
	@bash scripts/template-sync.sh setup

.PHONY: template-status
template-status:
	@bash scripts/template-sync.sh status

.PHONY: template-fetch
template-fetch:
	@bash scripts/template-sync.sh fetch

.PHONY: template-diff
template-diff:
	@bash scripts/template-sync.sh diff

.PHONY: template-sync
template-sync:
	@bash scripts/template-sync.sh sync

.PHONY: rename
rename:
ifndef NEW_NAME
	@echo "Error: NEW_NAME parameter is required"
	@echo "Usage: make rename NEW_NAME=my-new-service"
	@echo "   or: make rename NEW_NAME=github.com/your-org/my-new-service"
	@exit 1
endif
	@bash scripts/rename.sh "$(NEW_NAME)"
