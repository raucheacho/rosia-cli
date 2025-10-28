package plugins

import (
	"fmt"
	"sync"

	"github.com/raucheacho/rosia-cli/pkg/logger"
)

// PluginRegistry manages loaded plugins
type PluginRegistry interface {
	// Register adds a plugin to the registry
	Register(plugin Plugin) error

	// LoadAll loads all plugins from the specified directory
	LoadAll(dir string) error

	// Get retrieves a plugin by name
	Get(name string) (Plugin, error)

	// List returns all registered plugins
	List() []Plugin

	// Unregister removes a plugin from the registry
	Unregister(name string) error
}

// Registry is the default implementation of PluginRegistry
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
	loader  *Loader
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		loader:  NewLoader(),
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if plugin already exists
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already registered", name)
	}

	r.plugins[name] = plugin
	logger.Debug("Registered plugin: %s (version %s)", name, plugin.Version())
	return nil
}

// LoadAll loads all plugins from the specified directory
func (r *Registry) LoadAll(dir string) error {
	logger.Debug("Loading plugins from directory: %s", dir)

	plugins, err := r.loader.LoadAll(dir)
	if err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Register all loaded plugins
	for _, plugin := range plugins {
		if err := r.Register(plugin); err != nil {
			logger.Warn("Failed to register plugin %s: %v", plugin.Name(), err)
			// Continue loading other plugins
			continue
		}
	}

	logger.Info("Loaded %d plugins from %s", len(plugins), dir)
	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	delete(r.plugins, name)
	logger.Debug("Unregistered plugin: %s", name)
	return nil
}
