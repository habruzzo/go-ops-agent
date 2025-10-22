package responders

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/habruzzo/agent/core"
)

// LoggerResponder implements the DataResponder interface for logging
type LoggerResponder struct {
	name    string
	version string
	status  core.PluginStatus
	level   slog.Level
	mu      sync.RWMutex
}

// NewLoggerResponder creates a new logger responder plugin
func NewLoggerResponder(name string) *LoggerResponder {
	return &LoggerResponder{
		name:    name,
		version: "1.0.0",
		status:  core.PluginStatusStopped,
		level:   slog.LevelInfo,
	}
}

// Name returns the name of the plugin
func (l *LoggerResponder) Name() string {
	return l.name
}

// Type returns the type of plugin
func (l *LoggerResponder) Type() core.PluginType {
	return core.PluginTypeResponder
}

// Version returns the plugin version
func (l *LoggerResponder) Version() string {
	return l.version
}

// Configure initializes the plugin with configuration
func (l *LoggerResponder) Configure(config map[string]interface{}) error {
	if levelStr, ok := config["level"].(string); ok {
		switch levelStr {
		case "debug":
			l.level = slog.LevelDebug
		case "info":
			l.level = slog.LevelInfo
		case "warn", "warning":
			l.level = slog.LevelWarn
		case "error":
			l.level = slog.LevelError
		default:
			l.level = slog.LevelInfo
		}
	}

	return nil
}

// Start begins the plugin's operation
func (l *LoggerResponder) Start(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.status == core.PluginStatusRunning {
		return fmt.Errorf("responder is already running")
	}

	l.status = core.PluginStatusStarting
	slog.Info("Starting logger responder", "plugin", l.name, "type", l.Type())

	l.status = core.PluginStatusRunning
	slog.Info("Logger responder started", "plugin", l.name, "type", l.Type())
	return nil
}

// Stop gracefully stops the plugin
func (l *LoggerResponder) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.status != core.PluginStatusRunning {
		return fmt.Errorf("responder is not running")
	}

	l.status = core.PluginStatusStopping
	slog.Info("Stopping logger responder", "plugin", l.name, "type", l.Type())

	l.status = core.PluginStatusStopped
	slog.Info("Logger responder stopped", "plugin", l.name, "type", l.Type())
	return nil
}

// Status returns the current status of the plugin
func (l *LoggerResponder) Status() core.PluginStatus {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.status
}

// Health checks if the plugin is healthy
func (l *LoggerResponder) Health(ctx context.Context) error {
	// Logger is always healthy if it's running
	if l.status == core.PluginStatusRunning {
		return nil
	}
	return fmt.Errorf("responder is not running")
}

// GetCapabilities returns what this plugin can do
func (l *LoggerResponder) GetCapabilities() []string {
	return []string{
		"log_analysis",
		"format_output",
		"severity_filtering",
	}
}

// Respond logs the analysis result
func (l *LoggerResponder) Respond(ctx context.Context, analysis *core.Analysis) error {
	logger := slog.Default().With(
		"plugin", l.name,
		"analyzer", analysis.Source,
		"type", analysis.Type,
		"confidence", analysis.Confidence,
		"severity", analysis.Severity,
		"data_points", len(analysis.DataPoints),
	)

	message := fmt.Sprintf("[%s] %s", analysis.Type, analysis.Summary)

	switch analysis.Severity {
	case "critical":
		logger.Error(message)
	case "high":
		logger.Warn(message)
	case "medium":
		logger.Info(message)
	default:
		logger.Debug(message)
	}

	return nil
}

// CanHandle determines if this responder can handle the given analysis
func (l *LoggerResponder) CanHandle(analysis *core.Analysis) bool {
	// Logger can handle all analysis types
	return true
}
