package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/progress"
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

The scan command recursively traverses directories and identifies targets
that match cleaning patterns for various technologies (Node.js, Python, Rust, etc.).
Results show the path, type, and size of each cleanable target.

Flags:
  -d, --depth int           Maximum depth to scan (0 = unlimited)
  -H, --include-hidden      Include hidden files and directories
      --dry-run             Perform scan without making any changes

Examples:
  # Scan current directory
  rosia scan .

  # Scan specific directory
  rosia scan ~/projects

  # Scan multiple directories
  rosia scan ~/projects/app1 ~/projects/app2

  # Limit scan depth to 3 levels
  rosia scan . --depth 3

  # Include hidden files and directories
  rosia scan ~/projects --include-hidden

  # Dry run mode (no changes)
  rosia scan . --dry-run

Tips:
  • Use --depth to limit scanning in large directory trees
  • Combine with 'clean' command: rosia scan . && rosia clean .
  • Use --verbose flag for detailed logging`,
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

	// Use global configuration and profile loader
	cfg := GetGlobalConfig()
	profileLoader := GetGlobalProfileLoader()

	if profileLoader == nil {
		logger.Error("Profile loader not initialized")
		return fmt.Errorf("profile loader not initialized")
	}

	logger.Debug("Using %d profile(s)", len(profileLoader.GetProfiles()))

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

	// Perform scan with progress
	logger.Info("Scanning %d path(s)...", len(scanPaths))

	// Use async scan with progress bar
	targetChan, errorChan := scan.ScanAsync(ctx, scanPaths, opts)

	// Collect targets with progress indication
	targets := collectTargetsWithProgress(targetChan, errorChan)

	// Display results
	displayScanResults(targets)

	return nil
}

func collectTargetsWithProgress(targetChan <-chan types.Target, errorChan <-chan error) []types.Target {
	targets := make([]types.Target, 0)

	// Create a simple progress indicator
	fmt.Println("Scanning directories...")
	bar := progress.NewSimpleBar(100, "Progress", os.Stdout)

	targetCount := 0
	errorCount := 0
	done := false

	for !done {
		select {
		case target, ok := <-targetChan:
			if !ok {
				targetChan = nil
				if errorChan == nil {
					done = true
				}
				continue
			}
			targets = append(targets, target)
			targetCount++

			// Update progress bar label with current count
			bar.SetLabel(fmt.Sprintf("Found %d targets", targetCount))
			bar.IncrementBy(1)

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				if targetChan == nil {
					done = true
				}
				continue
			}
			if err != nil {
				logger.Warn("Scan error: %v", err)
				errorCount++
			}
		}
	}

	bar.Finish()

	if errorCount > 0 {
		logger.Warn("Completed with %d error(s)", errorCount)
	}

	return targets
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
