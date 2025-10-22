package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/responders"
)

// MockAIServer creates a mock AI API server for testing
func MockAIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the request
		var request struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Generate a mock response based on the query
		query := ""
		if len(request.Messages) > 0 {
			query = request.Messages[len(request.Messages)-1].Content
		}

		response := generateMockAIResponse(query)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func generateMockAIResponse(query string) map[string]interface{} {
	// Generate contextual responses based on the query
	response := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": generateResponseContent(query),
				},
			},
		},
		"usage": map[string]interface{}{
			"total_tokens": 150,
		},
	}
	return response
}

func generateResponseContent(query string) string {
	query = strings.ToLower(query)

	switch {
	case strings.Contains(query, "cpu") || strings.Contains(query, "performance"):
		return "I've analyzed the CPU metrics and detected a spike to 85% usage at 19:35:00. This is 3.2σ above normal and indicates potential performance issues. I recommend investigating recent deployments or increased user load."

	case strings.Contains(query, "memory") || strings.Contains(query, "leak"):
		return "Memory usage shows a gradual increase pattern consistent with a memory leak. Usage has grown from 40% to 65% over the past hour. I suggest checking for unclosed connections or growing data structures."

	case strings.Contains(query, "anomaly") || strings.Contains(query, "alert"):
		return "I've identified 3 anomalies in the last 10 minutes: 1 CPU spike (85%), 1 memory increase (65%), and 1 slow response (2.3s). The CPU spike is the most critical and requires immediate attention."

	case strings.Contains(query, "status") || strings.Contains(query, "health"):
		return "System status: Overall health is DEGRADED. CPU usage is elevated with recent spikes, memory shows leak patterns, but disk and network are normal. I recommend investigating the CPU and memory issues."

	case strings.Contains(query, "recommend") || strings.Contains(query, "suggest"):
		return "Based on the current metrics, I recommend: 1) Investigate the CPU spike source, 2) Check for memory leaks in recent deployments, 3) Monitor response times closely, 4) Consider scaling if patterns continue."

	default:
		return "I've analyzed the system metrics and found several areas of concern. CPU usage has spiked to 85%, memory shows gradual increase patterns, and response times are elevated. I recommend investigating these issues to prevent service degradation."
	}
}

// TestAIAgentWithMockServer tests the AI agent with a mock AI service
func TestAIAgentWithMockServer(t *testing.T) {
	// Start mock AI server
	mockServer := MockAIServer()
	defer mockServer.Close()

	t.Logf("Mock AI server running at: %s", mockServer.URL)

	// Create AI agent
	agent := agents.NewAIAgent("test-ai")
	agent.Configure(map[string]interface{}{
		"api_key": "test-key",
		"api_url": mockServer.URL,
		"model":   "gpt-3.5-turbo",
	})

	// Start the agent
	ctx := context.Background()
	err := agent.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start AI agent: %v", err)
	}
	defer agent.Stop()

	// Test queries
	testQueries := []string{
		"What's the current system status?",
		"I see CPU spikes, what should I do?",
		"Are there any memory leaks?",
		"Give me recommendations for the anomalies",
		"Analyze the performance issues",
	}

	for _, query := range testQueries {
		t.Logf("Testing query: %s", query)

		response, err := agent.ProcessQuery(ctx, query)
		if err != nil {
			t.Errorf("Query failed: %v", err)
			continue
		}

		if response == nil {
			t.Error("Expected response, got nil")
			continue
		}

		t.Logf("AI Response: %s", response.Response)

		// Verify response has content
		if response.Response == "" {
			t.Error("Expected non-empty response")
		}
	}

	t.Log("AI agent test completed successfully")
}

// TestAIAgentWithFramework tests the AI agent integrated with the framework
func TestAIAgentWithFramework(t *testing.T) {
	// Start mock AI server
	mockServer := MockAIServer()
	defer mockServer.Close()

	t.Log("Testing AI agent with framework integration")

	// Create framework
	cfg := &core.FrameworkConfig{
		LogLevel:     "info",
		LogFormat:    "text",
		LogOutput:    "stdout",
		DefaultAgent: "test-ai",
		Plugins:      []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Add AI agent
	agent := agents.NewAIAgent("test-ai")
	agent.Configure(map[string]interface{}{
		"api_key": "test-key",
		"api_url": mockServer.URL,
		"model":   "gpt-3.5-turbo",
	})
	framework.LoadPlugin(agent)

	// Add analyzer and responder
	analyzer := analyzers.NewAnomalyAnalyzer("test-analyzer")
	analyzer.Configure(map[string]interface{}{"threshold": 2.0})
	framework.LoadPlugin(analyzer)

	responder := responders.NewLoggerResponder("test-logger")
	responder.Configure(map[string]interface{}{"level": "info"})
	framework.LoadPlugin(responder)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Test AI agent queries through framework
	queries := []string{
		"What anomalies have been detected?",
		"Give me a system health summary",
		"What should I investigate first?",
	}

	for _, query := range queries {
		t.Logf("Framework query: %s", query)

		response, err := framework.QueryDefaultAgent(ctx, query)
		if err != nil {
			t.Errorf("Framework query failed: %v", err)
			continue
		}

		if response == nil {
			t.Error("Expected response, got nil")
			continue
		}

		t.Logf("Framework AI Response: %s", response.Response)
	}

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	t.Log("AI agent framework integration test completed successfully")
}

// TestAIAgentErrorHandling tests error scenarios
func TestAIAgentErrorHandling(t *testing.T) {
	// Create AI agent with invalid configuration
	agent := agents.NewAIAgent("test-ai")
	agent.Configure(map[string]interface{}{
		"api_key": "invalid-key",
		"api_url": "http://invalid-url:9999",
		"model":   "gpt-3.5-turbo",
	})

	// Start the agent - this should fail
	ctx := context.Background()
	err := agent.Start(ctx)
	if err == nil {
		t.Error("Expected error for invalid configuration, got nil")
		agent.Stop() // Clean up if it somehow started
		return
	}

	t.Logf("Expected error occurred: %v", err)

	// Test that agent is not running
	if agent.Status() == core.PluginStatusRunning {
		t.Error("Expected agent to not be running with invalid config")
	}

	t.Log("AI agent error handling test completed successfully")
}
