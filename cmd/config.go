package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage rosia configuration",
	Long: `Manage rosia configuration settings.

Configuration is stored in ~/.rosiarc.json and controls various aspects
of rosia's behavior including trash retention, concurrency, and telemetry.

Available Subcommands:
  show  - Display current configuration
  set   - Set a configuration value
  reset - Reset configuration to defaults

Configuration File:
  Location: ~/.rosiarc.json

Examples:
  # Show current configuration
  rosia config show

  # Set trash retention to 7 days
  rosia config set trash_retention_days 7

  # Reset to defaults
  rosia config reset`,
}

// configShowCmd displays the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Display the current rosia configuration from ~/.rosiarc.json

Shows all configuration values in JSON format, including:
  • trash_retention_days: Days to keep items in trash
  • profiles: Enabled technology profiles
  • ignore_paths: Paths excluded from scanning
  • plugins: Enabled plugin names
  • concurrency: Worker pool size (0 = auto-detect)
  • telemetry_enabled: Anonymous statistics collection

Examples:
  # Display configuration
  rosia config show`,
	RunE: runConfigShow,
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value in ~/.rosiarc.json

Available Configuration Keys:
  trash_retention_days  Number of days to retain trashed items (integer > 0)
  concurrency           Number of concurrent operations (integer >= 0, 0 = auto)
  telemetry_enabled     Enable anonymous telemetry (true/false)
  profiles              Comma-separated list of enabled profiles
  ignore_paths          Comma-separated list of paths to ignore
  plugins               Comma-separated list of enabled plugins

Examples:
  # Set trash retention to 7 days
  rosia config set trash_retention_days 7

  # Set concurrency to 4 workers
  rosia config set concurrency 4

  # Enable telemetry
  rosia config set telemetry_enabled true

  # Set enabled profiles
  rosia config set profiles "node,python,rust"

  # Add ignore paths
  rosia config set ignore_paths "/tmp,/var"

Tips:
  • Use 0 for concurrency to auto-detect based on CPU cores
  • Telemetry is disabled by default and stored locally
  • Changes take effect immediately`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configResetCmd resets configuration to defaults
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long: `Reset the rosia configuration to default values

This command overwrites ~/.rosiarc.json with default settings:
  • trash_retention_days: 3
  • profiles: ["node", "python", "rust", "flutter", "go"]
  • ignore_paths: []
  • plugins: []
  • concurrency: 0 (auto-detect)
  • telemetry_enabled: false

Examples:
  # Reset configuration
  rosia config reset

Warning:
  This will overwrite your current configuration. Make a backup if needed.`,
	RunE: runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Use global configuration
	cfg := GetGlobalConfig()

	// Get config path from global config manager
	configPath := "~/.rosiarc.json"
	if globalConfigManager != nil {
		configPath = globalConfigManager.GetConfigPath()
	}

	// Display configuration
	fmt.Printf("Configuration file: %s\n\n", configPath)

	// Pretty print JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format configuration: %w", err)
	}

	fmt.Println(string(data))

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Use global configuration manager
	if globalConfigManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	cfg, err := globalConfigManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set the value based on key
	switch key {
	case "trash_retention_days":
		days, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid value for trash_retention_days: must be an integer")
		}
		if days <= 0 {
			return fmt.Errorf("trash_retention_days must be greater than 0")
		}
		cfg.TrashRetentionDays = days

	case "concurrency":
		concurrency, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid value for concurrency: must be an integer")
		}
		if concurrency < 0 {
			return fmt.Errorf("concurrency must be non-negative")
		}
		cfg.Concurrency = concurrency

	case "telemetry_enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for telemetry_enabled: must be true or false")
		}
		cfg.TelemetryEnabled = enabled

	case "profiles":
		// Parse comma-separated list
		profiles := strings.Split(value, ",")
		for i := range profiles {
			profiles[i] = strings.TrimSpace(profiles[i])
		}
		cfg.Profiles = profiles

	case "ignore_paths":
		// Parse comma-separated list
		paths := strings.Split(value, ",")
		for i := range paths {
			paths[i] = strings.TrimSpace(paths[i])
		}
		cfg.IgnorePaths = paths

	case "plugins":
		// Parse comma-separated list
		plugins := strings.Split(value, ",")
		for i := range plugins {
			plugins[i] = strings.TrimSpace(plugins[i])
		}
		cfg.Plugins = plugins

	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Validate configuration
	if err := globalConfigManager.Validate(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := globalConfigManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Configuration updated: %s = %s\n", key, value)
	fmt.Printf("Configuration saved to: %s\n", globalConfigManager.GetConfigPath())

	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	// Use global configuration manager
	if globalConfigManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	// Get default configuration
	cfg := globalConfigManager.GetDefault()

	// Save default configuration
	if err := globalConfigManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Configuration reset to defaults\n")
	fmt.Printf("Configuration saved to: %s\n", globalConfigManager.GetConfigPath())

	// Display the default configuration
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format configuration: %w", err)
	}

	fmt.Println("\nDefault configuration:")
	fmt.Println(string(data))

	return nil
}
