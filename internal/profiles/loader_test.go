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

func TestLoadProfile_InvalidJSON(t *testing.T) {
	loader := NewLoader()

	// Create a temporary invalid JSON file
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Should fail to load
	_, err := loader.LoadProfile(invalidPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}
}

func TestLoadProfile_MissingFields(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "missing name",
			content: `{"version": "1.0", "patterns": ["test"], "detect": ["test.txt"], "enabled": true}`,
		},
		{
			name:    "missing version",
			content: `{"name": "Test", "patterns": ["test"], "detect": ["test.txt"], "enabled": true}`,
		},
		{
			name:    "missing patterns",
			content: `{"name": "Test", "version": "1.0", "detect": ["test.txt"], "enabled": true}`,
		},
		{
			name:    "missing detect",
			content: `{"name": "Test", "version": "1.0", "patterns": ["test"], "enabled": true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profilePath := filepath.Join(tmpDir, tt.name+".json")
			if err := os.WriteFile(profilePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, err := loader.LoadProfile(profilePath)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestLoadProfile_InvalidPattern(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create profile with invalid glob pattern
	content := `{
		"name": "Test",
		"version": "1.0",
		"patterns": ["[invalid"],
		"detect": ["test.txt"],
		"enabled": true
	}`

	profilePath := filepath.Join(tmpDir, "invalid_pattern.json")
	if err := os.WriteFile(profilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := loader.LoadProfile(profilePath)
	if err == nil {
		t.Error("Expected error for invalid glob pattern, got nil")
	}
}

func TestMatchProfile_Caching(t *testing.T) {
	loader := NewLoader()

	// Load profiles
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Create a temporary directory with package.json
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// First match - should populate cache
	profile1, err := loader.MatchProfile(tmpDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	if profile1 == nil {
		t.Fatal("Expected to match Node.js profile")
	}

	// Second match - should use cache
	profile2, err := loader.MatchProfile(tmpDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	// Should return the same profile instance (from cache)
	if profile1 != profile2 {
		t.Error("Expected cached profile to be the same instance")
	}

	// Clear cache
	loader.ClearCache()

	// Third match - should re-match after cache clear
	profile3, err := loader.MatchProfile(tmpDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	if profile3 == nil {
		t.Fatal("Expected to match Node.js profile after cache clear")
	}

	if profile3.Name != profile1.Name {
		t.Error("Expected same profile name after cache clear")
	}
}

func TestMatchProfile_DisabledProfile(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create a custom profile that's disabled
	content := `{
		"name": "Disabled",
		"version": "1.0",
		"patterns": ["test_dir"],
		"detect": ["marker.txt"],
		"enabled": false
	}`

	profilePath := filepath.Join(tmpDir, "disabled.json")
	if err := os.WriteFile(profilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Load the profile
	_, err := loader.LoadAll(tmpDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Create test directory with marker
	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	markerFile := filepath.Join(testDir, "marker.txt")
	if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create marker: %v", err)
	}

	// Should not match disabled profile
	profile, err := loader.MatchProfile(testDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	if profile != nil {
		t.Error("Should not match disabled profile")
	}
}

func TestGetProfile(t *testing.T) {
	loader := NewLoader()

	// Load profiles
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Test getting existing profile
	profile, err := loader.GetProfile("Node.js")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if profile.Name != "Node.js" {
		t.Errorf("Expected profile name 'Node.js', got '%s'", profile.Name)
	}

	// Test getting non-existent profile
	_, err = loader.GetProfile("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent profile, got nil")
	}
}

func TestMatchProfile_NoMatch(t *testing.T) {
	loader := NewLoader()

	// Load profiles
	profilesDir := filepath.Join("..", "..", "profiles")
	_, err := loader.LoadAll(profilesDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Create directory with no matching markers
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "empty-project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Should not match any profile
	profile, err := loader.MatchProfile(testDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	if profile != nil {
		t.Errorf("Expected no match, got profile: %s", profile.Name)
	}
}

func TestMatchProfile_GlobPattern(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create a profile with glob pattern in detect
	content := `{
		"name": "GlobTest",
		"version": "1.0",
		"patterns": ["target"],
		"detect": ["*.toml"],
		"enabled": true
	}`

	profilePath := filepath.Join(tmpDir, "glob.json")
	if err := os.WriteFile(profilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Load the profile
	_, err := loader.LoadAll(tmpDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Create test directory with matching file
	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	cargoFile := filepath.Join(testDir, "Cargo.toml")
	if err := os.WriteFile(cargoFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	// Should match via glob pattern
	profile, err := loader.MatchProfile(testDir)
	if err != nil {
		t.Fatalf("MatchProfile failed: %v", err)
	}

	if profile == nil {
		t.Fatal("Expected to match profile with glob pattern")
	}

	if profile.Name != "GlobTest" {
		t.Errorf("Expected profile 'GlobTest', got '%s'", profile.Name)
	}
}
