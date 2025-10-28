package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/raucheacho/rosia-cli/internal/profiles"
)

func TestScan(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a Node.js project structure
	nodeProject := filepath.Join(tmpDir, "my-project")
	if err := os.MkdirAll(nodeProject, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create package.json to trigger Node.js profile
	packageJSON := filepath.Join(nodeProject, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create node_modules directory
	nodeModules := filepath.Join(nodeProject, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}

	// Create a dummy file in node_modules
	dummyFile := filepath.Join(nodeModules, "test.js")
	if err := os.WriteFile(dummyFile, []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Scan the directory
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scanner.Scan(ctx, []string{tmpDir}, opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify we found the node_modules directory
	if len(targets) == 0 {
		t.Fatal("Expected to find at least one target")
	}

	found := false
	for _, target := range targets {
		if filepath.Base(target.Path) == "node_modules" {
			found = true
			if target.ProfileName != "Node.js" {
				t.Errorf("Expected profile name 'Node.js', got '%s'", target.ProfileName)
			}
			if target.Size <= 0 {
				t.Errorf("Expected size > 0, got %d", target.Size)
			}
			if !target.IsDirectory {
				t.Error("Expected IsDirectory to be true")
			}
		}
	}

	if !found {
		t.Error("Expected to find node_modules target")
	}
}

func TestScanWithMaxDepth(t *testing.T) {
	// Create a nested directory structure
	tmpDir := t.TempDir()

	// Create nested directories
	level1 := filepath.Join(tmpDir, "level1")
	level2 := filepath.Join(level1, "level2")
	level3 := filepath.Join(level2, "level3")

	if err := os.MkdirAll(level3, 0755); err != nil {
		t.Fatalf("Failed to create nested dirs: %v", err)
	}

	// Create package.json at level1
	packageJSON := filepath.Join(level1, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create node_modules at level3
	nodeModules := filepath.Join(level3, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Scan with MaxDepth = 2 (should not find level3/node_modules)
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      2,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scanner.Scan(ctx, []string{tmpDir}, opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should not find node_modules at level3 due to depth limit
	for _, target := range targets {
		if filepath.Base(target.Path) == "node_modules" {
			t.Error("Should not find node_modules at level3 with MaxDepth=2")
		}
	}
}

func TestScanWithIgnorePaths(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create two projects
	project1 := filepath.Join(tmpDir, "project1")
	project2 := filepath.Join(tmpDir, "project2")

	for _, proj := range []string{project1, project2} {
		if err := os.MkdirAll(proj, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}

		packageJSON := filepath.Join(proj, "package.json")
		if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create package.json: %v", err)
		}

		nodeModules := filepath.Join(proj, "node_modules")
		if err := os.MkdirAll(nodeModules, 0755); err != nil {
			t.Fatalf("Failed to create node_modules: %v", err)
		}
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Scan with project2 ignored
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{project2},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scanner.Scan(ctx, []string{tmpDir}, opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find node_modules in project1
	foundProject1 := false
	for _, target := range targets {
		if filepath.Base(target.Path) == "node_modules" {
			if filepath.Dir(target.Path) == project1 {
				foundProject1 = true
			}
			if filepath.Dir(target.Path) == project2 {
				t.Error("Should not find node_modules in ignored project2")
			}
		}
	}

	if !foundProject1 {
		t.Error("Expected to find node_modules in project1")
	}
}

func TestScanAsync(t *testing.T) {
	// Create a temporary directory structure with multiple projects
	tmpDir := t.TempDir()

	// Create 5 Node.js projects
	for i := 0; i < 5; i++ {
		projectDir := filepath.Join(tmpDir, "project"+string(rune('0'+i)))
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}

		packageJSON := filepath.Join(projectDir, "package.json")
		if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create package.json: %v", err)
		}

		nodeModules := filepath.Join(projectDir, "node_modules")
		if err := os.MkdirAll(nodeModules, 0755); err != nil {
			t.Fatalf("Failed to create node_modules: %v", err)
		}

		// Create some files in node_modules
		testFile := filepath.Join(nodeModules, "test.js")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Scan asynchronously
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targetChan, errorChan := scanner.ScanAsync(ctx, []string{tmpDir}, opts)

	// Collect results
	var targets []string
	var errors []error

	done := make(chan bool)
	go func() {
		for target := range targetChan {
			targets = append(targets, target.Path)
		}
		done <- true
	}()

	go func() {
		for err := range errorChan {
			errors = append(errors, err)
		}
	}()

	<-done

	// Verify results
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	if len(targets) != 5 {
		t.Errorf("Expected 5 targets, got %d", len(targets))
	}

	// Verify all targets are node_modules
	for _, target := range targets {
		if filepath.Base(target) != "node_modules" {
			t.Errorf("Expected node_modules, got %s", filepath.Base(target))
		}
	}
}

func TestScanWithContextCancellation(t *testing.T) {
	// Create a large directory structure
	tmpDir := t.TempDir()

	// Create many projects to ensure scan takes some time
	for i := 0; i < 100; i++ {
		projectDir := filepath.Join(tmpDir, "project"+string(rune('0'+i%10)), "sub"+string(rune('0'+i/10)))
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}

		packageJSON := filepath.Join(projectDir, "package.json")
		if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create package.json: %v", err)
		}

		nodeModules := filepath.Join(projectDir, "node_modules")
		if err := os.MkdirAll(nodeModules, 0755); err != nil {
			t.Fatalf("Failed to create node_modules: %v", err)
		}
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Start scan in goroutine
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	done := make(chan error)
	go func() {
		_, err := scanner.Scan(ctx, []string{tmpDir}, opts)
		done <- err
	}()

	// Cancel context immediately
	cancel()

	// Wait for scan to complete
	err = <-done

	// Should get context.Canceled error
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestScanWithHiddenFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a project with hidden directories
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	packageJSON := filepath.Join(projectDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create visible node_modules
	nodeModules := filepath.Join(projectDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}

	// Create hidden .cache directory
	cacheDir := filepath.Join(projectDir, ".cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create .cache: %v", err)
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Test 1: Scan without hidden files
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scanner.Scan(ctx, []string{tmpDir}, opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find node_modules but not .cache
	foundNodeModules := false
	foundCache := false
	for _, target := range targets {
		if filepath.Base(target.Path) == "node_modules" {
			foundNodeModules = true
		}
		if filepath.Base(target.Path) == ".cache" {
			foundCache = true
		}
	}

	if !foundNodeModules {
		t.Error("Expected to find node_modules")
	}
	if foundCache {
		t.Error("Should not find .cache when IncludeHidden is false")
	}

	// Test 2: Scan with hidden files
	opts.IncludeHidden = true
	targets, err = scanner.Scan(ctx, []string{tmpDir}, opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find both node_modules and .cache
	foundNodeModules = false
	foundCache = false
	for _, target := range targets {
		if filepath.Base(target.Path) == "node_modules" {
			foundNodeModules = true
		}
		if filepath.Base(target.Path) == ".cache" {
			foundCache = true
		}
	}

	if !foundNodeModules {
		t.Error("Expected to find node_modules")
	}
	if !foundCache {
		t.Error("Expected to find .cache when IncludeHidden is true")
	}
}

func TestScanMultipleProfiles(t *testing.T) {
	// Create a temporary directory with multiple project types
	tmpDir := t.TempDir()

	// Create Node.js project
	nodeProject := filepath.Join(tmpDir, "node-project")
	if err := os.MkdirAll(nodeProject, 0755); err != nil {
		t.Fatalf("Failed to create node project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeProject, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nodeProject, "node_modules"), 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}

	// Create Python project
	pythonProject := filepath.Join(tmpDir, "python-project")
	if err := os.MkdirAll(pythonProject, 0755); err != nil {
		t.Fatalf("Failed to create python project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pythonProject, "requirements.txt"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pythonProject, "venv"), 0755); err != nil {
		t.Fatalf("Failed to create venv: %v", err)
	}

	// Create Rust project
	rustProject := filepath.Join(tmpDir, "rust-project")
	if err := os.MkdirAll(rustProject, 0755); err != nil {
		t.Fatalf("Failed to create rust project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rustProject, "Cargo.toml"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(rustProject, "target"), 0755); err != nil {
		t.Fatalf("Failed to create target: %v", err)
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Create scanner
	scanner := NewScanner(loader)

	// Scan all projects
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	targets, err := scanner.Scan(ctx, []string{tmpDir}, opts)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find targets from all three profiles
	profilesFound := make(map[string]bool)
	for _, target := range targets {
		profilesFound[target.ProfileName] = true
	}

	expectedProfiles := []string{"Node.js", "Python", "Rust"}
	for _, profile := range expectedProfiles {
		if !profilesFound[profile] {
			t.Errorf("Expected to find target with profile %s", profile)
		}
	}
}

// Benchmark tests

func BenchmarkScanner_SmallDirectory(b *testing.B) {
	// Create a small directory structure (10 projects)
	tmpDir := b.TempDir()

	for i := 0; i < 10; i++ {
		projectDir := filepath.Join(tmpDir, "project"+string(rune('0'+i)))
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			b.Fatalf("Failed to create project: %v", err)
		}

		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}"), 0644); err != nil {
			b.Fatalf("Failed to create package.json: %v", err)
		}

		nodeModules := filepath.Join(projectDir, "node_modules")
		if err := os.MkdirAll(nodeModules, 0755); err != nil {
			b.Fatalf("Failed to create node_modules: %v", err)
		}

		// Create some files in node_modules
		for j := 0; j < 10; j++ {
			testFile := filepath.Join(nodeModules, "file"+string(rune('0'+j))+".js")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				b.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	if _, err := loader.LoadAll(profilesDir); err != nil {
		b.Fatalf("Failed to load profiles: %v", err)
	}

	scanner := NewScanner(loader)
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(ctx, []string{tmpDir}, opts)
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

func BenchmarkScanner_LargeDirectory(b *testing.B) {
	// Create a larger directory structure (100 projects)
	tmpDir := b.TempDir()

	for i := 0; i < 100; i++ {
		projectDir := filepath.Join(tmpDir, "project"+string(rune('0'+i%10)), "sub"+string(rune('0'+i/10)))
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			b.Fatalf("Failed to create project: %v", err)
		}

		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}"), 0644); err != nil {
			b.Fatalf("Failed to create package.json: %v", err)
		}

		nodeModules := filepath.Join(projectDir, "node_modules")
		if err := os.MkdirAll(nodeModules, 0755); err != nil {
			b.Fatalf("Failed to create node_modules: %v", err)
		}

		// Create some files
		for j := 0; j < 5; j++ {
			testFile := filepath.Join(nodeModules, "file"+string(rune('0'+j))+".js")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				b.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	if _, err := loader.LoadAll(profilesDir); err != nil {
		b.Fatalf("Failed to load profiles: %v", err)
	}

	scanner := NewScanner(loader)
	ctx := context.Background()
	opts := ScanOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		IgnorePaths:   []string{},
		DryRun:        false,
		Concurrency:   4,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(ctx, []string{tmpDir}, opts)
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

func BenchmarkScanner_ConcurrentVsSequential(b *testing.B) {
	// Create directory structure
	tmpDir := b.TempDir()

	for i := 0; i < 50; i++ {
		projectDir := filepath.Join(tmpDir, "project"+string(rune('0'+i%10)), "sub"+string(rune('0'+i/10)))
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			b.Fatalf("Failed to create project: %v", err)
		}

		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}"), 0644); err != nil {
			b.Fatalf("Failed to create package.json: %v", err)
		}

		nodeModules := filepath.Join(projectDir, "node_modules")
		if err := os.MkdirAll(nodeModules, 0755); err != nil {
			b.Fatalf("Failed to create node_modules: %v", err)
		}
	}

	// Load profiles
	loader := profiles.NewLoader()
	profilesDir := filepath.Join("..", "..", "profiles")
	if _, err := loader.LoadAll(profilesDir); err != nil {
		b.Fatalf("Failed to load profiles: %v", err)
	}

	scanner := NewScanner(loader)
	ctx := context.Background()

	b.Run("Sequential", func(b *testing.B) {
		opts := ScanOptions{
			MaxDepth:      10,
			IncludeHidden: false,
			IgnorePaths:   []string{},
			DryRun:        false,
			Concurrency:   1,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := scanner.Scan(ctx, []string{tmpDir}, opts)
			if err != nil {
				b.Fatalf("Scan failed: %v", err)
			}
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		opts := ScanOptions{
			MaxDepth:      10,
			IncludeHidden: false,
			IgnorePaths:   []string{},
			DryRun:        false,
			Concurrency:   4,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := scanner.Scan(ctx, []string{tmpDir}, opts)
			if err != nil {
				b.Fatalf("Scan failed: %v", err)
			}
		}
	})
}
