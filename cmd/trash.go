package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	trashCleanAll bool
)

// trashCmd represents the trash command
var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "Manage the trash directory",
	Long: `Manage the trash directory where cleaned files are stored.

The trash command provides subcommands to list and clean trashed items.
Items in trash can be restored using the 'restore' command.

Subcommands:
  list   - List all items in trash
  clean  - Remove old items from trash

Examples:
  # List all trashed items
  rosia trash list

  # Clean items older than retention period
  rosia trash clean

  # Clean all items (permanent deletion)
  rosia trash clean --all`,
}

// trashListCmd lists all trashed items
var trashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all items in trash",
	Long: `List all items currently stored in the trash directory.

Shows details including:
  • Trash ID (used for restoration)
  • Original path
  • Size
  • Deletion date

Examples:
  # List all trashed items
  rosia trash list

  # List with verbose output
  rosia trash list --verbose`,
	RunE: runTrashList,
}

// trashCleanCmd cleans old items from trash
var trashCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean old items from trash",
	Long: `Remove old items from the trash directory.

By default, removes items older than the retention period
(configured by trash_retention_days in ~/.rosiarc.json).

Use --all to remove all items immediately (permanent deletion).

Examples:
  # Clean items older than retention period
  rosia trash clean

  # Clean all items immediately
  rosia trash clean --all

Warning:
  This permanently deletes files. They cannot be restored.`,
	RunE: runTrashClean,
}

func init() {
	rootCmd.AddCommand(trashCmd)
	trashCmd.AddCommand(trashListCmd)
	trashCmd.AddCommand(trashCleanCmd)

	// Clean flags
	trashCleanCmd.Flags().BoolVar(&trashCleanAll, "all", false, "Remove all items regardless of retention period")
}

func runTrashList(cmd *cobra.Command, args []string) error {
	// Initialize trash system
	trashSystem, err := trash.NewDefaultSystem()
	if err != nil {
		return fmt.Errorf("failed to initialize trash system: %w", err)
	}

	// List items
	items, err := trashSystem.List()
	if err != nil {
		return fmt.Errorf("failed to list trash items: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("Trash is empty.")
		return nil
	}

	// Display items
	fmt.Printf("Trash Directory: %s\n", trashSystem.GetTrashDir())
	fmt.Printf("Found %d trashed item(s):\n\n", len(items))

	fmt.Printf("%-40s %-30s %-12s %-20s\n", "TRASH ID", "ORIGINAL PATH", "SIZE", "DELETED AT")
	fmt.Println(string(make([]byte, 110, 110)))
	for i := range make([]byte, 110) {
		_ = i
		fmt.Print("-")
	}
	fmt.Println()

	var totalSize int64
	for _, item := range items {
		trashPath := filepath.Join(trashSystem.GetTrashDir(), item.Name)
		info, err := os.Stat(trashPath)
		if err != nil {
			continue
		}

		size := info.Size()
		totalSize += size

		// Read original path if available
		pathFile := trashPath + ".path"
		originalPath := "unknown"
		if data, err := os.ReadFile(pathFile); err == nil {
			originalPath = string(data)
			if len(originalPath) > 28 {
				originalPath = "..." + originalPath[len(originalPath)-25:]
			}
		}

		fmt.Printf("%-40s %-30s %-12s %-20s\n",
			item.Name,
			originalPath,
			formatSize(size),
			item.DeletedAt.Format("2006-01-02 15:04"),
		)
	}

	for i := range make([]byte, 110) {
		_ = i
		fmt.Print("-")
	}
	fmt.Println()
	fmt.Printf("Total: %s across %d item(s)\n", formatSize(totalSize), len(items))
	fmt.Println("\nTo restore an item, use: rosia restore <trash-id>")

	return nil
}

func runTrashClean(cmd *cobra.Command, args []string) error {
	// Initialize trash system
	trashSystem, err := trash.NewDefaultSystem()
	if err != nil {
		return fmt.Errorf("failed to initialize trash system: %w", err)
	}

	// Get retention period from config
	cfg := GetGlobalConfig()
	retentionDays := cfg.TrashRetentionDays

	items, err := trashSystem.List()
	if err != nil {
		return fmt.Errorf("failed to list trash items: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("Trash is already empty.")
		return nil
	}

	var toDelete []trash.TrashItem
	var cutoff time.Time

	if trashCleanAll {
		// Delete all items
		toDelete = items
	} else {
		// Delete items older than retention period
		retention := time.Duration(retentionDays) * 24 * time.Hour
		cutoff = time.Now().Add(-retention)

		for _, item := range items {
			if item.DeletedAt.Before(cutoff) {
				toDelete = append(toDelete, item)
			}
		}
	}

	if len(toDelete) == 0 {
		fmt.Printf("No items to clean. All items are within the %d-day retention period.\n", retentionDays)
		return nil
	}

	// Confirmation if not --all
	if !trashCleanAll {
		fmt.Printf("This will permanently delete %d item(s) older than %d days.\n", len(toDelete), retentionDays)
		fmt.Print("Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete items
	deleted := 0
	var freed int64

	for _, item := range toDelete {
		trashPath := filepath.Join(trashSystem.GetTrashDir(), item.Name)

		// Get size before deleting
		if info, err := os.Stat(trashPath); err == nil {
			freed += info.Size()
		}

		// Remove item and sidecar file
		if err := os.RemoveAll(trashPath); err != nil {
			logger.Warn("Failed to delete %s: %v", item.Name, err)
			continue
		}
		os.Remove(trashPath + ".path")

		deleted++
	}

	fmt.Printf("✓ Removed %d item(s) from trash (%s freed)\n", deleted, formatSize(freed))

	return nil
}
