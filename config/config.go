package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/habruzzo/agent/core"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML file with environment variable overrides
func LoadConfig(filename string) (*core.FrameworkConfig, error) {
	// Start with default configuration
	config := DefaultConfig()

	// Load from YAML file if it exists
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, core.NewConfigurationError("config", "load", fmt.Sprintf("failed to read config file: %v", err))
		}

		// Parse YAML into the config struct
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, core.NewConfigurationError("config", "parse", fmt.Sprintf("failed to parse config file: %v", err))
		}
	}

	// Override with environment variables (this handles the env tags)
	if err := env.Parse(config); err != nil {
		return nil, core.NewConfigurationError("config", "env-parse", fmt.Sprintf("failed to parse environment variables: %v", err))
	}

	// Validate configuration
	if err := core.ValidateFrameworkConfig(config); err != nil {
		return nil, err
	}

	// Load plugin configurations
	if err := loadPluginConfigs(config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfigFromEnv loads configuration from environment variables only
func LoadConfigFromEnv() (*core.FrameworkConfig, error) {
	// Start with default configuration
	config := DefaultConfig()

	// Override with environment variables
	if err := env.Parse(config); err != nil {
		return nil, core.NewConfigurationError("config", "env-parse", fmt.Sprintf("failed to parse environment variables: %v", err))
	}

	// Validate configuration
	if err := core.ValidateFrameworkConfig(config); err != nil {
		return nil, err
	}

	// Load plugin configurations
	if err := loadPluginConfigs(config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *core.FrameworkConfig, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return core.NewConfigurationError("config", "save", fmt.Sprintf("failed to marshal config: %v", err))
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return core.NewConfigurationError("config", "save", fmt.Sprintf("failed to write config file: %v", err))
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *core.FrameworkConfig {
	config := &core.FrameworkConfig{
		LogLevel:           "info",
		LogFormat:          "text",
		LogOutput:          "stdout",
		ServerHost:         "0.0.0.0",
		ServerPort:         9090,
		DefaultAgent:       "",
		AIAPIKey:           "",
		AIAPIURL:           "https://api.openai.com/v1",
		PrometheusEnabled:  true,
		PrometheusURL:      "http://localhost:9090",
		PluginConfigFile:   "plugins.yaml",
		HealthCheckTimeout: 5 * time.Second,
		DataChannelSize:    100,
		WorkerPoolSize:     4,
		ShutdownTimeout:    30 * time.Second,
		Plugins:            getDefaultPluginConfigs(),
	}

	return config
}

// loadPluginConfigs loads plugin configurations from file or environment
func loadPluginConfigs(config *core.FrameworkConfig) error {
	// If plugins are already configured in the config, use them
	if len(config.Plugins) > 0 {
		return nil
	}

	// Try to load from file first
	if _, err := os.Stat(config.PluginConfigFile); err == nil {
		plugins, err := LoadPluginConfigsFromFile(config.PluginConfigFile)
		if err != nil {
			return err
		}
		config.Plugins = plugins
		return nil
	}

	// Fall back to environment-based plugin configuration
	plugins, err := getPluginConfigsFromEnv()
	if err != nil {
		return err
	}
	config.Plugins = plugins

	return nil
}

// getPluginConfigsFromEnv creates plugin configurations from environment variables
func getPluginConfigsFromEnv() ([]core.PluginConfig, error) {
	var plugins []core.PluginConfig

	// Check for individual plugin environment variables
	pluginTypes := []string{"collector", "analyzer", "responder", "agent"}

	for _, pluginType := range pluginTypes {
		envVar := fmt.Sprintf("AGENT_%s_PLUGINS", strings.ToUpper(pluginType))
		if pluginList := os.Getenv(envVar); pluginList != "" {
			pluginNames := strings.Split(pluginList, ",")
			for _, name := range pluginNames {
				name = strings.TrimSpace(name)
				if name != "" {
					config, err := createPluginConfigFromEnv(name, pluginType)
					if err != nil {
						return nil, err
					}
					plugins = append(plugins, config)
				}
			}
		}
	}

	// If no plugins configured via environment, return default set
	if len(plugins) == 0 {
		plugins = getDefaultPluginConfigs()
	}

	return plugins, nil
}

// createPluginConfigFromEnv creates a plugin configuration using caarlos0/env
func createPluginConfigFromEnv(pluginName, pluginType string) (core.PluginConfig, error) {
	config := core.PluginConfig{
		Name:    pluginName,
		Type:    pluginType,
		Enabled: true,
	}

	// Create the appropriate config struct based on plugin type and name
	var pluginConfig interface{}
	switch pluginType {
	case "collector":
		if pluginName == "prometheus" {
			pluginConfig = &core.PrometheusCollectorConfig{}
		} else {
			// Generic collector config
			pluginConfig = make(map[string]interface{})
		}
	case "analyzer":
		if pluginName == "anomaly" {
			pluginConfig = &core.AnomalyAnalyzerConfig{}
		} else {
			// Generic analyzer config
			pluginConfig = make(map[string]interface{})
		}
	case "responder":
		if pluginName == "logger" {
			pluginConfig = &core.LoggerResponderConfig{}
		} else {
			// Generic responder config
			pluginConfig = make(map[string]interface{})
		}
	case "agent":
		if pluginName == "ai" {
			pluginConfig = &core.AIAgentConfig{}
		} else if pluginName == "rag" {
			pluginConfig = &core.RAGAgentConfig{}
		} else {
			// Generic agent config
			pluginConfig = make(map[string]interface{})
		}
	default:
		pluginConfig = make(map[string]interface{})
	}

	// Parse environment variables into the config struct (only for structured configs)
	if _, ok := pluginConfig.(map[string]interface{}); !ok {
		if err := env.Parse(pluginConfig); err != nil {
			return config, core.NewConfigurationError("config", "plugin-env-parse",
				fmt.Sprintf("failed to parse environment variables for plugin %s: %v", pluginName, err))
		}
	}

	// Validate the plugin config if it has validation tags
	if validator, ok := pluginConfig.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return config, core.NewValidationError("config", "plugin-validate",
				fmt.Sprintf("validation failed for plugin %s: %v", pluginName, err))
		}
	}

	config.Config = pluginConfig
	return config, nil
}

// getDefaultPluginConfigs returns a default set of plugin configurations
func getDefaultPluginConfigs() []core.PluginConfig {
	return []core.PluginConfig{
		{
			Name: "prometheus-collector",
			Type: "collector",
			Config: &core.PrometheusCollectorConfig{
				URL:            "http://localhost:9090",
				ScrapeInterval: 30 * time.Second,
				Timeout:        10 * time.Second,
			},
			Enabled: true,
		},
		{
			Name: "anomaly-analyzer",
			Type: "analyzer",
			Config: &core.AnomalyAnalyzerConfig{
				Threshold:  0.8,
				WindowSize: 100,
				Algorithm:  "statistical",
			},
			Enabled: true,
		},
		{
			Name: "logger-responder",
			Type: "responder",
			Config: &core.LoggerResponderConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			Enabled: true,
		},
		{
			Name: "ai-agent",
			Type: "agent",
			Config: &core.AIAgentConfig{
				Model:       "gpt-4",
				MaxTokens:   1000,
				Temperature: 0.7,
				APIKey:      "", // Will be loaded from environment
				APIURL:      "https://api.openai.com/v1",
			},
			Enabled: true,
		},
	}
}

// GetConfigSummary returns a summary of the current configuration
func GetConfigSummary(config *core.FrameworkConfig) map[string]interface{} {
	return map[string]interface{}{
		"logging": map[string]interface{}{
			"level":  config.LogLevel,
			"format": config.LogFormat,
			"output": config.LogOutput,
		},
		"server": map[string]interface{}{
			"host": config.ServerHost,
			"port": config.ServerPort,
		},
		"agent": map[string]interface{}{
			"default_agent": config.DefaultAgent,
			"ai_api_url":    config.AIAPIURL,
			"has_api_key":   config.AIAPIKey != "",
		},
		"prometheus": map[string]interface{}{
			"enabled": config.PrometheusEnabled,
			"url":     config.PrometheusURL,
		},
		"processing": map[string]interface{}{
			"data_channel_size": config.DataChannelSize,
			"worker_pool_size":  config.WorkerPoolSize,
			"shutdown_timeout":  config.ShutdownTimeout.String(),
		},
		"plugins": map[string]interface{}{
			"count": len(config.Plugins),
			"types": getPluginTypeCounts(config.Plugins),
		},
	}
}

// getPluginTypeCounts returns a count of plugins by type
func getPluginTypeCounts(plugins []core.PluginConfig) map[string]int {
	counts := make(map[string]int)
	for _, plugin := range plugins {
		counts[plugin.Type]++
	}
	return counts
}

// Legacy functions for backward compatibility

// LoadPluginConfigsFromFile loads plugin configurations from a YAML file
func LoadPluginConfigsFromFile(filename string) ([]core.PluginConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, core.NewConfigurationError("config", "load-plugins", fmt.Sprintf("failed to read plugin config file: %v", err))
	}

	var plugins []core.PluginConfig
	if err := yaml.Unmarshal(data, &plugins); err != nil {
		return nil, core.NewConfigurationError("config", "parse-plugins", fmt.Sprintf("failed to parse plugin config file: %v", err))
	}

	return plugins, nil
}
