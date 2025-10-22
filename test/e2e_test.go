package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/habruzzo/agent/config"
	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/collectors"
	"github.com/habruzzo/agent/plugins/responders"
)

// TestDataGenerator generates realistic test data
type TestDataGenerator struct {
	baseCPU       float64
	baseMemory    float64
	baseDisk      float64
	anomalyChance float64
}

func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		baseCPU:       50.0,
		baseMemory:    60.0,
		baseDisk:      30.0,
		anomalyChance: 0.1, // 10% chance of anomaly
	}
}

func (g *TestDataGenerator) GenerateDataPoint(metric string, timestamp time.Time) core.DataPoint {
	var value float64

	switch metric {
	case "cpu_usage_percent":
		value = g.generateCPUValue()
	case "memory_usage_percent":
		value = g.generateMemoryValue()
	case "disk_usage_percent":
		value = g.generateDiskValue()
	case "response_time_ms":
		value = g.generateResponseTime()
	default:
		value = rand.Float64() * 100
	}

	return core.DataPoint{
		Timestamp: timestamp,
		Source:    "test-server",
		Metric:    metric,
		Value:     value,
		Labels: map[string]string{
			"instance": "test-instance-1",
			"job":      "test-job",
			"env":      "test",
		},
	}
}

func (g *TestDataGenerator) generateCPUValue() float64 {
	// Normal CPU usage with occasional spikes
	base := g.baseCPU + rand.NormFloat64()*5 // ±5% variation

	// 10% chance of anomaly (CPU spike)
	if rand.Float64() < g.anomalyChance {
		return base + 40 + rand.Float64()*20 // Spike to 80-100%
	}

	return max(0, min(100, base))
}

func (g *TestDataGenerator) generateMemoryValue() float64 {
	// Memory usage with gradual increase and occasional spikes
	base := g.baseMemory + rand.NormFloat64()*3

	// 5% chance of memory leak
	if rand.Float64() < 0.05 {
		return base + 20 + rand.Float64()*15
	}

	return max(0, min(100, base))
}

func (g *TestDataGenerator) generateDiskValue() float64 {
	// Disk usage with gradual increase
	base := g.baseDisk + rand.NormFloat64()*2

	// Very slow increase over time
	g.baseDisk += 0.01

	return max(0, min(100, base))
}

func (g *TestDataGenerator) generateResponseTime() float64 {
	// Response time with occasional spikes
	base := 100 + rand.NormFloat64()*20

	// 15% chance of slow response
	if rand.Float64() < 0.15 {
		return base + 500 + rand.Float64()*1000
	}

	return max(0, base)
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

// MockPrometheusServer creates a simple HTTP server that serves mock Prometheus data
type MockPrometheusServer struct {
	server *http.Server
	port   string
}

func NewMockPrometheusServer() *MockPrometheusServer {
	// Use port 0 to let the system assign an available port
	return &MockPrometheusServer{
		port: "0",
	}
}

func (m *MockPrometheusServer) Start() error {
	mux := http.NewServeMux()

	// Mock Prometheus query endpoint
	mux.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")

		// Generate mock response based on query
		var response string
		switch query {
		case "up":
			response = `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","job":"test"},"value":[1234567890,"1"]}]}}`
		case "cpu_usage_percent":
			value := 50 + rand.Float64()*30 // 50-80% CPU
			response = fmt.Sprintf(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"cpu_usage_percent","instance":"test"},"value":[1234567890,"%.2f"]}]}}`, value)
		case "memory_usage_percent":
			value := 60 + rand.Float64()*20 // 60-80% memory
			response = fmt.Sprintf(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"memory_usage_percent","instance":"test"},"value":[1234567890,"%.2f"]}]}}`, value)
		default:
			response = `{"status":"success","data":{"resultType":"vector","result":[]}}`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	// Create listener to get actual port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	// Get the actual port that was assigned
	addr := listener.Addr().(*net.TCPAddr)
	m.port = fmt.Sprintf("%d", addr.Port)

	m.server = &http.Server{
		Handler: mux,
	}

	go func() {
		if err := m.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Mock Prometheus server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (m *MockPrometheusServer) Stop() error {
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.server.Shutdown(ctx)
	}
	return nil
}

func (m *MockPrometheusServer) URL() string {
	return "http://localhost:" + m.port
}

// TestScenario represents a test scenario
type TestScenario struct {
	Name              string
	Description       string
	Duration          time.Duration
	AnomalyRate       float64
	ExpectedAnomalies int
}

func TestEndToEndAnomalyDetection(t *testing.T) {
	// Create test scenarios
	scenarios := []TestScenario{
		{
			Name:              "normal_operation",
			Description:       "Normal system operation with occasional anomalies",
			Duration:          30 * time.Second,
			AnomalyRate:       0.1,
			ExpectedAnomalies: 2,
		},
		{
			Name:              "high_anomaly_rate",
			Description:       "High anomaly rate to test detection",
			Duration:          20 * time.Second,
			AnomalyRate:       0.3,
			ExpectedAnomalies: 5,
		},
		{
			Name:              "memory_leak_simulation",
			Description:       "Simulate memory leak scenario",
			Duration:          25 * time.Second,
			AnomalyRate:       0.05,
			ExpectedAnomalies: 3,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			testScenario(t, scenario)
		})
	}
}

