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
