package config

import (
	"os"
	"testing"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expected: map[string]interface{}{
				"log_level":            "info",
				"log_format":           "text",
				"log_output":           "stdout",
				"server_host":          "0.0.0.0",
				"server_port":          9090,
				"default_agent":        "",
				"ai_api_key":           "",
				"ai_api_url":           "https://api.openai.com/v1",
				"prometheus_enabled":   true,
				"prometheus_url":       "http://localhost:9090",
				"plugin_config_file":   "plugins.yaml",
				"health_check_timeout": 5 * time.Second,
				"data_channel_size":    100,
				"worker_pool_size":     4,
				"shutdown_timeout":     30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"AGENT_LOG_LEVEL":          "debug",
				"AGENT_LOG_FORMAT":         "json",
				"AGENT_SERVER_PORT":        "8080",
				"AGENT_DEFAULT_AGENT":      "test-agent",
				"AGENT_AI_API_KEY":         "test-key",
				"AGENT_PROMETHEUS_ENABLED": "false",
				"AGENT_HEALTH_TIMEOUT":     "10s",
				"AGENT_DATA_CHANNEL_SIZE":  "200",
				"AGENT_WORKER_POOL_SIZE":   "8",
				"AGENT_SHUTDOWN_TIMEOUT":   "60s",
			},
			expected: map[string]interface{}{
				"log_level":            "debug",
				"log_format":           "json",
				"log_output":           "stdout",
				"server_host":          "0.0.0.0",
				"server_port":          8080,
				"default_agent":        "test-agent",
				"ai_api_key":           "test-key",
				"ai_api_url":           "https://api.openai.com/v1",
				"prometheus_enabled":   false,
				"prometheus_url":       "http://localhost:9090",
				"plugin_config_file":   "plugins.yaml",
				"health_check_timeout": 10 * time.Second,
				"data_channel_size":    200,
				"worker_pool_size":     8,
				"shutdown_timeout":     60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"AGENT_LOG_LEVEL": "invalid",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"AGENT_SERVER_PORT": "99999",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up environment variables after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Load configuration
			config, err := LoadConfigFromEnv()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)

				// Check individual fields
				assert.Equal(t, tt.expected["log_level"], config.LogLevel)
				assert.Equal(t, tt.expected["log_format"], config.LogFormat)
				assert.Equal(t, tt.expected["log_output"], config.LogOutput)
				assert.Equal(t, tt.expected["server_host"], config.ServerHost)
				assert.Equal(t, tt.expected["server_port"], config.ServerPort)
				assert.Equal(t, tt.expected["default_agent"], config.DefaultAgent)
				assert.Equal(t, tt.expected["ai_api_key"], config.AIAPIKey)
				assert.Equal(t, tt.expected["ai_api_url"], config.AIAPIURL)
				assert.Equal(t, tt.expected["prometheus_enabled"], config.PrometheusEnabled)
				assert.Equal(t, tt.expected["prometheus_url"], config.PrometheusURL)
				assert.Equal(t, tt.expected["plugin_config_file"], config.PluginConfigFile)
				assert.Equal(t, tt.expected["health_check_timeout"], config.HealthCheckTimeout)
				assert.Equal(t, tt.expected["data_channel_size"], config.DataChannelSize)
				assert.Equal(t, tt.expected["worker_pool_size"], config.WorkerPoolSize)
				assert.Equal(t, tt.expected["shutdown_timeout"], config.ShutdownTimeout)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "text", config.LogFormat)
	assert.Equal(t, "stdout", config.LogOutput)
	assert.Equal(t, "0.0.0.0", config.ServerHost)
	assert.Equal(t, 9090, config.ServerPort)
	assert.Equal(t, "", config.DefaultAgent)
	assert.Equal(t, "", config.AIAPIKey)
	assert.Equal(t, "https://api.openai.com/v1", config.AIAPIURL)
	assert.Equal(t, true, config.PrometheusEnabled)
	assert.Equal(t, "http://localhost:9090", config.PrometheusURL)
	assert.Equal(t, "plugins.yaml", config.PluginConfigFile)
	assert.Equal(t, 5*time.Second, config.HealthCheckTimeout)
	assert.Equal(t, 100, config.DataChannelSize)
	assert.Equal(t, 4, config.WorkerPoolSize)
	assert.Equal(t, 30*time.Second, config.ShutdownTimeout)
	assert.Len(t, config.Plugins, 4) // Default plugins (including ai-agent)
}

