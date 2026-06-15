.PHONY: all build build-all test bench clean fmt lint vet mod coverage release release-snapshot help

# Variables
BINARY_NAME := pptx-to-md
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT) -w -s"

# Go tools versions
GOLANGCI_LINT_VERSION := v2.11.4

# Default target
all: clean lint test build

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^##' Makefile | sed -E 's/^## /  /'

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Multi-platform build complete"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race -timeout 30s ./...

## test-short: Run short tests
test-short:
	@echo "Running short tests..."
	@go test -short -v ./...

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./...

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -v -race -coverprofile='coverage.out' -covermode='atomic' './...'
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

## lint: Run linters
lint:
	@echo "Running linters..."
	@GOBIN=$$(go env GOPATH)/bin; \
	if ! command -v $$GOBIN/golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$GOBIN $(GOLANGCI_LINT_VERSION); \
	fi; \
	$$GOBIN/golangci-lint run --timeout 5m

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## mod: Download and tidy modules
mod:
	@echo "Downloading and tidying modules..."
	@go mod download
	@go mod tidy
	@go mod verify

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) dist/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## release: Create release with goreleaser
release:
	@echo "Running goreleaser..."
	@curl -sSfL https://goreleaser.com/static/run | bash -s -- release --clean

## release-snapshot: Create snapshot release
release-snapshot:
	@echo "Running goreleaser snapshot..."
	@curl -sSfL https://goreleaser.com/static/run | bash -s -- release --snapshot --clean

## ci: Run CI pipeline locally
ci: clean mod fmt vet lint test build
	@echo "CI pipeline complete"

## check: Quick check before commit
check: fmt vet test-short
	@echo "Pre-commit checks complete"
