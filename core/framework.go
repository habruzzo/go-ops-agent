package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Framework manages all plugins and orchestrates their interactions
type Framework struct {
	plugins     map[string]Plugin
	collectors  []DataCollector
	analyzers   []DataAnalyzer
	responders  []DataResponder
	agents      []AgentPlugin
	config      *FrameworkConfig
	running     bool
	mu          sync.RWMutex
	dataChannel chan []DataPoint
	wg          sync.WaitGroup
	shutdown    bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewFramework creates a new framework instance
func NewFramework(config *FrameworkConfig) *Framework {
	// Initialize global logger with configuration
	InitLogger(config.Logging)

	return &Framework{
		plugins:     make(map[string]Plugin),
		collectors:  make([]DataCollector, 0),
		analyzers:   make([]DataAnalyzer, 0),
		responders:  make([]DataResponder, 0),
		agents:      make([]AgentPlugin, 0),
		config:      config,
		running:     false,
		dataChannel: make(chan []DataPoint, 100),
		wg:          sync.WaitGroup{}, // Initialize WaitGroup for graceful shutdown
	}
}

// LoadPlugin loads a plugin into the framework
func (f *Framework) LoadPlugin(plugin Plugin) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already loaded", plugin.Name())
	}

	f.plugins[plugin.Name()] = plugin

	// Add to appropriate category
	switch plugin.Type() {
	case PluginTypeCollector:
		if collector, ok := plugin.(DataCollector); ok {
			f.collectors = append(f.collectors, collector)
		}
	case PluginTypeAnalyzer:
		if analyzer, ok := plugin.(DataAnalyzer); ok {
			f.analyzers = append(f.analyzers, analyzer)
		}
	case PluginTypeResponder:
		if responder, ok := plugin.(DataResponder); ok {
			f.responders = append(f.responders, responder)
		}
	case PluginTypeAgent:
		if agent, ok := plugin.(AgentPlugin); ok {
			f.agents = append(f.agents, agent)
		}
	}

	slog.Info("Plugin loaded", "plugin", plugin.Name(), "type", plugin.Type())
	return nil
}

// UnloadPlugin removes a plugin from the framework
func (f *Framework) UnloadPlugin(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	plugin, exists := f.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Stop the plugin if it's running
	if plugin.Status() == PluginStatusRunning {
		plugin.Stop()
	}

	// Remove from appropriate category
	switch plugin.Type() {
	case PluginTypeCollector:
		f.removeCollector(name)
	case PluginTypeAnalyzer:
		f.removeAnalyzer(name)
	case PluginTypeResponder:
		f.removeResponder(name)
	case PluginTypeAgent:
		f.removeAgent(name)
	}

	delete(f.plugins, name)
	slog.Info("Plugin unloaded", "plugin", name)
	return nil
}

// Start begins the framework's operation
func (f *Framework) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return fmt.Errorf("framework is already running")
	}

	f.running = true
	f.ctx, f.cancel = context.WithCancel(ctx)
	slog.Info("Starting framework...")

	// Start all plugins
	for _, plugin := range f.plugins {
		if err := plugin.Start(f.ctx); err != nil {
			slog.Error("Failed to start plugin", "plugin", plugin.Name(), "error", err)
			continue
		}
	}

	// Start data collection workers
	for _, collector := range f.collectors {
		f.wg.Add(1)
		go f.collectorWorker(f.ctx, collector)
	}

	// Start data processing worker
	f.wg.Add(1)
	go f.dataProcessor(f.ctx)

	slog.Info("Framework started successfully")
	return nil
}

// Stop gracefully stops the framework
func (f *Framework) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return fmt.Errorf("framework is not running")
	}

	f.running = false
	f.shutdown = true
	slog.Info("Stopping framework...")

	// Cancel context to signal workers to stop
	if f.cancel != nil {
		f.cancel()
	}

	// Wait for all workers to finish with timeout
	slog.Info("Waiting for workers to finish...")
	done := make(chan struct{})
	go func() {
		f.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("All workers finished gracefully")
	case <-time.After(30 * time.Second):
		slog.Warn("Timeout waiting for workers to finish")
	}

	// Stop all plugins
	for _, plugin := range f.plugins {
		if err := plugin.Stop(); err != nil {
			slog.Error("Failed to stop plugin", "plugin", plugin.Name(), "error", err)
		}
	}

	slog.Info("Framework stopped")
	return nil
}

