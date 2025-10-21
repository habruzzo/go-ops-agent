# Microservice End-to-End Testing Guide

This guide shows you how to test your observability framework with a **real microservice** that generates realistic Prometheus metrics and anomalies.

## What We Built

### **Real Microservice** (`test/microservice/`)
A complete web service that simulates a real application with:

- **HTTP API endpoints** (users, products, orders)
- **Prometheus metrics** (CPU, memory, response time, business metrics)
- **Anomaly simulation** (CPU spikes, memory leaks, slow responses)
- **Admin controls** (toggle anomaly mode)
- **Health checks** and status endpoints

### **Realistic Metrics**
- `cpu_usage_percent` - CPU usage with occasional spikes
- `memory_usage_percent` - Memory usage with leak simulation
- `response_time_ms` - HTTP response times with delays
- `http_requests_total` - Request counters by endpoint/status
- `active_users` - Simulated user activity patterns
- `orders_processed_total` - Business metrics
- `order_value_dollars` - Order value distribution

## Quick Start

### **1. Start the Microservice**
```bash
cd test
./start_microservice.sh
```

This will:
- Build and start the microservice on port 8080
- Show you all available endpoints
- Keep it running until you press Ctrl+C

### **2. Run End-to-End Tests**
```bash
cd test
./run_microservice_tests.sh
```

This will:
- Start the microservice automatically
- Run all tests with real metrics
- Test anomaly detection
- Run performance benchmarks
- Clean up when done

### **3. Manual Testing**
```bash
# Start microservice
cd test
./start_microservice.sh

# In another terminal, run your framework
cd ..
go run cli/main.go -config test/test-config.yaml
```

## Available Endpoints

### **API Endpoints**
- `GET /api/users` - Get user count
- `GET /api/products` - Get product count  
- `GET /api/orders` - Get order count
- `POST /api/orders` - Create new order

### **Admin Endpoints**
- `GET /admin/status` - Get service status
- `POST /admin/anomaly` - Toggle anomaly mode
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

### **Example Usage**
```bash
# Check health
curl http://localhost:8080/health

# Get metrics
curl http://localhost:8080/metrics

# Enable anomalies
curl -X POST http://localhost:8080/admin/anomaly

# Check status
curl http://localhost:8080/admin/status

# Generate load
curl http://localhost:8080/api/users
curl http://localhost:8080/api/products
```

## Test Scenarios

### **1. Normal Operation Test**
```bash
go test -v ./test/e2e_microservice_test.go -run TestMicroserviceEndToEnd
```
- Runs framework with normal microservice load
- Tests metric collection and analysis
- Duration: 30 seconds

### **2. Anomaly Detection Test**
```bash
go test -v ./test/e2e_microservice_test.go -run TestMicroserviceWithAnomalies
```
- Starts with normal load
- Enables anomaly mode mid-test
- Tests anomaly detection
- Duration: 40 seconds

### **3. Performance Test**
```bash
go test -v ./test/e2e_microservice_test.go -run TestMicroservicePerformance
```
- Tests with multiple collectors/analyzers
- Measures framework performance
- Duration: 30 seconds

## Configuration

### **Framework Config** (`test/test-config.yaml`)
```yaml
logging:
  level: info
  format: text
  output: stdout

agent:
  default_agent: "test-ai"

plugins:
  - name: "prometheus-collector"
    type: "prometheus"
    config:
      url: "http://localhost:9090"  # Prometheus server
      interval: "5s"
      queries: 
        - "cpu_usage_percent"
        - "memory_usage_percent"
        - "response_time_ms"
        - "http_requests_total"
        - "active_users"
  
  - name: "anomaly-analyzer"
    type: "anomaly"
    config:
      threshold: 2.0
  
  - name: "logger-responder"
    type: "logger"
    config:
      level: "info"
```

### **Prometheus Config** (`test/prometheus.yml`)
```yaml
scrape_configs:
  - job_name: 'test-microservice'
    static_configs:
      - targets: ['test-microservice:8080']
    scrape_interval: 2s
    metrics_path: '/metrics'
```

## Docker Testing

### **Start All Services**
```bash
cd test
docker-compose -f docker-compose.test.yml up -d
```

