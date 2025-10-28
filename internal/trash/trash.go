// Package trash provides trash system functionality for safe file deletion.
//
// The trash system moves deleted files to a temporary location (~/.rosia/trash/)
// before permanent removal, enabling restoration if needed. It maintains metadata
// for each trashed item and supports automatic cleanup based on retention periods.
//
// Example usage:
//
//	system, err := trash.NewSystem("~/.rosia/trash")
//	id, err := system.Move(target)
//	// Later, if needed:
//	err = system.Restore(id)
package trash

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// System manages the trash directory and operations.
//
// The System handles moving files to trash, restoring them, listing trashed items,
// and automatic cleanup of old items based on retention policies.
type System struct {
	trashDir string
}

// NewSystem creates a new trash system with the specified trash directory
func NewSystem(trashDir string) (*System, error) {
	// Ensure trash directory exists
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create trash directory: %w", err)
	}

	return &System{
		trashDir: trashDir,
	}, nil
}

// NewDefaultSystem creates a new trash system with the default location
// Uses platform-specific paths (XDG on Linux, ~/Library on macOS, %LOCALAPPDATA% on Windows)
func NewDefaultSystem() (*System, error) {
	trashDir, err := getDefaultTrashDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get default trash directory: %w", err)
	}
	return NewSystem(trashDir)
}

// getDefaultTrashDir returns the platform-specific default trash directory
func getDefaultTrashDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// For backward compatibility, keep using ~/.rosia/trash
	// In the future, this could use fsutils.GetTrashDir() for platform-specific paths
	trashDir := filepath.Join(homeDir, ".rosia", "trash")
	return trashDir, nil
}

// Move relocates a target to the trash with a timestamp-based ID
func (s *System) Move(target types.Target) (string, error) {
	// Generate unique ID: YYYYMMDD_HHMMSS_<basename>
	timestamp := time.Now().Format("20060102_150405")
	basename := filepath.Base(target.Path)
	id := fmt.Sprintf("%s_%s", timestamp, basename)

	// Create trash item directory
	itemDir := filepath.Join(s.trashDir, id)
	if err := os.MkdirAll(itemDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create trash item directory: %w", err)
	}

	// Create metadata
	metadata := types.TrashMetadata{
		ID:           id,
		OriginalPath: target.Path,
		Size:         target.Size,
		DeletedAt:    time.Now(),
		ProfileName:  target.ProfileName,
	}

	// Write metadata.json
	metadataPath := filepath.Join(itemDir, "metadata.json")
	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return "", fmt.Errorf("failed to write metadata: %w", err)
	}

	// Move the actual content
	contentPath := filepath.Join(itemDir, "content")
	if err := os.Rename(target.Path, contentPath); err != nil {
		// Clean up metadata if move fails
		os.RemoveAll(itemDir)
		return "", fmt.Errorf("failed to move target to trash: %w", err)
	}

	return id, nil
}

// Restore moves an item back to its original location
func (s *System) Restore(id string) error {
	// Get metadata to find original path
	metadata, err := s.GetMetadata(id)
	if err != nil {
		return fmt.Errorf("failed to get metadata for trash item %s: %w", id, err)
	}

	// Check if original path already exists (conflict)
	if _, err := os.Stat(metadata.OriginalPath); err == nil {
		return fmt.Errorf("cannot restore trash item %s: path already exists: %s", id, metadata.OriginalPath)
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(metadata.OriginalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		if os.IsPermission(err) {
			return types.ErrPermissionDenied{Path: parentDir}
		}
		return fmt.Errorf("failed to create parent directory %s for restore: %w", parentDir, err)
	}

	// Move content back to original location
	itemDir := filepath.Join(s.trashDir, id)
	contentPath := filepath.Join(itemDir, "content")

	if err := os.Rename(contentPath, metadata.OriginalPath); err != nil {
		if os.IsPermission(err) {
			return types.ErrPermissionDenied{Path: metadata.OriginalPath}
		}
		return fmt.Errorf("failed to restore item %s to %s: %w", id, metadata.OriginalPath, err)
	}

	// Remove trash item directory
	if err := os.RemoveAll(itemDir); err != nil {
		// Log warning but don't fail - the item was restored successfully
		fmt.Fprintf(os.Stderr, "warning: failed to clean up trash directory %s: %v\n", itemDir, err)
	}

	return nil
}

// GetMetadata reads and returns the metadata for a trashed item
func (s *System) GetMetadata(id string) (*types.TrashMetadata, error) {
	metadataPath := filepath.Join(s.trashDir, id, "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.ErrPathNotFound{Path: metadataPath}
		}
		if os.IsPermission(err) {
			return nil, types.ErrPermissionDenied{Path: metadataPath}
		}
		return nil, fmt.Errorf("failed to read metadata for trash item %s: %w", id, err)
	}

	var metadata types.TrashMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata for trash item %s: %w", id, err)
	}

	return &metadata, nil
}

// List returns all trashed items
func (s *System) List() ([]types.TrashItem, error) {
	entries, err := os.ReadDir(s.trashDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.TrashItem{}, nil
		}
		return nil, fmt.Errorf("failed to read trash directory: %w", err)
	}

	var items []types.TrashItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		metadata, err := s.GetMetadata(id)
		if err != nil {
			// Skip items with invalid metadata
			fmt.Fprintf(os.Stderr, "warning: skipping item with invalid metadata: %s: %v\n", id, err)
			continue
		}

		items = append(items, types.TrashItem{
			ID:           metadata.ID,
			OriginalPath: metadata.OriginalPath,
			Size:         metadata.Size,
			DeletedAt:    metadata.DeletedAt,
			TrashPath:    filepath.Join(s.trashDir, id),
		})
	}

	return items, nil
}

// Clean removes trashed items older than the specified retention period
func (s *System) Clean(retentionPeriod time.Duration) error {
	items, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list trash items: %w", err)
	}

	cutoffTime := time.Now().Add(-retentionPeriod)
	var errors []error

	for _, item := range items {
		if item.DeletedAt.Before(cutoffTime) {
			itemDir := filepath.Join(s.trashDir, item.ID)
			if err := os.RemoveAll(itemDir); err != nil {
				errors = append(errors, fmt.Errorf("failed to remove %s: %w", item.ID, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to clean some items: %v", errors)
	}

	return nil
}

// GetTrashDir returns the trash directory path
func (s *System) GetTrashDir() string {
	return s.trashDir
}
