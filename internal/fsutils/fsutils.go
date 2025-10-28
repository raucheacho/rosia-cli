// Package fsutils provides cross-platform file system utilities.
//
// This package handles platform-specific file operations including path
// normalization, permission checks, and safe file operations. It ensures
// consistent behavior across Linux, macOS, and Windows.
//
// Example usage:
//
//	path := fsutils.NormalizePath("~/projects/app")
//	if fsutils.CanDelete(path) {
//	    fsutils.RemoveAll(path)
//	}
package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizePath normalizes a path for the current platform.
//
// It handles Windows path formats (C:\, UNC paths) and Unix paths,
// ensuring consistent path representation across platforms.
func NormalizePath(path string) string {
	// Clean the path using filepath which handles platform-specific separators
	cleaned := filepath.Clean(path)

	// On Windows, ensure proper handling of UNC paths
	if runtime.GOOS == "windows" {
		// UNC paths start with \\
		if strings.HasPrefix(path, "\\\\") && !strings.HasPrefix(cleaned, "\\\\") {
			cleaned = "\\\\" + strings.TrimPrefix(cleaned, "\\")
		}
	}

	return cleaned
}

// IsValidPath checks if a path is valid for the current platform
func IsValidPath(path string) bool {
	if path == "" {
		return false
	}

	// Check for invalid characters based on platform
	if runtime.GOOS == "windows" {
		// Windows invalid characters: < > : " | ? *
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		// Extract just the path part (not the drive letter)
		pathPart := path
		if len(path) >= 2 && path[1] == ':' {
			pathPart = path[2:]
		}
		for _, char := range invalidChars {
			if strings.Contains(pathPart, char) {
				return false
			}
		}
	}

	return true
}

// IsAbsolutePath checks if a path is absolute for the current platform
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// JoinPath joins path elements using the platform-specific separator
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// SplitPath splits a path into directory and file components
func SplitPath(path string) (dir, file string) {
	return filepath.Split(path)
}

// GetPathSeparator returns the platform-specific path separator
func GetPathSeparator() string {
	return string(filepath.Separator)
}

// CanDelete checks if a path can be deleted (has write permissions)
func CanDelete(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// Check write permission on parent directory
	parent := filepath.Dir(path)
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("failed to stat parent directory: %w", err)
	}

	// Platform-specific permission checks
	if runtime.GOOS == "windows" {
		// On Windows, check if we can open the file for writing
		// This is a simplified check; full Windows ACL checking would be more complex
		if info.IsDir() {
			// For directories, try to create a temp file inside
			testFile := filepath.Join(path, ".rosia_test_write")
			f, err := os.Create(testFile)
			if err != nil {
				return fmt.Errorf("no write permission: %w", err)
			}
			f.Close()
			os.Remove(testFile)
		} else {
			// For files, check parent directory
			testFile := filepath.Join(parent, ".rosia_test_write")
			f, err := os.Create(testFile)
			if err != nil {
				return fmt.Errorf("no write permission on parent directory: %w", err)
			}
			f.Close()
			os.Remove(testFile)
		}
	} else {
		// On Unix-like systems, check permission bits
		if parentInfo.Mode().Perm()&0200 == 0 {
			return fmt.Errorf("no write permission on parent directory: %s", parent)
		}
	}

	return nil
}

// IsHidden checks if a file or directory is hidden
func IsHidden(path string) (bool, error) {
	// Get the base name
	base := filepath.Base(path)

	if runtime.GOOS == "windows" {
		// On Windows, check the hidden attribute
		_, err := os.Stat(path)
		if err != nil {
			return false, err
		}

		// Check if file has hidden attribute (this is a simplified check)
		// Full implementation would use syscall to check FILE_ATTRIBUTE_HIDDEN
		// For now, also treat dot-prefixed files as hidden on Windows
		if strings.HasPrefix(base, ".") {
			return true, nil
		}

		// Try to get Windows attributes
		attrs, err := getWindowsAttributes(path)
		if err == nil && attrs&0x2 != 0 { // FILE_ATTRIBUTE_HIDDEN = 0x2
			return true, nil
		}

		return false, nil
	}

	// On Unix-like systems, files starting with . are hidden
	return strings.HasPrefix(base, "."), nil
}

// getWindowsAttributes attempts to get Windows file attributes
// Returns 0 if not on Windows or if unable to get attributes
func getWindowsAttributes(path string) (uint32, error) {
	if runtime.GOOS != "windows" {
		return 0, fmt.Errorf("not on Windows")
	}

	// This is a placeholder - full implementation would use:
	// syscall.GetFileAttributes on Windows
	// For now, return 0 to indicate no special attributes
	return 0, fmt.Errorf("Windows attribute checking not fully implemented")
}

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

// ResolveSymlink resolves a symbolic link to its target
func ResolveSymlink(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
