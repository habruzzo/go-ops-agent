package agents

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/habruzzo/agent/core"
)

// AgentOrchestrator coordinates multiple agents to work together
type AgentOrchestrator struct {
	name         string
	status       core.PluginStatus
	agents       map[string]Agent
	workflows    map[string]*Workflow
	messageBus   *MessageBus
	stateManager *StateManager
	monitor      *AgentMonitor
	mu           sync.RWMutex
}

// Agent represents any agent that can be orchestrated
type Agent interface {
	GetName() string
	GetStatus() core.PluginStatus
	ProcessMessage(ctx context.Context, msg *Message) (*Message, error)
	GetCapabilities() []string
	Start(ctx context.Context) error
	Stop() error
}

// Message represents communication between agents
type Message struct {
	ID         string                 `json:"id"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Type       string                 `json:"type"`
	Content    string                 `json:"content"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	Priority   int                    `json:"priority"`
	ResponseTo string                 `json:"response_to,omitempty"`
}

// Workflow represents a sequence of agent actions
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
	Agent      string                 `json:"agent"`
	Action     string                 `json:"action"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
	Condition  string                 `json:"condition,omitempty"`
	Timeout    time.Duration          `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	Status     StepStatus             `json:"status"`
}

// Trigger represents what starts a workflow
type Trigger struct {
	Type      string                 `json:"type"`
	Condition string                 `json:"condition"`
	Data      map[string]interface{} `json:"data"`
	Enabled   bool                   `json:"enabled"`
}

// WorkflowState represents the current state of a workflow
type WorkflowState int

const (
	WorkflowStatePending WorkflowState = iota
	WorkflowStateRunning
	WorkflowStatePaused
	WorkflowStateCompleted
	WorkflowStateFailed
	WorkflowStateCancelled
)

// StepStatus represents the status of a workflow step
type StepStatus int

const (
	StepStatusPending StepStatus = iota
	StepStatusRunning
	StepStatusCompleted
	StepStatusFailed
	StepStatusSkipped
)

// MessageBus handles communication between agents
type MessageBus struct {
	channels    map[string]chan *Message
	subscribers map[string][]string
	mu          sync.RWMutex
}

// StateManager manages workflow and agent state
type StateManager struct {
	workflows   map[string]*Workflow
	agentStates map[string]map[string]interface{}
	mu          sync.RWMutex
}

// AgentMonitor monitors agent performance and health
type AgentMonitor struct {
	metrics map[string]*AgentMetrics
	alerts  []Alert
	mu      sync.RWMutex
}

// AgentMetrics tracks agent performance
type AgentMetrics struct {
	AgentName         string        `json:"agent_name"`
	MessagesProcessed int64         `json:"messages_processed"`
	AverageLatency    time.Duration `json:"average_latency"`
	SuccessRate       float64       `json:"success_rate"`
	ErrorCount        int64         `json:"error_count"`
	LastActivity      time.Time     `json:"last_activity"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID           string                 `json:"id"`
	AgentName    string                 `json:"agent_name"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
	Acknowledged bool                   `json:"acknowledged"`
}

// NewAgentOrchestrator creates a new agent orchestrator
func NewAgentOrchestrator(name string) *AgentOrchestrator {
	return &AgentOrchestrator{
		name:         name,
		status:       core.PluginStatusStopped,
		agents:       make(map[string]Agent),
		workflows:    make(map[string]*Workflow),
		messageBus:   NewMessageBus(),
		stateManager: NewStateManager(),
		monitor:      NewAgentMonitor(),
	}
}

// RegisterAgent registers an agent with the orchestrator
func (o *AgentOrchestrator) RegisterAgent(agent Agent) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.agents[agent.GetName()] = agent
	slog.Info("Agent registered with orchestrator",
		"orchestrator", o.name,
		"agent", agent.GetName(),
		"capabilities", agent.GetCapabilities())

	return nil
}

// StartWorkflow starts a workflow execution
func (o *AgentOrchestrator) StartWorkflow(ctx context.Context, workflowID string) error {
	o.mu.RLock()
	workflow, exists := o.workflows[workflowID]
	o.mu.RUnlock()

	if !exists {
		return core.NewPluginError("orchestrator", "execute-workflow", fmt.Sprintf("workflow %s not found", workflowID))
	}

	workflow.State = WorkflowStateRunning
	workflow.UpdatedAt = time.Now()

	slog.Info("Starting workflow",
		"orchestrator", o.name,
		"workflow", workflowID,
		"steps", len(workflow.Steps))

	// Execute workflow steps
	go o.executeWorkflow(ctx, workflow)

	return nil
}

// executeWorkflow executes a workflow step by step
func (o *AgentOrchestrator) executeWorkflow(ctx context.Context, workflow *Workflow) {
	for i, step := range workflow.Steps {
		select {
		case <-ctx.Done():
			workflow.State = WorkflowStateCancelled
			return
		default:
			// Execute step
			err := o.executeStep(ctx, workflow, &workflow.Steps[i])
			if err != nil {
				slog.Error("Workflow step failed",
					"orchestrator", o.name,
					"workflow", workflow.ID,
					"step", step.ID,
					"error", err)

				workflow.State = WorkflowStateFailed
				return
			}
		}
	}

	workflow.State = WorkflowStateCompleted
	workflow.UpdatedAt = time.Now()

	slog.Info("Workflow completed",
		"orchestrator", o.name,
		"workflow", workflow.ID)
}

