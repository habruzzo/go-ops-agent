package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/habruzzo/agent/core"
	"github.com/habruzzo/agent/plugins/agents"
)

func orchestratorMain() {
	fmt.Println("Agent Orchestration Demo")
	fmt.Println("=======================")

	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: OPENAI_API_KEY environment variable is required")
		fmt.Println("Get your API key from: https://platform.openai.com/api-keys")
		os.Exit(1)
	}

	// Create orchestrator
	orchestrator := agents.NewAgentOrchestrator("incident-response-orchestrator")

	// Create specialized agents
	detector := NewAnomalyDetectorAgent("anomaly-detector")
	analyzer := NewRootCauseAgent("root-cause-analyzer", apiKey)
	responder := NewRemediationAgent("remediation-agent", apiKey)
	notifier := NewNotificationAgent("notification-agent")

	// Register agents with orchestrator
	orchestrator.RegisterAgent(detector)
	orchestrator.RegisterAgent(analyzer)
	orchestrator.RegisterAgent(responder)
	orchestrator.RegisterAgent(notifier)

	// Start all agents
	ctx := context.Background()

	fmt.Println("Starting agents...")
	detector.Start(ctx)
	analyzer.Start(ctx)
	responder.Start(ctx)
	notifier.Start(ctx)

	// Create incident response workflow
	workflow := &agents.Workflow{
		ID:   "incident-response",
		Name: "Incident Response Workflow",
		Steps: []agents.WorkflowStep{
			{
				ID:     "detect-anomaly",
				Name:   "Detect System Anomaly",
				Agent:  "anomaly-detector",
				Action: "detect",
				Input: map[string]interface{}{
					"threshold": 2.0,
					"metrics":   []string{"cpu", "memory", "response_time"},
				},
				Timeout: 30 * time.Second,
			},
			{
				ID:     "analyze-root-cause",
				Name:   "Analyze Root Cause",
				Agent:  "root-cause-analyzer",
				Action: "analyze",
				Input: map[string]interface{}{
					"anomaly_data": "{{detect-anomaly.output}}",
				},
				Timeout: 60 * time.Second,
			},
			{
				ID:     "remediate-issue",
				Name:   "Remediate Issue",
				Agent:  "remediation-agent",
				Action: "remediate",
				Input: map[string]interface{}{
					"root_cause": "{{analyze-root-cause.output}}",
				},
				Timeout: 120 * time.Second,
			},
			{
				ID:     "notify-team",
				Name:   "Notify Team",
				Agent:  "notification-agent",
				Action: "notify",
				Input: map[string]interface{}{
					"incident_summary": "{{remediate-issue.output}}",
				},
				Timeout: 10 * time.Second,
			},
		},
		State:     agents.WorkflowStatePending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add workflow to orchestrator
	orchestrator.AddWorkflow(workflow)

	fmt.Println("Workflow created:")
	fmt.Printf("  ID: %s\n", workflow.ID)
	fmt.Printf("  Name: %s\n", workflow.Name)
	fmt.Printf("  Steps: %d\n", len(workflow.Steps))
	fmt.Println("")

	// Simulate incident
	fmt.Println("Simulating system incident...")

	// Start workflow
	err := orchestrator.StartWorkflow(ctx, "incident-response")
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	// Monitor workflow execution
	fmt.Println("Monitoring workflow execution...")

	for i := 0; i < 10; i++ {
		status := orchestrator.GetStatus()
		fmt.Printf("Status check %d:\n", i+1)

		if workflows, ok := status["workflows"].(map[string]interface{}); ok {
			if incidentResp, ok := workflows["incident-response"].(map[string]interface{}); ok {
				fmt.Printf("  Workflow State: %v\n", incidentResp["state"])
			}
		}

		if metrics, ok := status["metrics"].(map[string]*agents.AgentMetrics); ok {
			for agentName, agentMetrics := range metrics {
				fmt.Printf("  %s: %d messages, %.2f%% success rate\n",
					agentName, agentMetrics.MessagesProcessed, agentMetrics.SuccessRate*100)
			}
		}

		fmt.Println("")
		time.Sleep(2 * time.Second)
	}

	// Show final status
	fmt.Println("Final orchestrator status:")
	finalStatus := orchestrator.GetStatus()
	fmt.Printf("  Orchestrator: %s\n", finalStatus["orchestrator"])
	fmt.Printf("  Status: %v\n", finalStatus["status"])

	if agents, ok := finalStatus["agents"].(map[string]interface{}); ok {
		fmt.Printf("  Registered Agents: %d\n", len(agents))
		for name, agentInfo := range agents {
			if info, ok := agentInfo.(map[string]interface{}); ok {
				fmt.Printf("    - %s: %v\n", name, info["status"])
			}
		}
	}

	// Stop all agents
	fmt.Println("\nStopping agents...")
	detector.Stop()
	analyzer.Stop()
	responder.Stop()
	notifier.Stop()

	fmt.Println("Orchestration demo completed!")
}

