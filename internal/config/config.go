package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config represents user configuration
type Config struct {
	TrashRetentionDays int      `json:"trash_retention_days"`
	Profiles           []string `json:"profiles"`
	IgnorePaths        []string `json:"ignore_paths"`
	Plugins            []string `json:"plugins"`
	Concurrency        int      `json:"concurrency"`
	TelemetryEnabled   bool     `json:"telemetry_enabled"`
}

// Manager handles configuration loading and saving
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".rosiarc.json")
	return &Manager{
		configPath: configPath,
	}, nil
}

// NewManagerWithPath creates a new configuration manager with a custom path
func NewManagerWithPath(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load reads configuration from ~/.rosiarc.json
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return m.GetDefault(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes configuration to ~/.rosiarc.json
func (m *Manager) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefault returns the default configuration
func (m *Manager) GetDefault() *Config {
	return &Config{
		TrashRetentionDays: 3,
		Profiles:           []string{"node", "python", "rust", "flutter", "go"},
		IgnorePaths:        []string{},
		Plugins:            []string{},
		Concurrency:        0, // 0 means auto-detect (NumCPU * 2)
		TelemetryEnabled:   false,
	}
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// Validate validates the configuration and applies defaults
func (m *Manager) Validate(config *Config) error {
	// Validate retention days > 0
	if config.TrashRetentionDays <= 0 {
		return fmt.Errorf("trash_retention_days must be greater than 0, got %d", config.TrashRetentionDays)
	}

	// Validate ignore paths are absolute
	for _, path := range config.IgnorePaths {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("ignore path must be absolute: %s", path)
		}
	}

	// Set concurrency to NumCPU * 2 if 0
	if config.Concurrency == 0 {
		config.Concurrency = runtime.NumCPU() * 2
	}

	// Validate concurrency is positive
	if config.Concurrency < 0 {
		return fmt.Errorf("concurrency must be non-negative, got %d", config.Concurrency)
	}

	return nil
}

// LoadAndValidate loads the configuration and validates it
func (m *Manager) LoadAndValidate() (*Config, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	if err := m.Validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}
