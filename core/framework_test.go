package core

import (
	"context"
	"testing"
	"time"
)

func TestFramework_NewFramework(t *testing.T) {
	config := &FrameworkConfig{
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []PluginConfig{},
	}

	framework := NewFramework(config)

	if framework == nil {
		t.Fatal("Expected framework to be created")
	}

	if framework.running {
		t.Error("Expected framework to not be running initially")
	}

	if len(framework.plugins) != 0 {
		t.Error("Expected no plugins initially")
	}
}

func TestFramework_LoadPlugin(t *testing.T) {
	config := &FrameworkConfig{
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []PluginConfig{},
	}

	framework := NewFramework(config)

	// Create a mock plugin
	mockPlugin := &MockPlugin{
		name:       "test-plugin",
		pluginType: PluginTypeCollector,
		status:     PluginStatusStopped,
	}

	err := framework.LoadPlugin(mockPlugin)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(framework.plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(framework.plugins))
	}

	if framework.plugins["test-plugin"] == nil {
		t.Error("Expected plugin to be loaded")
	}
}

func TestFramework_StartStop(t *testing.T) {
	config := &FrameworkConfig{
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []PluginConfig{},
	}

	framework := NewFramework(config)

	// Add a mock collector
	mockCollector := &MockCollector{
		MockPlugin: MockPlugin{
			name:       "test-collector",
			pluginType: PluginTypeCollector,
			status:     PluginStatusStopped,
		},
		interval: 100 * time.Millisecond,
	}

	err := framework.LoadPlugin(mockCollector)
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = framework.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	if !framework.running {
		t.Error("Expected framework to be running")
	}

	// Wait a bit to let workers run
	time.Sleep(500 * time.Millisecond)

	// Stop framework
	err = framework.Stop()
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	if framework.running {
		t.Error("Expected framework to be stopped")
	}
}

func TestFramework_GetStatus(t *testing.T) {
	config := &FrameworkConfig{
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Plugins: []PluginConfig{},
	}

	framework := NewFramework(config)

	// Add some mock plugins
	framework.LoadPlugin(&MockCollector{
		MockPlugin: MockPlugin{
			name:       "collector-1",
			pluginType: PluginTypeCollector,
			status:     PluginStatusRunning,
		},
		interval: 30 * time.Second,
	})

	framework.LoadPlugin(&MockAnalyzer{
		MockPlugin: MockPlugin{
			name:       "analyzer-1",
			pluginType: PluginTypeAnalyzer,
			status:     PluginStatusRunning,
		},
	})

	status := framework.GetStatus()

	if status["running"] != false {
		t.Error("Expected framework to not be running")
	}

	if status["total_plugins"] != 2 {
		t.Errorf("Expected 2 plugins, got %v", status["total_plugins"])
	}

	if status["collectors"] != 1 {
		t.Errorf("Expected 1 collector, got %v", status["collectors"])
	}

	if status["analyzers"] != 1 {
		t.Errorf("Expected 1 analyzer, got %v", status["analyzers"])
	}
}

// Mock implementations for testing

type MockPlugin struct {
	name       string
	pluginType PluginType
	status     PluginStatus
}

func (m *MockPlugin) Name() string                                  { return m.name }
func (m *MockPlugin) Type() PluginType                              { return m.pluginType }
func (m *MockPlugin) Version() string                               { return "1.0.0" }
func (m *MockPlugin) Configure(config map[string]interface{}) error { return nil }
func (m *MockPlugin) Start(ctx context.Context) error {
	m.status = PluginStatusRunning
	return nil
}
func (m *MockPlugin) Stop() error {
	m.status = PluginStatusStopped
	return nil
}
func (m *MockPlugin) Status() PluginStatus             { return m.status }
func (m *MockPlugin) Health(ctx context.Context) error { return nil }
func (m *MockPlugin) GetCapabilities() []string        { return []string{"test"} }

type MockCollector struct {
	MockPlugin
	interval time.Duration
}

func (m *MockCollector) Collect(ctx context.Context) ([]DataPoint, error) {
	return []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    m.name,
			Metric:    "test_metric",
			Value:     1.0,
			Labels:    map[string]string{"test": "true"},
		},
	}, nil
}

func (m *MockCollector) GetCollectionInterval() time.Duration {
	return m.interval
}

type MockAnalyzer struct {
	MockPlugin
}

func (m *MockAnalyzer) Analyze(data []DataPoint) (*Analysis, error) {
	return nil, nil
}

func (m *MockAnalyzer) CanAnalyze(data []DataPoint) bool {
	return len(data) > 0
}
