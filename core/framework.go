package core

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Framework manages all plugins and orchestrates their interactions
type Framework struct {
	registry         PluginRegistry
	factory          PluginFactory
	configManager    ConfigurationManager
	healthChecker    HealthChecker
	metricsCollector MetricsCollector
	eventBus         EventBus
	config           *FrameworkConfig
	running          bool
	mu               sync.RWMutex
	dataChannel      chan []DataPoint
	wg               sync.WaitGroup
	shutdown         bool
	ctx              context.Context
	cancel           context.CancelFunc
	startTime        time.Time
}

// NewFramework creates a new framework instance with default dependencies
func NewFramework(config *FrameworkConfig) *Framework {
	// Initialize global logger with configuration
	InitLogger(config)

	// Create default dependencies
	registry := NewDefaultPluginRegistry()
	factory := NewDefaultPluginFactory()

	framework := &Framework{
		registry:    registry,
		factory:     factory,
		config:      config,
		running:     false,
		dataChannel: make(chan []DataPoint, config.DataChannelSize),
		wg:          sync.WaitGroup{},
	}

	// Create health checker with framework reference
	healthChecker := NewFrameworkHealthChecker(framework, config.HealthCheckTimeout)
	framework.healthChecker = healthChecker

	return framework
}

// NewFrameworkWithDependencies creates a new framework instance with custom dependencies
func NewFrameworkWithDependencies(
	config *FrameworkConfig,
	registry PluginRegistry,
	factory PluginFactory,
	configManager ConfigurationManager,
	healthChecker HealthChecker,
	metricsCollector MetricsCollector,
	eventBus EventBus,
) *Framework {
	// Initialize global logger with configuration
	InitLogger(config)

	return &Framework{
		registry:         registry,
		factory:          factory,
		configManager:    configManager,
		healthChecker:    healthChecker,
		metricsCollector: metricsCollector,
		eventBus:         eventBus,
		config:           config,
		running:          false,
		dataChannel:      make(chan []DataPoint, config.DataChannelSize),
		wg:               sync.WaitGroup{},
	}
}

// LoadPlugin loads a plugin into the framework
func (f *Framework) LoadPlugin(plugin Plugin) error {
	if err := f.registry.RegisterPlugin(plugin); err != nil {
		return WrapError(err, ErrorTypePlugin, "framework", "load", "failed to register plugin")
	}

	// Publish plugin loaded event
	if f.eventBus != nil {
		event := Event{
			Type:      "plugin_loaded",
			Source:    "framework",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin_name": plugin.Name(),
				"plugin_type": plugin.Type(),
			},
		}
		f.eventBus.Publish(event)
	}

	slog.Info("Plugin loaded", "plugin", plugin.Name(), "type", plugin.Type())
	return nil
}

// LoadPluginFromConfig loads a plugin from configuration
func (f *Framework) LoadPluginFromConfig(config PluginConfig) error {
	plugin, err := f.factory.CreatePlugin(config)
	if err != nil {
		return WrapError(err, ErrorTypePlugin, "framework", "load", "failed to create plugin from config")
	}

	return f.LoadPlugin(plugin)
}

// UnloadPlugin removes a plugin from the framework
func (f *Framework) UnloadPlugin(name string) error {
	plugin, err := f.registry.GetPlugin(name)
	if err != nil {
		return WrapError(err, ErrorTypePlugin, "framework", "unload", "plugin not found")
	}

	// Stop the plugin if it's running
	if plugin.Status() == PluginStatusRunning {
		if err := plugin.Stop(); err != nil {
			slog.Error("Failed to stop plugin during unload", "plugin", name, "error", err)
		}
	}

	// Unregister from registry
	if err := f.registry.UnregisterPlugin(name); err != nil {
		return WrapError(err, ErrorTypePlugin, "framework", "unload", "failed to unregister plugin")
	}

	// Publish plugin unloaded event
	if f.eventBus != nil {
		event := Event{
			Type:      "plugin_unloaded",
			Source:    "framework",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin_name": name,
				"plugin_type": plugin.Type(),
			},
		}
		f.eventBus.Publish(event)
	}

	slog.Info("Plugin unloaded", "plugin", name)
	return nil
}

