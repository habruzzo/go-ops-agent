package collectors

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// PrometheusCollector implements the DataCollector interface for Prometheus
type PrometheusCollector struct {
	name     string
	version  string
	status   core.PluginStatus
	client   v1.API
	queries  []string
	interval time.Duration
	mu       sync.RWMutex
}

// NewPrometheusCollector creates a new Prometheus collector plugin
func NewPrometheusCollector(name string) *PrometheusCollector {
	return &PrometheusCollector{
		name:     name,
		version:  "1.0.0",
		status:   core.PluginStatusStopped,
		interval: 30 * time.Second,
	}
}

// Name returns the name of the plugin
func (p *PrometheusCollector) Name() string {
	return p.name
}

// Type returns the type of plugin
func (p *PrometheusCollector) Type() core.PluginType {
	return core.PluginTypeCollector
}

// Version returns the plugin version
func (p *PrometheusCollector) Version() string {
	return p.version
}

// Configure initializes the plugin with configuration
func (p *PrometheusCollector) Configure(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok {
		return fmt.Errorf("prometheus URL not specified")
	}

	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	p.client = v1.NewAPI(client)

	// Get queries from config
	if queries, ok := config["queries"].([]interface{}); ok {
		p.queries = make([]string, len(queries))
		for i, q := range queries {
			if query, ok := q.(string); ok {
				p.queries[i] = query
			}
		}
	} else {
		// Default queries
		p.queries = []string{
			"up",
			"cpu_usage_percent",
			"memory_usage_percent",
		}
	}

	// Get interval from config
	if intervalStr, ok := config["interval"].(string); ok {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			p.interval = interval
		}
	}

	return nil
}

// Start begins the plugin's operation
func (p *PrometheusCollector) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == core.PluginStatusRunning {
		return fmt.Errorf("collector is already running")
	}

	p.status = core.PluginStatusStarting
	slog.Info("Starting Prometheus collector", "plugin", p.name, "type", p.Type())

	// Test connectivity
	if err := p.Health(ctx); err != nil {
		p.status = core.PluginStatusError
		return fmt.Errorf("health check failed: %w", err)
	}

	p.status = core.PluginStatusRunning
	slog.Info("Prometheus collector started", "plugin", p.name, "type", p.Type())
	return nil
}

// Stop gracefully stops the plugin
func (p *PrometheusCollector) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != core.PluginStatusRunning {
		return fmt.Errorf("collector is not running")
	}

	p.status = core.PluginStatusStopping
	slog.Info("Stopping Prometheus collector", "plugin", p.name, "type", p.Type())

	p.status = core.PluginStatusStopped
	slog.Info("Prometheus collector stopped", "plugin", p.name, "type", p.Type())
	return nil
}

// Status returns the current status of the plugin
func (p *PrometheusCollector) Status() core.PluginStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// Health checks if the plugin is healthy
func (p *PrometheusCollector) Health(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("prometheus client not configured")
	}

	// Try a simple query to check connectivity
	_, _, err := p.client.Query(ctx, "up", time.Now())
	return err
}

// GetCapabilities returns what this plugin can do
func (p *PrometheusCollector) GetCapabilities() []string {
	return []string{
		"collect_metrics",
		"query_prometheus",
		"health_check",
	}
}

// Collect gathers data points from Prometheus
func (p *PrometheusCollector) Collect(ctx context.Context) ([]core.DataPoint, error) {
	if p.client == nil {
		return nil, fmt.Errorf("prometheus client not configured")
	}

	var dataPoints []core.DataPoint

	for _, query := range p.queries {
		result, warnings, err := p.client.Query(ctx, query, time.Now())
		if err != nil {
			slog.Error("Failed to query Prometheus", "plugin", p.name, "query", query, "error", err)
			continue
		}

		if warnings != nil {
			slog.Warn("Prometheus query warnings", "plugin", p.name, "warnings", warnings)
		}

		points := p.convertResultToDataPoints(result, query)
		dataPoints = append(dataPoints, points...)
	}

	return dataPoints, nil
}

// GetCollectionInterval returns how often this collector should run
func (p *PrometheusCollector) GetCollectionInterval() time.Duration {
	return p.interval
}

// convertResultToDataPoints converts Prometheus query result to DataPoints
func (p *PrometheusCollector) convertResultToDataPoints(result interface{}, query string) []core.DataPoint {
	// This is a simplified conversion - in practice, you'd handle different result types
	// based on the Prometheus query result structure
	return []core.DataPoint{
		{
			Timestamp: time.Now(),
			Source:    p.name,
			Metric:    query,
			Value:     1.0, // Placeholder - would extract actual value from result
			Labels:    map[string]string{"query": query},
		},
	}
}
