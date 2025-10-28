package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Contains(t, manager.configPath, ".rosiarc.json")
}

func TestGetDefault(t *testing.T) {
	manager := &Manager{}
	config := manager.GetDefault()

	assert.Equal(t, 3, config.TrashRetentionDays)
	assert.Equal(t, []string{"node", "python", "rust", "flutter", "go"}, config.Profiles)
	assert.Equal(t, []string{}, config.IgnorePaths)
	assert.Equal(t, []string{}, config.Plugins)
	assert.Equal(t, 0, config.Concurrency)
	assert.False(t, config.TelemetryEnabled)
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".rosiarc.json")
	manager := NewManagerWithPath(configPath)

	// Create test config
	testConfig := &Config{
		TrashRetentionDays: 7,
		Profiles:           []string{"node", "python"},
		IgnorePaths:        []string{"/tmp", "/var"},
		Plugins:            []string{"docker"},
		Concurrency:        4,
		TelemetryEnabled:   true,
	}

	// Save config
	err := manager.Save(testConfig)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load config
	loadedConfig, err := manager.Load()
	require.NoError(t, err)

	// Verify loaded config matches saved config
	assert.Equal(t, testConfig.TrashRetentionDays, loadedConfig.TrashRetentionDays)
	assert.Equal(t, testConfig.Profiles, loadedConfig.Profiles)
	assert.Equal(t, testConfig.IgnorePaths, loadedConfig.IgnorePaths)
	assert.Equal(t, testConfig.Plugins, loadedConfig.Plugins)
	assert.Equal(t, testConfig.Concurrency, loadedConfig.Concurrency)
	assert.Equal(t, testConfig.TelemetryEnabled, loadedConfig.TelemetryEnabled)
}

func TestLoad_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")
	manager := NewManagerWithPath(configPath)

	// Load should return default config when file doesn't exist
	config, err := manager.Load()
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Should match default config
	defaultConfig := manager.GetDefault()
	assert.Equal(t, defaultConfig.TrashRetentionDays, config.TrashRetentionDays)
	assert.Equal(t, defaultConfig.Profiles, config.Profiles)
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".rosiarc.json")
	manager := NewManagerWithPath(configPath)

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	// Load should fail
	_, err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestValidate_RetentionDays(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		name          string
		retentionDays int
		expectError   bool
	}{
		{"valid positive", 3, false},
		{"valid large", 30, false},
		{"invalid zero", 0, true},
		{"invalid negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				TrashRetentionDays: tt.retentionDays,
				Concurrency:        1,
			}

			err := manager.Validate(config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_IgnorePaths(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		name        string
		ignorePaths []string
		expectError bool
	}{
		{"empty paths", []string{}, false},
		{"absolute paths", []string{"/tmp", "/var/log"}, false},
		{"relative path", []string{"relative/path"}, true},
		{"mixed paths", []string{"/tmp", "relative"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				TrashRetentionDays: 3,
				IgnorePaths:        tt.ignorePaths,
				Concurrency:        1,
			}

			err := manager.Validate(config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "ignore path must be absolute")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_Concurrency(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		name        string
		concurrency int
		expected    int
		expectError bool
	}{
		{"zero auto-detect", 0, runtime.NumCPU() * 2, false},
		{"positive value", 4, 4, false},
		{"negative value", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				TrashRetentionDays: 3,
				Concurrency:        tt.concurrency,
			}

			err := manager.Validate(config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config.Concurrency)
			}
		})
	}
}

func TestLoadAndValidate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".rosiarc.json")
	manager := NewManagerWithPath(configPath)

	// Save valid config
	validConfig := &Config{
		TrashRetentionDays: 5,
		Profiles:           []string{"node"},
		IgnorePaths:        []string{"/tmp"},
		Plugins:            []string{},
		Concurrency:        0,
		TelemetryEnabled:   false,
	}

	err := manager.Save(validConfig)
	require.NoError(t, err)

	// Load and validate
	config, err := manager.LoadAndValidate()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, runtime.NumCPU()*2, config.Concurrency) // Should be auto-set
}

func TestLoadAndValidate_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".rosiarc.json")
	manager := NewManagerWithPath(configPath)

	// Save invalid config (retention days = 0)
	invalidConfig := &Config{
		TrashRetentionDays: 0,
		Profiles:           []string{"node"},
		IgnorePaths:        []string{},
		Plugins:            []string{},
		Concurrency:        1,
		TelemetryEnabled:   false,
	}

	err := manager.Save(invalidConfig)
	require.NoError(t, err)

	// Load and validate should fail
	_, err = manager.LoadAndValidate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config validation failed")
}

func TestLoadAndValidate_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")
	manager := NewManagerWithPath(configPath)

	// Should return validated default config
	config, err := manager.LoadAndValidate()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 3, config.TrashRetentionDays)
	assert.Equal(t, runtime.NumCPU()*2, config.Concurrency)
}
