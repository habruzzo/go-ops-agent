package analyzers

import (
	"context"
	"testing"
	"time"

	"github.com/holden/agent/core"
)

func TestAnomalyAnalyzer_NewAnomalyAnalyzer(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")

	if analyzer.Name() != "test-analyzer" {
		t.Errorf("Expected name 'test-analyzer', got %s", analyzer.Name())
	}

	if analyzer.Type() != core.PluginTypeAnalyzer {
		t.Errorf("Expected type 'analyzer', got %s", analyzer.Type())
	}

	if analyzer.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", analyzer.Version())
	}
}

func TestAnomalyAnalyzer_Configure(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")

	config := map[string]interface{}{
		"threshold": 3.0,
	}

	err := analyzer.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with invalid config
	invalidConfig := map[string]interface{}{
		"threshold": "invalid",
	}

	err = analyzer.Configure(invalidConfig)
	if err != nil {
		t.Fatalf("Expected no error for invalid config, got %v", err)
	}
}

func TestAnomalyAnalyzer_StartStop(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")

	ctx := context.Background()

	// Test start
	err := analyzer.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start analyzer: %v", err)
	}

	if analyzer.Status() != core.PluginStatusRunning {
		t.Errorf("Expected status 'running', got %s", analyzer.Status())
	}

	// Test stop
	err = analyzer.Stop()
	if err != nil {
		t.Fatalf("Failed to stop analyzer: %v", err)
	}

	if analyzer.Status() != core.PluginStatusStopped {
		t.Errorf("Expected status 'stopped', got %s", analyzer.Status())
	}
}

func TestAnomalyAnalyzer_Analyze(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")
	analyzer.Configure(map[string]interface{}{"threshold": 1.5})

	// Start the analyzer
	ctx := context.Background()
	err := analyzer.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start analyzer: %v", err)
	}
	defer analyzer.Stop()

	// Test with normal data (no anomalies)
	normalData := []core.DataPoint{
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 50.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 52.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 48.0, Labels: map[string]string{}},
	}

	analysis, err := analyzer.Analyze(normalData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if analysis != nil {
		t.Error("Expected no analysis for normal data")
	}

	// Test with anomalous data - more extreme values
	anomalousData := []core.DataPoint{
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 50.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 52.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 48.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 51.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 200.0, Labels: map[string]string{}}, // Extreme anomaly
	}

	analysis, err = analyzer.Analyze(anomalousData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if analysis == nil {
		t.Error("Expected analysis for anomalous data")
		return
	}

	if analysis.Type != core.AnalysisTypeAnomaly {
		t.Errorf("Expected analysis type 'anomaly', got %s", analysis.Type)
	}

	if analysis.Confidence <= 0 {
		t.Error("Expected confidence > 0")
	}
}

func TestAnomalyAnalyzer_CanAnalyze(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")

	// Test with insufficient data
	insufficientData := []core.DataPoint{
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 50.0, Labels: map[string]string{}},
	}

	if analyzer.CanAnalyze(insufficientData) {
		t.Error("Expected CanAnalyze to return false for insufficient data")
	}

	// Test with sufficient data
	sufficientData := []core.DataPoint{
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 50.0, Labels: map[string]string{}},
		{Timestamp: time.Now(), Source: "test", Metric: "cpu", Value: 52.0, Labels: map[string]string{}},
	}

	if !analyzer.CanAnalyze(sufficientData) {
		t.Error("Expected CanAnalyze to return true for sufficient data")
	}
}

func TestAnomalyAnalyzer_Health(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")

	ctx := context.Background()

	// Test health when not running
	err := analyzer.Health(ctx)
	if err == nil {
		t.Error("Expected health check to fail when not running")
	}

	// Test health when running
	analyzer.Start(ctx)
	err = analyzer.Health(ctx)
	if err != nil {
		t.Errorf("Expected health check to pass when running, got %v", err)
	}
}

func TestAnomalyAnalyzer_GetCapabilities(t *testing.T) {
	analyzer := NewAnomalyAnalyzer("test-analyzer")

	capabilities := analyzer.GetCapabilities()

	expected := []string{"detect_anomalies", "statistical_analysis", "threshold_detection"}

	if len(capabilities) != len(expected) {
		t.Errorf("Expected %d capabilities, got %d", len(expected), len(capabilities))
	}

	for _, expectedCap := range expected {
		found := false
		for _, cap := range capabilities {
			if cap == expectedCap {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected capability '%s' not found", expectedCap)
		}
	}
}
