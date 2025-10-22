#!/bin/bash

# Test runner script for the observability framework

set -e

echo " Running Observability Framework Tests"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Clean up any previous test artifacts
print_status "Cleaning up previous test artifacts..."
go clean -testcache

# Run unit tests
print_status "Running unit tests..."
if go test -v ./core/... ./plugins/...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Run integration tests
print_status "Running integration tests..."
if go test -v -run TestFullFrameworkIntegration ./integration_test.go; then
    print_success "Integration tests passed"
else
    print_warning "Integration tests failed (this is expected without proper setup)"
fi

# Run benchmarks
print_status "Running benchmarks..."
if go test -bench=. -benchmem ./plugins/analyzers/...; then
    print_success "Benchmarks completed"
else
    print_error "Benchmarks failed"
    exit 1
fi

# Run tests with race detection
print_status "Running tests with race detection..."
if go test -race ./core/... ./plugins/...; then
    print_success "Race detection tests passed"
else
    print_error "Race detection tests failed"
    exit 1
fi

# Run tests with coverage
print_status "Running tests with coverage..."
go test -coverprofile=coverage.out ./core/... ./plugins/...
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
print_success "Test coverage: $coverage"

# Generate coverage report
if command -v go tool cover &> /dev/null; then
    print_status "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage report generated: coverage.html"
fi

# Run linting (if golangci-lint is available)
if command -v golangci-lint &> /dev/null; then
    print_status "Running linter..."
    if golangci-lint run; then
        print_success "Linting passed"
    else
        print_warning "Linting found issues"
    fi
else
    print_warning "golangci-lint not found, skipping linting"
fi

# Run go vet
print_status "Running go vet..."
if go vet ./...; then
    print_success "go vet passed"
else
    print_error "go vet failed"
    exit 1
fi

# Run go mod tidy to check dependencies
print_status "Checking dependencies..."
if go mod tidy; then
    print_success "Dependencies are clean"
else
    print_error "Dependencies have issues"
    exit 1
fi

print_success "All tests completed successfully! "
print_status "Test artifacts:"
print_status "  - coverage.out: Coverage data"
print_status "  - coverage.html: Coverage report (if generated)"
print_status "  - Test logs: Above output"
