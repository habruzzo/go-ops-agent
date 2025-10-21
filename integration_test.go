package main

import (
	"context"
	"testing"
	"time"

	"github.com/holden/agent/config"
	"github.com/holden/agent/core"
	"github.com/holden/agent/plugins/agents"
	"github.com/holden/agent/plugins/analyzers"
	"github.com/holden/agent/plugins/collectors"
	"github.com/holden/agent/plugins/responders"
)

func TestFullFrameworkIntegration(t *testing.T) {
	// Create a test configuration
	cfg := &core.FrameworkConfig{
		Logging: core.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []core.PluginConfig{},
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Add a mock collector
	collector := collectors.NewPrometheusCollector("test-prometheus")
	collector.Configure(map[string]interface{}{
		"url":      "http://localhost:9090",
		"interval": "100ms",
		"queries":  []string{"up"},
	})
	framework.LoadPlugin(collector)

	// Add an anomaly analyzer
	analyzer := analyzers.NewAnomalyAnalyzer("test-anomaly")
	analyzer.Configure(map[string]interface{}{
		"threshold": 2.0,
	})
	framework.LoadPlugin(analyzer)

	// Add a logger responder
	responder := responders.NewLoggerResponder("test-logger")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Verify framework is running
	status := framework.GetStatus()
	if status["running"] != true {
		t.Error("Expected framework to be running")
	}

	if status["total_plugins"] != 3 {
		t.Errorf("Expected 3 plugins, got %v", status["total_plugins"])
	}

	// Wait for some processing
	time.Sleep(500 * time.Millisecond)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	// Verify framework is stopped
	status = framework.GetStatus()
	if status["running"] != false {
		t.Error("Expected framework to be stopped")
	}
}

func TestFrameworkWithAIAgent(t *testing.T) {
	// Skip this test if no API key is provided
	if testing.Short() {
		t.Skip("Skipping AI agent test in short mode")
	}

	cfg := &core.FrameworkConfig{
		Logging: core.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Agent: core.AgentConfig{
			DefaultAgent: "test-ai",
		},
		Plugins: []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Add AI agent (this will fail without API key, but we can test the structure)
	agent := agents.NewAIAgent("test-ai")
	agent.Configure(map[string]interface{}{
		"api_key": "test-key",
		"model":   "gpt-3.5-turbo",
	})
	framework.LoadPlugin(agent)

	// Test agent query (will fail due to invalid API key, but tests the flow)
	ctx := context.Background()
	_, err := framework.QueryAgent(ctx, "test-ai", "What's the system status?")
	if err == nil {
		t.Error("Expected error due to invalid API key")
	}

	// Test default agent query
	_, err = framework.QueryDefaultAgent(ctx, "What's the system status?")
	if err == nil {
		t.Error("Expected error due to invalid API key")
	}
}

func TestFrameworkConfigurationLoading(t *testing.T) {
	// Test loading configuration from file
	cfg, err := config.LoadConfig("framework.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Logging.Level == "" {
		t.Error("Expected logging level to be set")
	}

	if len(cfg.Plugins) == 0 {
		t.Error("Expected plugins to be configured")
	}

	// Test default configuration
	defaultCfg := config.DefaultConfig()
	if defaultCfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", defaultCfg.Logging.Level)
	}
}

func TestFrameworkPluginLifecycle(t *testing.T) {
	cfg := &core.FrameworkConfig{
		Logging: core.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Test loading and unloading plugins
	analyzer := analyzers.NewAnomalyAnalyzer("test-analyzer")
	framework.LoadPlugin(analyzer)

	if len(framework.GetStatus()["plugins"].(map[string]interface{})) != 1 {
		t.Error("Expected 1 plugin after loading")
	}

	// Test unloading plugin
	err := framework.UnloadPlugin("test-analyzer")
	if err != nil {
		t.Fatalf("Failed to unload plugin: %v", err)
	}

	if len(framework.GetStatus()["plugins"].(map[string]interface{})) != 0 {
		t.Error("Expected 0 plugins after unloading")
	}
}

func TestFrameworkConcurrentAccess(t *testing.T) {
	cfg := &core.FrameworkConfig{
		Logging: core.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []core.PluginConfig{},
	}

	framework := core.NewFramework(cfg)

	// Test concurrent plugin loading
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			analyzer := analyzers.NewAnomalyAnalyzer("test-analyzer-" + string(rune(id)))
			framework.LoadPlugin(analyzer)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	status := framework.GetStatus()
	if status["total_plugins"] != 10 {
		t.Errorf("Expected 10 plugins, got %v", status["total_plugins"])
	}
}
