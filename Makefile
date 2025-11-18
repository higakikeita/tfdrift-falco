.PHONY: build test clean install lint fmt help

# Variables
BINARY_NAME=tfdrift
VERSION?=0.1.0
BUILD_DIR=./bin
GO=go
GOFLAGS=-v

# Build information
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

## help: Display this help message
help:
	@echo "TFDrift-Falco Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /' | column -t -s ':'

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tfdrift

## build-all: Build binaries for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/tfdrift

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/tfdrift
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/tfdrift

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/tfdrift

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) ./cmd/tfdrift

## test: Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -cover -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-coverage-threshold: Run tests and check coverage threshold
test-coverage-threshold:
	@echo "Running tests with coverage threshold check..."
	$(GO) test -coverprofile=coverage.out -covermode=atomic ./...
	@COVERAGE=$$($(GO) tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${COVERAGE}%"; \
	THRESHOLD=30.0; \
	if [ $$(echo "$${COVERAGE} < $${THRESHOLD}" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $${COVERAGE}% is below threshold $${THRESHOLD}%"; \
		exit 1; \
	else \
		echo "✅ Coverage $${COVERAGE}% meets threshold $${THRESHOLD}%"; \
	fi

## test-race: Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GO) test -race ./...

## test-short: Run short tests only
test-short:
	@echo "Running short tests..."
	$(GO) test -short -v ./...

## lint: Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not found, skipping..."

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t tfdrift-falco:$(VERSION) -t tfdrift-falco:latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -v $(PWD)/config.yaml:/config.yaml tfdrift-falco:latest --config /config.yaml

## run: Run the application locally
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) --config examples/config.yaml

## run-dry: Run in dry-run mode
run-dry: build
	@echo "Running $(BINARY_NAME) in dry-run mode..."
	$(BUILD_DIR)/$(BINARY_NAME) --config examples/config.yaml --dry-run

## init: Initialize development environment
init:
	@echo "Initializing development environment..."
	$(GO) mod download
	@echo "Installing development tools..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment ready!"

## check: Run all checks (fmt, lint, test)
check: fmt lint test
	@echo "All checks passed!"

## ci: Run all CI checks locally
ci: deps fmt lint test-coverage-threshold test-race
	@echo "✅ All CI checks passed!"

## ci-local: Quick CI checks without race detector (faster)
ci-local: fmt lint test-coverage
	@echo "✅ Local CI checks passed!"

.DEFAULT_GOAL := help
