package core

import (
	"context"
	"time"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeCollector PluginType = "collector"
	PluginTypeAnalyzer  PluginType = "analyzer"
	PluginTypeResponder PluginType = "responder"
	PluginTypeAgent     PluginType = "agent"
)

// PluginStatus represents the current status of a plugin
type PluginStatus string

const (
	PluginStatusStopped  PluginStatus = "stopped"
	PluginStatusStarting PluginStatus = "starting"
	PluginStatusRunning  PluginStatus = "running"
	PluginStatusStopping PluginStatus = "stopping"
	PluginStatusError    PluginStatus = "error"
)

// Plugin defines the base interface that all plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string

	// Type returns the type of plugin
	Type() PluginType

	// Version returns the plugin version
	Version() string

	// Configure initializes the plugin with the given configuration
	Configure(config map[string]interface{}) error

	// Start begins the plugin's operation
	Start(ctx context.Context) error

	// Stop gracefully stops the plugin
	Stop() error

	// Status returns the current status of the plugin
	Status() PluginStatus

	// Health checks if the plugin is healthy
	Health(ctx context.Context) error

	// GetCapabilities returns what this plugin can do
	GetCapabilities() []string
}

// DataCollector defines the interface for plugins that collect data
type DataCollector interface {
	Plugin

	// Collect gathers data points from the source
	Collect(ctx context.Context) ([]DataPoint, error)

	// GetCollectionInterval returns how often this collector should run
	GetCollectionInterval() time.Duration
}

// DataAnalyzer defines the interface for plugins that analyze data
type DataAnalyzer interface {
	Plugin

	// Analyze processes data points and returns analysis results
	Analyze(data []DataPoint) (*Analysis, error)

	// CanAnalyze determines if this analyzer can process the given data
	CanAnalyze(data []DataPoint) bool
}

// DataResponder defines the interface for plugins that respond to analysis results
type DataResponder interface {
	Plugin

	// Respond takes action based on the analysis result
	Respond(ctx context.Context, analysis *Analysis) error

	// CanHandle determines if this responder can handle the given analysis
	CanHandle(analysis *Analysis) bool
}

// AgentPlugin defines the interface for AI agent plugins
type AgentPlugin interface {
	Plugin

	// ProcessQuery handles user queries and returns responses
	ProcessQuery(ctx context.Context, query string) (*AgentResponse, error)

	// SetContext provides the agent with current system data
	SetContext(data []DataPoint)

	// GetAvailableQueries returns what types of queries this agent can handle
	GetAvailableQueries() []string
}

// AgentResponse represents a response from an agent plugin
type AgentResponse struct {
	Query      string                 `json:"query"`
	Response   string                 `json:"response"`
	Confidence float64                `json:"confidence"`
	Actions    []AgentAction          `json:"actions,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// AgentAction represents an action the agent can take
type AgentAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// PluginConfig represents configuration for a plugin
type PluginConfig struct {
	Name    string                 `yaml:"name"`
	Type    string                 `yaml:"type"`
	Config  map[string]interface{} `yaml:"config"`
	Enabled bool                   `yaml:"enabled"`
}

// FrameworkConfig represents the configuration for the entire framework
type FrameworkConfig struct {
	Plugins []PluginConfig `yaml:"plugins"`
	Logging LoggingConfig  `yaml:"logging"`
	Agent   AgentConfig    `yaml:"agent,omitempty"`
}

// LoggingConfig represents configuration for logging
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // text, json
	Output string `yaml:"output"` // stdout, stderr, file path
}

// AgentConfig represents configuration for agent interactions
type AgentConfig struct {
	DefaultAgent string                 `yaml:"default_agent"`
	Queries      []string               `yaml:"queries,omitempty"`
	Config       map[string]interface{} `yaml:"config,omitempty"`
}
