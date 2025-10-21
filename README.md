# AI Agent for Observability

A simple, modular AI agent built in Go for observability tasks. The agent follows a plugin architecture with collectors, analyzers, and responders.

## Architecture

```
Data Collection → Preprocessing → AI Analysis → Decision Engine → Actions
```

### Components

- **Collectors**: Gather data from observability sources (Prometheus, ELK, etc.)
- **Analyzers**: Process data and detect patterns/anomalies
- **Responders**: Take actions based on analysis results (alerts, logs, etc.)

## Quick Start

1. **Create a default configuration:**
   ```bash
   go run cli/main.go -create-config
   ```

2. **Run the agent:**
   ```bash
   go run cli/main.go
   ```

3. **Run with custom config:**
   ```bash
   go run cli/main.go -config my-config.yaml
   ```

## Configuration

The agent uses YAML configuration files. See `agent.yaml` for an example.

### Collector Configuration
```yaml
collectors:
  - name: prometheus
    type: prometheus
    interval: 30s
    config:
      url: http://localhost:9090
      queries:
        - up
        - cpu_usage_percent
```

### Analyzer Configuration
```yaml
analyzers:
  - name: anomaly-detector
    type: anomaly
    config:
      threshold: 2.0
```

### Responder Configuration
```yaml
responders:
  - name: logger
    type: log
    config:
      level: info
```

## Building

```bash
go mod tidy
go build -o agent cli/main.go
```

## Extending the Agent

### Adding a New Collector

1. Implement the `core.Collector` interface
2. Add a case in `createCollector()` function
3. Update configuration schema

### Adding a New Analyzer

1. Implement the `core.Analyzer` interface
2. Add a case in `createAnalyzer()` function
3. Update configuration schema

### Adding a New Responder

1. Implement the `core.Responder` interface
2. Add a case in `createResponder()` function
3. Update configuration schema

## Example Usage

```go
// Create agent with configuration
config := config.LoadConfig("agent.yaml")
agent := core.NewAgent(config)

// Add components
agent.AddCollector(collectors.NewPrometheusCollector("prometheus"))
agent.AddAnalyzer(analyzers.NewAnomalyAnalyzer("anomaly-detector"))
agent.AddResponder(responders.NewLoggerResponder("logger"))

// Start agent
agent.Start(context.Background())
```

## Logging

The framework uses Go's standard `slog` library for structured logging with configurable options:

- **Consistent formatting**: All logs follow the same format
- **Centralized control**: Change log level, format, and output for the entire system
- **Plugin context**: Each log entry includes plugin name and type
- **Standard library**: No external dependencies

### Configuration

```yaml
logging:
  level: info      # debug, info, warn, error
  format: text     # text, json
  output: stdout   # stdout, stderr, or file path
```

### Usage

```go
// All plugins use slog directly
slog.Info("Starting collector", "plugin", "prometheus", "type", "collector")
slog.Warn("API rate limit approaching", "plugin", "ai-agent", "type", "agent")
```

## Project Structure

```
agent/
├── core/                    # Framework and plugin interfaces
├── plugins/                 # All plugins organized by type
│   ├── collectors/         # Data collection plugins
│   ├── analyzers/          # Data analysis plugins
│   ├── responders/         # Response/action plugins
│   └── agents/             # AI agent plugins
├── config/                 # Configuration management
├── cli/                    # Interactive CLI
├── examples/               # Example applications
├── framework.yaml          # Configuration file
└── go.mod
```

## Dependencies

- `github.com/prometheus/client_golang` - Prometheus client
- `gopkg.in/yaml.v3` - YAML configuration
- `log/slog` - Structured logging (Go standard library)
