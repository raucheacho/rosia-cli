package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/raucheacho/rosia-cli/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleaner_Clean(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")
	targetDir := filepath.Join(tmpDir, "target")

	// Create target directory with some content
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	testFile := filepath.Join(targetDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Create target
	target := types.Target{
		Path:        targetDir,
		Size:        100,
		Type:        "directory",
		ProfileName: "test",
		IsDirectory: true,
	}

	t.Run("Clean with trash", func(t *testing.T) {
		// Clean with trash enabled
		report, err := cleaner.Clean(context.Background(), []types.Target{target}, CleanOptions{
			UseTrash: true,
		})

		require.NoError(t, err)
		assert.Equal(t, 1, report.FilesDeleted)
		assert.Equal(t, int64(100), report.TotalSize)
		assert.Len(t, report.TrashedItems, 1)
		assert.Empty(t, report.Errors)

		// Verify target was moved to trash
		_, err = os.Stat(targetDir)
		assert.True(t, os.IsNotExist(err), "target should not exist after cleaning")

		// Verify trash item exists
		items, err := trashSystem.List()
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})
}

func TestCleaner_Clean_WithoutTrash(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")
	targetDir := filepath.Join(tmpDir, "target")

	// Create target directory
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Create target
	target := types.Target{
		Path:        targetDir,
		Size:        50,
		Type:        "directory",
		ProfileName: "test",
		IsDirectory: true,
	}

	// Clean without trash
	report, err := cleaner.Clean(context.Background(), []types.Target{target}, CleanOptions{
		UseTrash: false,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, report.FilesDeleted)
	assert.Equal(t, int64(50), report.TotalSize)
	assert.Empty(t, report.TrashedItems)
	assert.Empty(t, report.Errors)

	// Verify target was deleted
	_, err = os.Stat(targetDir)
	assert.True(t, os.IsNotExist(err))
}

func TestCleaner_Clean_PermissionError(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Create target with non-existent path
	target := types.Target{
		Path:        "/nonexistent/path",
		Size:        100,
		Type:        "directory",
		ProfileName: "test",
		IsDirectory: true,
	}

	// Clean should handle error gracefully
	report, err := cleaner.Clean(context.Background(), []types.Target{target}, CleanOptions{
		UseTrash: true,
	})

	require.NoError(t, err)
	assert.Equal(t, 0, report.FilesDeleted)
	assert.Len(t, report.Errors, 1)
	assert.Contains(t, report.Errors[0].Error.Error(), "does not exist")
}

func TestCleaner_CleanAsync(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create multiple target directories
	var targets []types.Target
	for i := 0; i < 5; i++ {
		targetDir := filepath.Join(tmpDir, "target", string(rune('a'+i)))
		require.NoError(t, os.MkdirAll(targetDir, 0755))

		targets = append(targets, types.Target{
			Path:        targetDir,
			Size:        int64(100 * (i + 1)),
			Type:        "directory",
			ProfileName: "test",
			IsDirectory: true,
		})
	}

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Clean asynchronously
	startTime := time.Now()
	progressCh, err := cleaner.CleanAsync(context.Background(), targets, CleanOptions{
		UseTrash:    true,
		Concurrency: 2,
	})
	require.NoError(t, err)

	// Generate report from progress
	report := GenerateReportFromProgress(progressCh, startTime)

	assert.Equal(t, 5, report.FilesDeleted)
	assert.Equal(t, int64(1500), report.TotalSize) // 100+200+300+400+500
	assert.Empty(t, report.Errors)
}

func TestCleaner_canDelete(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	cleaner := New(trashSystem)

	t.Run("Valid path", func(t *testing.T) {
		validPath := filepath.Join(tmpDir, "valid")
		require.NoError(t, os.MkdirAll(validPath, 0755))

		err := cleaner.canDelete(validPath)
		assert.NoError(t, err)
	})

	t.Run("Non-existent path", func(t *testing.T) {
		err := cleaner.canDelete("/nonexistent/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}
