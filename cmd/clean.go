package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raucheacho/rosia-cli/internal/cleaner"
	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/internal/telemetry"
	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cleanYes           bool
	cleanNoTrash       bool
	cleanRescan        bool
	cleanDepth         int
	cleanIncludeHidden bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean [paths...]",
	Short: "Clean detected targets by moving them to trash",
	Long: `Clean one or more directories by removing cleanable files and caches.
Files are moved to trash by default for safety and can be restored later.

Examples:
  rosia clean .
  rosia clean ~/projects --yes
  rosia clean . --no-trash
  rosia clean ~/projects --rescan --depth 3`,
	Args: cobra.MinimumNArgs(1),
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	// Clean-specific flags
	cleanCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "skip confirmation prompt")
	cleanCmd.Flags().BoolVar(&cleanNoTrash, "no-trash", false, "delete directly without moving to trash")
	cleanCmd.Flags().BoolVar(&cleanRescan, "rescan", false, "rescan directories before cleaning")
	cleanCmd.Flags().IntVarP(&cleanDepth, "depth", "d", 0, "maximum depth to scan (0 = unlimited)")
	cleanCmd.Flags().BoolVarP(&cleanIncludeHidden, "include-hidden", "H", false, "include hidden files and directories")
}

func runClean(cmd *cobra.Command, args []string) error {
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

	// Initialize trash system
	logger.Debug("Initializing trash system")
	trashSystem, err := trash.NewDefaultSystem()
	if err != nil {
		logger.Error("Failed to initialize trash system: %v", err)
		return fmt.Errorf("failed to initialize trash system: %w", err)
	}

	// Load profiles
	logger.Debug("Loading profiles")
	profileLoader, err := loadProfiles(cfg)
	if err != nil {
		logger.Error("Failed to load profiles: %v", err)
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Create scanner
	scan := scanner.NewScanner(profileLoader)

	// Initialize telemetry if enabled
	var telemetryStore telemetry.TelemetryStore
	if cfg.TelemetryEnabled {
		statsPath, err := getTelemetryStatsPath()
		if err == nil {
			if store, err := initTelemetryStore(statsPath); err == nil {
				telemetryStore = store
				scan.SetTelemetryStore(store)
				logger.Debug("Telemetry enabled for scanner")
			}
		}
	}

	// Prepare scan options
	opts := scanner.ScanOptions{
		MaxDepth:      cleanDepth,
		IncludeHidden: cleanIncludeHidden,
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

	if len(targets) == 0 {
		fmt.Println("No cleanable targets found.")
		return nil
	}

	// Calculate total size
	var totalSize int64
	for _, target := range targets {
		totalSize += target.Size
	}

	// Display targets
	fmt.Printf("\nFound %d cleanable target(s):\n\n", len(targets))
	fmt.Printf("%-50s %-15s %-15s\n", "PATH", "TYPE", "SIZE")
	fmt.Println(strings.Repeat("-", 80))

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
	fmt.Printf("Total: %s across %d target(s)\n\n", formatSize(totalSize), len(targets))

	// Confirmation prompt (unless --yes flag is set)
	if !cleanYes {
		if !confirmClean(totalSize, len(targets)) {
			fmt.Println("Clean operation cancelled.")
			return nil
		}
	}

	// Create cleaner
	clean := cleaner.New(trashSystem)

	// Set telemetry store if enabled
	if telemetryStore != nil {
		clean.SetTelemetryStore(telemetryStore)
		logger.Debug("Telemetry enabled for cleaner")
	}

	// Prepare clean options
	cleanOpts := cleaner.CleanOptions{
		SkipConfirmation: cleanYes,
		UseTrash:         !cleanNoTrash,
		Concurrency:      cfg.Concurrency,
	}

	// Perform cleaning
	fmt.Println("\nCleaning targets...")
	logger.Info("Starting clean operation for %d targets", len(targets))
	report, err := clean.Clean(ctx, targets, cleanOpts)
	if err != nil {
		logger.Error("Clean failed: %v", err)
		return fmt.Errorf("clean failed: %w", err)
	}

	// Display report
	displayCleanReport(report)

	if len(report.Errors) > 0 {
		logger.Warn("Clean completed with %d errors", len(report.Errors))
		// Return error if all targets failed
		if report.FilesDeleted == 0 {
			return fmt.Errorf("clean failed: all targets failed to clean")
		}
		// Partial success - don't return error
	} else {
		logger.Info("Clean completed successfully")
	}

	return nil
}

func confirmClean(totalSize int64, targetCount int) bool {
	fmt.Printf("This will clean %s across %d target(s).\n", formatSize(totalSize), targetCount)
	if cleanNoTrash {
		fmt.Println("WARNING: Files will be permanently deleted (--no-trash is set).")
	} else {
		fmt.Println("Files will be moved to trash and can be restored later.")
	}
	fmt.Print("\nDo you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func displayCleanReport(report *types.CleanReport) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("CLEAN REPORT")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("Files Deleted:  %d\n", report.FilesDeleted)
	fmt.Printf("Space Reclaimed: %s\n", formatSize(report.TotalSize))
	fmt.Printf("Duration:       %s\n", report.Duration)

	if len(report.TrashedItems) > 0 {
		fmt.Printf("Trashed Items:  %d\n", len(report.TrashedItems))
		if verbose {
			fmt.Println("\nTrashed IDs:")
			for _, id := range report.TrashedItems {
				fmt.Printf("  - %s\n", id)
			}
		}
	}

	if len(report.Errors) > 0 {
		fmt.Printf("\nErrors:         %d\n", len(report.Errors))
		fmt.Println("\nFailed targets:")
		for _, cleanErr := range report.Errors {
			fmt.Printf("  - %s: %v\n", cleanErr.Target.Path, cleanErr.Error)
		}
	}

	fmt.Println(strings.Repeat("=", 80))

	if len(report.TrashedItems) > 0 && !cleanNoTrash {
		fmt.Println("\nTo restore a trashed item, use: rosia restore <trash-id>")
		fmt.Println("To list all trashed items, use: rosia restore --list")
	}
}
