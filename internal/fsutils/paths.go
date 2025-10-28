package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the platform-specific configuration directory
// - Linux: $XDG_CONFIG_HOME/rosia or ~/.config/rosia
// - macOS: ~/Library/Application Support/rosia
// - Windows: %APPDATA%/rosia
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "linux":
		// Check XDG_CONFIG_HOME first
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, "rosia")
		} else {
			// Fall back to ~/.config/rosia
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			configDir = filepath.Join(homeDir, ".config", "rosia")
		}

	case "darwin":
		// macOS: ~/Library/Application Support/rosia
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "rosia")

	case "windows":
		// Windows: %APPDATA%/rosia
		if appData := os.Getenv("APPDATA"); appData != "" {
			configDir = filepath.Join(appData, "rosia")
		} else {
			// Fall back to user home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			configDir = filepath.Join(homeDir, "AppData", "Roaming", "rosia")
		}

	default:
		// Default to ~/.rosia for unknown platforms
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".rosia")
	}

	return configDir, nil
}

// GetConfigFilePath returns the full path to the configuration file
func GetConfigFilePath() (string, error) {
	// For backward compatibility, keep config file in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".rosiarc.json"), nil
}

// GetDataDir returns the platform-specific data directory
// - Linux: $XDG_DATA_HOME/rosia or ~/.local/share/rosia
// - macOS: ~/Library/Application Support/rosia
// - Windows: %LOCALAPPDATA%/rosia
func GetDataDir() (string, error) {
	var dataDir string

	switch runtime.GOOS {
	case "linux":
		// Check XDG_DATA_HOME first
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			dataDir = filepath.Join(xdgData, "rosia")
		} else {
			// Fall back to ~/.local/share/rosia
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			dataDir = filepath.Join(homeDir, ".local", "share", "rosia")
		}

	case "darwin":
		// macOS: ~/Library/Application Support/rosia
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		dataDir = filepath.Join(homeDir, "Library", "Application Support", "rosia")

	case "windows":
		// Windows: %LOCALAPPDATA%/rosia
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			dataDir = filepath.Join(localAppData, "rosia")
		} else {
			// Fall back to user home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			dataDir = filepath.Join(homeDir, "AppData", "Local", "rosia")
		}

	default:
		// Default to ~/.rosia for unknown platforms
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		dataDir = filepath.Join(homeDir, ".rosia")
	}

	return dataDir, nil
}

// GetTrashDir returns the platform-specific trash directory
func GetTrashDir() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "trash"), nil
}

// GetPluginsDir returns the platform-specific plugins directory
func GetPluginsDir() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "plugins"), nil
}

// GetStatsFilePath returns the platform-specific stats file path
func GetStatsFilePath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "stats.json"), nil
}

// GetCacheDir returns the platform-specific cache directory
// - Linux: $XDG_CACHE_HOME/rosia or ~/.cache/rosia
// - macOS: ~/Library/Caches/rosia
// - Windows: %LOCALAPPDATA%/rosia/cache
func GetCacheDir() (string, error) {
	var cacheDir string

	switch runtime.GOOS {
	case "linux":
		// Check XDG_CACHE_HOME first
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			cacheDir = filepath.Join(xdgCache, "rosia")
		} else {
			// Fall back to ~/.cache/rosia
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			cacheDir = filepath.Join(homeDir, ".cache", "rosia")
		}

	case "darwin":
		// macOS: ~/Library/Caches/rosia
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, "Library", "Caches", "rosia")

	case "windows":
		// Windows: %LOCALAPPDATA%/rosia/cache
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			cacheDir = filepath.Join(localAppData, "rosia", "cache")
		} else {
			// Fall back to user home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			cacheDir = filepath.Join(homeDir, "AppData", "Local", "rosia", "cache")
		}

	default:
		// Default to ~/.rosia/cache for unknown platforms
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".rosia", "cache")
	}

	return cacheDir, nil
}