// Start begins the framework's operation
func (f *Framework) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return NewInternalError("framework", "start", "framework is already running")
	}

	f.running = true
	f.startTime = time.Now()
	f.ctx, f.cancel = context.WithCancel(ctx)
	slog.Info("Starting framework...")

	// Start all plugins
	plugins := f.registry.ListPlugins()
	for _, plugin := range plugins {
		if err := plugin.Start(f.ctx); err != nil {
			slog.Error("Failed to start plugin", "plugin", plugin.Name(), "error", err)
			continue
		}
	}

	// Start data collection workers for collectors
	collectors := f.registry.ListPluginsByType(PluginTypeCollector)
	for _, plugin := range collectors {
		if collector, ok := plugin.(DataCollector); ok {
			f.wg.Add(1)
			go f.collectorWorker(f.ctx, collector)
		}
	}

	// Start data processing worker
	f.wg.Add(1)
	go f.dataProcessor(f.ctx)

	// Start health endpoints
	f.wg.Add(1)
	go f.startHealthEndpoints(f.ctx)

	// Publish framework started event
	if f.eventBus != nil {
		event := Event{
			Type:      "framework_started",
			Source:    "framework",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin_count": len(plugins),
			},
		}
		f.eventBus.Publish(event)
	}

	slog.Info("Framework started successfully", "plugin_count", len(plugins))
	return nil
}

// Stop gracefully stops the framework
func (f *Framework) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return NewInternalError("framework", "stop", "framework is not running")
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
	plugins := f.registry.ListPlugins()
	for _, plugin := range plugins {
		if err := plugin.Stop(); err != nil {
			slog.Error("Failed to stop plugin", "plugin", plugin.Name(), "error", err)
		}
	}

	// Publish framework stopped event
	if f.eventBus != nil {
		event := Event{
			Type:      "framework_stopped",
			Source:    "framework",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"uptime": time.Since(f.startTime),
			},
		}
		f.eventBus.Publish(event)
	}

	slog.Info("Framework stopped")
	return nil
}

// QueryAgent processes a query through the specified agent
func (f *Framework) QueryAgent(ctx context.Context, agentName, query string) (*AgentResponse, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	plugin, err := f.registry.GetPlugin(agentName)
	if err != nil {
		return nil, NewPluginError("framework", "query", fmt.Sprintf("agent %s not found", agentName))
	}

	agentPlugin, ok := plugin.(AgentPlugin)
	if !ok {
		return nil, NewPluginError("framework", "query", fmt.Sprintf("plugin %s is not an agent", agentName))
	}

	return agentPlugin.ProcessQuery(ctx, query)
}

// QueryDefaultAgent processes a query through the default agent
func (f *Framework) QueryDefaultAgent(ctx context.Context, query string) (*AgentResponse, error) {
	if f.config.DefaultAgent == "" {
		return nil, NewConfigurationError("framework", "query", "no default agent configured")
	}

	return f.QueryAgent(ctx, f.config.DefaultAgent, query)
}

// GetDataChannel returns the data channel for testing purposes
func (f *Framework) GetDataChannel() chan []DataPoint {
	return f.dataChannel
}

// GetStatus returns the current status of the framework
func (f *Framework) GetStatus() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	plugins := f.registry.ListPlugins()
	pluginStatus := make(map[string]interface{})

	for _, plugin := range plugins {
		pluginStatus[plugin.Name()] = map[string]interface{}{
			"type":   plugin.Type(),
			"status": plugin.Status(),
		}
	}

	status := map[string]interface{}{
		"running":       f.running,
		"total_plugins": len(plugins),
		"collectors":    f.registry.GetPluginCountByType(PluginTypeCollector),
		"analyzers":     f.registry.GetPluginCountByType(PluginTypeAnalyzer),
		"responders":    f.registry.GetPluginCountByType(PluginTypeResponder),
		"agents":        f.registry.GetPluginCountByType(PluginTypeAgent),
		"plugins":       pluginStatus,
	}

	if f.startTime != (time.Time{}) {
		status["uptime"] = time.Since(f.startTime)
	}

	return status
}

