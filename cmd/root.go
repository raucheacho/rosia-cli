package cmd

import (
	"fmt"
	"strings"

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
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "rosia",
	Short: "Clean development dependencies and caches",
	Long: `Rosia is a universal, fast, and secure CLI tool for cleaning 
development dependencies, builds, and caches across multiple project types.

It helps developers reclaim disk space by safely removing cleanable files
like node_modules, target/, build/, and other technology-specific artifacts.`,
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

	// Set up logger based on verbose flag
	cobra.OnInitialize(initLogger)

	// Add version command
	rootCmd.AddCommand(versionCmd)
}

// initLogger initializes the logger with the verbose flag
func initLogger() {
	logger.SetVerbose(verbose)
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
