package analyzers

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/habruzzo/agent/core"
)

// AnomalyAnalyzer implements the DataAnalyzer interface for anomaly detection
type AnomalyAnalyzer struct {
	name      string
	version   string
	status    core.PluginStatus
	threshold float64
	mu        sync.RWMutex
}

// NewAnomalyAnalyzer creates a new anomaly analyzer plugin
func NewAnomalyAnalyzer(name string) *AnomalyAnalyzer {
	return &AnomalyAnalyzer{
		name:      name,
		version:   "1.0.0",
		status:    core.PluginStatusStopped,
		threshold: 2.0,
	}
}

// Name returns the name of the plugin
func (a *AnomalyAnalyzer) Name() string {
	return a.name
}

// Type returns the type of plugin
func (a *AnomalyAnalyzer) Type() core.PluginType {
	return core.PluginTypeAnalyzer
}

// Version returns the plugin version
func (a *AnomalyAnalyzer) Version() string {
	return a.version
}

// Configure initializes the plugin with configuration
func (a *AnomalyAnalyzer) Configure(config map[string]interface{}) error {
	if threshold, ok := config["threshold"].(float64); ok {
		a.threshold = threshold
	}

	return nil
}

// Start begins the plugin's operation
func (a *AnomalyAnalyzer) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status == core.PluginStatusRunning {
		return fmt.Errorf("analyzer is already running")
	}

	a.status = core.PluginStatusStarting
	slog.Info("Starting anomaly analyzer", "plugin", a.name, "type", a.Type())

	a.status = core.PluginStatusRunning
	slog.Info("Anomaly analyzer started", "plugin", a.name, "type", a.Type())
	return nil
}

// Stop gracefully stops the plugin
func (a *AnomalyAnalyzer) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status != core.PluginStatusRunning {
		return fmt.Errorf("analyzer is not running")
	}

	a.status = core.PluginStatusStopping
	slog.Info("Stopping anomaly analyzer", "plugin", a.name, "type", a.Type())

	a.status = core.PluginStatusStopped
	slog.Info("Anomaly analyzer stopped", "plugin", a.name, "type", a.Type())
	return nil
}

// Status returns the current status of the plugin
func (a *AnomalyAnalyzer) Status() core.PluginStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

// Health checks if the plugin is healthy
func (a *AnomalyAnalyzer) Health(ctx context.Context) error {
	// Analyzer is always healthy if it's running
	if a.status == core.PluginStatusRunning {
		return nil
	}
	return fmt.Errorf("analyzer is not running")
}

// GetCapabilities returns what this plugin can do
func (a *AnomalyAnalyzer) GetCapabilities() []string {
	return []string{
		"detect_anomalies",
		"statistical_analysis",
		"threshold_detection",
	}
}

// Analyze detects anomalies in the data points
func (a *AnomalyAnalyzer) Analyze(data []core.DataPoint) (*core.Analysis, error) {
	if len(data) < 2 {
		return nil, nil // Need at least 2 points for comparison
	}

	// Simple statistical anomaly detection
	mean, stdDev := a.calculateStats(data)

	var anomalies []core.DataPoint
	for _, point := range data {
		if math.Abs(point.Value-mean) > a.threshold*stdDev {
			anomalies = append(anomalies, point)
		}
	}

	if len(anomalies) == 0 {
		return nil, nil // No anomalies detected
	}

	// Calculate confidence based on how far the anomaly is from the mean
	maxDeviation := 0.0
	for _, point := range anomalies {
		deviation := math.Abs(point.Value-mean) / stdDev
		if deviation > maxDeviation {
			maxDeviation = deviation
		}
	}

	confidence := math.Min(maxDeviation/a.threshold, 1.0)
	severity := a.determineSeverity(confidence)

	return &core.Analysis{
		Type:       core.AnalysisTypeAnomaly,
		Confidence: confidence,
		Severity:   severity,
		Summary:    fmt.Sprintf("Detected %d anomalies with max deviation of %.2fσ", len(anomalies), maxDeviation),
		Details: map[string]interface{}{
			"anomaly_count": len(anomalies),
			"mean":          mean,
			"std_dev":       stdDev,
			"threshold":     a.threshold,
		},
		DataPoints: anomalies,
		Timestamp:  time.Now(),
		Source:     a.name,
	}, nil
}

// CanAnalyze determines if this analyzer can process the given data
func (a *AnomalyAnalyzer) CanAnalyze(data []core.DataPoint) bool {
	return len(data) >= 2
}

// calculateStats calculates mean and standard deviation
func (a *AnomalyAnalyzer) calculateStats(data []core.DataPoint) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, point := range data {
		sum += point.Value
	}
	mean := sum / float64(len(data))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, point := range data {
		diff := point.Value - mean
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(data)))

	return mean, stdDev
}

// determineSeverity determines severity based on confidence
func (a *AnomalyAnalyzer) determineSeverity(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "critical"
	case confidence >= 0.7:
		return "high"
	case confidence >= 0.5:
		return "medium"
	default:
		return "low"
	}
}
