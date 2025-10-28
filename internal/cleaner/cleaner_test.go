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
	assert.Contains(t, report.Errors[0].Error.Error(), "path not found")
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
		assert.Contains(t, err.Error(), "path not found")
	})
}

func TestCleaner_ErrorIsolation(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create multiple targets, some valid and some invalid
	validTarget1 := filepath.Join(tmpDir, "valid1")
	validTarget2 := filepath.Join(tmpDir, "valid2")
	require.NoError(t, os.MkdirAll(validTarget1, 0755))
	require.NoError(t, os.MkdirAll(validTarget2, 0755))

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Create targets with mix of valid and invalid paths
	targets := []types.Target{
		{
			Path:        validTarget1,
			Size:        100,
			Type:        "directory",
			ProfileName: "test",
			IsDirectory: true,
		},
		{
			Path:        "/nonexistent/path",
			Size:        200,
			Type:        "directory",
			ProfileName: "test",
			IsDirectory: true,
		},
		{
			Path:        validTarget2,
			Size:        300,
			Type:        "directory",
			ProfileName: "test",
			IsDirectory: true,
		},
	}

	// Clean with error isolation
	report, err := cleaner.Clean(context.Background(), targets, CleanOptions{
		UseTrash: true,
	})

	// Should not return error (errors are isolated)
	require.NoError(t, err)

	// Should have cleaned 2 valid targets
	assert.Equal(t, 2, report.FilesDeleted)
	assert.Equal(t, int64(400), report.TotalSize) // 100 + 300

	// Should have 1 error for the invalid path
	assert.Len(t, report.Errors, 1)
	assert.Contains(t, report.Errors[0].Error.Error(), "path not found")

	// Should have 2 trashed items
	assert.Len(t, report.TrashedItems, 2)
}