// Mock agents for demonstration

type AnomalyDetectorAgent struct {
	name   string
	status core.PluginStatus
}

func NewAnomalyDetectorAgent(name string) *AnomalyDetectorAgent {
	return &AnomalyDetectorAgent{
		name:   name,
		status: core.PluginStatusStopped,
	}
}

func (a *AnomalyDetectorAgent) GetName() string              { return a.name }
func (a *AnomalyDetectorAgent) GetStatus() core.PluginStatus { return a.status }
func (a *AnomalyDetectorAgent) GetCapabilities() []string {
	return []string{"detect", "monitor", "alert"}
}

func (a *AnomalyDetectorAgent) Start(ctx context.Context) error {
	a.status = core.PluginStatusRunning
	fmt.Printf("  %s started\n", a.name)
	return nil
}

func (a *AnomalyDetectorAgent) Stop() error {
	a.status = core.PluginStatusStopped
	fmt.Printf("  %s stopped\n", a.name)
	return nil
}

func (a *AnomalyDetectorAgent) ProcessMessage(ctx context.Context, msg *agents.Message) (*agents.Message, error) {
	fmt.Printf("  %s processing: %s\n", a.name, msg.Content)

	// Simulate anomaly detection
	time.Sleep(1 * time.Second)

	return &agents.Message{
		ID:      fmt.Sprintf("response-%d", time.Now().Unix()),
		From:    a.name,
		To:      msg.From,
		Type:    "detection_result",
		Content: "Anomaly detected: CPU spike to 85%",
		Data: map[string]interface{}{
			"anomaly_type": "cpu_spike",
			"severity":     "high",
			"value":        85.0,
			"threshold":    80.0,
		},
		Timestamp: time.Now(),
	}, nil
}

type RootCauseAgent struct {
	name   string
	status core.PluginStatus
	apiKey string
}

func NewRootCauseAgent(name, apiKey string) *RootCauseAgent {
	return &RootCauseAgent{
		name:   name,
		status: core.PluginStatusStopped,
		apiKey: apiKey,
	}
}

func (a *RootCauseAgent) GetName() string              { return a.name }
func (a *RootCauseAgent) GetStatus() core.PluginStatus { return a.status }
func (a *RootCauseAgent) GetCapabilities() []string {
	return []string{"analyze", "diagnose", "investigate"}
}

func (a *RootCauseAgent) Start(ctx context.Context) error {
	a.status = core.PluginStatusRunning
	fmt.Printf("  %s started\n", a.name)
	return nil
}

func (a *RootCauseAgent) Stop() error {
	a.status = core.PluginStatusStopped
	fmt.Printf("  %s stopped\n", a.name)
	return nil
}

