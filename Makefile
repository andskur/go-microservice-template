APP:=template-service
COMMON_PATH	?= $(shell pwd)
APP_ENTRY_POINT:=cmd/template-service.go
BUILD_OUT_DIR:=./

GOOS	:=
GOARCH	:=
ifeq ($(OS),Windows_NT)
	GOOS =windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		OSFLAG =amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		OSFLAG =ia32
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		GOOS =linux
	endif
	ifeq ($(UNAME_S),Darwin)
		GOOS =darwin
	endif
		UNAME_P := $(shell uname -m)
	ifeq ($(UNAME_P),x86_64)
		GOARCH =amd64
	endif
	ifneq ($(filter %86,$(UNAME_P)),)
		GOARCH =386
	endif
	ifneq ($(filter arm%,$(UNAME_P)),)
		GOARCH =arm64
	endif
endif

RELEASE :=
ifeq ($(TAG),)
	RELEASE := $(COMMIT)
else
	RELEASE := $(TAG)
endif

all: tidy build

tidy:
	go mod tidy

update:
	go get -u ./...

run:
	MallocNanoZone=0 go run -race $(APP_ENTRY_POINT) serve

build:
	env CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s ${LDFLAGS}" -o $(BUILD_OUT_DIR)/$(APP) $(APP_ENTRY_POINT)

test:
	go test ./...
