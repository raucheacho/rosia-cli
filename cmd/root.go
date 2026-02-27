package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raucheacho/rosia-cli/internal/config"
	"github.com/raucheacho/rosia-cli/internal/profiles"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/types"
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
	globalConfig        *config.Config
	globalConfigManager *config.Manager
	globalProfileLoader *profiles.Loader
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
  • Cross-platform support (Linux, macOS, Windows)

Common Workflows:
  1. Quick scan and clean:
     $ rosia scan ~/projects
     $ rosia clean ~/projects --yes

  2. Interactive mode:
     $ rosia ui ~/projects

  3. Restore accidentally deleted files:
     $ rosia restore <trash-id>

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

}

// findProfilesDirectory locates the profiles directory
func findProfilesDirectory() string {
	// Try current directory first (development mode)
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

	// Use home directory - initialize default profiles if needed
	homeDir, err := os.UserHomeDir()
	if err == nil {
		rosiaDir := filepath.Join(homeDir, ".rosia")
		profilesFile := filepath.Join(rosiaDir, "profiles.json")
		
		// Create default profiles if file doesn't exist
		if _, err := os.Stat(profilesFile); os.IsNotExist(err) {
			if err := initDefaultProfiles(profilesFile); err != nil {
				logger.Debug("Failed to create default profiles: %v", err)
			}
		}
		return rosiaDir
	}

	// Fallback to current directory
	return "profiles"
}

// initDefaultProfiles creates the default profiles.json file
func initDefaultProfiles(profilesFile string) error {
	// Ensure directory exists
	dir := filepath.Dir(profilesFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .rosia directory: %w", err)
	}

	defaultProfiles := []types.Profile{
		{
			Name:        "Node.js",
			Version:     "1.0.0",
			Patterns:    []string{"node_modules", "dist", "build", ".next", ".cache", "coverage"},
			Detect:      []string{"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml"},
			Description: "Cleans Node.js project artifacts",
			Enabled:     true,
		},
		{
			Name:        "Python",
			Version:     "1.0.0",
			Patterns:    []string{"venv", "__pycache__", ".pytest_cache", ".tox", ".mypy_cache", ".ruff_cache"},
			Detect:      []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile"},
			Description: "Cleans Python virtual environments and caches",
			Enabled:     true,
		},
		{
			Name:        "Rust",
			Version:     "1.0.0",
			Patterns:    []string{"target"},
			Detect:      []string{"Cargo.toml", "Cargo.lock"},
			Description: "Cleans Rust build artifacts",
			Enabled:     true,
		},
		{
			Name:        "Go",
			Version:     "1.0.0",
			Patterns:    []string{"vendor", "bin"},
			Detect:      []string{"go.mod", "go.sum"},
			Description: "Cleans Go vendor and build directories",
			Enabled:     true,
		},
		{
			Name:        "Flutter",
			Version:     "1.0.0",
			Patterns:    []string{"build", ".dart_tool", ".flutter-plugins"},
			Detect:      []string{"pubspec.yaml", "pubspec.lock"},
			Description: "Cleans Flutter build artifacts",
			Enabled:     true,
		},
	}

	data, err := json.MarshalIndent(defaultProfiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(profilesFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write profiles file: %w", err)
	}

	logger.Debug("Created default profiles at %s", profilesFile)
	return nil
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *config.Config {
	if globalConfig == nil {
		if globalConfigManager != nil {
			return globalConfigManager.GetDefault()
		}
		return &config.Config{
			TrashRetentionDays: 3,
			IgnorePaths:        []string{},
		}
	}
	return globalConfig
}

// GetGlobalProfileLoader returns the global profile loader
func GetGlobalProfileLoader() *profiles.Loader {
	return globalProfileLoader
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
