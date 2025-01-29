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
	@for GOOS in $(PLATFORMS); do \
		for GOARCH in $(ARCHITECTURES); do \
			echo "Building for $$GOOS-$$GOARCH"; \
			ifeq ($$GOOS,windows) \
				$(MAKE) build-platform GOOS=$$GOOS GOARCH=$$GOARCH; \
			else \
				GOOS=$$GOOS GOARCH=$$GOARCH $(MAKE) build-platform; \
			endif \
		done \
	done

build-platform:
	@echo "Building for platform: $(GOOS)-$(GOARCH)"
	OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH); \
	if [ "$(GOOS)" = "windows" ]; then OUTPUT=$$OUTPUT.exe; fi; \
	go build -o $$OUTPUT .

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
