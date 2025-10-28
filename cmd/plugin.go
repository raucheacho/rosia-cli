package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/raucheacho/rosia-cli/internal/plugins"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long: `Manage Rosia plugins for extended functionality.

Plugins extend Rosia's capabilities by adding support for additional tools
and technologies not covered by default profiles. Plugins can be written in
Go or any language that supports JSON-RPC communication.

Available Subcommands:
  list        List all loaded plugins
  info        Show detailed information about a plugin

Plugin Directory:
  Plugins are loaded from: ~/.rosia/plugins/

Examples:
  # List all loaded plugins
  rosia plugin list

  # Show details about a specific plugin
  rosia plugin info rosia-docker`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all loaded plugins",
	Long: `Display a list of all currently loaded plugins with their versions.

This command scans the plugin directory (~/.rosia/plugins/) and displays
information about each successfully loaded plugin.

Examples:
  # List all plugins
  rosia plugin list

Output includes:
  • Plugin name
  • Version
  • Description`,
	RunE: runPluginList,
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show detailed information about a plugin",
	Long: `Display detailed information about a specific plugin.

Shows comprehensive information including name, version, description,
and any additional metadata provided by the plugin.

Examples:
  # Show info for docker plugin
  rosia plugin info rosia-docker

  # Show info for xcode plugin
  rosia plugin info rosia-xcode`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginInfo,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
}

// runPluginList lists all loaded plugins
func runPluginList(cmd *cobra.Command, args []string) error {
	// Get plugin directory
	pluginDir, err := getPluginDirectory()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	// Create plugin registry and load plugins
	registry := plugins.NewRegistry()
	if err := registry.LoadAll(pluginDir); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Get all plugins
	allPlugins := registry.List()

	if len(allPlugins) == 0 {
		logger.Info("No plugins loaded")
		logger.Info("Plugin directory: %s", pluginDir)
		return nil
	}

	// Display plugins in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t-----------")

	for _, plugin := range allPlugins {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			plugin.Name(),
			plugin.Version(),
			truncateString(plugin.Description(), 50),
		)
	}

	w.Flush()

	logger.Info("\nTotal plugins: %d", len(allPlugins))
	logger.Info("Plugin directory: %s", pluginDir)

	return nil
}

// runPluginInfo displays detailed information about a specific plugin
func runPluginInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	// Get plugin directory
	pluginDir, err := getPluginDirectory()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	// Create plugin registry and load plugins
	registry := plugins.NewRegistry()
	if err := registry.LoadAll(pluginDir); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Get the specific plugin
	plugin, err := registry.Get(pluginName)
	if err != nil {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}

	// Display plugin information
	fmt.Printf("Plugin: %s\n", plugin.Name())
	fmt.Printf("Version: %s\n", plugin.Version())
	fmt.Printf("Description: %s\n", plugin.Description())

	return nil
}

// getPluginDirectory returns the plugin directory path
func getPluginDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	pluginDir := filepath.Join(homeDir, ".rosia", "plugins")
	return pluginDir, nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
