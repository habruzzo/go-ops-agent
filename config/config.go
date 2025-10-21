package config

import (
	"fmt"
	"os"

	"github.com/holden/agent/core"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*core.FrameworkConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config core.FrameworkConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	return &config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *core.FrameworkConfig, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *core.FrameworkConfig {
	return &core.FrameworkConfig{
		Logging: core.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []core.PluginConfig{
			{
				Name:    "prometheus",
				Type:    "prometheus",
				Enabled: true,
				Config: map[string]interface{}{
					"url":      "http://localhost:9090",
					"interval": "30s",
					"queries": []string{
						"up",
						"cpu_usage_percent",
						"memory_usage_percent",
					},
				},
			},
			{
				Name:    "anomaly-detector",
				Type:    "anomaly",
				Enabled: true,
				Config: map[string]interface{}{
					"threshold": 2.0,
				},
			},
			{
				Name:    "logger",
				Type:    "log",
				Enabled: true,
				Config: map[string]interface{}{
					"level": "info",
				},
			},
			{
				Name:    "ai-agent",
				Type:    "ai",
				Enabled: false, // Disabled by default since it needs API key
				Config: map[string]interface{}{
					"api_key": "your-openai-api-key-here",
					"model":   "gpt-3.5-turbo",
				},
			},
		},
		Agent: core.AgentConfig{
			DefaultAgent: "ai-agent",
		},
	}
}
