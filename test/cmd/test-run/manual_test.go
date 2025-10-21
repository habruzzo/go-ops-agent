package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/holden/agent/config"
	"github.com/holden/agent/core"
	"github.com/holden/agent/plugins/agents"
	"github.com/holden/agent/plugins/analyzers"
	"github.com/holden/agent/plugins/collectors"
	"github.com/holden/agent/plugins/responders"
)

func main() {
	fmt.Println("🚀 Starting Manual End-to-End Test")
	fmt.Println("==================================")

	// Load configuration
	cfg, err := config.LoadConfig("test-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Add Prometheus collector
	collector := collectors.NewPrometheusCollector("prometheus-collector")
	collector.Configure(map[string]interface{}{
		"url":      "http://localhost:9090",
		"interval": "10s",
		"queries":  []string{"up", "cpu_usage_percent", "memory_usage_percent"},
	})
	framework.LoadPlugin(collector)

	// Add anomaly analyzer
	analyzer := analyzers.NewAnomalyAnalyzer("anomaly-analyzer")
	analyzer.Configure(map[string]interface{}{
		"threshold": 2.0,
	})
	framework.LoadPlugin(analyzer)

	// Add logger responder
	responder := responders.NewLoggerResponder("logger-responder")
	responder.Configure(map[string]interface{}{
		"level": "info",
	})
	framework.LoadPlugin(responder)

	// Add AI agent
	agent := agents.NewAIAgent("ai-agent")
	agent.Configure(map[string]interface{}{
		"api_key":    "test-key",
		"model":      "gpt-3.5-turbo",
		"max_tokens": 150,
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
	fmt.Println("Press Ctrl+C to stop...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Print status every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				status := framework.GetStatus()
				fmt.Printf("\n📊 Framework Status: %+v\n", status)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n🛑 Shutting down framework...")

	// Stop framework
	err = framework.Stop()
	if err != nil {
		log.Printf("Error stopping framework: %v", err)
	} else {
		fmt.Println("✅ Framework stopped successfully")
	}
}
