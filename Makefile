# Makefile for Observability Framework

.PHONY: help test test-unit test-integration test-bench test-race test-coverage lint vet clean build run

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-unit     - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-bench    - Run benchmarks"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  vet           - Run go vet"
	@echo "  clean         - Clean build artifacts"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  run-interactive - Run in interactive mode"

# Run all tests
test: test-unit test-integration test-bench test-race test-coverage

# Unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./core/... ./plugins/...

# Integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -run TestFullFrameworkIntegration ./integration_test.go

# Benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./plugins/analyzers/...

# Race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./core/... ./plugins/...

# Coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./core/... ./plugins/...
	go tool cover -func=coverage.out | grep total
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Linting
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

# Go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean
clean:
	@echo "Cleaning up..."
	go clean -testcache
	rm -f coverage.out coverage.html
	rm -f agent

# Build
build:
	@echo "Building application..."
	go build -o agent ./cli/main.go

# Run
run: build
	@echo "Running application..."
	./agent -config framework.yaml

# Run interactive
run-interactive: build
	@echo "Running application in interactive mode..."
	./agent -config framework.yaml -interactive

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

# Quick test (unit tests only)
quick-test:
	@echo "Running quick tests..."
	go test -short ./core/... ./plugins/...

# Test specific package
test-pkg:
	@echo "Running tests for package: $(PKG)"
	go test -v ./$(PKG)/...

# Benchmark specific package
bench-pkg:
	@echo "Running benchmarks for package: $(PKG)"
	go test -bench=. -benchmem ./$(PKG)/...

# Example usage:
# make test-pkg PKG=core
# make bench-pkg PKG=plugins/analyzers
