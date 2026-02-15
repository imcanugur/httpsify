.PHONY: build clean test run install lint fmt help

# Build variables
BINARY_NAME := httpsify
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -s -w"

# Go variables
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/httpsify

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/httpsify
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/httpsify
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/httpsify
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/httpsify
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/httpsify

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -rf cert/

## test: Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet only..."; \
		go vet ./...; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## run: Run the server with self-signed certs
run: build
	@echo "Starting httpsify..."
	sudo ./$(BINARY_NAME) --self-signed --verbose

## run-nonroot: Run on port 8443 without root
run-nonroot: build
	@echo "Starting httpsify on :8443..."
	./$(BINARY_NAME) --self-signed --listen :8443 --verbose

## install: Install to GOBIN
install:
	@echo "Installing $(BINARY_NAME) to $(GOBIN)..."
	go install $(LDFLAGS) ./cmd/httpsify

## cert: Generate certificates using mkcert
cert:
	@echo "Generating certificates with mkcert..."
	@mkdir -p cert
	mkcert -cert-file cert/localhost.pem -key-file cert/localhost-key.pem \
		localhost "*.localhost" localtest.me "*.localtest.me" 127.0.0.1 ::1
	@echo "Certificates generated in cert/"

## cert-self: Generate self-signed certificates
cert-self: build
	@echo "Generating self-signed certificates..."
	./$(BINARY_NAME) --self-signed --listen :0 &
	@sleep 2
	@pkill -f "$(BINARY_NAME) --self-signed" || true
	@echo "Self-signed certificates generated in cert/"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

## version: Show version
version:
	@echo "$(VERSION)"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'

# Install tools
.PHONY: tools
tools:
	@echo "Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
