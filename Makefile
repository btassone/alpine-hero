.PHONY: all build clean test fmt lint run help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
BINARY_NAME=alpine-template
BINARY_UNIX=$(BINARY_NAME)_unix

# Colors for help output
YELLOW := \033[1;33m
NC := \033[0m # No Color

# Default target when just running 'make'
.DEFAULT_GOAL := help

# Help target
help:
	@echo "Available targets:"
	@echo "${YELLOW}make help${NC}        - Show this help message"
	@echo "${YELLOW}make all${NC}         - Run tests and build the application"
	@echo "${YELLOW}make build${NC}       - Build the application"
	@echo "${YELLOW}make clean${NC}       - Clean build artifacts"
	@echo "${YELLOW}make test${NC}        - Run tests"
	@echo "${YELLOW}make fmt${NC}         - Format Go code"
	@echo "${YELLOW}make lint${NC}        - Run linter"
	@echo "${YELLOW}make run${NC}         - Generate answers.txt file"
	@echo "${YELLOW}make build-linux${NC} - Cross compile for Linux ARM64 (Raspberry Pi)"
	@echo "${YELLOW}make deps${NC}        - Install dependencies"
	@echo "\nExample usage:"
	@echo "  make build && ./$(BINARY_NAME) > answers.txt"

# Build all
all: test build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f answers.txt

# Run tests
test:
	$(GOTEST) -v ./...

# Format code
fmt:
	go fmt ./...

# Run linting
lint:
	if [ ! -f $(GOPATH)/bin/golangci-lint ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin; \
	fi
	golangci-lint run

# Generate answers file
run:
	$(GORUN) main.go > answers.txt

# Cross compilation for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINARY_UNIX) -v

# Install dependencies
deps:
	$(GOGET) -v ./...