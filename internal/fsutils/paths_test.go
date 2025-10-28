package fsutils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDir(t *testing.T) {
	// Save original env vars
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupEnv    func()
		expectedDir string
		skipOS      string
	}{
		{
			name: "Linux with XDG_CONFIG_HOME",
			setupEnv: func() {
				os.Setenv("XDG_CONFIG_HOME", "/custom/config")
			},
			expectedDir: "/custom/config/rosia",
			skipOS:      "darwin,windows",
		},
		{
			name: "Linux without XDG_CONFIG_HOME",
			setupEnv: func() {
				os.Unsetenv("XDG_CONFIG_HOME")
			},
			expectedDir: filepath.Join(homeDir, ".config", "rosia"),
			skipOS:      "darwin,windows",
		},
		{
			name:        "macOS",
			setupEnv:    func() {},
			expectedDir: filepath.Join(homeDir, "Library", "Application Support", "rosia"),
			skipOS:      "linux,windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS != "" && contains(tt.skipOS, runtime.GOOS) {
				t.Skip("Skipping on " + runtime.GOOS)
			}

			tt.setupEnv()
			configDir, err := GetConfigDir()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedDir, configDir)
		})
	}
}

func TestGetConfigFilePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	configPath, err := GetConfigFilePath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(homeDir, ".rosiarc.json"), configPath)
}

func TestGetDataDir(t *testing.T) {
	// Save original env vars
	originalXDGData := os.Getenv("XDG_DATA_HOME")
	originalLocalAppData := os.Getenv("LOCALAPPDATA")
	defer func() {
		os.Setenv("XDG_DATA_HOME", originalXDGData)
		os.Setenv("LOCALAPPDATA", originalLocalAppData)
	}()

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupEnv    func()
		expectedDir string
		skipOS      string
	}{
		{
			name: "Linux with XDG_DATA_HOME",
			setupEnv: func() {
				os.Setenv("XDG_DATA_HOME", "/custom/data")
			},
			expectedDir: "/custom/data/rosia",
			skipOS:      "darwin,windows",
		},
		{
			name: "Linux without XDG_DATA_HOME",
			setupEnv: func() {
				os.Unsetenv("XDG_DATA_HOME")
			},
			expectedDir: filepath.Join(homeDir, ".local", "share", "rosia"),
			skipOS:      "darwin,windows",
		},
		{
			name:        "macOS",
			setupEnv:    func() {},
			expectedDir: filepath.Join(homeDir, "Library", "Application Support", "rosia"),
			skipOS:      "linux,windows",
		},
		{
			name: "Windows with LOCALAPPDATA",
			setupEnv: func() {
				os.Setenv("LOCALAPPDATA", "C:\\Users\\Test\\AppData\\Local")
			},
			expectedDir: "C:\\Users\\Test\\AppData\\Local\\rosia",
			skipOS:      "linux,darwin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS != "" && contains(tt.skipOS, runtime.GOOS) {
				t.Skip("Skipping on " + runtime.GOOS)
			}

			tt.setupEnv()
			dataDir, err := GetDataDir()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedDir, dataDir)
		})
	}
}

func TestGetTrashDir(t *testing.T) {
	trashDir, err := GetTrashDir()
	assert.NoError(t, err)
	assert.Contains(t, trashDir, "trash")

	// Verify it's under the data directory
	dataDir, err := GetDataDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dataDir, "trash"), trashDir)
}

func TestGetPluginsDir(t *testing.T) {
	pluginsDir, err := GetPluginsDir()
	assert.NoError(t, err)
	assert.Contains(t, pluginsDir, "plugins")

	// Verify it's under the data directory
	dataDir, err := GetDataDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dataDir, "plugins"), pluginsDir)
}

func TestGetStatsFilePath(t *testing.T) {
	statsPath, err := GetStatsFilePath()
	assert.NoError(t, err)
	assert.Contains(t, statsPath, "stats.json")

	// Verify it's under the data directory
	dataDir, err := GetDataDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dataDir, "stats.json"), statsPath)
}

func TestGetCacheDir(t *testing.T) {
	// Save original env vars
	originalXDGCache := os.Getenv("XDG_CACHE_HOME")
	originalLocalAppData := os.Getenv("LOCALAPPDATA")
	defer func() {
		os.Setenv("XDG_CACHE_HOME", originalXDGCache)
		os.Setenv("LOCALAPPDATA", originalLocalAppData)
	}()

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupEnv    func()
		expectedDir string
		skipOS      string
	}{
		{
			name: "Linux with XDG_CACHE_HOME",
			setupEnv: func() {
				os.Setenv("XDG_CACHE_HOME", "/custom/cache")
			},
			expectedDir: "/custom/cache/rosia",
			skipOS:      "darwin,windows",
		},
		{
			name: "Linux without XDG_CACHE_HOME",
			setupEnv: func() {
				os.Unsetenv("XDG_CACHE_HOME")
			},
			expectedDir: filepath.Join(homeDir, ".cache", "rosia"),
			skipOS:      "darwin,windows",
		},
		{
			name:        "macOS",
			setupEnv:    func() {},
			expectedDir: filepath.Join(homeDir, "Library", "Caches", "rosia"),
			skipOS:      "linux,windows",
		},
		{
			name: "Windows with LOCALAPPDATA",
			setupEnv: func() {
				os.Setenv("LOCALAPPDATA", "C:\\Users\\Test\\AppData\\Local")
			},
			expectedDir: "C:\\Users\\Test\\AppData\\Local\\rosia\\cache",
			skipOS:      "linux,darwin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS != "" && contains(tt.skipOS, runtime.GOOS) {
				t.Skip("Skipping on " + runtime.GOOS)
			}

			tt.setupEnv()
			cacheDir, err := GetCacheDir()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedDir, cacheDir)
		})
	}
}

func TestPlatformSpecificPaths(t *testing.T) {
	// This test verifies that all path functions return valid paths
	t.Run("All paths are valid", func(t *testing.T) {
		paths := []struct {
			name string
			fn   func() (string, error)
		}{
			{"ConfigDir", GetConfigDir},
			{"ConfigFilePath", GetConfigFilePath},
			{"DataDir", GetDataDir},
			{"TrashDir", GetTrashDir},
			{"PluginsDir", GetPluginsDir},
			{"StatsFilePath", GetStatsFilePath},
			{"CacheDir", GetCacheDir},
		}

		for _, p := range paths {
			t.Run(p.name, func(t *testing.T) {
				path, err := p.fn()
				assert.NoError(t, err)
				assert.NotEmpty(t, path)
				assert.True(t, IsValidPath(path), "Path should be valid: %s", path)
			})
		}
	})
}
