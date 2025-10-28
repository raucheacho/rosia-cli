package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/raucheacho/rosia-cli/internal/cleaner"
	"github.com/raucheacho/rosia-cli/internal/profiles"
	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/raucheacho/rosia-cli/internal/ui"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui [paths...]",
	Short: "Launch interactive TUI for cleaning",
	Long: `Launch an interactive terminal user interface (TUI) for selecting and cleaning targets.

The TUI provides a visual, keyboard-driven interface for managing cleanable
files. It scans directories, displays results in a list, and allows you to
select which targets to clean using keyboard navigation.

Features:
  • Visual list of all cleanable targets with sizes
  • Keyboard navigation (↑/↓ arrows)
  • Individual selection (Space)
  • Batch operations (a: select all, n: deselect all)
  • Real-time scan progress
  • Confirmation dialog before cleaning
  • Post-clean summary

Keyboard Controls:
  ↑/↓         Navigate up/down
  Space       Toggle selection
  a           Select all targets
  n           Deselect all targets
  Enter       Confirm and clean selected
  q           Quit without cleaning

Examples:
  # Launch TUI for current directory
  rosia ui

  # Launch TUI for specific directory
  rosia ui ~/projects

  # Launch TUI for multiple directories
  rosia ui ~/projects/app1 ~/projects/app2

Tips:
  • Use 'a' to quickly select all targets
  • Review total size before confirming
  • Files are moved to trash and can be restored
  • Press 'q' at any time to quit safely`,
	RunE: runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine scan paths
	scanPaths := args
	if len(scanPaths) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		scanPaths = []string{cwd}
	}

	// Validate paths
	for _, path := range scanPaths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path %s is not a directory", path)
		}
	}

	// Load profiles
	profileLoader := profiles.NewLoader()
	_, err := profileLoader.LoadAll("profiles")
	if err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Initialize scanner
	scannerInstance := scanner.NewScanner(profileLoader)

	// Initialize trash system
	trashSystem, err := trash.NewDefaultSystem()
	if err != nil {
		return fmt.Errorf("failed to initialize trash system: %w", err)
	}

	// Initialize cleaner
	cleanerInstance := cleaner.New(trashSystem)

	// Run TUI
	logger.Debug("Starting TUI for paths: %v", scanPaths)
	if err := ui.Run(ctx, scannerInstance, cleanerInstance, scanPaths); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
