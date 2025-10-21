package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics
var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Business metrics
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)

	ordersProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_processed_total",
			Help: "Total number of orders processed",
		},
	)

	orderValue = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "order_value_dollars",
			Help:    "Order value in dollars",
			Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		},
	)

	// System metrics
	cpuUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cpu_usage_percent",
			Help: "CPU usage percentage",
		},
	)

	memoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_percent",
			Help: "Memory usage percentage",
		},
	)

	diskUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "disk_usage_percent",
			Help: "Disk usage percentage",
		},
	)

	responseTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "response_time_ms",
			Help: "Average response time in milliseconds",
		},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "severity"},
	)
)

// Microservice represents our test microservice
type Microservice struct {
	server      *http.Server
	mu          sync.RWMutex
	running     bool
	anomalyMode bool
	baseCPU     float64
	baseMemory  float64
	baseDisk    float64
}

// NewMicroservice creates a new microservice instance
func NewMicroservice(port string) *Microservice {
	return &Microservice{
		baseCPU:    30.0,
		baseMemory: 40.0,
		baseDisk:   25.0,
	}
}

// Start starts the microservice
func (m *Microservice) Start(port string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("microservice is already running")
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", m.healthHandler)

	// API endpoints
	mux.HandleFunc("/api/users", m.usersHandler)
	mux.HandleFunc("/api/orders", m.ordersHandler)
	mux.HandleFunc("/api/products", m.productsHandler)

	// Admin endpoints
	mux.HandleFunc("/admin/anomaly", m.anomalyHandler)
	mux.HandleFunc("/admin/status", m.statusHandler)

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	m.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start metrics simulation
	go m.simulateMetrics()

	// Start the server
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	m.running = true
	log.Printf("Microservice started on port %s", port)
	return nil
}

// Stop stops the microservice
func (m *Microservice) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("microservice is not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	m.running = false
	log.Println("Microservice stopped")
	return nil
}

// simulateMetrics simulates realistic system metrics
func (m *Microservice) simulateMetrics() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.RLock()
		running := m.running
		anomalyMode := m.anomalyMode
		m.mu.RUnlock()

		if !running {
			break
		}

		// Simulate CPU usage
		cpu := m.simulateCPUUsage(anomalyMode)
		cpuUsage.Set(cpu)

		// Simulate memory usage
		memory := m.simulateMemoryUsage(anomalyMode)
		memoryUsage.Set(memory)

		// Simulate disk usage
		disk := m.simulateDiskUsage()
		diskUsage.Set(disk)

		// Simulate response time
		response := m.simulateResponseTime(anomalyMode)
		responseTime.Set(response)

		// Simulate active users
		users := m.simulateActiveUsers()
		activeUsers.Set(users)
	}
}

func (m *Microservice) simulateCPUUsage(anomalyMode bool) float64 {
	base := m.baseCPU + rand.NormFloat64()*5 // ±5% variation

	if anomalyMode {
		// 30% chance of CPU spike in anomaly mode
		if rand.Float64() < 0.3 {
			return base + 40 + rand.Float64()*30 // Spike to 70-100%
		}
	} else {
		// 5% chance of CPU spike in normal mode
		if rand.Float64() < 0.05 {
			return base + 20 + rand.Float64()*15 // Spike to 50-70%
		}
	}

	return max(0, min(100, base))
}

func (m *Microservice) simulateMemoryUsage(anomalyMode bool) float64 {
	base := m.baseMemory + rand.NormFloat64()*3

	if anomalyMode {
		// 20% chance of memory leak in anomaly mode
		if rand.Float64() < 0.2 {
			m.baseMemory += 2 // Gradual increase
			return base + 15 + rand.Float64()*10
		}
	} else {
		// 2% chance of memory leak in normal mode
		if rand.Float64() < 0.02 {
			m.baseMemory += 0.5 // Very gradual increase
			return base + 5 + rand.Float64()*5
		}
	}

	return max(0, min(100, base))
}

func (m *Microservice) simulateDiskUsage() float64 {
	// Disk usage increases very slowly
	m.baseDisk += 0.01
	base := m.baseDisk + rand.NormFloat64()*1
	return max(0, min(100, base))
}

