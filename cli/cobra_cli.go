package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/habruzzo/agent/config"
	"github.com/habruzzo/agent/core"
	"github.com/spf13/cobra"
)

// CLI represents the command-line interface
type CLI struct {
	rootCmd *cobra.Command
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{}
	cli.setupCommands()
	return cli
}

// setupCommands sets up all CLI commands
func (c *CLI) setupCommands() {
	c.rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "AI Agent Framework - Production-ready observability and automation",
		Long: `AI Agent Framework is a modular, extensible system for building AI-powered 
observability agents. It follows modern software engineering principles including 
dependency injection, interface-based design, and comprehensive error handling.`,
		Version: "0.2.0",
	}

	// Add subcommands
	c.rootCmd.AddCommand(c.createStartCommand())
	c.rootCmd.AddCommand(c.createConfigCommand())
	c.rootCmd.AddCommand(c.createVersionCommand())
	c.rootCmd.AddCommand(c.createInteractiveCommand())
	c.rootCmd.AddCommand(c.createHealthCommand())
	c.rootCmd.AddCommand(c.createStatusCommand())
}

// createStartCommand creates the start command
func (c *CLI) createStartCommand() *cobra.Command {
	var configFile string
	var useEnv bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the agent framework",
		Long:  "Start the agent framework with the specified configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFramework(configFile, useEnv)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "framework.yaml", "Path to configuration file")
	cmd.Flags().BoolVarP(&useEnv, "env", "e", false, "Use environment variables for configuration")

	return cmd
}

// createConfigCommand creates the config command
func (c *CLI) createConfigCommand() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Manage framework configuration files and environment variables",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createDefaultConfig(outputFile)
		},
	}
	createCmd.Flags().StringVarP(&outputFile, "output", "o", "framework.yaml", "Output file path")

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.validateConfig(outputFile)
		},
	}
	validateCmd.Flags().StringVarP(&outputFile, "config", "c", "framework.yaml", "Configuration file to validate")

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.showConfig()
		},
	}

	cmd.AddCommand(createCmd, validateCmd, showCmd)
	return cmd
}

// createVersionCommand creates the version command
func (c *CLI) createVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("AI Agent Framework v0.2.0")
			fmt.Println("Built with Go", "1.21+")
		},
	}
}

// createInteractiveCommand creates the interactive command
func (c *CLI) createInteractiveCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Start in interactive mode",
		Long:  "Start the framework in interactive mode for testing and development",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startInteractive(configFile)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "framework.yaml", "Path to configuration file")

	return cmd
}

// createHealthCommand creates the health command
func (c *CLI) createHealthCommand() *cobra.Command {
	var host string
	var port int
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check framework health",
		Long:  "Check the health status of a running framework instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.checkHealth(host, port, timeout)
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "Framework host")
	cmd.Flags().IntVar(&port, "port", 9090, "Framework port")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "Request timeout")

	return cmd
}

// createStatusCommand creates the status command
func (c *CLI) createStatusCommand() *cobra.Command {
	var host string
	var port int

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show framework status",
		Long:  "Show detailed status information about a running framework instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.showStatus(host, port)
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "Framework host")
	cmd.Flags().IntVar(&port, "port", 9090, "Framework port")

	return cmd
}

// startFramework starts the framework
func (c *CLI) startFramework(configFile string, useEnv bool) error {
	var frameworkConfig *core.FrameworkConfig
	var err error

	if useEnv {
		// Load configuration from environment variables only
		frameworkConfig, err = config.LoadConfigFromEnv()
		if err != nil {
			return fmt.Errorf("failed to load config from environment: %w", err)
		}
	} else {
		// Load configuration from file (with environment variable overrides)
		frameworkConfig, err = config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Create framework
	framework := core.NewFramework(frameworkConfig)

	// Register plugin creators
	registerPluginCreators(framework.GetFactory())

	// Load plugins
	if err := loadPluginsFromConfig(framework, frameworkConfig); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping framework...")
		cancel()
	}()

	// Start framework
	if err := framework.Start(ctx); err != nil {
		return fmt.Errorf("failed to start framework: %w", err)
	}

	// Wait for shutdown
	<-ctx.Done()

	// Stop framework
	if err := framework.Stop(); err != nil {
		return fmt.Errorf("failed to stop framework: %w", err)
	}

	fmt.Println("Framework stopped successfully")
	return nil
}

