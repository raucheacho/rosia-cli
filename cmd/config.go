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

Available subcommands:
  show  - Display current configuration
  set   - Set a configuration value
  reset - Reset configuration to defaults`,
}

// configShowCmd displays the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Display the current rosia configuration from ~/.rosiarc.json`,
	RunE:  runConfigShow,
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value in ~/.rosiarc.json

Available keys:
  trash_retention_days - Number of days to retain trashed items (integer)
  concurrency          - Number of concurrent operations (integer, 0 = auto)
  telemetry_enabled    - Enable telemetry (true/false)

Examples:
  rosia config set trash_retention_days 7
  rosia config set concurrency 4
  rosia config set telemetry_enabled false`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configResetCmd resets configuration to defaults
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long:  `Reset the rosia configuration to default values`,
	RunE:  runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr, err := loadConfigManager()
	if err != nil {
		return fmt.Errorf("failed to load config manager: %w", err)
	}

	cfg, err := configMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Display configuration
	fmt.Printf("Configuration file: %s\n\n", configMgr.GetConfigPath())

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

	// Load configuration
	configMgr, err := loadConfigManager()
	if err != nil {
		return fmt.Errorf("failed to load config manager: %w", err)
	}

	cfg, err := configMgr.Load()
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
	if err := configMgr.Validate(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := configMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Configuration updated: %s = %s\n", key, value)
	fmt.Printf("Configuration saved to: %s\n", configMgr.GetConfigPath())

	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	// Load configuration manager
	configMgr, err := loadConfigManager()
	if err != nil {
		return fmt.Errorf("failed to load config manager: %w", err)
	}

	// Get default configuration
	cfg := configMgr.GetDefault()

	// Save default configuration
	if err := configMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Configuration reset to defaults\n")
	fmt.Printf("Configuration saved to: %s\n", configMgr.GetConfigPath())

	// Display the default configuration
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format configuration: %w", err)
	}

	fmt.Println("\nDefault configuration:")
	fmt.Println(string(data))

	return nil
}
