package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAll(t *testing.T) {
	loader := NewLoader()

	// Get the profiles directory relative to the project root
	profilesDir := filepath.Join("..", "..", "profiles")

	profiles, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(profiles) == 0 {
		t.Fatal("Expected at least one profile to be loaded")
	}

	// Check that we have the expected profiles
	expectedProfiles := map[string]bool{
		"Node.js": false,
		"Python":  false,
		"Rust":    false,
		"Flutter": false,
		"Go":      false,
	}

	for _, profile := range profiles {
		if _, exists := expectedProfiles[profile.Name]; exists {
			expectedProfiles[profile.Name] = true
		}
	}

	for name, found := range expectedProfiles {
		if !found {
			t.Errorf("Expected profile %s not found", name)
		}
	}
}

func TestLoadProfile(t *testing.T) {
	loader := NewLoader()

	profilePath := filepath.Join("..", "..", "profiles", "node.json")
	profile, err := loader.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadProfile failed: %v", err)
	}

	if profile.Name != "Node.js" {
		t.Errorf("Expected profile name 'Node.js', got '%s'", profile.Name)
	}

	if len(profile.Patterns) == 0 {
		t.Error("Expected at least one pattern")
	}

	if len(profile.Detect) == 0 {
		t.Error("Expected at least one detect pattern")
	}
}

func TestMatchProfile(t *testing.T) {
	loader := NewLoader()

	// Load profiles first
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Create a temporary directory with a package.json file
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Test matching
	profile, err := loader.MatchProfile(tmpDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	if profile == nil {
		t.Fatal("Expected to match Node.js profile")
	}

	if profile.Name != "Node.js" {
		t.Errorf("Expected Node.js profile, got %s", profile.Name)
	}
}

func TestMatchesPattern(t *testing.T) {
	loader := NewLoader()

	profilePath := filepath.Join("..", "..", "profiles", "node.json")
	profile, err := loader.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadProfile failed: %v", err)
	}

	tests := []struct {
		name     string
		expected bool
	}{
		{"node_modules", true},
		{"dist", true},
		{"build", true},
		{".next", true},
		{"src", false},
		{"index.js", false},
	}

	for _, tt := range tests {
		result := loader.MatchesPattern(tt.name, profile)
		if result != tt.expected {
			t.Errorf("MatchesPattern(%s) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}
