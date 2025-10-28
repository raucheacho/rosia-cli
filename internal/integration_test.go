package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/raucheacho/rosia-cli/internal/cleaner"
	"github.com/raucheacho/rosia-cli/internal/profiles"
	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScanCleanRestoreFlow tests the complete workflow: scan → clean → restore
func TestScanCleanRestoreFlow(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create a Node.js project
	projectDir := filepath.Join(tmpDir, "my-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	packageJSON := filepath.Join(projectDir, "package.json")
	require.NoError(t, os.WriteFile(packageJSON, []byte("{}"), 0644))

	nodeModules := filepath.Join(projectDir, "node_modules")
	require.NoError(t, os.MkdirAll(nodeModules, 0755))

	// Create some files in node_modules
	testFile := filepath.Join(nodeModules, "test.js")
	require.NoError(t, os.WriteFile(testFile, []byte("console.log('test');"), 0644))

	// Step 1: Scan
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	require.NoError(t, err)

	scannerInstance := scanner.NewScanner(loader)
	ctx := context.Background()
	opts := scanner.ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scannerInstance.Scan(ctx, []string{tmpDir}, opts)
	require.NoError(t, err)
	require.Len(t, targets, 1, "Expected to find 1 target (node_modules)")

	target := targets[0]
	assert.Equal(t, "node_modules", filepath.Base(target.Path))
	assert.Equal(t, "Node.js", target.ProfileName)

	// Step 2: Clean
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	cleanerInstance := cleaner.New(trashSystem)
	cleanOpts := cleaner.CleanOptions{
		UseTrash: true,
	}

	report, err := cleanerInstance.Clean(ctx, targets, cleanOpts)
	require.NoError(t, err)
	assert.Equal(t, 1, report.FilesDeleted)
	assert.Len(t, report.TrashedItems, 1)
	assert.Empty(t, report.Errors)

	// Verify node_modules is gone
	_, err = os.Stat(nodeModules)
	assert.True(t, os.IsNotExist(err), "node_modules should not exist after cleaning")

	// Step 3: Restore
	trashID := report.TrashedItems[0]
	err = trashSystem.Restore(trashID)
	require.NoError(t, err)

	// Verify node_modules is restored
	_, err = os.Stat(nodeModules)
	assert.NoError(t, err, "node_modules should exist after restore")

	// Verify nested file is restored
	_, err = os.Stat(testFile)
	assert.NoError(t, err, "nested file should exist after restore")
}

// TestMultiProfileDetection tests scanning a directory with multiple project types
func TestMultiProfileDetection(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create Node.js project
	nodeProject := filepath.Join(tmpDir, "node-project")
	require.NoError(t, os.MkdirAll(nodeProject, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nodeProject, "package.json"), []byte("{}"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(nodeProject, "node_modules"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(nodeProject, "dist"), 0755))

	// Create Python project
	pythonProject := filepath.Join(tmpDir, "python-project")
	require.NoError(t, os.MkdirAll(pythonProject, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(pythonProject, "requirements.txt"), []byte(""), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(pythonProject, "venv"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(pythonProject, "__pycache__"), 0755))

	// Create Rust project
	rustProject := filepath.Join(tmpDir, "rust-project")
	require.NoError(t, os.MkdirAll(rustProject, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(rustProject, "Cargo.toml"), []byte(""), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(rustProject, "target"), 0755))

	// Scan all projects
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	require.NoError(t, err)

	scannerInstance := scanner.NewScanner(loader)
	ctx := context.Background()
	opts := scanner.ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scannerInstance.Scan(ctx, []string{tmpDir}, opts)
	require.NoError(t, err)

	// Should find targets from all three profiles
	profilesFound := make(map[string]int)
	for _, target := range targets {
		profilesFound[target.ProfileName]++
	}

	// Verify we found targets from each profile
	assert.Greater(t, profilesFound["Node.js"], 0, "Should find Node.js targets")
	assert.Greater(t, profilesFound["Python"], 0, "Should find Python targets")
	assert.Greater(t, profilesFound["Rust"], 0, "Should find Rust targets")

	// Verify specific targets
	targetNames := make(map[string]bool)
	for _, target := range targets {
		targetNames[filepath.Base(target.Path)] = true
	}

	assert.True(t, targetNames["node_modules"], "Should find node_modules")
	assert.True(t, targetNames["dist"], "Should find dist")
	assert.True(t, targetNames["venv"], "Should find venv")
	assert.True(t, targetNames["__pycache__"], "Should find __pycache__")
	assert.True(t, targetNames["target"], "Should find target")
}

// TestScanWithIgnorePaths tests scanning with ignored paths
func TestScanWithIgnorePaths_Integration(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create multiple projects
	project1 := filepath.Join(tmpDir, "project1")
	project2 := filepath.Join(tmpDir, "project2")
	project3 := filepath.Join(tmpDir, "project3")

	for _, proj := range []string{project1, project2, project3} {
		require.NoError(t, os.MkdirAll(proj, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(proj, "package.json"), []byte("{}"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(proj, "node_modules"), 0755))
	}

	// Scan with project2 ignored
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	require.NoError(t, err)

	scannerInstance := scanner.NewScanner(loader)
	ctx := context.Background()
	opts := scanner.ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{project2},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scannerInstance.Scan(ctx, []string{tmpDir}, opts)
	require.NoError(t, err)

	// Should find targets in project1 and project3, but not project2
	foundInProject1 := false
	foundInProject2 := false
	foundInProject3 := false

	for _, target := range targets {
		if filepath.Dir(target.Path) == project1 {
			foundInProject1 = true
		}
		if filepath.Dir(target.Path) == project2 {
			foundInProject2 = true
		}
		if filepath.Dir(target.Path) == project3 {
			foundInProject3 = true
		}
	}

	assert.True(t, foundInProject1, "Should find targets in project1")
	assert.False(t, foundInProject2, "Should not find targets in ignored project2")
	assert.True(t, foundInProject3, "Should find targets in project3")
}

// TestConcurrentScanAndClean tests concurrent operations
func TestConcurrentScanAndClean(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	// Create multiple projects with different target names to avoid ID collisions
	for i := 0; i < 4; i++ {
		projectDir := filepath.Join(tmpDir, "project"+string(rune('0'+i)))
		require.NoError(t, os.MkdirAll(projectDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}"), 0644))

		// Use different directory names to avoid trash ID collisions (use only non-hidden patterns)
		targetNames := []string{"node_modules", "dist", "build", "coverage"}
		targetDir := filepath.Join(projectDir, targetNames[i])
		require.NoError(t, os.MkdirAll(targetDir, 0755))
	}

	// Scan with high concurrency
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	require.NoError(t, err)

	scannerInstance := scanner.NewScanner(loader)
	ctx := context.Background()
	opts := scanner.ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   4,
	}

	targets, err := scannerInstance.Scan(ctx, []string{tmpDir}, opts)
	require.NoError(t, err)
	assert.Len(t, targets, 4, "Should find 4 target directories")

	// Clean with high concurrency
	trashSystem, err := trash.NewSystem(trashDir)
	require.NoError(t, err)

	cleanerInstance := cleaner.New(trashSystem)
	cleanOpts := cleaner.CleanOptions{
		UseTrash:    true,
		Concurrency: 4,
	}

	progressCh, err := cleanerInstance.CleanAsync(ctx, targets, cleanOpts)
	require.NoError(t, err)

	// Collect results
	var successCount int
	var errorCount int
	for progress := range progressCh {
		if progress.Error == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	assert.Equal(t, 4, successCount, "Should successfully clean all 4 targets")
	assert.Equal(t, 0, errorCount, "Should have no errors")

	// Verify all targets were deleted
	for _, target := range targets {
		_, err := os.Stat(target.Path)
		assert.True(t, os.IsNotExist(err), "Target should not exist: %s", target.Path)
	}
}
