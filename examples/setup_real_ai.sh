#!/bin/bash

# Setup script for Real AI Agent

set -e

echo "Setting up Real AI Agent for Observability"
echo "=========================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

print_success "Go is installed: $(go version)"

# Check if we're in the right directory
if [ ! -f "real_ai_agent.go" ]; then
    print_error "real_ai_agent.go not found. Please run this script from the examples directory."
    exit 1
fi

# Check for OpenAI API key
if [ -z "$OPENAI_API_KEY" ]; then
    print_warning "OPENAI_API_KEY environment variable is not set."
    echo ""
    echo "To get your API key:"
    echo "1. Go to https://platform.openai.com/api-keys"
    echo "2. Sign up or log in"
    echo "3. Create a new API key"
    echo "4. Copy the key (starts with 'sk-')"
    echo ""
    echo "Then run:"
    echo "export OPENAI_API_KEY=sk-your-actual-key-here"
    echo ""
    echo "Or add it to your ~/.bashrc or ~/.zshrc for persistence."
    echo ""
    read -p "Do you want to continue without an API key? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    print_success "OpenAI API key is set"
fi

# Install dependencies
print_status "Installing dependencies..."
go mod tidy

# Build the agent
print_status "Building AI agent..."
go build -o real_ai_agent real_ai_agent.go

print_success "AI agent built successfully!"

# Check if Prometheus is running
print_status "Checking for Prometheus..."
if curl -s http://localhost:9090/api/v1/query?query=up > /dev/null 2>&1; then
    print_success "Prometheus is running on port 9090"
    PROMETHEUS_AVAILABLE=true
else
    print_warning "Prometheus not found on port 9090"
    print_status "You can start Prometheus with:"
    echo "  docker run -p 9090:9090 prom/prometheus"
    PROMETHEUS_AVAILABLE=false
fi

# Check if test microservice is running
print_status "Checking for test microservice..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    print_success "Test microservice is running on port 8080"
    MICROSERVICE_AVAILABLE=true
else
    print_warning "Test microservice not found on port 8080"
    print_status "You can start it with:"
    echo "  cd ../test && ./start_microservice.sh"
    MICROSERVICE_AVAILABLE=false
fi

echo ""
echo "Setup Complete!"
echo "==============="
echo ""
echo "To run your AI agent:"
echo "  ./real_ai_agent"
echo ""
echo "Available data sources:"
if [ "$PROMETHEUS_AVAILABLE" = true ]; then
    echo "  ✓ Prometheus metrics (port 9090)"
else
    echo "  ✗ Prometheus metrics (not running)"
fi

if [ "$MICROSERVICE_AVAILABLE" = true ]; then
    echo "  ✓ Test microservice (port 8080)"
else
    echo "  ✗ Test microservice (not running)"
fi

echo ""
echo "Example queries to try:"
echo "  - 'What's the current system status?'"
echo "  - 'Are there any anomalies?'"
echo "  - 'Give me performance recommendations'"
echo "  - 'Analyze the system health'"
echo ""

if [ -z "$OPENAI_API_KEY" ]; then
    print_warning "Remember to set your OPENAI_API_KEY before running the agent!"
fi

echo "For more information, see: REAL_AI_AGENT_GUIDE.md"
