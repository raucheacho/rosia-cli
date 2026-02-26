// Package trash provides simple trash functionality for safe file deletion.
//
// The trash system moves deleted files to ~/.rosia/trash/ before permanent removal.
// Items are stored with descriptive names containing timestamp and original path.
package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// System manages the trash directory.
type System struct {
	trashDir string
}

// NewSystem creates a new trash system.
func NewSystem(trashDir string) (*System, error) {
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create trash directory: %w", err)
	}
	return &System{trashDir: trashDir}, nil
}

// NewDefaultSystem creates a trash system with default location.
func NewDefaultSystem() (*System, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return NewSystem(filepath.Join(homeDir, ".rosia", "trash"))
}

// Move relocates a target to trash.
// Returns the trash item name.
func (s *System) Move(target types.Target) (string, error) {
	// Create descriptive name: YYYYMMDD_HHMMSS_basename
	timestamp := time.Now().Format("20060102_150405")
	basename := filepath.Base(target.Path)
	itemName := fmt.Sprintf("%s_%s_%d", timestamp, basename, time.Now().UnixNano())
	
	// Replace problematic characters
	itemName = strings.ReplaceAll(itemName, "/", "_")
	itemName = strings.ReplaceAll(itemName, "\\", "_")
	itemName = strings.ReplaceAll(itemName, " ", "_")
	
	trashPath := filepath.Join(s.trashDir, itemName)
	
	if err := os.Rename(target.Path, trashPath); err != nil {
		return "", fmt.Errorf("failed to move to trash: %w", err)
	}
	
	// Store original path in a sidecar file
	pathFile := trashPath + ".path"
	if err := os.WriteFile(pathFile, []byte(target.Path), 0644); err != nil {
		// If we can't write the path file, try to move back and return error
		os.Rename(trashPath, target.Path)
		return "", fmt.Errorf("failed to store original path: %w", err)
	}
	
	return itemName, nil
}

// Restore moves an item back to its original location.
func (s *System) Restore(itemName string) error {
	trashPath := filepath.Join(s.trashDir, itemName)
	
	// Read original path from sidecar file
	pathFile := trashPath + ".path"
	originalPathBytes, err := os.ReadFile(pathFile)
	var originalPath string
	if err != nil {
		// Fallback: restore to current directory with original name (without timestamp prefix)
		originalPath = s.extractBaseName(itemName)
	} else {
		originalPath = string(originalPathBytes)
	}
	
	// Check if destination exists
	if _, err := os.Stat(originalPath); err == nil {
		return fmt.Errorf("cannot restore: path already exists: %s", originalPath)
	}
	
	// Ensure parent directory exists
	parentDir := filepath.Dir(originalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	
	// Move back
	if err := os.Rename(trashPath, originalPath); err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}
	
	// Clean up sidecar file
	os.Remove(pathFile)
	
	return nil
}

// List returns all trashed items.
func (s *System) List() ([]TrashItem, error) {
	entries, err := os.ReadDir(s.trashDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TrashItem{}, nil
		}
		return nil, err
	}
	
	var items []TrashItem
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			items = append(items, TrashItem{
				Name:      entry.Name(),
				DeletedAt: info.ModTime(),
			})
		}
	}
	
	return items, nil
}

// Clean removes items older than retention period.
func (s *System) Clean(retentionPeriod time.Duration) error {
	items, err := s.List()
	if err != nil {
		return err
	}
	
	cutoff := time.Now().Add(-retentionPeriod)
	for _, item := range items {
		if item.DeletedAt.Before(cutoff) {
			path := filepath.Join(s.trashDir, item.Name)
			os.RemoveAll(path)
			// Also remove sidecar file if exists
			os.Remove(path + ".path")
		}
	}
	
	return nil
}

// GetTrashDir returns the trash directory path.
func (s *System) GetTrashDir() string {
	return s.trashDir
}

// TrashItem represents a trashed item.
type TrashItem struct {
	Name      string
	DeletedAt time.Time
}

// extractBaseName extracts the original basename from trash item name.
// Format: YYYYMMDD_HHMMSS_basename_unixnano
func (s *System) extractBaseName(itemName string) string {
	parts := strings.Split(itemName, "_")
	if len(parts) >= 3 {
		// Return third part (basename)
		return parts[2]
	}
	return itemName
}
