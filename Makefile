# Makefile for Observability Framework

.PHONY: bench-pkg build clean deps dev-tools help lint quick-test run run-interactive test test-bench test-coverage test-integration test-pkg test-race test-unit vet

# Default target
help:
	@echo "Available targets:"
	@echo "  bench-pkg     - Run benchmarks for specific package (use PKG=package)"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  dev-tools     - Install development tools"
	@echo "  lint          - Run linter"
	@echo "  quick-test    - Run quick tests (unit tests only)"
	@echo "  run           - Run the application"
	@echo "  run-interactive - Run in interactive mode"
	@echo "  test          - Run all tests"
	@echo "  test-bench    - Run benchmarks"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-pkg      - Run tests for specific package (use PKG=package)"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-unit     - Run unit tests only"
	@echo "  vet           - Run go vet"

# Benchmark specific package
bench-pkg:
	@echo "Running benchmarks for package: $(PKG)"
	go test -bench=. -benchmem ./$(PKG)/...

# Build
build:
	@echo "Building application..."
	go build -o agent ./cli/main.go

# Clean
clean:
	@echo "Cleaning up..."
	go clean -testcache
	rm -f coverage.out coverage.html
	rm -f agent

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Linting
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

# Quick test (unit tests only)
quick-test:
	@echo "Running quick tests..."
	go test -short ./core/... ./plugins/...

# Run
run: build
	@echo "Running application..."
	./agent -config framework.yaml

# Run interactive
run-interactive: build
	@echo "Running application in interactive mode..."
	./agent -config framework.yaml -interactive

# Run all tests
test: test-unit test-integration test-bench test-race test-coverage

# Benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./plugins/analyzers/...

# Coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./core/... ./plugins/...
	go tool cover -func=coverage.out | grep total
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -run TestFullFrameworkIntegration ./integration_test.go

# Test specific package
test-pkg:
	@echo "Running tests for package: $(PKG)"
	go test -v ./$(PKG)/...

# Race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./core/... ./plugins/...

# Unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./core/... ./plugins/...

# Go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Example usage:
# make test-pkg PKG=core
# make bench-pkg PKG=plugins/analyzers
