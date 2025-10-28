package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raucheacho/rosia-cli/internal/config"
	"github.com/raucheacho/rosia-cli/internal/profiles"
	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	scanDepth         int
	scanIncludeHidden bool
	scanDryRun        bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [paths...]",
	Short: "Scan directories for cleanable files and caches",
	Long: `Scan one or more directories to identify cleanable files and caches
based on loaded technology profiles.

Examples:
  rosia scan .
  rosia scan ~/projects
  rosia scan ~/projects/app1 ~/projects/app2
  rosia scan . --depth 3 --include-hidden`,
	Args: cobra.MinimumNArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Scan-specific flags
	scanCmd.Flags().IntVarP(&scanDepth, "depth", "d", 0, "maximum depth to scan (0 = unlimited)")
	scanCmd.Flags().BoolVarP(&scanIncludeHidden, "include-hidden", "H", false, "include hidden files and directories")
	scanCmd.Flags().BoolVar(&scanDryRun, "dry-run", false, "perform scan without making any changes")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	logger.Debug("Loading configuration")
	configMgr, err := loadConfigManager()
	if err != nil {
		logger.Error("Failed to load config manager: %v", err)
		return fmt.Errorf("failed to load config manager: %w", err)
	}

	cfg, err := configMgr.LoadAndValidate()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Load profiles
	logger.Debug("Loading profiles")
	profileLoader, err := loadProfiles(cfg)
	if err != nil {
		logger.Error("Failed to load profiles: %v", err)
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	logger.Info("Loaded %d profiles", len(profileLoader.GetProfiles()))

	// Create scanner
	scan := scanner.NewScanner(profileLoader)

	// Initialize telemetry if enabled
	if cfg.TelemetryEnabled {
		statsPath, err := getTelemetryStatsPath()
		if err == nil {
			if store, err := initTelemetryStore(statsPath); err == nil {
				scan.SetTelemetryStore(store)
				logger.Debug("Telemetry enabled for scanner")
			}
		}
	}

	// Prepare scan options
	opts := scanner.ScanOptions{
		MaxDepth:      scanDepth,
		IncludeHidden: scanIncludeHidden,
		DryRun:        scanDryRun,
		IgnorePaths:   cfg.IgnorePaths,
		Concurrency:   cfg.Concurrency,
	}

	// Resolve and validate paths
	scanPaths := make([]string, 0, len(args))
	for _, path := range args {
		absPath, err := filepath.Abs(path)
		if err != nil {
			logger.Error("Failed to resolve path %s: %v", path, err)
			return fmt.Errorf("failed to resolve path %s: %w", path, err)
		}

		// Check if path exists
		if _, err := os.Stat(absPath); err != nil {
			logger.Error("Path does not exist: %s", path)
			return fmt.Errorf("path does not exist: %s", path)
		}

		scanPaths = append(scanPaths, absPath)
	}

	// Perform scan
	logger.Info("Scanning %d path(s)...", len(scanPaths))

	targets, err := scan.Scan(ctx, scanPaths, opts)
	if err != nil {
		logger.Error("Scan failed: %v", err)
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display results
	displayScanResults(targets)

	return nil
}

func displayScanResults(targets []types.Target) {
	if len(targets) == 0 {
		fmt.Println("No cleanable targets found.")
		return
	}

	fmt.Printf("\nFound %d cleanable target(s):\n\n", len(targets))

	// Calculate total size
	var totalSize int64
	for _, target := range targets {
		totalSize += target.Size
	}

	// Display table header
	fmt.Printf("%-50s %-15s %-15s\n", "PATH", "TYPE", "SIZE")
	fmt.Println(strings.Repeat("-", 80))

	// Display each target
	for _, target := range targets {
		path := target.Path
		if len(path) > 48 {
			path = "..." + path[len(path)-45:]
		}

		fmt.Printf("%-50s %-15s %-15s\n",
			path,
			target.ProfileName,
			formatSize(target.Size),
		)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %s across %d target(s)\n", formatSize(totalSize), len(targets))
	fmt.Println("\nTo clean these targets, run: rosia clean")
}

// loadConfigManager creates a config manager based on flags
func loadConfigManager() (*config.Manager, error) {
	if configPath != "" {
		return config.NewManagerWithPath(configPath), nil
	}
	return config.NewManager()
}

// loadProfiles loads profiles based on configuration
func loadProfiles(cfg *config.Config) (*profiles.Loader, error) {
	loader := profiles.NewLoader()

	// Determine profiles directory
	profilesDir := "profiles"
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		// Try relative to executable
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			profilesDir = filepath.Join(execDir, "profiles")
		}
	}

	// Load all profiles
	loadedProfiles, err := loader.LoadAll(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load profiles: %w", err)
	}

	if verbose {
		fmt.Printf("Loaded profiles from: %s\n", profilesDir)
		for _, p := range loadedProfiles {
			fmt.Printf("  - %s (v%s): %s\n", p.Name, p.Version, p.Description)
		}
	}

	return loader, nil
}
