package profiles

import (
	"os"
	"path/filepath"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// MatchProfile detects the technology type by checking detect patterns
// Returns the first matching profile or nil if no match found
func (l *Loader) MatchProfile(dirPath string) (*types.Profile, error) {
	// Check cache first
	l.cacheMutex.RLock()
	if cached, exists := l.matchCache[dirPath]; exists {
		l.cacheMutex.RUnlock()
		return cached, nil
	}
	l.cacheMutex.RUnlock()

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	// Try to match against each profile
	for i := range l.profiles {
		profile := &l.profiles[i]

		// Skip disabled profiles
		if !profile.Enabled {
			continue
		}

		// Check if any detect pattern matches
		if l.matchesDetectPatterns(dirPath, profile.Detect) {
			// Cache the result
			l.cacheMutex.Lock()
			l.matchCache[dirPath] = profile
			l.cacheMutex.Unlock()

			return profile, nil
		}
	}

	// No match found, cache nil result
	l.cacheMutex.Lock()
	l.matchCache[dirPath] = nil
	l.cacheMutex.Unlock()

	return nil, nil
}

// matchesDetectPatterns checks if any detect pattern exists in the directory
func (l *Loader) matchesDetectPatterns(dirPath string, detectPatterns []string) bool {
	for _, pattern := range detectPatterns {
		// Check if file/directory exists in the directory
		targetPath := filepath.Join(dirPath, pattern)
		if _, err := os.Stat(targetPath); err == nil {
			return true
		}

		// Also try glob matching for patterns with wildcards
		if hasGlobChars(pattern) {
			matches, err := filepath.Glob(filepath.Join(dirPath, pattern))
			if err == nil && len(matches) > 0 {
				return true
			}
		}
	}

	return false
}

// MatchesPattern checks if a file or directory name matches any of the profile's patterns
func (l *Loader) MatchesPattern(name string, profile *types.Profile) bool {
	for _, pattern := range profile.Patterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}

		// Also check if the name contains the pattern (for paths like "node_modules")
		if name == pattern {
			return true
		}
	}

	return false
}

// hasGlobChars checks if a string contains glob wildcard characters
func hasGlobChars(s string) bool {
	return containsAny(s, "*?[]")
}

// containsAny checks if string contains any of the characters
func containsAny(s string, chars string) bool {
	for _, c := range chars {
		for _, sc := range s {
			if c == sc {
				return true
			}
		}
	}
	return false
}

// ClearCache clears the match cache
func (l *Loader) ClearCache() {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()
	l.matchCache = make(map[string]*types.Profile)
}
