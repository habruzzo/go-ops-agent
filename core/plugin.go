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
	Name    string      `yaml:"name" env:"AGENT_PLUGIN_NAME" validate:"required"`
	Type    string      `yaml:"type" env:"AGENT_PLUGIN_TYPE" validate:"required,oneof=collector analyzer responder agent"`
	Config  interface{} `yaml:"config"`
	Enabled bool        `yaml:"enabled" env:"AGENT_PLUGIN_ENABLED" envDefault:"true"`
}

// PrometheusCollectorConfig represents configuration for Prometheus collector
type PrometheusCollectorConfig struct {
	URL            string        `yaml:"url" env:"AGENT_PROMETHEUS_URL" envDefault:"http://localhost:9090" validate:"required,url"`
	ScrapeInterval time.Duration `yaml:"scrape_interval" env:"AGENT_PROMETHEUS_SCRAPE_INTERVAL" envDefault:"30s" validate:"min=1s"`
	Timeout        time.Duration `yaml:"timeout" env:"AGENT_PROMETHEUS_TIMEOUT" envDefault:"10s" validate:"min=1s"`
}

// AnomalyAnalyzerConfig represents configuration for anomaly analyzer
type AnomalyAnalyzerConfig struct {
	Threshold  float64 `yaml:"threshold" env:"AGENT_ANOMALY_THRESHOLD" envDefault:"0.8" validate:"min=0,max=1"`
	WindowSize int     `yaml:"window_size" env:"AGENT_ANOMALY_WINDOW_SIZE" envDefault:"100" validate:"min=1"`
	Algorithm  string  `yaml:"algorithm" env:"AGENT_ANOMALY_ALGORITHM" envDefault:"statistical" validate:"oneof=statistical machine_learning"`
}

// LoggerResponderConfig represents configuration for logger responder
type LoggerResponderConfig struct {
	Level  string `yaml:"level" env:"AGENT_LOGGER_LEVEL" envDefault:"info" validate:"oneof=debug info warn error"`
	Format string `yaml:"format" env:"AGENT_LOGGER_FORMAT" envDefault:"json" validate:"oneof=text json"`
	Output string `yaml:"output" env:"AGENT_LOGGER_OUTPUT" envDefault:"stdout"`
}

// AIAgentConfig represents configuration for AI agent
type AIAgentConfig struct {
	Model       string  `yaml:"model" env:"AGENT_AI_MODEL" envDefault:"gpt-4" validate:"required"`
	MaxTokens   int     `yaml:"max_tokens" env:"AGENT_AI_MAX_TOKENS" envDefault:"1000" validate:"min=1,max=4000"`
	Temperature float64 `yaml:"temperature" env:"AGENT_AI_TEMPERATURE" envDefault:"0.7" validate:"min=0,max=2"`
	APIKey      string  `yaml:"api_key" env:"AGENT_AI_API_KEY"`
	APIURL      string  `yaml:"api_url" env:"AGENT_AI_API_URL" envDefault:"https://api.openai.com/v1" validate:"required,url"`
}

// RAGAgentConfig represents configuration for RAG agent
type RAGAgentConfig struct {
	Model               string  `yaml:"model" env:"AGENT_RAG_MODEL" envDefault:"gpt-4" validate:"required"`
	MaxTokens           int     `yaml:"max_tokens" env:"AGENT_RAG_MAX_TOKENS" envDefault:"1000" validate:"min=1,max=4000"`
	Temperature         float64 `yaml:"temperature" env:"AGENT_RAG_TEMPERATURE" envDefault:"0.7" validate:"min=0,max=2"`
	APIKey              string  `yaml:"api_key" env:"AGENT_AI_API_KEY"`
	APIURL              string  `yaml:"api_url" env:"AGENT_AI_API_URL" envDefault:"https://api.openai.com/v1" validate:"required,url"`
	KnowledgeBasePath   string  `yaml:"knowledge_base_path" env:"AGENT_RAG_KNOWLEDGE_PATH" envDefault:"./knowledge" validate:"required"`
	MaxContextLength    int     `yaml:"max_context_length" env:"AGENT_RAG_MAX_CONTEXT" envDefault:"4000" validate:"min=1,max=8000"`
	EmbeddingModel      string  `yaml:"embedding_model" env:"AGENT_RAG_EMBEDDING_MODEL" envDefault:"text-embedding-ada-002" validate:"required"`
	SimilarityThreshold float64 `yaml:"similarity_threshold" env:"AGENT_RAG_SIMILARITY_THRESHOLD" envDefault:"0.7" validate:"min=0,max=1"`
	MaxDocuments        int     `yaml:"max_documents" env:"AGENT_RAG_MAX_DOCUMENTS" envDefault:"5" validate:"min=1,max=20"`
}

// FrameworkConfig represents the configuration for the entire framework
type FrameworkConfig struct {
	// Logging configuration
	LogLevel  string `yaml:"log_level" env:"AGENT_LOG_LEVEL" validate:"oneof=debug info warn error"`
	LogFormat string `yaml:"log_format" env:"AGENT_LOG_FORMAT" validate:"oneof=text json"`
	LogOutput string `yaml:"log_output" env:"AGENT_LOG_OUTPUT"`

	// Server configuration
	ServerHost string `yaml:"server_host" env:"AGENT_SERVER_HOST" envDefault:"0.0.0.0" validate:"required"`
	ServerPort int    `yaml:"server_port" env:"AGENT_SERVER_PORT" envDefault:"9090" validate:"min=1,max=65535"`

	// Agent configuration
	DefaultAgent string `yaml:"default_agent" env:"AGENT_DEFAULT_AGENT" envDefault:""`
	AIAPIKey     string `yaml:"ai_api_key" env:"AGENT_AI_API_KEY" envDefault:""`
	AIAPIURL     string `yaml:"ai_api_url" env:"AGENT_AI_API_URL" envDefault:"https://api.openai.com/v1"`

	// Prometheus configuration
	PrometheusEnabled bool   `yaml:"prometheus_enabled" env:"AGENT_PROMETHEUS_ENABLED" envDefault:"true"`
	PrometheusURL     string `yaml:"prometheus_url" env:"AGENT_PROMETHEUS_URL" envDefault:"http://localhost:9090"`

	// Plugin configuration
	PluginConfigFile string `yaml:"plugin_config_file" env:"AGENT_PLUGIN_CONFIG" envDefault:"plugins.yaml"`

	// Health check configuration
	HealthCheckTimeout time.Duration `yaml:"health_check_timeout" env:"AGENT_HEALTH_TIMEOUT" envDefault:"5s" validate:"min=1s"`

	// Data processing configuration
	DataChannelSize int           `yaml:"data_channel_size" env:"AGENT_DATA_CHANNEL_SIZE" envDefault:"100" validate:"min=1"`
	WorkerPoolSize  int           `yaml:"worker_pool_size" env:"AGENT_WORKER_POOL_SIZE" envDefault:"4" validate:"min=1"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"AGENT_SHUTDOWN_TIMEOUT" envDefault:"30s" validate:"min=1s"`

	// Plugin configurations (loaded from file or environment)
	Plugins []PluginConfig `yaml:"plugins,omitempty"`
}