func (a *RootCauseAgent) ProcessMessage(ctx context.Context, msg *agents.Message) (*agents.Message, error) {
	fmt.Printf("  %s processing: %s\n", a.name, msg.Content)

	// Simulate root cause analysis
	time.Sleep(2 * time.Second)

	return &agents.Message{
		ID:      fmt.Sprintf("response-%d", time.Now().Unix()),
		From:    a.name,
		To:      msg.From,
		Type:    "analysis_result",
		Content: "Root cause identified: Memory leak in user service",
		Data: map[string]interface{}{
			"root_cause":     "memory_leak",
			"service":        "user-service",
			"confidence":     0.85,
			"recommendation": "Restart user service pods",
		},
		Timestamp: time.Now(),
	}, nil
}

type RemediationAgent struct {
	name   string
	status core.PluginStatus
	apiKey string
}

func NewRemediationAgent(name, apiKey string) *RemediationAgent {
	return &RemediationAgent{
		name:   name,
		status: core.PluginStatusStopped,
		apiKey: apiKey,
	}
}

func (a *RemediationAgent) GetName() string              { return a.name }
func (a *RemediationAgent) GetStatus() core.PluginStatus { return a.status }
func (a *RemediationAgent) GetCapabilities() []string {
	return []string{"remediate", "fix", "restart", "scale"}
}

func (a *RemediationAgent) Start(ctx context.Context) error {
	a.status = core.PluginStatusRunning
	fmt.Printf("  %s started\n", a.name)
	return nil
}

func (a *RemediationAgent) Stop() error {
	a.status = core.PluginStatusStopped
	fmt.Printf("  %s stopped\n", a.name)
	return nil
}

func (a *RemediationAgent) ProcessMessage(ctx context.Context, msg *agents.Message) (*agents.Message, error) {
	fmt.Printf("  %s processing: %s\n", a.name, msg.Content)

	// Simulate remediation
	time.Sleep(3 * time.Second)

	return &agents.Message{
		ID:      fmt.Sprintf("response-%d", time.Now().Unix()),
		From:    a.name,
		To:      msg.From,
		Type:    "remediation_result",
		Content: "Remediation completed: User service pods restarted",
		Data: map[string]interface{}{
			"action_taken":   "restart_pods",
			"service":        "user-service",
			"pods_restarted": 3,
			"success":        true,
			"duration":       "3s",
		},
		Timestamp: time.Now(),
	}, nil
}

type NotificationAgent struct {
	name   string
	status core.PluginStatus
}

func NewNotificationAgent(name string) *NotificationAgent {
	return &NotificationAgent{
		name:   name,
		status: core.PluginStatusStopped,
	}
}

func (a *NotificationAgent) GetName() string              { return a.name }
func (a *NotificationAgent) GetStatus() core.PluginStatus { return a.status }
func (a *NotificationAgent) GetCapabilities() []string {
	return []string{"notify", "alert", "email", "slack"}
}

func (a *NotificationAgent) Start(ctx context.Context) error {
	a.status = core.PluginStatusRunning
	fmt.Printf("  %s started\n", a.name)
	return nil
}

func (a *NotificationAgent) Stop() error {
	a.status = core.PluginStatusStopped
	fmt.Printf("  %s stopped\n", a.name)
	return nil
}

func (a *NotificationAgent) ProcessMessage(ctx context.Context, msg *agents.Message) (*agents.Message, error) {
	fmt.Printf("  %s processing: %s\n", a.name, msg.Content)

	// Simulate notification
	time.Sleep(1 * time.Second)

	return &agents.Message{
		ID:      fmt.Sprintf("response-%d", time.Now().Unix()),
		From:    a.name,
		To:      msg.From,
		Type:    "notification_result",
		Content: "Team notified: Incident resolved",
		Data: map[string]interface{}{
			"channels":   []string{"slack", "email"},
			"recipients": []string{"devops-team", "on-call"},
			"message":    "Incident resolved: CPU spike fixed by restarting user service pods",
			"success":    true,
		},
		Timestamp: time.Now(),
	}, nil
}
