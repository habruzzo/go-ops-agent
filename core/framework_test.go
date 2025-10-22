package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFramework_NewFramework(t *testing.T) {
	config := &FrameworkConfig{
		LogLevel:  "debug",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []PluginConfig{},
	}

	framework := NewFramework(config)

	require.NotNil(t, framework, "Expected framework to be created")
	assert.False(t, framework.running, "Expected framework to not be running initially")
	assert.Equal(t, 0, framework.registry.GetPluginCount(), "Expected no plugins initially")
}

func TestFramework_LoadPlugin(t *testing.T) {
	config := &FrameworkConfig{
		LogLevel:  "info",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []PluginConfig{},
	}

	framework := NewFramework(config)

	// Create a mock plugin
	mockPlugin := &MockPlugin{
		name:       "test-plugin",
		pluginType: PluginTypeCollector,
		status:     PluginStatusStopped,
	}

	err := framework.LoadPlugin(mockPlugin)
	require.NoError(t, err, "Expected no error when loading plugin")

	assert.Equal(t, 1, framework.registry.GetPluginCount(), "Expected 1 plugin after loading")

	plugin, err := framework.registry.GetPlugin("test-plugin")
	require.NoError(t, err, "Expected no error when getting plugin")
	require.NotNil(t, plugin, "Expected plugin to be loaded")
}

func TestFramework_StartStop(t *testing.T) {
	config := &FrameworkConfig{
		LogLevel:  "info",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []PluginConfig{},
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
	require.NoError(t, err, "Failed to load plugin")

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = framework.Start(ctx)
	require.NoError(t, err, "Failed to start framework")
	assert.True(t, framework.running, "Expected framework to be running")

	// Wait a bit to let workers run
	time.Sleep(500 * time.Millisecond)

	// Stop framework
	err = framework.Stop()
	require.NoError(t, err, "Failed to stop framework")

	assert.False(t, framework.running, "Expected framework to be stopped")
}

func TestFramework_GetStatus(t *testing.T) {
	config := &FrameworkConfig{
		LogLevel:  "info",
		LogFormat: "text",
		LogOutput: "stdout",
		Plugins:   []PluginConfig{},
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

	assert.Equal(t, false, status["running"], "Expected framework to not be running")
	assert.Equal(t, 2, status["total_plugins"], "Expected 2 plugins")
	assert.Equal(t, 1, status["collectors"], "Expected 1 collector")
	assert.Equal(t, 1, status["analyzers"], "Expected 1 analyzer")
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
