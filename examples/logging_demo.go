package main

import (
	"context"
	"time"

	"github.com/holden/agent/core"
	"github.com/holden/agent/plugins/analyzers"
	"github.com/holden/agent/plugins/collectors"
)

func main() {
	// Create a simple config
	cfg := &core.FrameworkConfig{
		Logging: core.LoggingConfig{
			Level:  "debug",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []core.PluginConfig{
			{
				Name:    "prometheus",
				Type:    "prometheus",
				Enabled: true,
				Config: map[string]interface{}{
					"url": "http://localhost:9090",
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
		},
	}

	// Create framework
	framework := core.NewFramework(cfg)

	// Load plugins
	prometheus := collectors.NewPrometheusCollector("prometheus")
	prometheus.Configure(cfg.Plugins[0].Config)
	framework.LoadPlugin(prometheus)

	anomaly := analyzers.NewAnomalyAnalyzer("anomaly-detector")
	anomaly.Configure(cfg.Plugins[1].Config)
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
