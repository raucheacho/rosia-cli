package trash

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

func TestSystem_Move(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create a test file to move
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create target
	target := types.Target{
		Path:        testFile,
		Size:        int64(len(content)),
		ProfileName: "test",
	}

	// Move to trash
	id, err := sys.Move(target)
	if err != nil {
		t.Fatalf("failed to move to trash: %v", err)
	}

	// Verify original file is gone
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("original file still exists after move")
	}

	// Verify trash item exists
	metadata, err := sys.GetMetadata(id)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}

	if metadata.OriginalPath != testFile {
		t.Errorf("expected original path %s, got %s", testFile, metadata.OriginalPath)
	}

	if metadata.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), metadata.Size)
	}
}

func TestSystem_Restore(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create a test file to move
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create target and move to trash
	target := types.Target{
		Path:        testFile,
		Size:        int64(len(content)),
		ProfileName: "test",
	}

	id, err := sys.Move(target)
	if err != nil {
		t.Fatalf("failed to move to trash: %v", err)
	}

	// Restore from trash
	if err := sys.Restore(id); err != nil {
		t.Fatalf("failed to restore: %v", err)
	}

	// Verify file is restored
	restoredContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}

	if string(restoredContent) != string(content) {
		t.Errorf("expected content %s, got %s", content, restoredContent)
	}

	// Verify trash item is removed
	if _, err := sys.GetMetadata(id); err == nil {
		t.Error("trash item still exists after restore")
	}
}

func TestSystem_List(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create and move multiple test files
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		target := types.Target{
			Path:        testFile,
			Size:        7,
			ProfileName: "test",
		}

		if _, err := sys.Move(target); err != nil {
			t.Fatalf("failed to move to trash: %v", err)
		}
	}

	// List trash items
	items, err := sys.List()
	if err != nil {
		t.Fatalf("failed to list trash items: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestSystem_Clean(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create and move a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	target := types.Target{
		Path:        testFile,
		Size:        7,
		ProfileName: "test",
	}

	id, err := sys.Move(target)
	if err != nil {
		t.Fatalf("failed to move to trash: %v", err)
	}

	// Clean items older than 0 seconds (should remove all)
	if err := sys.Clean(0); err != nil {
		t.Fatalf("failed to clean trash: %v", err)
	}

	// Verify item is removed
	if _, err := sys.GetMetadata(id); err == nil {
		t.Error("trash item still exists after clean")
	}

	// Verify list is empty
	items, err := sys.List()
	if err != nil {
		t.Fatalf("failed to list trash items: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items after clean, got %d", len(items))
	}
}

