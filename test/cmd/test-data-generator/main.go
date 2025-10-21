package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := generateMetrics()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metrics))
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Test data generator starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func generateMetrics() string {
	now := time.Now().Unix()

	// Generate realistic system metrics with occasional anomalies
	cpuUsage := generateCPUUsage()
	memoryUsage := generateMemoryUsage()
	diskUsage := generateDiskUsage()
	responseTime := generateResponseTime()

	metrics := fmt.Sprintf(`# HELP cpu_usage_percent CPU usage percentage
# TYPE cpu_usage_percent gauge
cpu_usage_percent{instance="test-instance",job="test-job"} %.2f %d

# HELP memory_usage_percent Memory usage percentage
# TYPE memory_usage_percent gauge
memory_usage_percent{instance="test-instance",job="test-job"} %.2f %d

# HELP disk_usage_percent Disk usage percentage
# TYPE disk_usage_percent gauge
disk_usage_percent{instance="test-instance",job="test-job"} %.2f %d

# HELP response_time_ms HTTP response time in milliseconds
# TYPE response_time_ms gauge
response_time_ms{instance="test-instance",job="test-job",endpoint="/api/users"} %.2f %d

# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{instance="test-instance",job="test-job",method="GET",status="200"} %d %d
http_requests_total{instance="test-instance",job="test-job",method="GET",status="500"} %d %d

# HELP up Whether the service is up
# TYPE up gauge
up{instance="test-instance",job="test-job"} 1 %d
`,
		cpuUsage, now*1000,
		memoryUsage, now*1000,
		diskUsage, now*1000,
		responseTime, now*1000,
		rand.Intn(1000)+500, now*1000,
		rand.Intn(10), now*1000,
		now*1000,
	)

	return metrics
}

func generateCPUUsage() float64 {
	// Normal CPU usage with occasional spikes
	base := 50.0 + rand.NormFloat64()*5 // ±5% variation

	// 10% chance of CPU spike
	if rand.Float64() < 0.1 {
		return base + 40 + rand.Float64()*20 // Spike to 80-100%
	}

	// Keep within bounds
	if base < 0 {
		return 0
	}
	if base > 100 {
		return 100
	}
	return base
}

func generateMemoryUsage() float64 {
	// Memory usage with gradual increase and occasional spikes
	base := 60.0 + rand.NormFloat64()*3

	// 5% chance of memory leak
	if rand.Float64() < 0.05 {
		return base + 20 + rand.Float64()*15
	}

	// Keep within bounds
	if base < 0 {
		return 0
	}
	if base > 100 {
		return 100
	}
	return base
}

func generateDiskUsage() float64 {
	// Disk usage with gradual increase
	base := 30.0 + rand.NormFloat64()*2

	// Keep within bounds
	if base < 0 {
		return 0
	}
	if base > 100 {
		return 100
	}
	return base
}

func generateResponseTime() float64 {
	// Response time with occasional spikes
	base := 100.0 + rand.NormFloat64()*20

	// 15% chance of slow response
	if rand.Float64() < 0.15 {
		return base + 500 + rand.Float64()*1000
	}

	// Keep positive
	if base < 0 {
		return 0
	}
	return base
}
