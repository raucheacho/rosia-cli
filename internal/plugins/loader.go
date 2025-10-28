package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/types"
)

// Loader handles loading Go plugins from .so files
type Loader struct{}

// NewLoader creates a new plugin loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadAll loads all plugins from the specified directory
func (l *Loader) LoadAll(dir string) ([]Plugin, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Debug("Plugin directory does not exist: %s", dir)
		return []Plugin{}, nil
	}

	// Find all .so files in the directory
	soFiles, err := filepath.Glob(filepath.Join(dir, "*.so"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob plugin directory: %w", err)
	}

	if len(soFiles) == 0 {
		logger.Debug("No plugin files found in %s", dir)
		return []Plugin{}, nil
	}

	plugins := make([]Plugin, 0, len(soFiles))

	// Load each plugin file
	for _, soFile := range soFiles {
		logger.Debug("Loading plugin from: %s", soFile)

		plugin, err := l.Load(soFile)
		if err != nil {
			logger.Warn("Failed to load plugin %s: %v", soFile, err)
			// Continue loading other plugins
			continue
		}

		plugins = append(plugins, plugin)
		logger.Info("Successfully loaded plugin: %s (version %s)", plugin.Name(), plugin.Version())
	}

	return plugins, nil
}

// Load loads a single plugin from the specified .so file
func (l *Loader) Load(path string) (Plugin, error) {
	pluginName := filepath.Base(path)

	// Open the plugin file
	p, err := plugin.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.ErrPluginLoadFailed{
				PluginName: pluginName,
				Reason:     types.ErrPathNotFound{Path: path},
			}
		}
		if os.IsPermission(err) {
			return nil, types.ErrPluginLoadFailed{
				PluginName: pluginName,
				Reason:     types.ErrPermissionDenied{Path: path},
			}
		}
		return nil, types.ErrPluginLoadFailed{
			PluginName: pluginName,
			Reason:     fmt.Errorf("failed to open plugin file: %w", err),
		}
	}

	// Look up the Plugin symbol
	symPlugin, err := p.Lookup("Plugin")
	if err != nil {
		return nil, types.ErrPluginLoadFailed{
			PluginName: pluginName,
			Reason:     fmt.Errorf("plugin does not export 'Plugin' symbol: %w", err),
		}
	}

	// Type assert to Plugin interface
	pluginInstance, ok := symPlugin.(Plugin)
	if !ok {
		return nil, types.ErrPluginLoadFailed{
			PluginName: pluginName,
			Reason:     fmt.Errorf("exported 'Plugin' symbol does not implement Plugin interface"),
		}
	}

	// Validate plugin
	if err := l.validate(pluginInstance); err != nil {
		return nil, types.ErrPluginLoadFailed{
			PluginName: pluginInstance.Name(),
			Reason:     fmt.Errorf("plugin validation failed: %w", err),
		}
	}

	return pluginInstance, nil
}

// validate checks if a plugin is valid
func (l *Loader) validate(p Plugin) error {
	if p.Name() == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if p.Version() == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	return nil
}
