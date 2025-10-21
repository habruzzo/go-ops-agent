package responders

import (
	"context"
	"testing"
	"time"

	"github.com/holden/agent/core"
)

func TestLoggerResponder_NewLoggerResponder(t *testing.T) {
	responder := NewLoggerResponder("test-logger")

	if responder.Name() != "test-logger" {
		t.Errorf("Expected name 'test-logger', got %s", responder.Name())
	}

	if responder.Type() != core.PluginTypeResponder {
		t.Errorf("Expected type 'responder', got %s", responder.Type())
	}

	if responder.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", responder.Version())
	}
}

func TestLoggerResponder_Configure(t *testing.T) {
	responder := NewLoggerResponder("test-logger")

	config := map[string]interface{}{
		"level": "debug",
	}

	err := responder.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with invalid level
	invalidConfig := map[string]interface{}{
		"level": "invalid",
	}

	err = responder.Configure(invalidConfig)
	if err != nil {
		t.Fatalf("Expected no error for invalid level, got %v", err)
	}
}

func TestLoggerResponder_StartStop(t *testing.T) {
	responder := NewLoggerResponder("test-logger")

	ctx := context.Background()

	// Test start
	err := responder.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start responder: %v", err)
	}

	if responder.Status() != core.PluginStatusRunning {
		t.Errorf("Expected status 'running', got %s", responder.Status())
	}

	// Test stop
	err = responder.Stop()
	if err != nil {
		t.Fatalf("Failed to stop responder: %v", err)
	}

	if responder.Status() != core.PluginStatusStopped {
		t.Errorf("Expected status 'stopped', got %s", responder.Status())
	}
}

func TestLoggerResponder_Respond(t *testing.T) {
	responder := NewLoggerResponder("test-logger")
	responder.Start(context.Background())

	// Test with different severity levels
	testCases := []struct {
		severity string
		expected string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
	}

	for _, tc := range testCases {
		analysis := &core.Analysis{
			Type:       core.AnalysisTypeAnomaly,
			Confidence: 0.8,
			Severity:   tc.severity,
			Summary:    "Test analysis",
			Details:    map[string]interface{}{"test": true},
			DataPoints: []core.DataPoint{},
			Timestamp:  time.Now(),
			Source:     "test-analyzer",
		}

		err := responder.Respond(context.Background(), analysis)
		if err != nil {
			t.Errorf("Expected no error for severity %s, got %v", tc.severity, err)
		}
	}
}

func TestLoggerResponder_CanHandle(t *testing.T) {
	responder := NewLoggerResponder("test-logger")

	// Logger should handle all analysis types
	analysis := &core.Analysis{
		Type: core.AnalysisTypeAnomaly,
	}

	if !responder.CanHandle(analysis) {
		t.Error("Expected logger to handle all analysis types")
	}

	// Test with different analysis types
	analysisTypes := []core.AnalysisType{
		core.AnalysisTypeAnomaly,
		core.AnalysisTypeTrend,
		core.AnalysisTypeCorrelation,
		core.AnalysisTypeAlert,
	}

	for _, analysisType := range analysisTypes {
		analysis.Type = analysisType
		if !responder.CanHandle(analysis) {
			t.Errorf("Expected logger to handle analysis type %s", analysisType)
		}
	}
}

func TestLoggerResponder_Health(t *testing.T) {
	responder := NewLoggerResponder("test-logger")

	ctx := context.Background()

	// Test health when not running
	err := responder.Health(ctx)
	if err == nil {
		t.Error("Expected health check to fail when not running")
	}

	// Test health when running
	responder.Start(ctx)
	err = responder.Health(ctx)
	if err != nil {
		t.Errorf("Expected health check to pass when running, got %v", err)
	}
}

func TestLoggerResponder_GetCapabilities(t *testing.T) {
	responder := NewLoggerResponder("test-logger")

	capabilities := responder.GetCapabilities()

	expected := []string{"log_analysis", "format_output", "severity_filtering"}

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
