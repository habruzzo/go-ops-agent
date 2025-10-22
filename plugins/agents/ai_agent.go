package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/habruzzo/agent/core"
)

// AIAgent implements the AgentPlugin interface using external AI APIs
type AIAgent struct {
	name        string
	version     string
	status      core.PluginStatus
	apiKey      string
	apiURL      string
	model       string
	httpClient  *http.Client
	contextData []core.DataPoint
	mu          sync.RWMutex
}

// NewAIAgent creates a new AI agent plugin
func NewAIAgent(name string) *AIAgent {
	return &AIAgent{
		name:       name,
		version:    "1.0.0",
		status:     core.PluginStatusStopped,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the name of the plugin
func (a *AIAgent) Name() string {
	return a.name
}

// Type returns the type of plugin
func (a *AIAgent) Type() core.PluginType {
	return core.PluginTypeAgent
}

// Version returns the plugin version
func (a *AIAgent) Version() string {
	return a.version
}

// Configure initializes the plugin with configuration
func (a *AIAgent) Configure(config map[string]interface{}) error {
	apiKey, ok := config["api_key"].(string)
	if !ok {
		return fmt.Errorf("API key not specified")
	}
	a.apiKey = apiKey

	if apiURL, ok := config["api_url"].(string); ok {
		a.apiURL = apiURL
	} else {
		a.apiURL = "https://api.openai.com/v1/chat/completions"
	}

	if model, ok := config["model"].(string); ok {
		a.model = model
	} else {
		a.model = "gpt-3.5-turbo"
	}

	return nil
}

// Start begins the plugin's operation
func (a *AIAgent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status == core.PluginStatusRunning {
		return fmt.Errorf("agent is already running")
	}

	a.status = core.PluginStatusStarting
	slog.Info("Starting AI agent", "plugin", a.name, "type", a.Type())

	// Test API connectivity
	if err := a.Health(ctx); err != nil {
		a.status = core.PluginStatusError
		return fmt.Errorf("health check failed: %w", err)
	}

	a.status = core.PluginStatusRunning
	slog.Info("AI agent started", "plugin", a.name, "type", a.Type())
	return nil
}

// Stop gracefully stops the plugin
func (a *AIAgent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status != core.PluginStatusRunning {
		return fmt.Errorf("agent is not running")
	}

	a.status = core.PluginStatusStopping
	slog.Info("Stopping AI agent", "plugin", a.name, "type", a.Type())

	a.status = core.PluginStatusStopped
	slog.Info("AI agent stopped", "plugin", a.name, "type", a.Type())
	return nil
}

// Status returns the current status of the plugin
func (a *AIAgent) Status() core.PluginStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

// Health checks if the plugin is healthy
func (a *AIAgent) Health(ctx context.Context) error {
	if a.apiKey == "" {
		return fmt.Errorf("API key not configured")
	}

	// Test API connectivity with a simple request
	testRequest := map[string]interface{}{
		"model": a.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Hello",
			},
		},
		"max_tokens": 5,
	}

	_, err := a.callAIAPI(testRequest)
	return err
}

// GetCapabilities returns what this plugin can do
func (a *AIAgent) GetCapabilities() []string {
	return []string{
		"analyze_metrics",
		"detect_anomalies",
		"provide_recommendations",
		"troubleshoot_issues",
		"explain_patterns",
		"predict_trends",
	}
}

// ProcessQuery handles user queries and returns responses
func (a *AIAgent) ProcessQuery(ctx context.Context, query string) (*core.AgentResponse, error) {
	if a.status != core.PluginStatusRunning {
		return nil, fmt.Errorf("agent is not running")
	}

	// Prepare context-aware prompt
	prompt := a.buildPrompt(query)

	// Call AI API
	response, err := a.callAIAPI(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// Convert response to AgentResponse
	agentResponse := a.convertResponseToAgentResponse(response, query)
	return agentResponse, nil
}

// SetContext provides the agent with current system data
func (a *AIAgent) SetContext(data []core.DataPoint) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.contextData = data
}

