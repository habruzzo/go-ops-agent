package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/responders"
)

// TestSimpleMicroserviceEndToEnd tests the framework with direct microservice metrics
func TestSimpleMicroserviceEndToEnd(t *testing.T) {
	// Check if microservice is running
	if !isMicroserviceRunning() {
		t.Skip("Microservice not running, skipping test. Start with: cd test && ./start_microservice.sh")
	}

	t.Log("Running simple end-to-end test with microservice")

	// Create framework configuration
	cfg := &core.FrameworkConfig{
		LogLevel:  "info",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []core.PluginConfig{},
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Add anomaly analyzer
	analyzer := analyzers.NewAnomalyAnalyzer("microservice-analyzer")
	analyzer.Configure(map[string]interface{}{
		"threshold": 2.0,
	})
	framework.LoadPlugin(analyzer)

	// Add logger responder
	responder := responders.NewLoggerResponder("microservice-logger")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Generate some test data manually
	go func() {
		time.Sleep(2 * time.Second)

		// Create some test data points
		testData := []core.DataPoint{
			{
				Timestamp: time.Now(),
				Source:    "microservice",
				Metric:    "cpu_usage_percent",
				Value:     45.0,
				Labels:    map[string]string{"instance": "test"},
			},
			{
				Timestamp: time.Now(),
				Source:    "microservice",
				Metric:    "cpu_usage_percent",
				Value:     48.0,
				Labels:    map[string]string{"instance": "test"},
			},
			{
				Timestamp: time.Now(),
				Source:    "microservice",
				Metric:    "cpu_usage_percent",
				Value:     85.0, // This should be detected as an anomaly
				Labels:    map[string]string{"instance": "test"},
			},
		}

		// Send data to the framework's data channel
		// This simulates what a collector would do
		for _, point := range testData {
			framework.GetDataChannel() <- []core.DataPoint{point}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Let it run for a while
	time.Sleep(10 * time.Second)

	// Get status
	status := framework.GetStatus()
	t.Logf("Framework status: %+v", status)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	t.Log("Simple microservice end-to-end test completed successfully")
}

// TestMicroserviceMetrics tests that we can fetch metrics from the microservice
func TestMicroserviceMetrics(t *testing.T) {
	// Check if microservice is running
	if !isMicroserviceRunning() {
		t.Skip("Microservice not running, skipping test. Start with: cd test && ./start_microservice.sh")
	}

	t.Log("Testing microservice metrics endpoint")

	// Test health endpoint
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test metrics endpoint
	resp, err = http.Get("http://localhost:8080/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test admin status
	resp, err = http.Get("http://localhost:8080/admin/status")
	if err != nil {
		t.Fatalf("Failed to get admin status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Log("Microservice metrics test completed successfully")
}

// TestMicroserviceAnomalyMode tests the anomaly mode toggle
func TestMicroserviceAnomalyMode(t *testing.T) {
	// Check if microservice is running
	if !isMicroserviceRunning() {
		t.Skip("Microservice not running, skipping test. Start with: cd test && ./start_microservice.sh")
	}

	t.Log("Testing microservice anomaly mode")

	// Check initial status
	status, err := getMicroserviceStatus()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	t.Logf("Initial status: %+v", status)

	// Enable anomaly mode
	err = enableMicroserviceAnomalyMode(true)
	if err != nil {
		t.Fatalf("Failed to enable anomaly mode: %v", err)
	}

	// Check status after enabling
	status, err = getMicroserviceStatus()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	t.Logf("Status after enabling anomalies: %+v", status)

	// Disable anomaly mode
	err = enableMicroserviceAnomalyMode(false)
	if err != nil {
		t.Fatalf("Failed to disable anomaly mode: %v", err)
	}

	// Check final status
	status, err = getMicroserviceStatus()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	t.Logf("Final status: %+v", status)

	t.Log("Microservice anomaly mode test completed successfully")
}
