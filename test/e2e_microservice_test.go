package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/collectors"
	"github.com/habruzzo/agent/plugins/responders"
)

// TestMicroserviceEndToEnd tests the framework with a real microservice
func TestMicroserviceEndToEnd(t *testing.T) {
	// Check if microservice is running
	if !isMicroserviceRunning() {
		t.Skip("Microservice not running, skipping test. Start with: cd test/microservice && go run main.go")
	}

	t.Log("Running end-to-end test with real microservice")

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
	collector := collectors.NewPrometheusCollector("microservice-collector")
	collector.Configure(map[string]interface{}{
		"url":      "http://localhost:9090",
		"interval": "5s",
		"queries": []string{
			"cpu_usage_percent",
			"memory_usage_percent",
			"response_time_ms",
			"http_requests_total",
			"active_users",
			"orders_processed_total",
		},
	})
	framework.LoadPlugin(collector)

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

	// Add AI agent (skip if no valid API key available)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		agent := agents.NewAIAgent("microservice-ai")
		agent.Configure(map[string]interface{}{
			"api_key":    apiKey,
			"model":      "gpt-3.5-turbo",
			"max_tokens": 150,
		})
		framework.LoadPlugin(agent)
	} else {
		t.Log("Skipping AI agent - no OPENAI_API_KEY environment variable set")
	}

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Generate some load on the microservice
	go generateMicroserviceLoad()

	// Let it run for a while
	time.Sleep(30 * time.Second)

	// Get status
	status := framework.GetStatus()
	t.Logf("Framework status: %+v", status)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	t.Log("Microservice end-to-end test completed successfully")
}

// TestMicroserviceWithAnomalies tests anomaly detection with the microservice
func TestMicroserviceWithAnomalies(t *testing.T) {
	// Check if microservice is running
	if !isMicroserviceRunning() {
		t.Skip("Microservice not running, skipping test. Start with: cd test/microservice && go run main.go")
	}

	t.Log("Running anomaly detection test with microservice")

	// Create framework
	cfg := &core.FrameworkConfig{
		LogLevel:  "info",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Add Prometheus collector
	collector := collectors.NewPrometheusCollector("anomaly-collector")
	collector.Configure(map[string]interface{}{
		"url":      "http://localhost:9090",
		"interval": "3s",
		"queries": []string{
			"cpu_usage_percent",
			"memory_usage_percent",
			"response_time_ms",
		},
	})
	framework.LoadPlugin(collector)

	// Add anomaly analyzer with lower threshold for more sensitive detection
	analyzer := analyzers.NewAnomalyAnalyzer("anomaly-analyzer")
	analyzer.Configure(map[string]interface{}{
		"threshold": 1.5, // More sensitive
	})
	framework.LoadPlugin(analyzer)

	// Add logger responder
	responder := responders.NewLoggerResponder("anomaly-logger")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Generate normal load for 10 seconds
	t.Log("Generating normal load...")
	go generateMicroserviceLoad()
	time.Sleep(10 * time.Second)

	// Enable anomaly mode
	t.Log("Enabling anomaly mode...")
	err = enableMicroserviceAnomalyMode(true)
	if err != nil {
		t.Logf("Warning: Failed to enable anomaly mode: %v", err)
	}

	// Let it run with anomalies for 20 seconds
	time.Sleep(20 * time.Second)

	// Disable anomaly mode
	t.Log("Disabling anomaly mode...")
	err = enableMicroserviceAnomalyMode(false)
	if err != nil {
		t.Logf("Warning: Failed to disable anomaly mode: %v", err)
	}

	// Let it run normally for another 10 seconds
	time.Sleep(10 * time.Second)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	t.Log("Anomaly detection test completed successfully")
}

// TestMicroservicePerformance tests the framework performance with real microservice
func TestMicroservicePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Check if microservice is running
	if !isMicroserviceRunning() {
		t.Skip("Microservice not running, skipping test. Start with: cd test/microservice && go run main.go")
	}

	t.Log("Running performance test with microservice")

	// Create framework with multiple analyzers
	cfg := &core.FrameworkConfig{
		LogLevel:  "warn", // Reduce logging for performance
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Add multiple collectors and analyzers
	for i := 0; i < 3; i++ {
		collector := collectors.NewPrometheusCollector(fmt.Sprintf("collector-%d", i))
		collector.Configure(map[string]interface{}{
			"url":      "http://localhost:9090",
			"interval": "2s",
			"queries": []string{
				"cpu_usage_percent",
				"memory_usage_percent",
				"response_time_ms",
			},
		})
		framework.LoadPlugin(collector)

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Generate load
	go generateMicroserviceLoad()

	// Let it run
	time.Sleep(20 * time.Second)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	duration := time.Since(start)
	t.Logf("Performance test completed in %v", duration)

	// Performance assertions
	if duration > 30*time.Second {
		t.Errorf("Framework took too long to start/stop: %v", duration)
	}
}

// Helper functions

func isMicroserviceRunning() bool {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func generateMicroserviceLoad() {
	client := &http.Client{Timeout: 5 * time.Second}

	// Generate load for 60 seconds
	endTime := time.Now().Add(60 * time.Second)

	for time.Now().Before(endTime) {
		// Make requests to different endpoints
		endpoints := []string{
			"http://localhost:8080/api/users",
			"http://localhost:8080/api/products",
			"http://localhost:8080/api/orders",
		}

		for _, endpoint := range endpoints {
			resp, err := client.Get(endpoint)
			if err == nil {
				resp.Body.Close()
			}

			// Small delay between requests
			time.Sleep(100 * time.Millisecond)
		}

		// Occasionally create an order
		if time.Now().Unix()%5 == 0 {
			orderData := `{"product_id": 123, "quantity": 2, "price": 29.99}`
			resp, err := client.Post("http://localhost:8080/api/orders",
				"application/json", bytes.NewBufferString(orderData))
			if err == nil {
				resp.Body.Close()
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func enableMicroserviceAnomalyMode(enable bool) error {
	client := &http.Client{Timeout: 5 * time.Second}

	url := "http://localhost:8080/admin/anomaly"
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to toggle anomaly mode: %s", string(body))
	}

	return nil
}

func getMicroserviceStatus() (map[string]interface{}, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get("http://localhost:8080/admin/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Simple parsing (in a real test, you'd use JSON unmarshaling)
	status := map[string]interface{}{
		"response": string(body),
	}

	return status, nil
}
