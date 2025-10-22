package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
)

func ragMain() {
	fmt.Println("RAG (Retrieval-Augmented Generation) Demo")
	fmt.Println("========================================")

	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: OPENAI_API_KEY environment variable is required")
		fmt.Println("Get your API key from: https://platform.openai.com/api-keys")
		os.Exit(1)
	}

	// Create RAG agent
	ragAgent := agents.NewRAGAgent("demo-rag-agent")

	// Configure the agent
	ragAgent.Configure(map[string]interface{}{
		"api_key":    apiKey,
		"model":      "gpt-4",
		"max_tokens": 500,
	})

	// Start the agent
	ctx := context.Background()
	err := ragAgent.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start RAG agent: %v", err)
	}
	defer ragAgent.Stop()

	fmt.Println("RAG Agent started successfully!")
	fmt.Println("")

	// Add some sample knowledge to the knowledge base
	fmt.Println("Adding sample knowledge to the knowledge base...")

	// Add system documentation
	ragAgent.AddDocument(agents.Document{
		ID:      "system-docs-1",
		Content: "The production system runs on Kubernetes with 3 replicas. CPU threshold is 80%, memory threshold is 85%. The system uses Prometheus for monitoring and Grafana for visualization.",
		Metadata: map[string]interface{}{
			"type":     "documentation",
			"category": "system",
			"source":   "internal-docs",
		},
		Timestamp: time.Now(),
	})

	// Add incident history
	ragAgent.AddDocument(agents.Document{
		ID:      "incident-2024-01",
		Content: "Incident on 2024-01-15: CPU spike to 95% caused by memory leak in user service. Resolution: Restarted user service pods and updated memory limits. Downtime: 15 minutes.",
		Metadata: map[string]interface{}{
			"type":     "incident",
			"category": "system",
			"source":   "incident-reports",
			"severity": "high",
		},
		Timestamp: time.Now(),
	})

	// Add performance benchmarks
	ragAgent.AddDocument(agents.Document{
		ID:      "benchmark-2024",
		Content: "Performance benchmarks: Normal response time is 200ms, 95th percentile is 500ms. Normal CPU usage is 40-60%, memory usage is 50-70%. Database connection pool size is 100.",
		Metadata: map[string]interface{}{
			"type":     "benchmark",
			"category": "performance",
			"source":   "performance-tests",
		},
		Timestamp: time.Now(),
	})

	// Add some mock metrics data
	fmt.Println("Adding sample metrics data...")
	sampleMetrics := []core.DataPoint{
		{
			Metric:    "cpu_usage_percent",
			Value:     85.0,
			Source:    "prometheus",
			Timestamp: time.Now(),
		},
		{
			Metric:    "memory_usage_percent",
			Value:     75.0,
			Source:    "prometheus",
			Timestamp: time.Now(),
		},
		{
			Metric:    "response_time_ms",
			Value:     800.0,
			Source:    "prometheus",
			Timestamp: time.Now(),
		},
		{
			Metric:    "error_rate_percent",
			Value:     2.5,
			Source:    "prometheus",
			Timestamp: time.Now(),
		},
	}

	ragAgent.AddMetricsData(sampleMetrics)

	// Show knowledge base stats
	stats := ragAgent.GetKnowledgeBaseStats()
	fmt.Printf("Knowledge base stats: %+v\n", stats)
	fmt.Println("")

	// Demo queries
	queries := []string{
		"What are the normal performance thresholds for this system?",
		"I see CPU at 85% and memory at 75%, is this normal?",
		"Have we had similar incidents before?",
		"What should I do if response time is 800ms?",
		"Are there any performance issues I should be aware of?",
	}

	for i, query := range queries {
		fmt.Printf("--- Query %d ---\n", i+1)
		fmt.Printf("Question: %s\n", query)

		// Process query with RAG
		response, err := ragAgent.ProcessQueryWithRAG(ctx, query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if response != nil {
			fmt.Printf("RAG Response: %s\n", response.Response)
			fmt.Printf("Confidence: %.2f\n", response.Confidence)
			if sources, ok := response.Metadata["rag_sources"].([]string); ok {
				fmt.Printf("Sources: %v\n", sources)
			}
			if docCount, ok := response.Metadata["rag_documents_used"].(int); ok {
				fmt.Printf("Documents used: %d\n", docCount)
			}
		} else {
			fmt.Println("No response received")
		}

		fmt.Println("")
		time.Sleep(2 * time.Second) // Pause between queries
	}

	fmt.Println("RAG Demo completed!")
	fmt.Println("")
	fmt.Println("Key RAG Benefits Demonstrated:")
	fmt.Println("1. Context-aware responses using real system data")
	fmt.Println("2. Historical incident knowledge for better recommendations")
	fmt.Println("3. Performance benchmarks for accurate threshold analysis")
	fmt.Println("4. Source attribution for transparency")
	fmt.Println("5. Confidence scoring for response quality")
}
