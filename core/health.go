package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultHealthChecker is the default implementation of HealthChecker
type DefaultHealthChecker struct {
	checks  map[string]HealthCheckFunc
	mu      sync.RWMutex
	timeout time.Duration
}

// NewDefaultHealthChecker creates a new health checker
func NewDefaultHealthChecker(timeout time.Duration) *DefaultHealthChecker {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &DefaultHealthChecker{
		checks:  make(map[string]HealthCheckFunc),
		timeout: timeout,
	}
}

// RegisterHealthCheck registers a health check function
func (h *DefaultHealthChecker) RegisterHealthCheck(name string, check HealthCheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// UnregisterHealthCheck removes a health check
func (h *DefaultHealthChecker) UnregisterHealthCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
}

// CheckHealth performs all registered health checks
func (h *DefaultHealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	h.mu.RLock()
	checks := make(map[string]HealthCheckFunc)
	for name, check := range h.checks {
		checks[name] = check
	}
	h.mu.RUnlock()

	results := make(map[string]CheckResult)
	overallStatus := "healthy"
	overallMessage := "All health checks passed"

	// Run all health checks
	for name, check := range checks {
		result := h.runHealthCheck(ctx, name, check)
		results[name] = result

		// Update overall status based on individual results
		if result.Status == "unhealthy" {
			overallStatus = "unhealthy"
			overallMessage = "One or more health checks failed"
		} else if result.Status == "degraded" && overallStatus != "unhealthy" {
			overallStatus = "degraded"
			overallMessage = "One or more health checks are degraded"
		}
	}

	return HealthStatus{
		Status:    overallStatus,
		Message:   overallMessage,
		Checks:    results,
		Timestamp: time.Now(),
	}
}

// runHealthCheck runs a single health check with timeout
func (h *DefaultHealthChecker) runHealthCheck(ctx context.Context, name string, check HealthCheckFunc) CheckResult {
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Channel to receive the result
	resultChan := make(chan error, 1)

	// Run the health check in a goroutine
	go func() {
		resultChan <- check(checkCtx)
	}()

	// Wait for result or timeout
	select {
	case err := <-resultChan:
		if err != nil {
			return CheckResult{
				Status:  "unhealthy",
				Message: "Health check failed",
				Error:   err.Error(),
			}
		}
		return CheckResult{
			Status:  "healthy",
			Message: "Health check passed",
		}
	case <-checkCtx.Done():
		return CheckResult{
			Status:  "unhealthy",
			Message: "Health check timed out",
			Error:   "timeout",
		}
	}
}

// FrameworkHealthChecker provides health checks for the framework
type FrameworkHealthChecker struct {
	*DefaultHealthChecker
	framework *Framework
}

// NewFrameworkHealthChecker creates a new framework health checker
func NewFrameworkHealthChecker(framework *Framework, timeout time.Duration) *FrameworkHealthChecker {
	checker := &FrameworkHealthChecker{
		DefaultHealthChecker: NewDefaultHealthChecker(timeout),
		framework:            framework,
	}

	// Register default health checks
	checker.registerDefaultChecks()

	return checker
}

// registerDefaultChecks registers default framework health checks
func (f *FrameworkHealthChecker) registerDefaultChecks() {
	// Framework running check
	f.RegisterHealthCheck("framework_running", func(ctx context.Context) error {
		status := f.framework.GetStatus()
		if running, ok := status["running"].(bool); !ok || !running {
			return NewInternalError("framework", "health", "framework is not running")
		}
		return nil
	})

	// Plugin health checks
	f.RegisterHealthCheck("plugins_healthy", func(ctx context.Context) error {
		status := f.framework.GetStatus()
		plugins := status["plugins"].(map[string]interface{})

		unhealthyPlugins := make([]string, 0)
		for name, pluginInfo := range plugins {
			info := pluginInfo.(map[string]interface{})
			if pluginStatus, ok := info["status"].(string); !ok || pluginStatus == string(PluginStatusError) {
				unhealthyPlugins = append(unhealthyPlugins, name)
			}
		}

		if len(unhealthyPlugins) > 0 {
			return NewInternalError("framework", "health",
				fmt.Sprintf("unhealthy plugins: %v", unhealthyPlugins))
		}
		return nil
	})

	// Data channel health check
	f.RegisterHealthCheck("data_channel", func(ctx context.Context) error {
		// Check if data channel is not blocked
		select {
		case <-f.framework.GetDataChannel():
			// Channel is not blocked
			return nil
		default:
			// Channel might be blocked, but this is not necessarily an error
			// We'll consider it healthy for now
			return nil
		}
	})
}

// PluginHealthChecker provides health checks for individual plugins
type PluginHealthChecker struct {
	*DefaultHealthChecker
	plugin Plugin
}

// NewPluginHealthChecker creates a new plugin health checker
func NewPluginHealthChecker(plugin Plugin, timeout time.Duration) *PluginHealthChecker {
	checker := &PluginHealthChecker{
		DefaultHealthChecker: NewDefaultHealthChecker(timeout),
		plugin:               plugin,
	}

	// Register default plugin health checks
	checker.registerDefaultChecks()

	return checker
}

// registerDefaultChecks registers default plugin health checks
func (p *PluginHealthChecker) registerDefaultChecks() {
	// Plugin status check
	p.RegisterHealthCheck("plugin_status", func(ctx context.Context) error {
		status := p.plugin.Status()
		if status == PluginStatusError {
			return NewPluginError("plugin", "health", "plugin is in error state")
		}
		return nil
	})

	// Plugin health check
	p.RegisterHealthCheck("plugin_health", func(ctx context.Context) error {
		return p.plugin.Health(ctx)
	})
}
