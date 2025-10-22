package core

import (
	"fmt"
	"sync"
)

// PluginFactory creates plugin instances
type PluginFactory interface {
	CreatePlugin(config PluginConfig) (Plugin, error)
	RegisterPluginCreator(pluginType string, creator PluginCreator) error
	GetSupportedTypes() []string
}

// DefaultPluginFactory is the default implementation of PluginFactory
type DefaultPluginFactory struct {
	creators map[string]PluginCreator
	mu       sync.RWMutex
}

// PluginCreator is a function that creates a plugin instance
type PluginCreator func(config PluginConfig) (Plugin, error)

// NewDefaultPluginFactory creates a new default plugin factory
func NewDefaultPluginFactory() *DefaultPluginFactory {
	return &DefaultPluginFactory{
		creators: make(map[string]PluginCreator),
	}
}

// RegisterPluginCreator registers a plugin creator for a specific type
func (f *DefaultPluginFactory) RegisterPluginCreator(pluginType string, creator PluginCreator) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[pluginType] = creator
	return nil
}

// CreatePlugin creates a plugin instance based on the configuration
func (f *DefaultPluginFactory) CreatePlugin(config PluginConfig) (Plugin, error) {
	f.mu.RLock()
	creator, exists := f.creators[config.Type]
	f.mu.RUnlock()

	if !exists {
		return nil, NewPluginError("factory", "create", fmt.Sprintf("unknown plugin type: %s", config.Type))
	}

	plugin, err := creator(config)
	if err != nil {
		return nil, WrapError(err, ErrorTypePlugin, "factory", "create", fmt.Sprintf("failed to create plugin %s", config.Name))
	}

	return plugin, nil
}

// GetSupportedTypes returns all supported plugin types
func (f *DefaultPluginFactory) GetSupportedTypes() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]string, 0, len(f.creators))
	for pluginType := range f.creators {
		types = append(types, pluginType)
	}
	return types
}

// PluginRegistry manages plugin registration and discovery
type DefaultPluginRegistry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewDefaultPluginRegistry creates a new plugin registry
func NewDefaultPluginRegistry() *DefaultPluginRegistry {
	return &DefaultPluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

// RegisterPlugin registers a plugin in the registry
func (r *DefaultPluginRegistry) RegisterPlugin(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[plugin.Name()]; exists {
		return NewPluginError("registry", "register", fmt.Sprintf("plugin %s already registered", plugin.Name()))
	}

	r.plugins[plugin.Name()] = plugin
	return nil
}

// UnregisterPlugin removes a plugin from the registry
func (r *DefaultPluginRegistry) UnregisterPlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return NewPluginError("registry", "unregister", fmt.Sprintf("plugin %s not found", name))
	}

	delete(r.plugins, name)
	return nil
}

// GetPlugin retrieves a plugin by name
func (r *DefaultPluginRegistry) GetPlugin(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, NewPluginError("registry", "get", fmt.Sprintf("plugin %s not found", name))
	}

	return plugin, nil
}

// ListPlugins returns all registered plugins
func (r *DefaultPluginRegistry) ListPlugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// ListPluginsByType returns all plugins of a specific type
func (r *DefaultPluginRegistry) ListPluginsByType(pluginType PluginType) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plugins []Plugin
	for _, plugin := range r.plugins {
		if plugin.Type() == pluginType {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

// GetPluginCount returns the number of registered plugins
func (r *DefaultPluginRegistry) GetPluginCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// GetPluginCountByType returns the number of plugins of a specific type
func (r *DefaultPluginRegistry) GetPluginCountByType(pluginType PluginType) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, plugin := range r.plugins {
		if plugin.Type() == pluginType {
			count++
		}
	}
	return count
}