// executeStep executes a single workflow step
func (o *AgentOrchestrator) executeStep(ctx context.Context, workflow *Workflow, step *WorkflowStep) error {
	step.Status = StepStatusRunning

	// Get the agent for this step
	o.mu.RLock()
	agent, exists := o.agents[step.Agent]
	o.mu.RUnlock()

	if !exists {
		return core.NewPluginError("orchestrator", "execute-step", fmt.Sprintf("agent %s not found", step.Agent))
	}

	// Create message for the agent
	msg := &Message{
		ID:        fmt.Sprintf("%s-%s-%d", workflow.ID, step.ID, time.Now().Unix()),
		From:      o.name,
		To:        step.Agent,
		Type:      step.Action,
		Content:   step.Name,
		Data:      step.Input,
		Timestamp: time.Now(),
		Priority:  1,
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(ctx, step.Timeout)
	defer cancel()

	// Send message to agent
	response, err := agent.ProcessMessage(ctx, msg)
	if err != nil {
		step.Status = StepStatusFailed
		return err
	}

	// Store output
	step.Output = response.Data
	step.Status = StepStatusCompleted

	// Update metrics
	o.monitor.RecordMessage(step.Agent, time.Since(msg.Timestamp), err == nil)

	return nil
}

// SendMessage sends a message between agents
func (o *AgentOrchestrator) SendMessage(ctx context.Context, msg *Message) error {
	o.mu.RLock()
	_, exists := o.agents[msg.To]
	o.mu.RUnlock()

	if !exists {
		return core.NewPluginError("orchestrator", "send-message", fmt.Sprintf("agent %s not found", msg.To))
	}

	// Send message through message bus
	return o.messageBus.Send(msg)
}

// GetStatus returns the orchestrator status
func (o *AgentOrchestrator) GetStatus() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agentStatuses := make(map[string]interface{})
	for name, agent := range o.agents {
		agentStatuses[name] = map[string]interface{}{
			"status":       agent.GetStatus(),
			"capabilities": agent.GetCapabilities(),
		}
	}

	workflowStatuses := make(map[string]interface{})
	for id, workflow := range o.workflows {
		workflowStatuses[id] = map[string]interface{}{
			"name":  workflow.Name,
			"state": workflow.State,
			"steps": len(workflow.Steps),
		}
	}

	return map[string]interface{}{
		"orchestrator": o.name,
		"status":       o.status,
		"agents":       agentStatuses,
		"workflows":    workflowStatuses,
		"metrics":      o.monitor.GetMetrics(),
	}
}

// NewMessageBus creates a new message bus
func NewMessageBus() *MessageBus {
	return &MessageBus{
		channels:    make(map[string]chan *Message),
		subscribers: make(map[string][]string),
	}
}

// Send sends a message through the message bus
func (mb *MessageBus) Send(msg *Message) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	channel, exists := mb.channels[msg.To]
	if !exists {
		channel = make(chan *Message, 100)
		mb.channels[msg.To] = channel
	}

	select {
	case channel <- msg:
		return nil
	default:
		return core.NewPluginError("orchestrator", "send-message", fmt.Sprintf("message queue full for agent %s", msg.To))
	}
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		workflows:   make(map[string]*Workflow),
		agentStates: make(map[string]map[string]interface{}),
	}
}

// NewAgentMonitor creates a new agent monitor
func NewAgentMonitor() *AgentMonitor {
	return &AgentMonitor{
		metrics: make(map[string]*AgentMetrics),
		alerts:  make([]Alert, 0),
	}
}

// RecordMessage records a message processing event
func (am *AgentMonitor) RecordMessage(agentName string, latency time.Duration, success bool) {
	am.mu.Lock()
	defer am.mu.Unlock()

	metrics, exists := am.metrics[agentName]
	if !exists {
		metrics = &AgentMetrics{
			AgentName: agentName,
		}
		am.metrics[agentName] = metrics
	}

	metrics.MessagesProcessed++
	metrics.LastActivity = time.Now()

	if success {
		metrics.SuccessRate = (metrics.SuccessRate*float64(metrics.MessagesProcessed-1) + 1.0) / float64(metrics.MessagesProcessed)
	} else {
		metrics.ErrorCount++
		metrics.SuccessRate = (metrics.SuccessRate*float64(metrics.MessagesProcessed-1) + 0.0) / float64(metrics.MessagesProcessed)
	}

	// Update average latency
	metrics.AverageLatency = (metrics.AverageLatency*time.Duration(metrics.MessagesProcessed-1) + latency) / time.Duration(metrics.MessagesProcessed)
}

// GetMetrics returns all agent metrics
func (am *AgentMonitor) GetMetrics() map[string]*AgentMetrics {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make(map[string]*AgentMetrics)
	for name, metrics := range am.metrics {
		result[name] = metrics
	}
	return result
}

// AddWorkflow adds a workflow to the orchestrator
func (o *AgentOrchestrator) AddWorkflow(workflow *Workflow) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.workflows[workflow.ID] = workflow
	slog.Info("Workflow added to orchestrator",
		"orchestrator", o.name,
		"workflow", workflow.ID,
		"name", workflow.Name,
		"steps", len(workflow.Steps))
}