func (m *Microservice) simulateResponseTime(anomalyMode bool) float64 {
	base := 50.0 + rand.NormFloat64()*10

	if anomalyMode {
		// 25% chance of slow response in anomaly mode
		if rand.Float64() < 0.25 {
			return base + 200 + rand.Float64()*500
		}
	} else {
		// 3% chance of slow response in normal mode
		if rand.Float64() < 0.03 {
			return base + 100 + rand.Float64()*200
		}
	}

	return max(0, base)
}

func (m *Microservice) simulateActiveUsers() float64 {
	// Simulate daily user patterns
	hour := time.Now().Hour()
	var base float64

	if hour >= 9 && hour <= 17 {
		base = 1000 + rand.NormFloat64()*100 // Work hours
	} else if hour >= 18 && hour <= 22 {
		base = 800 + rand.NormFloat64()*80 // Evening
	} else {
		base = 200 + rand.NormFloat64()*50 // Night/early morning
	}

	return max(0, base)
}

// HTTP handlers
func (m *Microservice) healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, "/health").Observe(duration)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	httpRequestsTotal.WithLabelValues(r.Method, "/health", "200").Inc()
}

func (m *Microservice) usersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, "/api/users").Observe(duration)
	}()

	// Simulate some processing time
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)

	// Simulate occasional errors
	if rand.Float64() < 0.02 { // 2% error rate
		errorsTotal.WithLabelValues("database", "high").Inc()
		httpRequestsTotal.WithLabelValues(r.Method, "/api/users", "500").Inc()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"database connection failed"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users":` + strconv.Itoa(rand.Intn(1000)+100) + `}`))
	httpRequestsTotal.WithLabelValues(r.Method, "/api/users", "200").Inc()
}

func (m *Microservice) ordersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, "/api/orders").Observe(duration)
	}()

	if r.Method == "POST" {
		// Simulate order processing
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		// Simulate order value
		value := 10 + rand.Float64()*500
		orderValue.Observe(value)
		ordersProcessed.Inc()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"order_id":"%d","value":%.2f}`, rand.Intn(10000), value)))
		httpRequestsTotal.WithLabelValues(r.Method, "/api/orders", "201").Inc()
	} else {
		// GET orders
		time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"orders":` + strconv.Itoa(rand.Intn(100)+10) + `}`))
		httpRequestsTotal.WithLabelValues(r.Method, "/api/orders", "200").Inc()
	}
}

func (m *Microservice) productsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, "/api/products").Observe(duration)
	}()

	// Simulate processing time
	time.Sleep(time.Duration(15+rand.Intn(25)) * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"products":` + strconv.Itoa(rand.Intn(500)+50) + `}`))
	httpRequestsTotal.WithLabelValues(r.Method, "/api/products", "200").Inc()
}

func (m *Microservice) anomalyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Toggle anomaly mode
		m.mu.Lock()
		m.anomalyMode = !m.anomalyMode
		mode := m.anomalyMode
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		status := "disabled"
		if mode {
			status = "enabled"
		}
		w.Write([]byte(fmt.Sprintf(`{"anomaly_mode":"%s"}`, status)))
	} else {
		// GET current status
		m.mu.RLock()
		mode := m.anomalyMode
		m.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		status := "disabled"
		if mode {
			status = "enabled"
		}
		w.Write([]byte(fmt.Sprintf(`{"anomaly_mode":"%s"}`, status)))
	}
}

func (m *Microservice) statusHandler(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	running := m.running
	anomalyMode := m.anomalyMode
	m.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"running":%t,"anomaly_mode":%t,"timestamp":"%s"}`,
		running, anomalyMode, time.Now().Format(time.RFC3339))))
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create microservice
	service := NewMicroservice(port)

	// Start the service
	if err := service.Start(port); err != nil {
		log.Fatalf("Failed to start microservice: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Microservice started on port %s\n", port)
	fmt.Printf("Metrics available at: http://localhost:%s/metrics\n", port)
	fmt.Printf("Health check: http://localhost:%s/health\n", port)
	fmt.Printf("Admin panel: http://localhost:%s/admin/status\n", port)
	fmt.Printf("Toggle anomalies: POST http://localhost:%s/admin/anomaly\n", port)
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down microservice...")

	// Stop the service
	if err := service.Stop(); err != nil {
		log.Printf("Error stopping microservice: %v", err)
	} else {
		fmt.Println("Microservice stopped successfully")
	}
}
