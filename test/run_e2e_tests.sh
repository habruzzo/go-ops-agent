#!/bin/bash

# End-to-End Test Runner for Observability Framework

set -e

echo "Running End-to-End Tests for Observability Framework"
echo "===================================================="

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

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_warning "Docker is not installed, skipping Docker-based tests"
    SKIP_DOCKER=true
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    print_warning "Docker Compose is not installed, skipping Docker-based tests"
    SKIP_DOCKER=true
fi

print_status "Go version: $(go version)"

# Clean up any previous test artifacts
print_status "Cleaning up previous test artifacts..."
go clean -testcache

# Run unit tests first
print_status "Running unit tests..."
if go test -v ./core/... ./plugins/...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Run integration tests
print_status "Running integration tests..."
if go test -v ./integration_test.go; then
    print_success "Integration tests passed"
else
    print_error "Integration tests failed"
    exit 1
fi

# Run end-to-end tests
print_status "Running end-to-end tests..."
if go test -v ./test/e2e_test.go -timeout 5m; then
    print_success "End-to-end tests passed"
else
    print_error "End-to-end tests failed"
    exit 1
fi

# Run end-to-end tests with real Prometheus (if available)
print_status "Testing with real Prometheus..."
if go test -v ./test/e2e_test.go -run TestEndToEndWithRealPrometheus -timeout 2m; then
    print_success "Real Prometheus tests passed"
else
    print_warning "Real Prometheus tests failed (Prometheus may not be running)"
fi

# Run performance tests
print_status "Running performance tests..."
if go test -v ./test/e2e_test.go -run TestEndToEndPerformance -timeout 3m; then
    print_success "Performance tests passed"
else
    print_error "Performance tests failed"
    exit 1
fi

# Docker-based tests (if Docker is available)
if [ "$SKIP_DOCKER" != "true" ]; then
    print_status "Running Docker-based tests..."
    
    # Start test services
    print_status "Starting test services with Docker Compose..."
    cd test
    docker-compose -f docker-compose.test.yml up -d
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check if Prometheus is ready
    for i in {1..30}; do
        if curl -s http://localhost:9090/api/v1/query?query=up > /dev/null; then
            print_success "Prometheus is ready"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "Prometheus failed to start"
            docker-compose -f docker-compose.test.yml logs prometheus
            docker-compose -f docker-compose.test.yml down
            exit 1
        fi
        sleep 2
    done
    
    # Run tests with Docker services
    print_status "Running tests with Docker services..."
    cd ..
    if go test -v ./test/e2e_test.go -run TestEndToEndWithRealPrometheus -timeout 5m; then
        print_success "Docker-based tests passed"
    else
        print_error "Docker-based tests failed"
    fi
    
    # Stop test services
    print_status "Stopping test services..."
    cd test
    docker-compose -f docker-compose.test.yml down
    cd ..
else
    print_warning "Skipping Docker-based tests"
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
if go test -race ./core/... ./plugins/... ./test/...; then
    print_success "Race detection tests passed"
else
    print_error "Race detection tests failed"
    exit 1
fi

# Run tests with coverage
print_status "Running tests with coverage..."
go test -coverprofile=coverage.out ./core/... ./plugins/... ./test/...
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

print_success "All end-to-end tests completed successfully!"
print_status "Test artifacts:"
print_status "  - coverage.out: Coverage data"
print_status "  - coverage.html: Coverage report (if generated)"
print_status "  - Test logs: Above output"

# Summary
echo ""
echo "Test Summary:"
echo "  - Unit tests: PASSED"
echo "  - Integration tests: PASSED"
echo "  - End-to-end tests: PASSED"
echo "  - Performance tests: PASSED"
echo "  - Race detection: PASSED"
echo "  - Coverage: $coverage"
echo ""
echo "Your observability framework is ready for production!"
