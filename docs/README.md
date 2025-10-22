# AI Agent Framework

A production-ready, modular AI agent framework built in Go for observability and automation tasks. The framework has been **completely refactored** to follow modern software engineering principles with dependency injection, comprehensive error handling, and extensive configurability.

## Features

- **Modular Plugin Architecture**: Extensible plugin system with collectors, analyzers, responders, and AI agents
- **Dependency Injection**: Clean, testable architecture with interface-based design
- **Comprehensive Error Handling**: Structured error types with context and traceability
- **Health Monitoring**: Built-in health checks and monitoring endpoints
- **Configuration Management**: YAML and environment variable support with validation
- **Graceful Shutdown**: Proper resource cleanup and timeout handling
- **Production Ready**: Kubernetes deployment, metrics, and observability built-in

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Agent Framework                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Collectors  │  │ Analyzers   │  │ Responders  │         │
│  │             │  │             │  │             │         │
│  │ • Prometheus│  │ • Anomaly   │  │ • Logger    │         │
│  │ • Custom    │  │ • Trend     │  │ • Alert     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         │                │                │                │
│         └────────────────┼────────────────┘                │
│                          │                                 │
│  ┌─────────────────────────────────────────────────────────┤
│  │              AI Agents                                 │
│  │  • Query Processing  • Decision Making                 │
│  │  • Context Awareness • Action Planning                 │
│  └─────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────────┤
│  │              Core Framework                             │
│  │  • Plugin Registry  • Health Monitoring                │
│  │  • Event Bus       • Configuration Management          │
│  └─────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Build and Run

```bash
# Clone the repository
git clone <repository-url>
cd agent

# Build the application
make build

# Create default configuration
./agent -create-config

# Run the agent
./agent
```

### 2. Docker

```bash
# Build Docker image
docker build -t agent-framework .

# Run with Docker
docker run -p 9090:9090 agent-framework
```

### 3. Kubernetes

```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n agent-framework
```

## Configuration

### YAML Configuration

```yaml
# framework.yaml
logging:
  level: info
  format: text
  output: stdout

server:
  host: 0.0.0.0
  port: 9090

plugins:
  - name: prometheus
    type: prometheus
    enabled: true
    config:
      url: http://localhost:9090
      interval: 30s
      queries:
        - up
        - cpu_usage_percent

  - name: anomaly-detector
    type: anomaly
    enabled: true
    config:
      threshold: 2.0

  - name: ai-agent
    type: ai
    enabled: true
    config:
      api_key: ${AGENT_AI_API_KEY}
      model: gpt-3.5-turbo

agent:
  default_agent: ai-agent
```

### Environment Variables

```bash
# Logging
export AGENT_LOG_LEVEL=info
export AGENT_LOG_FORMAT=json

# Server
export AGENT_SERVER_HOST=0.0.0.0
export AGENT_SERVER_PORT=9090

# Plugins
export AGENT_PROMETHEUS_ENABLED=true
export AGENT_PROMETHEUS_URL=http://localhost:9090
export AGENT_AI_API_KEY=your-api-key-here
```

##  Plugin Development

### Creating a Custom Collector

```go
package main

import (
    "context"
    "time"
    "github.com/habruzzo/agent/core"
)

type CustomCollector struct {
    name     string
    status   core.PluginStatus
    interval time.Duration
}

func (c *CustomCollector) Name() string { return c.name }
func (c *CustomCollector) Type() core.PluginType { return core.PluginTypeCollector }
func (c *CustomCollector) Version() string { return "1.0.0" }

func (c *CustomCollector) Configure(config map[string]interface{}) error {
    // Configure the collector
    return nil
}

func (c *CustomCollector) Start(ctx context.Context) error {
    c.status = core.PluginStatusRunning
    return nil
}

func (c *CustomCollector) Stop() error {
    c.status = core.PluginStatusStopped
    return nil
}

func (c *CustomCollector) Status() core.PluginStatus { return c.status }
func (c *CustomCollector) Health(ctx context.Context) error { return nil }
func (c *CustomCollector) GetCapabilities() []string { return []string{"custom-data"} }

func (c *CustomCollector) Collect(ctx context.Context) ([]core.DataPoint, error) {
    // Collect data from your source
    return []core.DataPoint{
        {
            Timestamp: time.Now(),
            Source:    c.name,
            Metric:    "custom_metric",
            Value:     1.0,
            Labels:    map[string]string{"source": "custom"},
        },
    }, nil
}

func (c *CustomCollector) GetCollectionInterval() time.Duration {
    return c.interval
}
```

### Registering Your Plugin

```go
// In your main function or plugin registration
factory.RegisterPluginCreator("custom", func(config core.PluginConfig) (core.Plugin, error) {
    collector := &CustomCollector{
        name:     config.Name,
        interval: 30 * time.Second,
    }
    if err := collector.Configure(config.Config); err != nil {
        return nil, err
    }
    return collector, nil
})
```

## Health Monitoring

### Health Endpoints

- **`/health`**: Basic health check
- **`/ready`**: Readiness probe for Kubernetes
- **`/metrics`**: Prometheus metrics
- **`/status`**: Detailed status information

### Example Health Check Response

```json
{
  "status": "healthy",
  "message": "All health checks passed",
  "checks": {
    "framework_running": {
      "status": "healthy",
      "message": "Health check passed"
    },
    "plugins_healthy": {
      "status": "healthy",
      "message": "Health check passed"
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run benchmarks
make test-bench
```

### Writing Tests

```go
func TestCustomCollector(t *testing.T) {
    collector := &CustomCollector{
        name:     "test-collector",
        interval: 100 * time.Millisecond,
    }

    // Test configuration
    err := collector.Configure(map[string]interface{}{
        "endpoint": "http://test.com",
    })
    assert.NoError(t, err)

    // Test data collection
    ctx := context.Background()
    data, err := collector.Collect(ctx)
    assert.NoError(t, err)
    assert.Len(t, data, 1)
}
```

## Monitoring and Observability

### Metrics

The framework exposes Prometheus-compatible metrics:

```
# Framework metrics
framework_running 1
framework_total_plugins 4
framework_collectors 1
framework_analyzers 1
framework_responders 1
framework_agents 1
```

### Logging

Structured logging with context:

```go
slog.Info("Plugin started", 
    "plugin", "prometheus", 
    "type", "collector",
    "version", "1.0.0")
```

### Tracing

Built-in support for distributed tracing with OpenTelemetry (coming soon).

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o agent ./cli/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent .
COPY --from=builder /app/framework.yaml .
CMD ["./agent"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-framework
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agent-framework
  template:
    metadata:
      labels:
        app: agent-framework
    spec:
      containers:
      - name: agent
        image: agent-framework:latest
        ports:
        - containerPort: 9090
        env:
        - name: AGENT_AI_API_KEY
          valueFrom:
            secretKeyRef:
              name: agent-secrets
              key: api-key
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
        readinessProbe:
          httpGet:
            path: /ready
            port: 9090
```

### Development Setup

```bash
# Install development tools
make dev-tools

# Run linting
make lint

# Run tests
make test

# Build
make build
```
