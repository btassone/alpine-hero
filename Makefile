.PHONY: all build clean test fmt lint run help build-mac build-linux deps repomix coverage coverage-html coverage-race

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
BINARY_NAME=alpine-hero
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_MAC=$(BINARY_NAME)_mac

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
	@echo "${YELLOW}make build${NC}       - Build the application for current platform"
	@echo "${YELLOW}make clean${NC}       - Clean build artifacts"
	@echo "${YELLOW}make test${NC}        - Run tests"
	@echo "${YELLOW}make coverage${NC}    - Run tests with coverage analysis"
	@echo "${YELLOW}make coverage-html${NC} - Generate and open HTML coverage report"
	@echo "${YELLOW}make coverage-race${NC} - Run tests with race detection and coverage"
	@echo "${YELLOW}make fmt${NC}         - Format Go code"
	@echo "${YELLOW}make lint${NC}        - Run linter"
	@echo "${YELLOW}make run${NC}         - Generate answers.txt file"
	@echo "${YELLOW}make build-linux${NC} - Cross compile for Linux ARM64 (Raspberry Pi)"
	@echo "${YELLOW}make build-mac${NC}   - Cross compile for macOS (both AMD64 and ARM64)"
	@echo "${YELLOW}make deps${NC}        - Install dependencies"
	@echo "${YELLOW}make repomix${NC}     - Generate repomix output file"
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
	rm -f $(BINARY_MAC)*
	rm -f answers.txt
	rm -f repomix-output.txt
	rm -f coverage.out
	rm -f coverage.html

# Run tests
test:
	$(GOTEST) -v ./...

# Test coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: coverage
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Generated coverage report: coverage.html"
	@if [ "$(shell uname)" = "Darwin" ]; then \
		open coverage.html; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html in your browser"; \
	else \
		echo "Please open coverage.html in your browser"; \
	fi

# Run tests with race detection and coverage
coverage-race:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -func=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

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

# Cross compilation for macOS (both AMD64 and ARM64)
build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_MAC)_amd64 -v
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_MAC)_arm64 -v

# Install dependencies
deps:
	$(GOGET) -v ./...

# Generate repomix output
repomix:
	repomix