// startInteractive starts the framework in interactive mode
func (c *CLI) startInteractive(configFile string) error {
	// Load configuration
	frameworkConfig, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create framework
	framework := core.NewFramework(frameworkConfig)

	// Register plugin creators
	registerPluginCreators(framework.GetFactory())

	// Load plugins
	if err := loadPluginsFromConfig(framework, frameworkConfig); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Start framework
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := framework.Start(ctx); err != nil {
		return fmt.Errorf("failed to start framework: %w", err)
	}

	fmt.Println("Framework started in interactive mode. Type 'help' for commands, 'quit' to exit.")

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("agent> ")
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
		case "help":
			c.showInteractiveHelp()
		case "status":
			c.showFrameworkStatus(framework)
		case "plugins":
			c.showPluginStatus(framework)
		case "query":
			if len(parts) > 1 {
				query := strings.Join(parts[1:], " ")
				c.processQuery(framework, query)
			} else {
				fmt.Println("Usage: query <your question>")
			}
		case "quit", "exit":
			fmt.Println("Shutting down framework...")
			cancel()
			framework.Stop()
			return nil
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}

	return nil
}

// Execute runs the CLI
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}

// Helper functions for interactive mode
func (c *CLI) showInteractiveHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help     - Show this help message")
	fmt.Println("  status   - Show framework status")
	fmt.Println("  plugins  - Show plugin information")
	fmt.Println("  query    - Send a query to the default agent")
	fmt.Println("  quit     - Exit the framework")
}

func (c *CLI) showFrameworkStatus(framework *core.Framework) {
	status := framework.GetStatus()
	fmt.Printf("Framework Status: %v\n", status["running"])
	fmt.Printf("Uptime: %v\n", status["uptime"])
	fmt.Printf("Plugin Count: %v\n", status["total_plugins"])
}

func (c *CLI) showPluginStatus(framework *core.Framework) {
	status := framework.GetStatus()
	fmt.Println("Plugins:")
	if pluginCounts, ok := status["plugin_counts"].(map[string]int); ok {
		for pluginType, count := range pluginCounts {
			fmt.Printf("  %s: %d\n", pluginType, count)
		}
	}
}

func (c *CLI) processQuery(framework *core.Framework, query string) {
	response, err := framework.QueryDefaultAgent(context.Background(), query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", response.Response)
	if response.Confidence < 0.8 {
		fmt.Printf("Confidence: %.2f (low confidence)\n", response.Confidence)
	}
}

// checkHealth checks the health of a running framework
func (c *CLI) checkHealth(host string, port int, timeout time.Duration) error {
	// Implementation would make HTTP request to health endpoint
	fmt.Printf("Checking health at %s:%d (timeout: %v)\n", host, port, timeout)
	// TODO: Implement actual health check
	return nil
}

// showStatus shows the status of a running framework
func (c *CLI) showStatus(host string, port int) error {
	// Implementation would make HTTP request to status endpoint
	fmt.Printf("Getting status from %s:%d\n", host, port)
	// TODO: Implement actual status check
	return nil
}

// validateConfig validates a configuration file
func (c *CLI) validateConfig(configFile string) error {
	_, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		return err
	}
	fmt.Println("Configuration is valid")
	return nil
}

// showConfig shows the current configuration
func (c *CLI) showConfig() error {
	frameworkConfig, err := config.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	summary := config.GetConfigSummary(frameworkConfig)
	fmt.Println("Current Configuration:")
	// TODO: Pretty print the configuration summary
	fmt.Printf("%+v\n", summary)
	return nil
}