func TestCleaner_ContextCancellation(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create many targets
	var targets []types.Target
	for i := 0; i < 100; i++ {
		targetDir := filepath.Join(tmpDir, "target", string(rune('a'+i%26)))
		require.NoError(t, os.MkdirAll(targetDir, 0755))

		targets = append(targets, types.Target{
			Path:        targetDir,
			Size:        int64(100),
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

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start cleaning in goroutine
	done := make(chan *types.CleanReport)
	go func() {
		report, _ := cleaner.Clean(ctx, targets, CleanOptions{
			UseTrash: true,
		})
		done <- report
	}()

	// Cancel immediately
	cancel()

	// Wait for completion
	report := <-done

	// Should have stopped early (not all targets cleaned)
	assert.Less(t, report.FilesDeleted, len(targets))
}

func TestCleaner_ConcurrentCleaning(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create multiple targets
	var targets []types.Target
	for i := 0; i < 10; i++ {
		targetDir := filepath.Join(tmpDir, "target"+string(rune('0'+i)))
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

	// Clean with high concurrency
	startTime := time.Now()
	progressCh, err := cleaner.CleanAsync(context.Background(), targets, CleanOptions{
		UseTrash:    true,
		Concurrency: 4,
	})
	require.NoError(t, err)

	// Collect progress and generate report simultaneously
	report := GenerateReportFromProgress(progressCh, startTime)

	// Verify report
	assert.Equal(t, 10, report.FilesDeleted)
	assert.Equal(t, int64(5500), report.TotalSize) // Sum of 100+200+...+1000
	assert.Empty(t, report.Errors)

	// Verify all targets were deleted
	for _, target := range targets {
		_, err := os.Stat(target.Path)
		assert.True(t, os.IsNotExist(err), "Target should not exist: %s", target.Path)
	}
}

func TestCleaner_DryRun(t *testing.T) {
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
		Size:        100,
		Type:        "directory",
		ProfileName: "test",
		IsDirectory: true,
	}

	// Note: DryRun is handled at CLI level, not in cleaner
	// This test verifies normal operation
	report, err := cleaner.Clean(context.Background(), []types.Target{target}, CleanOptions{
		UseTrash: true,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, report.FilesDeleted)

	// Verify target was actually deleted
	_, err = os.Stat(targetDir)
	assert.True(t, os.IsNotExist(err))
}

func TestCleaner_MultipleTargetsSameDirectory(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create project with multiple cleanable directories
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	nodeModules := filepath.Join(projectDir, "node_modules")
	dist := filepath.Join(projectDir, "dist")
	build := filepath.Join(projectDir, "build")

	require.NoError(t, os.MkdirAll(nodeModules, 0755))
	require.NoError(t, os.MkdirAll(dist, 0755))
	require.NoError(t, os.MkdirAll(build, 0755))

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Create targets
	targets := []types.Target{
		{
			Path:        nodeModules,
			Size:        1000,
			Type:        "directory",
			ProfileName: "Node.js",
			IsDirectory: true,
		},
		{
			Path:        dist,
			Size:        500,
			Type:        "directory",
			ProfileName: "Node.js",
			IsDirectory: true,
		},
		{
			Path:        build,
			Size:        300,
			Type:        "directory",
			ProfileName: "Node.js",
			IsDirectory: true,
		},
	}

	// Clean all targets
	report, err := cleaner.Clean(context.Background(), targets, CleanOptions{
		UseTrash: true,
	})

	require.NoError(t, err)
	assert.Equal(t, 3, report.FilesDeleted)
	assert.Equal(t, int64(1800), report.TotalSize)
	assert.Len(t, report.TrashedItems, 3)
	assert.Empty(t, report.Errors)

	// Verify all targets were deleted
	for _, target := range targets {
		_, err := os.Stat(target.Path)
		assert.True(t, os.IsNotExist(err), "Target should not exist: %s", target.Path)
	}
}

// Benchmark tests

func BenchmarkCleaner_SmallBatch(b *testing.B) {
	// Create temporary directories
	tmpDir := b.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(b, err)

	// Create cleaner
	cleaner := New(trashSystem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create 10 targets
		var targets []types.Target
		for j := 0; j < 10; j++ {
			targetDir := filepath.Join(tmpDir, "batch"+string(rune('0'+i%10)), "target"+string(rune('0'+j)))
			require.NoError(b, os.MkdirAll(targetDir, 0755))

			targets = append(targets, types.Target{
				Path:        targetDir,
				Size:        int64(100),
				Type:        "directory",
				ProfileName: "test",
				IsDirectory: true,
			})
		}
		b.StartTimer()

		// Clean
		_, err := cleaner.Clean(context.Background(), targets, CleanOptions{
			UseTrash: true,
		})
		require.NoError(b, err)
	}
}

func BenchmarkCleaner_LargeBatch(b *testing.B) {
	// Create temporary directories
	tmpDir := b.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(b, err)

	// Create cleaner
	cleaner := New(trashSystem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create 100 targets
		var targets []types.Target
		for j := 0; j < 100; j++ {
			targetDir := filepath.Join(tmpDir, "batch"+string(rune('0'+i%10)), "target"+string(rune('0'+j%10)), "sub"+string(rune('0'+j/10)))
			require.NoError(b, os.MkdirAll(targetDir, 0755))

			targets = append(targets, types.Target{
				Path:        targetDir,
				Size:        int64(100),
				Type:        "directory",
				ProfileName: "test",
				IsDirectory: true,
			})
		}
		b.StartTimer()

		// Clean with concurrency
		_, err := cleaner.Clean(context.Background(), targets, CleanOptions{
			UseTrash: true,
		})
		require.NoError(b, err)
	}
}

func BenchmarkCleaner_ConcurrentVsSequential(b *testing.B) {
	// Create temporary directories
	tmpDir := b.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create trash system
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(b, err)

	// Create cleaner
	cleaner := New(trashSystem)

	// Create targets once
	var targets []types.Target
	for j := 0; j < 50; j++ {
		targetDir := filepath.Join(tmpDir, "targets", "target"+string(rune('0'+j%10)), "sub"+string(rune('0'+j/10)))
		require.NoError(b, os.MkdirAll(targetDir, 0755))

		targets = append(targets, types.Target{
			Path:        targetDir,
			Size:        int64(100),
			Type:        "directory",
			ProfileName: "test",
			IsDirectory: true,
		})
	}

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Recreate targets
			for _, target := range targets {
				require.NoError(b, os.MkdirAll(target.Path, 0755))
			}
			b.StartTimer()

			// Clean sequentially
			_, err := cleaner.Clean(context.Background(), targets, CleanOptions{
				UseTrash: true,
			})
			require.NoError(b, err)
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Recreate targets
			for _, target := range targets {
				require.NoError(b, os.MkdirAll(target.Path, 0755))
			}
			b.StartTimer()

			// Clean concurrently
			progressCh, err := cleaner.CleanAsync(context.Background(), targets, CleanOptions{
				UseTrash:    true,
				Concurrency: 4,
			})
			require.NoError(b, err)

			// Consume progress
			for range progressCh {
			}
		}
	})
}
