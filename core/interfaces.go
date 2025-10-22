package core

import (
	"context"
	"time"
)

// PluginRegistry manages plugin registration and discovery
type PluginRegistry interface {
	RegisterPlugin(plugin Plugin) error
	UnregisterPlugin(name string) error
	GetPlugin(name string) (Plugin, error)
	ListPlugins() []Plugin
	ListPluginsByType(pluginType PluginType) []Plugin
	GetPluginCount() int
	GetPluginCountByType(pluginType PluginType) int
}

// ConfigurationManager handles configuration loading, validation, and updates
type ConfigurationManager interface {
	LoadConfig(filename string) (*FrameworkConfig, error)
	ValidateConfig(config *FrameworkConfig) error
	WatchConfig(filename string) (<-chan *FrameworkConfig, error)
	SaveConfig(config *FrameworkConfig, filename string) error
}

// MetricsCollector collects and exposes framework metrics
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
	ObserveHistogram(name string, value float64, labels map[string]string)
	GetMetrics() map[string]interface{}
}

// HealthChecker provides health check functionality
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthStatus
	RegisterHealthCheck(name string, check HealthCheckFunc)
	UnregisterHealthCheck(name string)
}

// HealthCheckFunc represents a health check function
type HealthCheckFunc func(ctx context.Context) error

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string                 `json:"status"` // healthy, unhealthy, degraded
	Message   string                 `json:"message"`
	Checks    map[string]CheckResult `json:"checks"`
	Timestamp time.Time              `json:"timestamp"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// EventBus handles event publishing and subscription
type EventBus interface {
	Publish(event Event) error
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
}

// Event represents a framework event
type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler handles events
type EventHandler func(event Event) error

// ContextManager manages request context and tracing
type ContextManager interface {
	WithContext(ctx context.Context) context.Context
	WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc)
	WithCancel(ctx context.Context) (context.Context, context.CancelFunc)
	GetTraceID(ctx context.Context) string
	GetSpanID(ctx context.Context) string
}

// DataProcessor processes data through the pipeline
type DataProcessor interface {
	ProcessData(ctx context.Context, data []DataPoint) error
	AddProcessor(processor DataProcessorFunc) error
	RemoveProcessor(processor DataProcessorFunc) error
}

// DataProcessorFunc represents a data processing function
type DataProcessorFunc func(ctx context.Context, data []DataPoint) ([]DataPoint, error)

// Workflow represents a workflow definition
type Workflow struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Steps     []WorkflowStep         `json:"steps"`
	Triggers  []Trigger              `json:"triggers"`
	State     WorkflowState          `json:"state"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Config     map[string]interface{} `json:"config"`
	NextSteps  []string               `json:"next_steps"`
	Condition  string                 `json:"condition,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
}

// Trigger represents a workflow trigger
type Trigger struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// WorkflowState represents the state of a workflow
type WorkflowState string

const (
	WorkflowStatePending   WorkflowState = "pending"
	WorkflowStateRunning   WorkflowState = "running"
	WorkflowStateCompleted WorkflowState = "completed"
	WorkflowStateFailed    WorkflowState = "failed"
	WorkflowStateCancelled WorkflowState = "cancelled"
)

// WorkflowEngine manages workflow execution
type WorkflowEngine interface {
	CreateWorkflow(workflow *Workflow) error
	ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowResult, error)
	GetWorkflowStatus(workflowID string) (*WorkflowStatus, error)
	CancelWorkflow(workflowID string) error
}

// WorkflowResult represents the result of workflow execution
type WorkflowResult struct {
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
}

// WorkflowStatus represents the current status of a workflow
type WorkflowStatus struct {
	WorkflowID  string    `json:"workflow_id"`
	Status      string    `json:"status"`
	Progress    float64   `json:"progress"`
	CurrentStep string    `json:"current_step"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceDiscovery provides service discovery functionality
type ServiceDiscovery interface {
	RegisterService(service ServiceInfo) error
	UnregisterService(serviceID string) error
	DiscoverServices(serviceType string) ([]ServiceInfo, error)
	GetService(serviceID string) (*ServiceInfo, error)
}

// ServiceInfo represents service information
type ServiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	HealthCheck string            `json:"health_check"`
	Metadata    map[string]string `json:"metadata"`
	Tags        []string          `json:"tags"`
}

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker interface {
	Execute(ctx context.Context, operation func() error) error
	GetState() string
	Reset()
}

// RateLimiter provides rate limiting functionality
type RateLimiter interface {
	Allow() bool
	AllowN(n int) bool
	Wait(ctx context.Context) error
	WaitN(ctx context.Context, n int) error
}

// Cache provides caching functionality
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Size() int
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// RetryExecutor executes operations with retry logic
type RetryExecutor interface {
	Execute(ctx context.Context, operation func() error) error
	ExecuteWithPolicy(ctx context.Context, policy RetryPolicy, operation func() error) error
}
