package cmd

import (
	"fmt"

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
	Use:   "restore [item-name]",
	Short: "Restore a trashed item to its original location",
	Long: `Restore a previously trashed item back to its original location.

When you clean files with rosia, they are moved to ~/.rosia/trash instead
of being permanently deleted. This command allows you to restore those files
if you change your mind or accidentally deleted something important.

Flags:
  -l, --list                List all trashed items with their names
      --all                 Restore all trashed items

Examples:
  # List all trashed items
  rosia restore --list

  # Restore a specific item by name
  rosia restore 20250428_143022_node_modules_FROM_myproject_12345678

  # Restore all trashed items
  rosia restore --all

Trash Name Format:
  Items follow the format: YYYYMMDD_HHMMSS_basename_FROM_parent_timestamp
  Example: 20250428_143022_node_modules_FROM_myapp_12345678

Tips:
  • Use --list to see available items before restoring
  • Trash items are automatically cleaned after retention period (default: 3 days)
  • Restoration may fail if the original path already exists`,
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

	// Require item name argument if not using --list or --all
	if len(args) == 0 {
		logger.Error("Item name is required")
		return fmt.Errorf("item name is required (use --list to see available items)")
	}

	itemName := args[0]
	logger.Debug("Restoring item: %s", itemName)

	// Restore the item
	if err := trashSystem.Restore(itemName); err != nil {
		logger.Error("Failed to restore item %s: %v", itemName, err)
		return fmt.Errorf("failed to restore item: %w", err)
	}

	fmt.Printf("✓ Successfully restored: %s\n", itemName)
	logger.Info("Successfully restored: %s", itemName)

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

	// Display each item
	for _, item := range items {
		name := item.Name
		if len(name) > 60 {
			name = name[:57] + "..."
		}
		fmt.Printf("  %s  (deleted: %s)\n", name, item.DeletedAt.Format("2006-01-02 15:04"))
	}

	fmt.Printf("\nTo restore an item, use: rosia restore <item-name>\n")

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
		fmt.Printf("Restoring: %s... ", item.Name)

		if err := trashSystem.Restore(item.Name); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			logger.Error("Failed to restore %s: %v", item.Name, err)
			errorCount++
		} else {
			fmt.Println("✓ Success")
			logger.Debug("Restored %s", item.Name)
			successCount++
		}
	}

	fmt.Printf("\nRestored %d item(s), %d error(s)\n", successCount, errorCount)
	logger.Info("Restore all completed: %d success, %d errors", successCount, errorCount)

	return nil
}
