# End-to-End Testing Guide

This directory contains comprehensive end-to-end tests for the observability framework.

## Test Types

### 1. Unit Tests
- **Location**: `../core/`, `../plugins/`
- **Purpose**: Test individual components in isolation
- **Run**: `go test ./core/... ./plugins/...`

### 2. Integration Tests
- **Location**: `../integration_test.go`
- **Purpose**: Test component interactions
- **Run**: `go test ./integration_test.go`

### 3. End-to-End Tests
- **Location**: `e2e_test.go`
- **Purpose**: Test complete workflows with realistic scenarios
- **Run**: `go test ./test/e2e_test.go`

### 4. Performance Tests
- **Location**: `e2e_test.go` (TestEndToEndPerformance)
- **Purpose**: Test framework performance under load
- **Run**: `go test -run TestEndToEndPerformance ./test/e2e_test.go`

## Quick Start

### Run All Tests
```bash
# From project root
./test/run_e2e_tests.sh
```

### Run Specific Test Types
```bash
# Unit tests only
go test ./core/... ./plugins/...

# Integration tests only
go test ./integration_test.go

# End-to-end tests only
go test ./test/e2e_test.go

# Performance tests only
go test -run TestEndToEndPerformance ./test/e2e_test.go
```

## Docker-Based Testing

### Prerequisites
- Docker
- Docker Compose

### Start Test Services
```bash
cd test
docker-compose -f docker-compose.test.yml up -d
```

### Run Tests with Real Services
```bash
# Test with real Prometheus
go test -run TestEndToEndWithRealPrometheus ./test/e2e_test.go
```

### Stop Test Services
```bash
cd test
docker-compose -f docker-compose.test.yml down
```

## Test Scenarios

### 1. Normal Operation
- **Duration**: 30 seconds
- **Anomaly Rate**: 10%
- **Expected**: 2-3 anomalies detected

### 2. High Anomaly Rate
- **Duration**: 20 seconds
- **Anomaly Rate**: 30%
- **Expected**: 5-6 anomalies detected

### 3. Memory Leak Simulation
- **Duration**: 25 seconds
- **Anomaly Rate**: 5%
- **Expected**: 3-4 anomalies detected

### 4. Performance Test
- **Duration**: 15 seconds
- **Load**: 5 analyzers + 1 responder
- **Expected**: <15s total runtime

## Manual Testing

### Run Manual Test
```bash
cd test
go run manual_test.go
```

This will:
1. Load configuration from `test-config.yaml`
2. Start the framework with all plugins
3. Display status every 30 seconds
4. Allow graceful shutdown with Ctrl+C

### Test with Real Prometheus
1. Start Prometheus: `docker run -p 9090:9090 prom/prometheus`
2. Run manual test: `go run manual_test.go`
3. Watch logs for anomaly detection

## Performance Benchmarks

### Anomaly Analyzer Performance
```
BenchmarkAnomalyAnalyzer_Analyze-12                215,826    5,527 ns/op    0 B/op    0 allocs/op
BenchmarkAnomalyAnalyzer_AnalyzeSmallDataset-12    21,530,134 57.26 ns/op    0 B/op    0 allocs/op
BenchmarkAnomalyAnalyzer_AnalyzeLargeDataset-12    18,904     62,908 ns/op   0 B/op    0 allocs/op
```

### Memory Usage
- **Zero allocations** in hot paths
- **Efficient data structures** for analysis
- **Minimal memory footprint** for plugins

## Debugging Tests

### Enable Debug Logging
```bash
# Set log level to debug
export LOG_LEVEL=debug
go test -v ./test/e2e_test.go
```

### Run Single Test
```bash
# Run specific test
go test -run TestEndToEndAnomalyDetection ./test/e2e_test.go -v
```

### Run with Race Detection
```bash
go test -race ./test/e2e_test.go
```

### Run with Coverage
```bash
go test -coverprofile=coverage.out ./test/e2e_test.go
go tool cover -html=coverage.out -o coverage.html
```

## Test Checklist

### Before Running Tests
- [ ] Go 1.21+ installed
- [ ] Docker installed (for Docker tests)
- [ ] No Prometheus running on port 9090 (for mock tests)
- [ ] Sufficient disk space for test artifacts

### Test Validation
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All end-to-end tests pass
- [ ] Performance tests meet requirements
- [ ] Race detection shows no issues
- [ ] Coverage is >80%

### After Tests
- [ ] Test artifacts cleaned up
- [ ] Docker containers stopped
- [ ] Coverage report generated
- [ ] Performance benchmarks recorded

## Troubleshooting

### Common Issues

#### Prometheus Connection Refused
```
Error: dial tcp [::1]:9090: connect: connection refused
```
**Solution**: Start Prometheus or use mock server

#### Docker Permission Denied
```
Error: permission denied while trying to connect to Docker daemon
```
**Solution**: Add user to docker group or use sudo

#### Test Timeout
```
Error: test timed out after 5m0s
```
**Solution**: Increase timeout or check for hanging processes

#### Race Condition Detected
```
Error: race detected during execution
```
**Solution**: Check for concurrent access to shared resources

### Debug Commands
```bash
# Check if Prometheus is running
curl http://localhost:9090/api/v1/query?query=up

# Check Docker services
docker ps

# Check test logs
go test -v ./test/e2e_test.go 2>&1 | tee test.log
```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Framework Architecture](../README.md)
