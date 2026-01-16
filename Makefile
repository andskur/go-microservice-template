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