// QueryAgent processes a query through the specified agent
func (f *Framework) QueryAgent(ctx context.Context, agentName, query string) (*AgentResponse, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	agent, exists := f.plugins[agentName]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}

	agentPlugin, ok := agent.(AgentPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin %s is not an agent", agentName)
	}

	return agentPlugin.ProcessQuery(ctx, query)
}

// QueryDefaultAgent processes a query through the default agent
func (f *Framework) QueryDefaultAgent(ctx context.Context, query string) (*AgentResponse, error) {
	if f.config.Agent.DefaultAgent == "" {
		return nil, fmt.Errorf("no default agent configured")
	}

	return f.QueryAgent(ctx, f.config.Agent.DefaultAgent, query)
}

// GetDataChannel returns the data channel for testing purposes
func (f *Framework) GetDataChannel() chan []DataPoint {
	return f.dataChannel
}

// GetStatus returns the current status of the framework
func (f *Framework) GetStatus() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	pluginStatus := make(map[string]interface{})
	for name, plugin := range f.plugins {
		pluginStatus[name] = map[string]interface{}{
			"type":   plugin.Type(),
			"status": plugin.Status(),
		}
	}

	return map[string]interface{}{
		"running":       f.running,
		"total_plugins": len(f.plugins),
		"collectors":    len(f.collectors),
		"analyzers":     len(f.analyzers),
		"responders":    len(f.responders),
		"agents":        len(f.agents),
		"plugins":       pluginStatus,
	}
}

// collectorWorker runs a collector in a separate goroutine
func (f *Framework) collectorWorker(ctx context.Context, collector DataCollector) {
	defer f.wg.Done()

	interval := collector.GetCollectionInterval()
	if interval == 0 {
		interval = 30 * time.Second // default
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Collector worker stopping due to context cancellation", "collector", collector.Name())
			return
		case <-ticker.C:
			// Check if framework is shutting down
			f.mu.RLock()
			shutdown := f.shutdown
			f.mu.RUnlock()

			if shutdown {
				slog.Info("Collector worker stopping due to shutdown", "collector", collector.Name())
				return
			}

			data, err := collector.Collect(ctx)
			if err != nil {
				slog.Error("Failed to collect data", "collector", collector.Name(), "error", err)
				continue
			}

			if len(data) > 0 {
				// Send data to processing pipeline
				select {
				case f.dataChannel <- data:
				case <-ctx.Done():
					slog.Info("Collector worker stopping, dropping data", "collector", collector.Name())
					return
				}
			}
		}
	}
}

// dataProcessor processes collected data through analyzers and responders
func (f *Framework) dataProcessor(ctx context.Context) {
	defer f.wg.Done()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Data processor stopping due to context cancellation")
			return
		case data, ok := <-f.dataChannel:
			if !ok {
				slog.Info("Data processor stopping due to channel closure")
				return
			}
			f.processData(ctx, data)
		}
	}
}

// processData runs data through analyzers and triggers responders
func (f *Framework) processData(ctx context.Context, data []DataPoint) {
	// Update agent context
	for _, agent := range f.agents {
		agent.SetContext(data)
	}

	// Run analyzers
	for _, analyzer := range f.analyzers {
		if !analyzer.CanAnalyze(data) {
			continue
		}

		analysis, err := analyzer.Analyze(data)
		if err != nil {
			slog.Error("Failed to analyze data", "analyzer", analyzer.Name(), "error", err)
			continue
		}

		if analysis == nil {
			continue
		}

		// Trigger responders
		for _, responder := range f.responders {
			if !responder.CanHandle(analysis) {
				continue
			}

			if err := responder.Respond(ctx, analysis); err != nil {
				slog.Error("Failed to respond", "responder", responder.Name(), "error", err)
			}
		}
	}
}

// Helper methods to remove plugins from categories
func (f *Framework) removeCollector(name string) {
	for i, collector := range f.collectors {
		if collector.Name() == name {
			f.collectors = append(f.collectors[:i], f.collectors[i+1:]...)
			break
		}
	}
}

func (f *Framework) removeAnalyzer(name string) {
	for i, analyzer := range f.analyzers {
		if analyzer.Name() == name {
			f.analyzers = append(f.analyzers[:i], f.analyzers[i+1:]...)
			break
		}
	}
}

func (f *Framework) removeResponder(name string) {
	for i, responder := range f.responders {
		if responder.Name() == name {
			f.responders = append(f.responders[:i], f.responders[i+1:]...)
			break
		}
	}
}

func (f *Framework) removeAgent(name string) {
	for i, agent := range f.agents {
		if agent.Name() == name {
			f.agents = append(f.agents[:i], f.agents[i+1:]...)
			break
		}
	}
}
