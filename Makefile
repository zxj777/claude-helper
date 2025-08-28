# Claude Helper Makefile

.PHONY: build clean install test fmt vet deps help

# Binary name
BINARY_NAME=cchp

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-helper

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Clean build files
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Development build (with race detector)
dev-build:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -race -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-helper

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  deps       - Install dependencies"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build files"
	@echo "  install    - Install binary"
	@echo "  dev-build  - Build with race detector"
	@echo "  help       - Show this help"

# Default target
all: fmt vet test build