func TestSystem_RestorePathConflict(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create a test file to move
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Move to trash
	target := types.Target{
		Path:        testFile,
		Size:        8,
		ProfileName: "test",
	}

	id, err := sys.Move(target)
	if err != nil {
		t.Fatalf("failed to move to trash: %v", err)
	}

	// Create a new file at the same location
	if err := os.WriteFile(testFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	// Try to restore - should fail due to conflict
	if err := sys.Restore(id); err == nil {
		t.Error("expected error when restoring to existing path, got nil")
	}
}

func TestSystem_CleanRetentionPeriod(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create and move a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	target := types.Target{
		Path:        testFile,
		Size:        7,
		ProfileName: "test",
	}

	if _, err := sys.Move(target); err != nil {
		t.Fatalf("failed to move to trash: %v", err)
	}

	// Clean items older than 1 hour (should not remove anything)
	if err := sys.Clean(1 * time.Hour); err != nil {
		t.Fatalf("failed to clean trash: %v", err)
	}

	// Verify item still exists
	items, err := sys.List()
	if err != nil {
		t.Fatalf("failed to list trash items: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item after clean with retention, got %d", len(items))
	}
}

func TestSystem_MoveDirectory(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create a directory with nested content
	testDir := filepath.Join(tmpDir, "test_dir")
	if err := os.MkdirAll(filepath.Join(testDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "subdir", "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Move directory to trash
	target := types.Target{
		Path:        testDir,
		Size:        100,
		ProfileName: "test",
		IsDirectory: true,
	}

	id, err := sys.Move(target)
	if err != nil {
		t.Fatalf("failed to move directory to trash: %v", err)
	}

	// Verify original directory is gone
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("original directory still exists after move")
	}

	// Verify trash item exists with nested content
	contentPath := filepath.Join(trashDir, id, "content")
	nestedFile := filepath.Join(contentPath, "subdir", "file.txt")
	if _, err := os.Stat(nestedFile); err != nil {
		t.Errorf("nested file not found in trash: %v", err)
	}

	// Restore and verify nested content
	if err := sys.Restore(id); err != nil {
		t.Fatalf("failed to restore: %v", err)
	}

	restoredFile := filepath.Join(testDir, "subdir", "file.txt")
	if _, err := os.Stat(restoredFile); err != nil {
		t.Errorf("nested file not found after restore: %v", err)
	}
}

func TestSystem_MetadataPersistence(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create and move a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	target := types.Target{
		Path:        testFile,
		Size:        7,
		ProfileName: "Node.js",
	}

	id, err := sys.Move(target)
	if err != nil {
		t.Fatalf("failed to move to trash: %v", err)
	}

	// Get metadata
	metadata, err := sys.GetMetadata(id)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}

	// Verify all metadata fields
	if metadata.ID != id {
		t.Errorf("expected ID %s, got %s", id, metadata.ID)
	}

	if metadata.OriginalPath != testFile {
		t.Errorf("expected original path %s, got %s", testFile, metadata.OriginalPath)
	}

	if metadata.Size != 7 {
		t.Errorf("expected size 7, got %d", metadata.Size)
	}

	if metadata.ProfileName != "Node.js" {
		t.Errorf("expected profile name 'Node.js', got '%s'", metadata.ProfileName)
	}

	if metadata.DeletedAt.IsZero() {
		t.Error("expected non-zero DeletedAt timestamp")
	}

	// Verify metadata is persisted (create new system instance)
	sys2, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create second trash system: %v", err)
	}

	metadata2, err := sys2.GetMetadata(id)
	if err != nil {
		t.Fatalf("failed to get metadata from second system: %v", err)
	}

	if metadata2.OriginalPath != metadata.OriginalPath {
		t.Error("metadata not persisted correctly")
	}
}

func TestSystem_GetMetadata_NotFound(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Try to get metadata for non-existent item
	_, err = sys.GetMetadata("nonexistent_id")
	if err == nil {
		t.Error("expected error for non-existent trash item, got nil")
	}
}

func TestSystem_ListEmpty(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// List should return empty slice, not error
	items, err := sys.List()
	if err != nil {
		t.Fatalf("failed to list empty trash: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items in empty trash, got %d", len(items))
	}
}

func TestSystem_CleanMultipleItems(t *testing.T) {
	// Create temporary trash directory
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	// Create and move multiple test files
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		target := types.Target{
			Path:        testFile,
			Size:        7,
			ProfileName: "test",
		}

		if _, err := sys.Move(target); err != nil {
			t.Fatalf("failed to move to trash: %v", err)
		}
	}

	// Verify all items are in trash
	items, err := sys.List()
	if err != nil {
		t.Fatalf("failed to list trash items: %v", err)
	}

	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}

	// Clean all items (retention period = 0)
	if err := sys.Clean(0); err != nil {
		t.Fatalf("failed to clean trash: %v", err)
	}

	// Verify all items are removed
	items, err = sys.List()
	if err != nil {
		t.Fatalf("failed to list trash items: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items after clean, got %d", len(items))
	}
}

func TestSystem_GetTrashDir(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	sys, err := NewSystem(trashDir)
	if err != nil {
		t.Fatalf("failed to create trash system: %v", err)
	}

	if sys.GetTrashDir() != trashDir {
		t.Errorf("expected trash dir %s, got %s", trashDir, sys.GetTrashDir())
	}
}