This starts:
- **Prometheus** (port 9090) - Metrics collection
- **Node Exporter** (port 9100) - System metrics
- **Grafana** (port 3000) - Visualization
- **Test Microservice** (port 8080) - Your app

### **Run Tests with Docker**
```bash
# Test with real Prometheus
go test -run TestMicroserviceEndToEnd ./test/e2e_microservice_test.go

# Test with Docker services
docker-compose -f docker-compose.test.yml exec test-microservice curl localhost:8080/health
```

### **Stop Services**
```bash
cd test
docker-compose -f docker-compose.test.yml down
```

## What You'll See

### **Microservice Output**
```
Microservice started on port 8080
Metrics available at: http://localhost:8080/metrics
Health check: http://localhost:8080/health
Admin panel: http://localhost:8080/admin/status
Toggle anomalies: POST http://localhost:8080/admin/anomaly
Press Ctrl+C to stop...
```

### **Framework Output**
```
time=2025-10-20T18:34:21.060-07:00 level=INFO msg="Plugin loaded" plugin=microservice-collector type=collector
time=2025-10-20T18:34:21.060-07:00 level=INFO msg="Starting framework..."
time=2025-10-20T18:34:21.060-07:00 level=INFO msg="Framework started successfully"

# Anomaly detected!
time=2025-10-20T18:34:25.123-07:00 level=ERROR [anomaly] Detected 1 anomalies with max deviation of 3.2σ plugin=microservice-logger analyzer=microservice-analyzer type=anomaly confidence=0.8 severity=high data_points=1
```

### **Test Output**
```
=== RUN   TestMicroserviceEndToEnd
    e2e_microservice_test.go:25: Running end-to-end test with real microservice
    e2e_microservice_test.go:67: Framework status: map[running:true total_plugins:4 collectors:1 analyzers:1 responders:1 agents:1]
    e2e_microservice_test.go:75: Microservice end-to-end test completed successfully
--- PASS: TestMicroserviceEndToEnd (30.51s)
```

## Real-World Scenarios

### **1. CPU Spike Detection**
- Microservice simulates CPU spikes (30% chance in anomaly mode)
- Framework detects spikes >2σ from normal
- Logger responds with high severity alerts

### **2. Memory Leak Detection**
- Microservice gradually increases memory usage
- Framework detects trend anomalies
- Continuous monitoring of memory patterns

### **3. Response Time Anomalies**
- Microservice simulates slow API responses
- Framework detects response time spikes
- Real-time alerting for performance issues

### **4. Business Metric Monitoring**
- Order processing rates
- User activity patterns
- Revenue tracking

## Debugging

### **Check Microservice Status**
```bash
curl http://localhost:8080/admin/status
# {"running":true,"anomaly_mode":false,"timestamp":"2025-10-20T18:34:21Z"}
```

### **View Metrics**
```bash
curl http://localhost:8080/metrics | grep cpu_usage_percent
# cpu_usage_percent 45.2
```

### **Enable Anomalies**
```bash
curl -X POST http://localhost:8080/admin/anomaly
# {"anomaly_mode":"enabled"}
```

### **Check Framework Logs**
```bash
# Run with debug logging
LOG_LEVEL=debug go test -v ./test/e2e_microservice_test.go
```

## Production Ready

This microservice setup provides:

- **Real Metrics** - Actual Prometheus metrics, not mock data  
- **Realistic Anomalies** - CPU spikes, memory leaks, slow responses  
- **Business Logic** - Orders, users, products with real patterns  
- **Admin Controls** - Toggle anomalies, check status  
- **Docker Support** - Full containerized testing  
- **Performance Testing** - Load testing with multiple analyzers  
- **Integration Testing** - End-to-end workflow validation  

## Next Steps

1. **Start the microservice**: `./start_microservice.sh`
2. **Run the tests**: `./run_microservice_tests.sh`
3. **Explore the metrics**: Visit `http://localhost:8080/metrics`
4. **Test anomaly detection**: Enable anomalies and watch the alerts
5. **Scale up**: Add more microservices, test with multiple instances

Your observability framework is now tested with **real production-like scenarios**!
