package profiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// Loader handles loading and managing profiles
type Loader struct {
	profiles     []types.Profile
	profileCache map[string]*types.Profile
	matchCache   map[string]*types.Profile
	cacheMutex   sync.RWMutex
}

// NewLoader creates a new profile loader
func NewLoader() *Loader {
	return &Loader{
		profiles:     make([]types.Profile, 0),
		profileCache: make(map[string]*types.Profile),
		matchCache:   make(map[string]*types.Profile),
	}
}

// LoadAll reads all JSON profiles from the specified directory
func (l *Loader) LoadAll(dir string) ([]types.Profile, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("profiles directory does not exist: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	profiles := make([]types.Profile, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		profilePath := filepath.Join(dir, entry.Name())
		profile, err := l.LoadProfile(profilePath)
		if err != nil {
			// Log error but continue loading other profiles
			fmt.Fprintf(os.Stderr, "Warning: failed to load profile %s: %v\n", entry.Name(), err)
			continue
		}

		profiles = append(profiles, *profile)
	}

	l.profiles = profiles

	// Build profile cache
	l.cacheMutex.Lock()
	for i := range l.profiles {
		l.profileCache[l.profiles[i].Name] = &l.profiles[i]
	}
	l.cacheMutex.Unlock()

	return profiles, nil
}

// LoadProfile loads a single profile from a JSON file
func (l *Loader) LoadProfile(path string) (*types.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile types.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile JSON: %w", err)
	}

	// Validate profile
	if err := l.validateProfile(&profile); err != nil {
		return nil, fmt.Errorf("profile validation failed: %w", err)
	}

	return &profile, nil
}

// validateProfile checks if a profile has all required fields and valid patterns
func (l *Loader) validateProfile(profile *types.Profile) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.Version == "" {
		return fmt.Errorf("profile version is required")
	}

	if len(profile.Patterns) == 0 {
		return fmt.Errorf("profile must have at least one pattern")
	}

	if len(profile.Detect) == 0 {
		return fmt.Errorf("profile must have at least one detect pattern")
	}

	// Validate pattern syntax (basic glob validation)
	for _, pattern := range profile.Patterns {
		if pattern == "" {
			return fmt.Errorf("empty pattern found")
		}
		// Check for valid glob pattern
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}
	}

	// Validate detect patterns
	for _, detect := range profile.Detect {
		if detect == "" {
			return fmt.Errorf("empty detect pattern found")
		}
	}

	return nil
}

// GetProfiles returns all loaded profiles
func (l *Loader) GetProfiles() []types.Profile {
	return l.profiles
}

// GetProfile returns a profile by name
func (l *Loader) GetProfile(name string) (*types.Profile, error) {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	profile, exists := l.profileCache[name]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	return profile, nil
}