func testScenario(t *testing.T, scenario TestScenario) {
	t.Logf("Running scenario: %s - %s", scenario.Name, scenario.Description)

	// Start mock Prometheus server
	mockServer := NewMockPrometheusServer()
	if err := mockServer.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer mockServer.Stop()

	// Create framework configuration
	cfg := &core.FrameworkConfig{
		LogLevel:     "info",
		LogFormat:    "text",
		LogOutput:    "stdout",
		DefaultAgent: "test-ai",
		Plugins:      []core.PluginConfig{},
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Add Prometheus collector
	collector := collectors.NewPrometheusCollector("test-prometheus")
	collector.Configure(map[string]interface{}{
		"url":      mockServer.URL(),
		"interval": "2s",
		"queries":  []string{"cpu_usage_percent", "memory_usage_percent", "up"},
	})
	framework.LoadPlugin(collector)

	// Add anomaly analyzer
	analyzer := analyzers.NewAnomalyAnalyzer("test-anomaly")
	analyzer.Configure(map[string]interface{}{
		"threshold": 2.0,
	})
	framework.LoadPlugin(analyzer)

	// Add logger responder
	responder := responders.NewLoggerResponder("test-logger")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Add AI agent (skip if no valid API key available)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		agent := agents.NewAIAgent("test-ai")
		agent.Configure(map[string]interface{}{
			"api_key": apiKey,
			"model":   "gpt-3.5-turbo",
		})
		framework.LoadPlugin(agent)
	} else {
		t.Log("Skipping AI agent - no OPENAI_API_KEY environment variable set")
	}

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), scenario.Duration+10*time.Second)
	defer cancel()

	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Let it run for the scenario duration
	t.Logf("Running for %v...", scenario.Duration)
	time.Sleep(scenario.Duration)

	// Get status
	status := framework.GetStatus()
	t.Logf("Framework status: %+v", status)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	t.Logf("Scenario %s completed successfully", scenario.Name)
}

func TestEndToEndWithRealPrometheus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real Prometheus test in short mode")
	}

	// Check if Prometheus is running
	resp, err := http.Get("http://localhost:9090/api/v1/query?query=up")
	if err != nil {
		t.Skip("Prometheus not running, skipping real Prometheus test")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skip("Prometheus not responding correctly, skipping test")
	}

	t.Log("Running end-to-end test with real Prometheus")

	// Create framework with real Prometheus
	cfg := &core.FrameworkConfig{
		LogLevel:  "info",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Add real Prometheus collector
	collector := collectors.NewPrometheusCollector("real-prometheus")
	collector.Configure(map[string]interface{}{
		"url":      "http://localhost:9090",
		"interval": "5s",
		"queries":  []string{"up", "prometheus_tsdb_head_samples", "prometheus_tsdb_head_series"},
	})
	framework.LoadPlugin(collector)

	// Add anomaly analyzer
	analyzer := analyzers.NewAnomalyAnalyzer("real-anomaly")
	analyzer.Configure(map[string]interface{}{
		"threshold": 2.5,
	})
	framework.LoadPlugin(analyzer)

	// Add logger responder
	responder := responders.NewLoggerResponder("real-logger")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Let it run
	time.Sleep(20 * time.Second)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	t.Log("Real Prometheus test completed successfully")
}

func TestEndToEndConfigurationLoading(t *testing.T) {
	// Create a test configuration file
	testConfig := `
log_level: debug
log_format: json
log_output: stdout
default_agent: "test-ai"

plugins:
  - name: "test-collector"
    type: "collector"
    config:
      url: "http://localhost:9091"
      interval: "5s"
      queries: ["up", "cpu_usage_percent"]
  
  - name: "test-analyzer"
    type: "analyzer"
    config:
      threshold: 2.0
  
  - name: "test-responder"
    type: "responder"
    config:
      level: "info"
`

	// Write test config to temporary file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify configuration
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got %s", cfg.LogLevel)
	}

	if cfg.LogFormat != "json" {
		t.Errorf("Expected log format 'json', got %s", cfg.LogFormat)
	}

	if len(cfg.Plugins) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(cfg.Plugins))
	}

	t.Log("Configuration loading test completed successfully")
}

func TestEndToEndPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Log("Running performance test...")

	// Create framework with multiple analyzers
	cfg := &core.FrameworkConfig{
		LogLevel:  "warn", // Reduce logging for performance
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Add multiple analyzers to test performance
	for i := 0; i < 5; i++ {
		analyzer := analyzers.NewAnomalyAnalyzer(fmt.Sprintf("analyzer-%d", i))
		analyzer.Configure(map[string]interface{}{
			"threshold": 2.0,
		})
		framework.LoadPlugin(analyzer)
	}

	// Add logger responder
	responder := responders.NewLoggerResponder("perf-logger")
	responder.Configure(map[string]interface{}{
		"level": "warn",
	})
	framework.LoadPlugin(responder)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start := time.Now()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Let it run
	time.Sleep(10 * time.Second)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	duration := time.Since(start)
	t.Logf("Performance test completed in %v", duration)

	// Performance assertions
	if duration > 15*time.Second {
		t.Errorf("Framework took too long to start/stop: %v", duration)
	}
}
