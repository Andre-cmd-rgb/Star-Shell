BINARY_NAME = starshell
VERSION = 1.0.0
BUILD_DIR = dist
PLATFORMS = darwin linux windows
ARCHITECTURES = amd64 arm64

# Detect the current OS
ifeq ($(OS),Windows_NT)
	CURRENT_OS = windows
	BINARY_EXT = .exe
else
	CURRENT_OS = $(shell uname | tr '[:upper:]' '[:lower:]')
	BINARY_EXT =
endif

.PHONY: all deps build clean run install build-all help

all: deps build

deps:
	go mod tidy

build:
	@echo "Building for current platform: $(CURRENT_OS)"
	go build -o $(BINARY_NAME)$(BINARY_EXT) .

build-all:
	@echo "Building for darwin-amd64"
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_NAME).go
	@echo "Building for darwin-arm64"
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_NAME).go
	@echo "Building for linux-amd64"
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_NAME).go
	@echo "Building for linux-arm64"
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(BINARY_NAME).go
	@echo "Building for windows-amd64"
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_NAME).go
	@echo "Building for windows-arm64"
	GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(BINARY_NAME).go

clean:
	rm -rf $(BUILD_DIR) $(BINARY_NAME) *.exe

run: build
	./$(BINARY_NAME)$(BINARY_EXT)

install: build
	go install

help:
	@echo "Available targets:"
	@echo "  all       - Install dependencies and build"
	@echo "  deps      - Install dependencies"
	@echo "  build     - Build for current platform ($(CURRENT_OS))"
	@echo "  build-all - Build for all platforms/architectures"
	@echo "  clean     - Clean build artifacts"
	@echo "  run       - Build and run"
	@echo "  install   - Install binary"

.DEFAULT_GOAL := help
