package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/habruzzo/agent/config"
	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/collectors"
	"github.com/habruzzo/agent/plugins/responders"
)

func realAIMain() {
	fmt.Println("Starting Real AI Agent for Observability")
	fmt.Println("========================================")

	// Check for required environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: OPENAI_API_KEY environment variable is required")
		fmt.Println("Get your API key from: https://platform.openai.com/api-keys")
		fmt.Println("Then run: export OPENAI_API_KEY=your_key_here")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := loadRealConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Add Prometheus collector (if Prometheus is running)
	if isPrometheusRunning() {
		fmt.Println("Found Prometheus, adding collector...")
		collector := collectors.NewPrometheusCollector("real-prometheus")
		collector.Configure(map[string]interface{}{
			"url":      "http://localhost:9090",
			"interval": "10s",
			"queries": []string{
				"up",
				"prometheus_tsdb_head_samples",
				"prometheus_tsdb_head_series",
				"prometheus_http_requests_total",
			},
		})
		framework.LoadPlugin(collector)
	} else {
		fmt.Println("Prometheus not found, using mock data...")
		// Add a mock collector for demonstration
		addMockCollector(framework)
	}

	// Add anomaly analyzer
	analyzer := analyzers.NewAnomalyAnalyzer("real-analyzer")
	analyzer.Configure(map[string]interface{}{
		"threshold": 2.0,
	})
	framework.LoadPlugin(analyzer)

	// Add logger responder
	responder := responders.NewLoggerResponder("real-logger")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Add real AI agent
	agent := agents.NewAIAgent("real-ai")
	agent.Configure(map[string]interface{}{
		"api_key":    apiKey,
		"model":      "gpt-4", // Use GPT-4 for better analysis
		"max_tokens": 500,
	})
	framework.LoadPlugin(agent)

	// Start framework
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Starting framework...")
	err = framework.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start framework: %v", err)
	}

	fmt.Println("Framework started successfully!")
	fmt.Println("")
	fmt.Println("AI Agent is ready! Try these queries:")
	fmt.Println("  - 'What's the current system status?'")
	fmt.Println("  - 'Are there any anomalies?'")
	fmt.Println("  - 'Give me recommendations'")
	fmt.Println("  - 'Analyze the performance'")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Interactive query loop
	go interactiveQueryLoop(ctx, framework)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down framework...")

	// Stop framework
	err = framework.Stop()
	if err != nil {
		log.Printf("Error stopping framework: %v", err)
	} else {
		fmt.Println("Framework stopped successfully")
	}
}

func loadRealConfig() (*core.FrameworkConfig, error) {
	// Try to load from file first
	cfg, err := config.LoadConfig("real-config.yaml")
	if err == nil {
		return cfg, nil
	}

	// Create default config if file doesn't exist
	cfg = &core.FrameworkConfig{
		LogLevel:     "info",
		LogFormat:    "text",
		LogOutput:    "stdout",
		DefaultAgent: "real-ai",
		Plugins:      []core.PluginConfig{},
	}

	// Save default config for future use
	err = config.SaveConfig(cfg, "real-config.yaml")
	if err != nil {
		log.Printf("Warning: Could not save config file: %v", err)
	}

	return cfg, nil
}

func isPrometheusRunning() bool {
	// Simple check - in a real implementation, you'd use proper HTTP client
	// For now, just check if the port is likely in use
	return false // Simplified for demo
}

func addMockCollector(framework *core.Framework) {
	// Create a simple mock collector that generates realistic data
	// This would be a custom collector in a real implementation
	fmt.Println("Mock collector would be added here")
}

func interactiveQueryLoop(ctx context.Context, framework *core.Framework) {
	// In a real implementation, you'd have an interactive CLI
	// For now, we'll just demonstrate with some example queries

	time.Sleep(5 * time.Second) // Wait for framework to be ready

	exampleQueries := []string{
		"What's the current system status?",
		"Are there any performance issues?",
		"Give me recommendations for optimization",
		"Analyze the system health",
	}

	for i, query := range exampleQueries {
		select {
		case <-ctx.Done():
			return
		default:
			if i > 0 {
				time.Sleep(10 * time.Second) // Wait between queries
			}

			fmt.Printf("\n--- AI Query %d ---\n", i+1)
			fmt.Printf("Query: %s\n", query)

			response, err := framework.QueryDefaultAgent(ctx, query)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if response != nil {
				fmt.Printf("AI Response: %s\n", response.Response)
			} else {
				fmt.Println("No response received")
			}
		}
	}
}
