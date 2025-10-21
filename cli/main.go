package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/holden/agent/config"
	"github.com/holden/agent/core"
	"github.com/holden/agent/plugins/agents"
	"github.com/holden/agent/plugins/analyzers"
	"github.com/holden/agent/plugins/collectors"
	"github.com/holden/agent/plugins/responders"
)

func main() {
	var (
		configFile   = flag.String("config", "framework.yaml", "Path to configuration file")
		createConfig = flag.Bool("create-config", false, "Create a default configuration file")
		version      = flag.Bool("version", false, "Show version information")
		interactive  = flag.Bool("interactive", false, "Start in interactive mode")
	)
	flag.Parse()

	if *version {
		fmt.Println("Observability Framework v0.2.0")
		return
	}

	if *createConfig {
		if err := createDefaultConfig(*configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created default configuration file: %s\n", *configFile)
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Load plugins
	for _, pluginConfig := range cfg.Plugins {
		if !pluginConfig.Enabled {
			continue
		}

		plugin, err := createPlugin(pluginConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create plugin %s: %v\n", pluginConfig.Name, err)
			continue
		}

		if err := framework.LoadPlugin(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load plugin %s: %v\n", pluginConfig.Name, err)
		}
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down framework...")
		cancel()
	}()

	// Start framework
	if err := framework.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start framework: %v\n", err)
		os.Exit(1)
	}

	// Interactive mode
	if *interactive {
		runInteractiveMode(ctx, framework)
	} else {
		// Wait for shutdown
		<-ctx.Done()
	}

	// Stop framework
	if err := framework.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping framework: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Framework stopped")
}

func createDefaultConfig(filename string) error {
	cfg := config.DefaultConfig()
	return config.SaveConfig(cfg, filename)
}

func createPlugin(config core.PluginConfig) (core.Plugin, error) {
	switch config.Type {
	case "prometheus":
		plugin := collectors.NewPrometheusCollector(config.Name)
		if err := plugin.Configure(config.Config); err != nil {
			return nil, err
		}
		return plugin, nil
	case "anomaly":
		plugin := analyzers.NewAnomalyAnalyzer(config.Name)
		if err := plugin.Configure(config.Config); err != nil {
			return nil, err
		}
		return plugin, nil
	case "log":
		plugin := responders.NewLoggerResponder(config.Name)
		if err := plugin.Configure(config.Config); err != nil {
			return nil, err
		}
		return plugin, nil
	case "ai":
		plugin := agents.NewAIAgent(config.Name)
		if err := plugin.Configure(config.Config); err != nil {
			return nil, err
		}
		return plugin, nil
	default:
		return nil, fmt.Errorf("unknown plugin type: %s", config.Type)
	}
}

func runInteractiveMode(ctx context.Context, framework *core.Framework) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Observability Framework - Interactive Mode")
	fmt.Println("Type 'help' for available commands, 'quit' to exit")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]

		switch command {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "help":
			printHelp()
		case "status":
			printStatus(framework)
		case "query":
			if len(parts) < 2 {
				fmt.Println("Usage: query <your question>")
				continue
			}
			query := strings.Join(parts[1:], " ")
			handleQuery(ctx, framework, query)
		case "agents":
			listAgents(framework)
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help     - Show this help message")
	fmt.Println("  status   - Show framework status")
	fmt.Println("  query    - Ask a question to the AI agent")
	fmt.Println("  agents   - List available agents")
	fmt.Println("  quit     - Exit the framework")
}

func printStatus(framework *core.Framework) {
	status := framework.GetStatus()
	fmt.Printf("Framework Status: %v\n", status["running"])
	fmt.Printf("Total Plugins: %v\n", status["total_plugins"])
	fmt.Printf("Collectors: %v\n", status["collectors"])
	fmt.Printf("Analyzers: %v\n", status["analyzers"])
	fmt.Printf("Responders: %v\n", status["responders"])
	fmt.Printf("Agents: %v\n", status["agents"])
}

func handleQuery(ctx context.Context, framework *core.Framework, query string) {
	response, err := framework.QueryDefaultAgent(ctx, query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Agent Response (%.0f%% confidence):\n", response.Confidence*100)
	fmt.Println(response.Response)

	if len(response.Actions) > 0 {
		fmt.Println("\nSuggested Actions:")
		for i, action := range response.Actions {
			fmt.Printf("  %d. %s: %s\n", i+1, action.Type, action.Description)
		}
	}
}

func listAgents(framework *core.Framework) {
	status := framework.GetStatus()
	plugins := status["plugins"].(map[string]interface{})

	fmt.Println("Available Agents:")
	for name, pluginInfo := range plugins {
		info := pluginInfo.(map[string]interface{})
		if info["type"] == "agent" {
			fmt.Printf("  - %s (Status: %s)\n", name, info["status"])
		}
	}
}
