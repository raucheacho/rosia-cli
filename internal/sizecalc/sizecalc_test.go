package sizecalc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

func TestCalculate(t *testing.T) {
	// Create a temporary directory with some files
	tmpDir := t.TempDir()

	// Create a file with known size
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, World!")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create size calculator
	sc := NewSizeCalc(2)

	// Calculate size of the file
	size, err := sc.Calculate(testFile)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedSize := int64(len(content))
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestCalculateDirectory(t *testing.T) {
	// Create a temporary directory with multiple files
	tmpDir := t.TempDir()

	// Create multiple files
	files := []struct {
		name    string
		content string
	}{
		{"file1.txt", "Hello"},
		{"file2.txt", "World"},
		{"file3.txt", "Test"},
	}

	var expectedSize int64
	for _, f := range files {
		filePath := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(filePath, []byte(f.content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		expectedSize += int64(len(f.content))
	}

	// Create size calculator
	sc := NewSizeCalc(2)

	// Calculate size of the directory
	size, err := sc.Calculate(tmpDir)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestCalculateTargets(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()

	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Create files in each directory
	file1 := filepath.Join(dir1, "test1.txt")
	file2 := filepath.Join(dir2, "test2.txt")

	content1 := []byte("Content 1")
	content2 := []byte("Content 2 is longer")

	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create targets
	targets := []types.Target{
		{Path: dir1, IsDirectory: true},
		{Path: dir2, IsDirectory: true},
	}

	// Create size calculator
	sc := NewSizeCalc(2)

	// Calculate sizes
	ctx := context.Background()
	results, err := sc.CalculateTargets(ctx, targets)
	if err != nil {
		t.Fatalf("CalculateTargets failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Verify sizes
	if results[0].Size != int64(len(content1)) {
		t.Errorf("Expected size %d for dir1, got %d", len(content1), results[0].Size)
	}

	if results[1].Size != int64(len(content2)) {
		t.Errorf("Expected size %d for dir2, got %d", len(content2), results[1].Size)
	}
}

func TestCalculateSymlink(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a regular file
	targetFile := filepath.Join(tmpDir, "target.txt")
	content := []byte("Target content")
	if err := os.WriteFile(targetFile, content, 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create a symlink
	symlinkPath := filepath.Join(tmpDir, "symlink.txt")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	// Create size calculator
	sc := NewSizeCalc(2)

	// Calculate size of symlink (should be 0)
	size, err := sc.Calculate(symlinkPath)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected size 0 for symlink, got %d", size)
	}
}
