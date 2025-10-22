package cli

import (
	"github.com/habruzzo/agent/config"
	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/collectors"
	"github.com/habruzzo/agent/plugins/responders"
)

// registerPluginCreators registers all available plugin creators with the factory
func registerPluginCreators(factory core.PluginFactory) {
	// Register Prometheus collector
	factory.RegisterPluginCreator("prometheus", func(config core.PluginConfig) (core.Plugin, error) {
		plugin := collectors.NewPrometheusCollector(config.Name)
		if configMap, ok := config.Config.(map[string]interface{}); ok {
			if err := plugin.Configure(configMap); err != nil {
				return nil, err
			}
		}
		return plugin, nil
	})

	// Register anomaly analyzer
	factory.RegisterPluginCreator("anomaly", func(config core.PluginConfig) (core.Plugin, error) {
		plugin := analyzers.NewAnomalyAnalyzer(config.Name)
		if configMap, ok := config.Config.(map[string]interface{}); ok {
			if err := plugin.Configure(configMap); err != nil {
				return nil, err
			}
		}
		return plugin, nil
	})

	// Register logger responder
	factory.RegisterPluginCreator("log", func(config core.PluginConfig) (core.Plugin, error) {
		plugin := responders.NewLoggerResponder(config.Name)
		if configMap, ok := config.Config.(map[string]interface{}); ok {
			if err := plugin.Configure(configMap); err != nil {
				return nil, err
			}
		}
		return plugin, nil
	})

	// Register AI agent
	factory.RegisterPluginCreator("ai", func(config core.PluginConfig) (core.Plugin, error) {
		plugin := agents.NewAIAgent(config.Name)
		if configMap, ok := config.Config.(map[string]interface{}); ok {
			if err := plugin.Configure(configMap); err != nil {
				return nil, err
			}
		}
		return plugin, nil
	})

	// Register RAG agent
	factory.RegisterPluginCreator("rag", func(config core.PluginConfig) (core.Plugin, error) {
		plugin := agents.NewRAGAgent(config.Name)
		if configMap, ok := config.Config.(map[string]interface{}); ok {
			if err := plugin.Configure(configMap); err != nil {
				return nil, err
			}
		}
		return plugin, nil
	})

	// Note: Orchestrator agent registration removed as it doesn't implement Plugin interface correctly
}

// loadPluginsFromConfig loads plugins from the framework configuration
func loadPluginsFromConfig(framework *core.Framework, config *core.FrameworkConfig) error {
	for _, pluginConfig := range config.Plugins {
		if !pluginConfig.Enabled {
			continue
		}

		if err := framework.LoadPluginFromConfig(pluginConfig); err != nil {
			return err
		}
	}
	return nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(filename string) error {
	cfg := config.DefaultConfig()
	return config.SaveConfig(cfg, filename)
}