// GetHealthStatus returns the health status of the framework
func (f *Framework) GetHealthStatus(ctx context.Context) HealthStatus {
	if f.healthChecker != nil {
		return f.healthChecker.CheckHealth(ctx)
	}

	// Fallback health check
	return HealthStatus{
		Status:    "healthy",
		Message:   "Framework is running",
		Timestamp: time.Now(),
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
	agents := f.registry.ListPluginsByType(PluginTypeAgent)
	for _, plugin := range agents {
		if agent, ok := plugin.(AgentPlugin); ok {
			agent.SetContext(data)
		}
	}

	// Run analyzers
	analyzers := f.registry.ListPluginsByType(PluginTypeAnalyzer)
	for _, plugin := range analyzers {
		analyzer, ok := plugin.(DataAnalyzer)
		if !ok {
			continue
		}

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
		responders := f.registry.ListPluginsByType(PluginTypeResponder)
		for _, plugin := range responders {
			responder, ok := plugin.(DataResponder)
			if !ok {
				continue
			}

			if !responder.CanHandle(analysis) {
				continue
			}

			if err := responder.Respond(ctx, analysis); err != nil {
				slog.Error("Failed to respond", "responder", responder.Name(), "error", err)
			}
		}
	}
}

// GetRegistry returns the plugin registry
func (f *Framework) GetRegistry() PluginRegistry {
	return f.registry
}

// GetFactory returns the plugin factory
func (f *Framework) GetFactory() PluginFactory {
	return f.factory
}

// GetHealthChecker returns the health checker
func (f *Framework) GetHealthChecker() HealthChecker {
	return f.healthChecker
}

// startHealthEndpoints starts HTTP health check endpoints
func (f *Framework) startHealthEndpoints(ctx context.Context) {
	defer f.wg.Done()

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		f.mu.RLock()
		running := f.running && !f.shutdown
		f.mu.RUnlock()

		if running {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service Unavailable"))
		}
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		f.mu.RLock()
		running := f.running && !f.shutdown
		pluginCount := f.registry.GetPluginCount()
		f.mu.RUnlock()

		if running && pluginCount > 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	})

	// Metrics endpoint (basic)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		f.mu.RLock()
		status := f.GetStatus()
		f.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "# Agent Framework Metrics\n")
		fmt.Fprintf(w, "framework_running %t\n", status["running"])
		fmt.Fprintf(w, "framework_total_plugins %d\n", status["total_plugins"])
		fmt.Fprintf(w, "framework_collectors %d\n", status["collectors"])
		fmt.Fprintf(w, "framework_analyzers %d\n", status["analyzers"])
		fmt.Fprintf(w, "framework_responders %d\n", status["responders"])
		fmt.Fprintf(w, "framework_agents %d\n", status["agents"])
	})

	// Status endpoint (JSON)
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		f.mu.RLock()
		status := f.GetStatus()
		f.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON response
		fmt.Fprintf(w, `{
			"running": %t,
			"total_plugins": %d,
			"collectors": %d,
			"analyzers": %d,
			"responders": %d,
			"agents": %d,
			"uptime": "%v"
		}`,
			status["running"],
			status["total_plugins"],
			status["collectors"],
			status["analyzers"],
			status["responders"],
			status["agents"],
			status["uptime"])
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", f.config.ServerHost, f.config.ServerPort),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Health endpoints server error", "error", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Failed to shutdown health endpoints server", "error", err)
	}
}
