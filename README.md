# AI Agent Framework

A production-ready, modular AI agent framework built in Go for observability and automation tasks.

## Quick Start

```bash
# Build and run
make build
./agent -create-config
./agent

# Or with Docker
docker build -t agent-framework .
docker run -p 9090:9090 agent-framework

# Or deploy to Kubernetes
kubectl apply -f k8s/
```

## Documentation

All documentation is located in the [`docs/`](docs/) directory:

- **[Main Documentation](docs/README.md)** - Complete framework documentation
- **[Architecture Guide](docs/ARCHITECTURE.md)** - Comprehensive architecture documentation

## Key Features

- **Modular Plugin Architecture**: Extensible plugin system with collectors, analyzers, responders, and AI agents
- **Dependency Injection**: Clean, testable architecture with interface-based design
- **Comprehensive Error Handling**: Structured error types with context and traceability
- **Health Monitoring**: Built-in health checks and monitoring endpoints
- **Configuration Management**: YAML and environment variable support with validation
- **Graceful Shutdown**: Proper resource cleanup and timeout handling
- **Production Ready**: Kubernetes deployment, metrics, and observability built-in

## Testing

```bash
# Run all tests
make test

# Run specific test types
make test-unit
make test-integration
make test-coverage
```

**For complete documentation, see the [`docs/`](docs/) directory.**