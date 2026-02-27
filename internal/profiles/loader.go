// Package profiles provides profile loading and matching functionality.
//
// Profiles define technology-specific cleaning rules including patterns to match
// and detection criteria. The loader reads JSON profile definitions and provides
// caching for efficient profile matching during scanning.
//
// Example usage:
//
//	loader := profiles.NewLoader()
//	if err := loader.LoadAll("profiles/"); err != nil {
//	    log.Fatal(err)
//	}
//	profile, err := loader.MatchProfile("/path/to/project")
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

// Loader handles loading and managing profiles.
//
// The Loader reads profile definitions from JSON files, validates them,
// and provides efficient profile matching with caching support.
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

// init ensures matchCache is initialized (used after loading profiles)
func (l *Loader) init() {
	if l.matchCache == nil {
		l.matchCache = make(map[string]*types.Profile)
	}
}

// LoadAll reads profiles from the specified path.
// It supports:
// - A directory containing individual .json profile files
// - A single profiles.json file containing an array of profiles
func (l *Loader) LoadAll(path string) ([]types.Profile, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.ErrPathNotFound{Path: path}
		}
		if os.IsPermission(err) {
			return nil, types.ErrPermissionDenied{Path: path}
		}
		return nil, fmt.Errorf("failed to access profiles path %s: %w", path, err)
	}

	// If path is a file, load it directly
	if !info.IsDir() {
		return l.loadProfilesFromFile(path)
	}

	// If it's a directory, check for profiles.json first
	profilesFile := filepath.Join(path, "profiles.json")
	if _, err := os.Stat(profilesFile); err == nil {
		return l.loadProfilesFromFile(profilesFile)
	}

	// Otherwise, load all .json files from directory
	return l.loadProfilesFromDir(path)
}

// loadProfilesFromDir loads profiles from individual JSON files in a directory
func (l *Loader) loadProfilesFromDir(dir string) ([]types.Profile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsPermission(err) {
			return nil, types.ErrPermissionDenied{Path: dir}
		}
		return nil, fmt.Errorf("failed to read profiles directory %s: %w", dir, err)
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

	l.setProfiles(profiles)
	return profiles, nil
}

// loadProfilesFromFile loads profiles from a single JSON file (supports array or single object)
func (l *Loader) loadProfilesFromFile(path string) ([]types.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.ErrPathNotFound{Path: path}
		}
		if os.IsPermission(err) {
			return nil, types.ErrPermissionDenied{Path: path}
		}
		return nil, fmt.Errorf("failed to read profiles file %s: %w", path, err)
	}

	// Try to unmarshal as array first
	var profiles []types.Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		// Try to unmarshal as single profile object
		var singleProfile types.Profile
		if err := json.Unmarshal(data, &singleProfile); err != nil {
			return nil, fmt.Errorf("failed to parse profiles file %s: %w", path, err)
		}
		profiles = []types.Profile{singleProfile}
	}

	// Validate all profiles
	validProfiles := make([]types.Profile, 0, len(profiles))
	for i := range profiles {
		if err := l.validateProfile(&profiles[i]); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: profile validation failed for %s: %v\n", profiles[i].Name, err)
			continue
		}
		validProfiles = append(validProfiles, profiles[i])
	}

	l.setProfiles(validProfiles)
	return validProfiles, nil
}

// setProfiles updates the loader's profiles and rebuilds the cache
func (l *Loader) setProfiles(profiles []types.Profile) {
	l.profiles = profiles
	l.init()

	// Build profile cache
	l.cacheMutex.Lock()
	l.profileCache = make(map[string]*types.Profile)
	for i := range l.profiles {
		l.profileCache[l.profiles[i].Name] = &l.profiles[i]
	}
	l.cacheMutex.Unlock()
}

// LoadProfile loads a single profile from a JSON file
func (l *Loader) LoadProfile(path string) (*types.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.ErrPathNotFound{Path: path}
		}
		if os.IsPermission(err) {
			return nil, types.ErrPermissionDenied{Path: path}
		}
		return nil, fmt.Errorf("failed to read profile file %s: %w", path, err)
	}

	var profile types.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile JSON from %s: %w", path, err)
	}

	// Validate profile
	if err := l.validateProfile(&profile); err != nil {
		return nil, fmt.Errorf("profile validation failed for %s: %w", path, err)
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
