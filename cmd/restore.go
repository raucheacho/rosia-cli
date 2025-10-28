package cmd

import (
	"fmt"
	"strings"

	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	restoreList bool
	restoreAll  bool
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore [trash-id]",
	Short: "Restore a trashed item to its original location",
	Long: `Restore a previously trashed item back to its original location.

When you clean files with rosia, they are moved to ~/.rosia/trash instead
of being permanently deleted. This command allows you to restore those files
if you change your mind or accidentally deleted something important.

Flags:
  -l, --list                List all trashed items with their IDs
      --all                 Restore all trashed items

Examples:
  # List all trashed items
  rosia restore --list

  # Restore a specific item by ID
  rosia restore 20250428_143022_node_modules

  # Restore all trashed items
  rosia restore --all

Trash ID Format:
  Trash IDs follow the format: YYYYMMDD_HHMMSS_<basename>
  Example: 20250428_143022_node_modules

Tips:
  • Use --list to see available items before restoring
  • Trash items are automatically cleaned after retention period (default: 3 days)
  • Original paths must be available for restoration
  • If path conflicts exist, restoration will fail with an error`,
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Restore-specific flags
	restoreCmd.Flags().BoolVarP(&restoreList, "list", "l", false, "list all trashed items")
	restoreCmd.Flags().BoolVar(&restoreAll, "all", false, "restore all trashed items")
}

func runRestore(cmd *cobra.Command, args []string) error {
	// Initialize trash system
	logger.Debug("Initializing trash system")
	trashSystem, err := trash.NewDefaultSystem()
	if err != nil {
		logger.Error("Failed to initialize trash system: %v", err)
		return fmt.Errorf("failed to initialize trash system: %w", err)
	}

	// Handle --list flag
	if restoreList {
		return listTrashedItems(trashSystem)
	}

	// Handle --all flag
	if restoreAll {
		return restoreAllItems(trashSystem)
	}

	// Require trash ID argument if not using --list or --all
	if len(args) == 0 {
		logger.Error("Trash ID is required")
		return fmt.Errorf("trash ID is required (use --list to see available items)")
	}

	trashID := args[0]
	logger.Debug("Restoring trash ID: %s", trashID)

	// Get metadata to show what we're restoring
	metadata, err := trashSystem.GetMetadata(trashID)
	if err != nil {
		logger.Error("Failed to get trash metadata for %s: %v", trashID, err)
		return fmt.Errorf("failed to get trash metadata: %w", err)
	}

	logger.Info("Restoring: %s (size: %s)", metadata.OriginalPath, formatSize(metadata.Size))

	// Restore the item
	if err := trashSystem.Restore(trashID); err != nil {
		logger.Error("Failed to restore item %s: %v", trashID, err)
		return fmt.Errorf("failed to restore item: %w", err)
	}

	fmt.Printf("✓ Successfully restored: %s\n", metadata.OriginalPath)
	logger.Info("Successfully restored: %s", metadata.OriginalPath)

	return nil
}

func listTrashedItems(trashSystem *trash.System) error {
	logger.Debug("Listing trashed items")
	items, err := trashSystem.List()
	if err != nil {
		logger.Error("Failed to list trashed items: %v", err)
		return fmt.Errorf("failed to list trashed items: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("No trashed items found.")
		return nil
	}

	fmt.Printf("\nTrash Directory: %s\n", trashSystem.GetTrashDir())
	fmt.Printf("Found %d trashed item(s):\n\n", len(items))

	// Display table header
	fmt.Printf("%-40s %-40s %-15s %-20s\n", "TRASH ID", "ORIGINAL PATH", "SIZE", "DELETED AT")
	fmt.Println(strings.Repeat("-", 120))

	// Calculate total size
	var totalSize int64

	// Display each item
	for _, item := range items {
		totalSize += item.Size

		id := item.ID
		if len(id) > 38 {
			id = id[:35] + "..."
		}

		path := item.OriginalPath
		if len(path) > 38 {
			path = "..." + path[len(path)-35:]
		}

		deletedAt := item.DeletedAt.Format("2006-01-02 15:04:05")

		fmt.Printf("%-40s %-40s %-15s %-20s\n",
			id,
			path,
			formatSize(item.Size),
			deletedAt,
		)
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("Total: %s across %d item(s)\n", formatSize(totalSize), len(items))
	fmt.Println("\nTo restore an item, use: rosia restore <trash-id>")

	return nil
}

func restoreAllItems(trashSystem *trash.System) error {
	logger.Debug("Restoring all trashed items")
	items, err := trashSystem.List()
	if err != nil {
		logger.Error("Failed to list trashed items: %v", err)
		return fmt.Errorf("failed to list trashed items: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("No trashed items found.")
		return nil
	}

	fmt.Printf("Restoring %d item(s)...\n\n", len(items))
	logger.Info("Restoring %d items", len(items))

	successCount := 0
	errorCount := 0

	for _, item := range items {
		fmt.Printf("Restoring: %s... ", item.OriginalPath)

		if err := trashSystem.Restore(item.ID); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			logger.Error("Failed to restore %s: %v", item.OriginalPath, err)
			errorCount++
		} else {
			fmt.Println("✓ Success")
			logger.Debug("Restored %s", item.OriginalPath)
			successCount++
		}
	}

	fmt.Printf("\nRestored %d item(s), %d error(s)\n", successCount, errorCount)
	logger.Info("Restore all completed: %d success, %d errors", successCount, errorCount)

	return nil
}
