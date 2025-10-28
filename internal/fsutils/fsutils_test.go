package fsutils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		skipOS   string // Skip test on this OS
	}{
		{
			name:     "Unix absolute path",
			input:    "/home/user/project",
			expected: "/home/user/project",
			skipOS:   "windows",
		},
		{
			name:     "Unix path with dots",
			input:    "/home/user/../user/project",
			expected: "/home/user/project",
			skipOS:   "windows",
		},
		{
			name:     "Windows absolute path",
			input:    "C:\\Users\\user\\project",
			expected: "C:\\Users\\user\\project",
			skipOS:   "linux,darwin",
		},
		{
			name:     "Relative path",
			input:    "project/subdir",
			expected: "project/subdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS != "" && contains(tt.skipOS, runtime.GOOS) {
				t.Skip("Skipping on " + runtime.GOOS)
			}
			result := NormalizePath(tt.input)
			// Use filepath.Clean on expected to handle platform differences
			expected := filepath.Clean(tt.expected)
			assert.Equal(t, expected, result)
		})
	}
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Valid Unix path",
			path:     "/home/user/project",
			expected: true,
		},
		{
			name:     "Valid relative path",
			path:     "project/subdir",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
		skipOS   string
	}{
		{
			name:     "Unix absolute path",
			path:     "/home/user/project",
			expected: true,
			skipOS:   "windows",
		},
		{
			name:     "Unix relative path",
			path:     "project/subdir",
			expected: false,
		},
		{
			name:     "Windows absolute path",
			path:     "C:\\Users\\user",
			expected: true,
			skipOS:   "linux,darwin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS != "" && contains(tt.skipOS, runtime.GOOS) {
				t.Skip("Skipping on " + runtime.GOOS)
			}
			result := IsAbsolutePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinPath(t *testing.T) {
	result := JoinPath("home", "user", "project")
	expected := filepath.Join("home", "user", "project")
	assert.Equal(t, expected, result)
}

func TestSplitPath(t *testing.T) {
	dir, file := SplitPath("/home/user/file.txt")
	assert.Equal(t, "/home/user/", dir)
	assert.Equal(t, "file.txt", file)
}

func TestGetPathSeparator(t *testing.T) {
	sep := GetPathSeparator()
	assert.Equal(t, string(filepath.Separator), sep)
}

func TestCanDelete(t *testing.T) {
	t.Run("Non-existent path", func(t *testing.T) {
		err := CanDelete("/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("Valid writable path", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		err := CanDelete(testFile)
		assert.NoError(t, err)
	})

	t.Run("Valid writable directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "testdir")
		require.NoError(t, os.MkdirAll(testDir, 0755))

		err := CanDelete(testDir)
		assert.NoError(t, err)
	})
}

func TestIsHidden(t *testing.T) {
	t.Run("Hidden file (dot prefix)", func(t *testing.T) {
		tmpDir := t.TempDir()
		hiddenFile := filepath.Join(tmpDir, ".hidden")
		require.NoError(t, os.WriteFile(hiddenFile, []byte("test"), 0644))

		hidden, err := IsHidden(hiddenFile)
		assert.NoError(t, err)
		assert.True(t, hidden)
	})

	t.Run("Non-hidden file", func(t *testing.T) {
		tmpDir := t.TempDir()
		normalFile := filepath.Join(tmpDir, "normal.txt")
		require.NoError(t, os.WriteFile(normalFile, []byte("test"), 0644))

		hidden, err := IsHidden(normalFile)
		assert.NoError(t, err)
		assert.False(t, hidden)
	})
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")

	err := EnsureDir(newDir)
	assert.NoError(t, err)

	// Verify directory exists
	info, err := os.Stat(newDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestIsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows")
	}

	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	symlinkFile := filepath.Join(tmpDir, "link.txt")

	require.NoError(t, os.WriteFile(targetFile, []byte("test"), 0644))
	require.NoError(t, os.Symlink(targetFile, symlinkFile))

	t.Run("Symlink", func(t *testing.T) {
		isLink, err := IsSymlink(symlinkFile)
		assert.NoError(t, err)
		assert.True(t, isLink)
	})

	t.Run("Regular file", func(t *testing.T) {
		isLink, err := IsSymlink(targetFile)
		assert.NoError(t, err)
		assert.False(t, isLink)
	})
}

func TestResolveSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows")
	}

	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	symlinkFile := filepath.Join(tmpDir, "link.txt")

	require.NoError(t, os.WriteFile(targetFile, []byte("test"), 0644))
	require.NoError(t, os.Symlink(targetFile, symlinkFile))

	resolved, err := ResolveSymlink(symlinkFile)
	assert.NoError(t, err)

	// On macOS, /var is a symlink to /private/var, so resolve both paths for comparison
	expectedResolved, err := filepath.EvalSymlinks(targetFile)
	require.NoError(t, err)
	assert.Equal(t, expectedResolved, resolved)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for _, part := range splitString(s, ",") {
		if part == substr {
			return true
		}
	}
	return false
}

func splitString(s, sep string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if string(c) == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
