# Set APP to the name of the application
APP:=microservice-template

# Set APP_ENTRY_POINT to the main Go file for the application
APP_ENTRY_POINT:=cmd/microservice-template.go

# Set BUILD_OUT_DIR to the directory where the built executable should be placed
BUILD_OUT_DIR:=./

# path to version package
GITVER_PKG:=microservice-template/pkg/version

# Set GOOS and GOARCH to the current system values using the go env command
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# set git related vars for versioning
TAG 		:= $(shell git describe --abbrev=0 --tags)
COMMIT		:= $(shell git rev-parse HEAD)
BRANCH		?= $(shell git rev-parse --abbrev-ref HEAD)
REMOTE		:= $(shell git config --get remote.origin.url)
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

# The all target runs the tidy, build, and test targets
all: tidy build test

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
DATABASE_NAME ?= microservice_dev
DATABASE_SSL_MODE ?= disable

# Construct DATABASE_URL from parts
DATABASE_URL := postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)

.PHONY: migrate-install
migrate-install:
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)

.PHONY: migrate-create
migrate-create:
ifndef NAME
	@echo "Error: NAME parameter is required"
	@echo "Usage: make migrate-create NAME=add_user_roles"
	@exit 1
endif
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "Created migration files in $(MIGRATIONS_DIR)/"

.PHONY: migrate-up
migrate-up:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "Migrations applied successfully"

.PHONY: migrate-down
migrate-down:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "Last migration rolled back"

.PHONY: migrate-force
migrate-force:
ifndef VERSION
	@echo "Error: VERSION parameter is required"
	@echo "Usage: make migrate-force VERSION=2"
	@exit 1
endif
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "Migration version forced to $(VERSION)"

.PHONY: migrate-version
migrate-version:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: migrate-drop
migrate-drop:
	@echo "WARNING: This will drop all tables! Press Ctrl+C to cancel, Enter to continue..."
	@read _
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" drop -f
	@echo "All migrations dropped"

# The run target runs the application with race detection enabled
run:
	GODEBUG=xray_ptrace=1 go run -race $(APP_ENTRY_POINT) serve

# The build target builds the application for the current system
build:
	env CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s ${LDFLAGS}" -o $(BUILD_OUT_DIR)/$(APP) $(APP_ENTRY_POINT)

# The test target runs go test
test:
	go test ./...

# The clean target deletes the build output file
clean:
	rm $(BUILD_OUT_DIR)/$(APP)

# Docker Compose helpers
.PHONY: compose-up
compose-up:
	docker-compose up -d

.PHONY: compose-down
compose-down:
	docker-compose down

.PHONY: compose-restart
compose-restart:
	docker-compose down
	docker-compose up -d

# The test-coverage target runs go test with coverage enabled and generates a coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# The lint target runs golangci-lint to check for common style and code quality issues
lint:
	golangci-lint run ./...

# The lint-install target installs golangci-lint if not already installed
lint-install:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)

.PHONY: rename
rename:
ifndef NEW_NAME
	@echo "Error: NEW_NAME parameter is required"
	@echo "Usage: make rename NEW_NAME=my-new-service"
	@exit 1
endif
	@bash scripts/rename.sh "$(NEW_NAME)"