func TestGetConfigSummary(t *testing.T) {
	config := &core.FrameworkConfig{
		LogLevel:          "debug",
		LogFormat:         "json",
		LogOutput:         "stderr",
		ServerHost:        "localhost",
		ServerPort:        8080,
		DefaultAgent:      "test-agent",
		AIAPIKey:          "test-key",
		AIAPIURL:          "https://api.test.com/v1",
		PrometheusEnabled: true,
		PrometheusURL:     "http://localhost:9090",
		DataChannelSize:   200,
		WorkerPoolSize:    8,
		ShutdownTimeout:   60 * time.Second,
		Plugins: []core.PluginConfig{
			{Name: "test-collector", Type: "collector"},
			{Name: "test-analyzer", Type: "analyzer"},
		},
	}

	summary := GetConfigSummary(config)

	assert.IsType(t, map[string]interface{}{}, summary)
	assert.Contains(t, summary, "logging")
	assert.Contains(t, summary, "server")
	assert.Contains(t, summary, "agent")
	assert.Contains(t, summary, "prometheus")
	assert.Contains(t, summary, "processing")
	assert.Contains(t, summary, "plugins")

	logging := summary["logging"].(map[string]interface{})
	assert.Equal(t, "debug", logging["level"])
	assert.Equal(t, "json", logging["format"])
	assert.Equal(t, "stderr", logging["output"])

	server := summary["server"].(map[string]interface{})
	assert.Equal(t, "localhost", server["host"])
	assert.Equal(t, 8080, server["port"])

	agent := summary["agent"].(map[string]interface{})
	assert.Equal(t, "test-agent", agent["default_agent"])
	assert.Equal(t, "https://api.test.com/v1", agent["ai_api_url"])
	assert.Equal(t, true, agent["has_api_key"])

	prometheus := summary["prometheus"].(map[string]interface{})
	assert.Equal(t, true, prometheus["enabled"])
	assert.Equal(t, "http://localhost:9090", prometheus["url"])

	processing := summary["processing"].(map[string]interface{})
	assert.Equal(t, 200, processing["data_channel_size"])
	assert.Equal(t, 8, processing["worker_pool_size"])
	assert.Equal(t, "1m0s", processing["shutdown_timeout"])

	plugins := summary["plugins"].(map[string]interface{})
	assert.Equal(t, 2, plugins["count"])

	types := plugins["types"].(map[string]int)
	assert.Equal(t, 1, types["collector"])
	assert.Equal(t, 1, types["analyzer"])
}

func TestGetPluginConfigsFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected int
	}{
		{
			name:     "no plugin environment variables",
			envVars:  map[string]string{},
			expected: 4, // Default plugins (including ai-agent)
		},
		{
			name: "custom collectors",
			envVars: map[string]string{
				"AGENT_COLLECTOR_PLUGINS": "prometheus,custom-collector",
			},
			expected: 2, // Only the specified collectors
		},
		{
			name: "multiple plugin types",
			envVars: map[string]string{
				"AGENT_COLLECTOR_PLUGINS": "prometheus",
				"AGENT_ANALYZER_PLUGINS":  "anomaly,custom-analyzer",
				"AGENT_AGENT_PLUGINS":     "ai,rag",
			},
			expected: 5, // 1 collector + 2 analyzers + 2 agents
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up environment variables after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			plugins, err := getPluginConfigsFromEnv()

			require.NoError(t, err)
			assert.Len(t, plugins, tt.expected)
		})
	}
}

func TestGetDefaultPluginConfigs(t *testing.T) {
	plugins := getDefaultPluginConfigs()

	assert.Len(t, plugins, 4)

	// Check that we have the expected default plugins
	pluginNames := make(map[string]bool)
	pluginTypes := make(map[string]bool)

	for _, plugin := range plugins {
		pluginNames[plugin.Name] = true
		pluginTypes[plugin.Type] = true
		assert.True(t, plugin.Enabled)
	}

	assert.True(t, pluginNames["prometheus-collector"])
	assert.True(t, pluginNames["anomaly-analyzer"])
	assert.True(t, pluginNames["logger-responder"])
	assert.True(t, pluginNames["ai-agent"])

	assert.True(t, pluginTypes["collector"])
	assert.True(t, pluginTypes["analyzer"])
	assert.True(t, pluginTypes["responder"])
	assert.True(t, pluginTypes["agent"])
}