// GetAvailableQueries returns what types of queries this agent can handle
func (a *AIAgent) GetAvailableQueries() []string {
	return []string{
		"What's causing the high CPU usage?",
		"Show me the top 5 slowest endpoints",
		"Why did the error rate spike at 2pm?",
		"Are there any anomalies in the current metrics?",
		"What recommendations do you have for improving performance?",
		"Explain the current system health",
	}
}

// buildPrompt creates a context-aware prompt for the AI
func (a *AIAgent) buildPrompt(query string) map[string]interface{} {
	contextInfo := ""
	if len(a.contextData) > 0 {
		contextInfo = a.formatContextData()
	}

	systemPrompt := `You are an observability expert AI agent. You have access to real-time system metrics and can help with:
- Analyzing performance issues
- Detecting anomalies and patterns
- Providing troubleshooting recommendations
- Explaining system behavior
- Predicting trends and issues

Current system context:
` + contextInfo + `

Respond in a helpful, technical manner. If you need more specific data, ask for it.`

	return map[string]interface{}{
		"model": a.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": query,
			},
		},
		"temperature": 0.1,
	}
}

// callAIAPI makes the actual API call to the AI service
func (a *AIAgent) callAIAPI(request map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

// convertResponseToAgentResponse converts AI response to AgentResponse format
func (a *AIAgent) convertResponseToAgentResponse(response map[string]interface{}, query string) *core.AgentResponse {
	// Extract content from AI response
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return &core.AgentResponse{
			Query:      query,
			Response:   "I'm sorry, I couldn't process your request.",
			Confidence: 0.0,
			Timestamp:  time.Now(),
		}
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return &core.AgentResponse{
			Query:      query,
			Response:   "I'm sorry, I couldn't process your request.",
			Confidence: 0.0,
			Timestamp:  time.Now(),
		}
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return &core.AgentResponse{
			Query:      query,
			Response:   "I'm sorry, I couldn't process your request.",
			Confidence: 0.0,
			Timestamp:  time.Now(),
		}
	}

	content, ok := message["content"].(string)
	if !ok {
		return &core.AgentResponse{
			Query:      query,
			Response:   "I'm sorry, I couldn't process your request.",
			Confidence: 0.0,
			Timestamp:  time.Now(),
		}
	}

	// Calculate confidence based on response length and context
	confidence := 0.8
	if len(content) < 50 {
		confidence = 0.6
	}

	return &core.AgentResponse{
		Query:      query,
		Response:   content,
		Confidence: confidence,
		Actions:    a.extractActions(content),
		Metadata: map[string]interface{}{
			"model":     a.model,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}
}

// formatContextData formats the current context data for the AI
func (a *AIAgent) formatContextData() string {
	if len(a.contextData) == 0 {
		return "No current data available"
	}

	summary := make(map[string]interface{})
	metrics := make(map[string][]float64)

	for _, point := range a.contextData {
		metrics[point.Metric] = append(metrics[point.Metric], point.Value)
	}

	for metric, values := range metrics {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			summary[metric] = map[string]interface{}{
				"count":  len(values),
				"avg":    sum / float64(len(values)),
				"latest": values[len(values)-1],
			}
		}
	}

	jsonData, _ := json.MarshalIndent(summary, "", "  ")
	return string(jsonData)
}

// extractActions attempts to extract actionable items from the AI response
func (a *AIAgent) extractActions(content string) []core.AgentAction {
	// Simple action extraction - in practice, you'd use more sophisticated NLP
	actions := []core.AgentAction{}

	// Look for common action patterns
	if contains(content, "restart") || contains(content, "reboot") {
		actions = append(actions, core.AgentAction{
			Type:        "restart",
			Description: "Restart the service",
		})
	}

	if contains(content, "scale") || contains(content, "increase") {
		actions = append(actions, core.AgentAction{
			Type:        "scale",
			Description: "Scale up resources",
		})
	}

	return actions
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
