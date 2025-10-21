#!/bin/bash

# Script to start the test microservice

set -e

echo "Starting Test Microservice"
echo "=========================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "microservice/main.go" ]; then
    echo "Error: microservice/main.go not found. Please run this script from the test directory."
    exit 1
fi

# Check if port 8080 is already in use
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    print_warning "Port 8080 is already in use. Stopping existing process..."
    pkill -f "microservice" || true
    sleep 2
fi

# Install dependencies
print_status "Installing dependencies..."
cd microservice
go mod tidy

# Build the microservice
print_status "Building microservice..."
go build -o microservice main.go

# Start the microservice
print_status "Starting microservice on port 8080..."
./microservice &
MICROSERVICE_PID=$!

# Wait for microservice to start
print_status "Waiting for microservice to start..."
sleep 3

# Check if microservice is running
if curl -s http://localhost:8080/health > /dev/null; then
    print_success "Microservice started successfully!"
    echo ""
    echo "Available endpoints:"
    echo "  - Health check: http://localhost:8080/health"
    echo "  - Metrics: http://localhost:8080/metrics"
    echo "  - API Users: http://localhost:8080/api/users"
    echo "  - API Products: http://localhost:8080/api/products"
    echo "  - API Orders: http://localhost:8080/api/orders"
    echo "  - Admin Status: http://localhost:8080/admin/status"
    echo "  - Toggle Anomalies: POST http://localhost:8080/admin/anomaly"
    echo ""
    echo "Admin commands:"
    echo "  - Enable anomalies: curl -X POST http://localhost:8080/admin/anomaly"
    echo "  - Check status: curl http://localhost:8080/admin/status"
    echo ""
    echo "Press Ctrl+C to stop the microservice"
    
    # Set up signal handling
    trap 'echo ""; print_status "Stopping microservice..."; kill $MICROSERVICE_PID; wait $MICROSERVICE_PID; print_success "Microservice stopped"; exit 0' INT TERM
    
    # Wait for the microservice process
    wait $MICROSERVICE_PID
else
    print_warning "Failed to start microservice. Check the logs above."
    kill $MICROSERVICE_PID 2>/dev/null || true
    exit 1
fi
