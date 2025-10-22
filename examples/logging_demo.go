package main

import (
	"context"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/analyzers"
	"github.com/habruzzo/agent/plugins/collectors"
)

func main() {
	// Create a simple config
	cfg := &core.FrameworkConfig{
		LogLevel:  "debug",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins: []core.PluginConfig{
			{
				Name:    "prometheus",
				Type:    "collector",
				Enabled: true,
				Config: map[string]interface{}{
					"url": "http://localhost:9090",
				},
			},
			{
				Name:    "anomaly-detector",
				Type:    "analyzer",
				Enabled: true,
				Config: map[string]interface{}{
					"threshold": 2.0,
				},
			},
		},
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Load plugins
	prometheus := collectors.NewPrometheusCollector("prometheus")
	if configMap, ok := cfg.Plugins[0].Config.(map[string]interface{}); ok {
		prometheus.Configure(configMap)
	}
	framework.LoadPlugin(prometheus)

	anomaly := analyzers.NewAnomalyAnalyzer("anomaly-detector")
	if configMap, ok := cfg.Plugins[1].Config.(map[string]interface{}); ok {
		anomaly.Configure(configMap)
	}
	framework.LoadPlugin(anomaly)

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	framework.Start(ctx)

	// Wait a bit to see logs
	time.Sleep(2 * time.Second)

	// Stop framework
	framework.Stop()
}
