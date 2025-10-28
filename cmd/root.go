package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raucheacho/rosia-cli/internal/config"
	"github.com/raucheacho/rosia-cli/internal/plugins"
	"github.com/raucheacho/rosia-cli/internal/profiles"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose    bool
	configPath string

	// Build info (set via ldflags)
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Global components (initialized once)
	globalConfig         *config.Config
	globalConfigManager  *config.Manager
	globalProfileLoader  *profiles.Loader
	globalPluginRegistry plugins.PluginRegistry
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "rosia",
	Short: "Clean development dependencies and caches",
	Long: `Rosia is a universal, fast, and secure CLI tool for cleaning 
development dependencies, builds, and caches across multiple project types.

It helps developers reclaim disk space by safely removing cleanable files
like node_modules, target/, build/, and other technology-specific artifacts.

Supported Technologies:
  • Node.js (node_modules, dist, build, .next, coverage)
  • Python (venv, __pycache__, .pytest_cache, .tox)
  • Rust (target/)
  • Flutter (build/, .dart_tool/)
  • Go (vendor/, bin/)

Features:
  • Fast concurrent scanning with configurable worker pools
  • Safe deletion with trash system and restoration capability
  • Interactive TUI for visual selection
  • Extensible plugin system
  • Cross-platform support (Linux, macOS, Windows)

Common Workflows:
  1. Quick scan and clean:
     $ rosia scan ~/projects
     $ rosia clean ~/projects --yes

  2. Interactive mode:
     $ rosia ui ~/projects

  3. Restore accidentally deleted files:
     $ rosia restore <trash-id>

  4. View statistics:
     $ rosia stats

For more information, visit: https://github.com/raucheacho/rosia-cli`,
	SilenceUsage: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteWithExitCode runs the root command and returns appropriate exit code
func ExecuteWithExitCode() int {
	if err := Execute(); err != nil {
		// Check if it's a critical error
		if isCriticalError(err) {
			logger.Error("Critical error: %v", err)
			return 1
		}
		// Recoverable error
		logger.Warn("Command completed with errors: %v", err)
		return 0
	}
	return 0
}

// isCriticalError determines if an error should cause a non-zero exit code
func isCriticalError(err error) bool {
	if err == nil {
		return false
	}

	// Critical errors that should cause non-zero exit
	criticalPatterns := []string{
		"failed to load config",
		"failed to initialize",
		"scan failed",
		"clean failed",
		"permission denied",
		"path does not exist",
	}

	errMsg := err.Error()
	for _, pattern := range criticalPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path (default: ~/.rosiarc.json)")

	// Set up initialization hooks
	cobra.OnInitialize(initLogger, initComponents)

	// Add version command
	rootCmd.AddCommand(versionCmd)
}

// initLogger initializes the logger with the verbose flag
func initLogger() {
	logger.SetVerbose(verbose)
}

// initComponents initializes global components (config, profiles, plugins)
func initComponents() {
	// Initialize config manager
	var err error
	if configPath != "" {
		globalConfigManager = config.NewManagerWithPath(configPath)
		logger.Debug("Using custom config path: %s", configPath)
	} else {
		globalConfigManager, err = config.NewManager()
		if err != nil {
			logger.Warn("Failed to create config manager: %v", err)
			// Use default config - create a temporary manager to get defaults
			tempMgr := config.NewManagerWithPath("")
			globalConfig = tempMgr.GetDefault()
			return
		}
	}

	// Load and validate configuration
	globalConfig, err = globalConfigManager.LoadAndValidate()
	if err != nil {
		logger.Debug("Failed to load config, using defaults: %v", err)
		globalConfig = globalConfigManager.GetDefault()
	} else {
		logger.Debug("Configuration loaded successfully")
	}

	// Initialize profile loader
	globalProfileLoader = profiles.NewLoader()

	// Determine profiles directory
	profilesDir := findProfilesDirectory()

	// Load profiles
	loadedProfiles, err := globalProfileLoader.LoadAll(profilesDir)
	if err != nil {
		logger.Warn("Failed to load profiles: %v", err)
	} else {
		logger.Debug("Loaded %d profile(s) from %s", len(loadedProfiles), profilesDir)
		if verbose {
			for _, p := range loadedProfiles {
				logger.Debug("  - %s (v%s): %s", p.Name, p.Version, p.Description)
			}
		}
	}

	// Initialize plugin registry
	globalPluginRegistry = plugins.NewRegistry()

	// Load plugins if configured
	if len(globalConfig.Plugins) > 0 {
		pluginsDir := findPluginsDirectory()
		if pluginsDir != "" {
			err := globalPluginRegistry.LoadAll(pluginsDir)
			if err != nil {
				logger.Warn("Failed to load plugins: %v", err)
			} else {
				pluginList := globalPluginRegistry.List()
				logger.Debug("Loaded %d plugin(s)", len(pluginList))
				if verbose {
					for _, p := range pluginList {
						logger.Debug("  - %s (v%s): %s", p.Name(), p.Version(), p.Description())
					}
				}
			}
		}
	}
}

// findProfilesDirectory locates the profiles directory
func findProfilesDirectory() string {
	// Try current directory first
	if _, err := os.Stat("profiles"); err == nil {
		return "profiles"
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		profilesDir := filepath.Join(execDir, "profiles")
		if _, err := os.Stat(profilesDir); err == nil {
			return profilesDir
		}
	}

	// Try home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		profilesDir := filepath.Join(homeDir, ".rosia", "profiles")
		if _, err := os.Stat(profilesDir); err == nil {
			return profilesDir
		}
	}

	// Default to current directory
	return "profiles"
}

// findPluginsDirectory locates the plugins directory
func findPluginsDirectory() string {
	// Try home directory first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		pluginsDir := filepath.Join(homeDir, ".rosia", "plugins")
		if _, err := os.Stat(pluginsDir); err == nil {
			return pluginsDir
		}
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		pluginsDir := filepath.Join(execDir, "plugins")
		if _, err := os.Stat(pluginsDir); err == nil {
			return pluginsDir
		}
	}

	return ""
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *config.Config {
	if globalConfig == nil {
		// Return a default config if not initialized
		if globalConfigManager != nil {
			return globalConfigManager.GetDefault()
		}
		// Fallback to hardcoded defaults
		return &config.Config{
			TrashRetentionDays: 3,
			Profiles:           []string{"node", "python", "rust", "flutter", "go"},
			IgnorePaths:        []string{},
			Plugins:            []string{},
			Concurrency:        0,
			TelemetryEnabled:   false,
		}
	}
	return globalConfig
}

// GetGlobalProfileLoader returns the global profile loader
func GetGlobalProfileLoader() *profiles.Loader {
	return globalProfileLoader
}

// GetGlobalPluginRegistry returns the global plugin registry
func GetGlobalPluginRegistry() plugins.PluginRegistry {
	return globalPluginRegistry
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("rosia version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

// GetVerbose returns the verbose flag value
func GetVerbose() bool {
	return verbose
}

// GetConfigPath returns the config path flag value
func GetConfigPath() string {
	return configPath
}
