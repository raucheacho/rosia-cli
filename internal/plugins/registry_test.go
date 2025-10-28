package plugins

import (
	"context"
	"testing"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// mockPlugin is a mock implementation of the Plugin interface for testing
type mockPlugin struct {
	name        string
	version     string
	description string
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Version() string {
	return m.version
}

func (m *mockPlugin) Description() string {
	return m.description
}

func (m *mockPlugin) Scan(ctx context.Context) ([]types.Target, error) {
	return []types.Target{}, nil
}

func (m *mockPlugin) Clean(ctx context.Context, targets []types.Target) error {
	return nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if registry.plugins == nil {
		t.Error("Registry plugins map is nil")
	}
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()

	plugin := &mockPlugin{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin",
	}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify plugin was registered
	retrieved, err := registry.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name() != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", retrieved.Name())
	}
}

func TestRegisterDuplicate(t *testing.T) {
	registry := NewRegistry()

	plugin := &mockPlugin{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin",
	}

	// Register first time
	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("First Register failed: %v", err)
	}

	// Try to register again
	err = registry.Register(plugin)
	if err == nil {
		t.Error("Expected error when registering duplicate plugin")
	}
}

func TestRegisterNil(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(nil)
	if err == nil {
		t.Error("Expected error when registering nil plugin")
	}
}

func TestGet(t *testing.T) {
	registry := NewRegistry()

	plugin := &mockPlugin{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin",
	}

	registry.Register(plugin)

	// Test successful get
	retrieved, err := registry.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name() != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", retrieved.Name())
	}

	// Test get non-existent plugin
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent plugin")
	}
}

func TestList(t *testing.T) {
	registry := NewRegistry()

	// Empty list
	plugins := registry.List()
	if len(plugins) != 0 {
		t.Errorf("Expected empty list, got %d plugins", len(plugins))
	}

	// Add plugins
	plugin1 := &mockPlugin{name: "plugin1", version: "1.0.0", description: "Plugin 1"}
	plugin2 := &mockPlugin{name: "plugin2", version: "2.0.0", description: "Plugin 2"}

	registry.Register(plugin1)
	registry.Register(plugin2)

	plugins = registry.List()
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestUnregister(t *testing.T) {
	registry := NewRegistry()

	plugin := &mockPlugin{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin",
	}

	registry.Register(plugin)

	// Verify plugin exists
	_, err := registry.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Unregister
	err = registry.Unregister("test-plugin")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	// Verify plugin no longer exists
	_, err = registry.Get("test-plugin")
	if err == nil {
		t.Error("Expected error when getting unregistered plugin")
	}
}

func TestUnregisterNonExistent(t *testing.T) {
	registry := NewRegistry()

	err := registry.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent plugin")
	}
}
