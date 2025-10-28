// Package types provides core data types and error definitions used throughout Rosia CLI.
//
// This package defines the fundamental structures for representing cleanable targets,
// technology profiles, configuration settings, and custom error types. These types
// are used across all components of the application.
//
// Example usage:
//
//	target := types.Target{
//	    Path:        "/path/to/node_modules",
//	    Size:        524288000,
//	    Type:        "dependency",
//	    ProfileName: "node",
//	    IsDirectory: true,
//	}
package types

import "time"

// Target represents a cleanable file or directory detected during scanning.
//
// A Target contains all information needed to identify, display, and clean
// a specific file or directory. Targets are created by the scanner engine
// when matching profile patterns.
type Target struct {
	Path         string    // Absolute path to the target file or directory
	Size         int64     // Total size in bytes
	Type         string    // Type classification (e.g., "dependency", "build", "cache")
	ProfileName  string    // Name of the profile that matched this target
	LastAccessed time.Time // Last access timestamp
	IsDirectory  bool      // True if target is a directory
}

// Profile defines cleaning rules and detection patterns for a specific technology stack.
//
// Profiles are loaded from JSON files in the profiles/ directory and define:
//   - Patterns: directories/files to clean (supports glob patterns)
//   - Detect: files that indicate the technology is present
//
// Example profile for Node.js:
//
//	{
//	  "name": "Node.js",
//	  "patterns": ["node_modules", "dist", ".next"],
//	  "detect": ["package.json"],
//	  "enabled": true
//	}
type Profile struct {
	Name        string   `json:"name"`        // Display name of the technology
	Version     string   `json:"version"`     // Profile version (semver)
	Patterns    []string `json:"patterns"`    // Glob patterns for files/directories to clean
	Detect      []string `json:"detect"`      // Files that indicate technology presence
	Description string   `json:"description"` // Human-readable description
	Enabled     bool     `json:"enabled"`     // Whether profile is enabled
}

// Config represents user configuration loaded from ~/.rosiarc.json.
//
// The configuration file allows users to customize Rosia's behavior including
// trash retention, enabled profiles, ignored paths, and performance settings.
//
// Example configuration:
//
//	{
//	  "trash_retention_days": 3,
//	  "profiles": ["node", "python", "rust"],
//	  "ignore_paths": ["/usr/local"],
//	  "concurrency": 8
//	}
type Config struct {
	TrashRetentionDays int      `json:"trash_retention_days"` // Days to keep items in trash
	Profiles           []string `json:"profiles"`             // Enabled profile names
	IgnorePaths        []string `json:"ignore_paths"`         // Paths to exclude from scanning
	Plugins            []string `json:"plugins"`              // Enabled plugin names
	Concurrency        int      `json:"concurrency"`          // Worker pool size (0 = auto)
	TelemetryEnabled   bool     `json:"telemetry_enabled"`    // Enable anonymous statistics
}

// CleanReport summarizes the results of a cleaning operation.
//
// The report includes statistics about deleted files, total space reclaimed,
// any errors encountered, and items moved to trash for potential restoration.
type CleanReport struct {
	TotalSize    int64         // Total bytes deleted
	FilesDeleted int           // Number of files/directories deleted
	Errors       []CleanError  // Errors encountered during cleaning
	Duration     time.Duration // Time taken to complete operation
	TrashedItems []string      // IDs of items moved to trash
}

// CleanError represents an error that occurred while cleaning a specific target.
//
// Errors are isolated per target, allowing the cleaning operation to continue
// even if individual targets fail.
type CleanError struct {
	Target Target // The target that failed to clean
	Error  error  // The error that occurred
}

// TrashMetadata stores information about trashed items for restoration.
//
// Metadata is persisted as JSON alongside trashed items in ~/.rosia/trash/
// and enables restoration to the original location.
type TrashMetadata struct {
	ID           string    `json:"id"`            // Unique identifier (timestamp-based)
	OriginalPath string    `json:"original_path"` // Original location before deletion
	Size         int64     `json:"size"`          // Size in bytes
	DeletedAt    time.Time `json:"deleted_at"`    // Deletion timestamp
	ProfileName  string    `json:"profile_name"`  // Profile that matched this item
}

// TrashItem represents a trashed item with its metadata and current location.
//
// TrashItems are returned by the trash system's List() method and include
// both the metadata and the current trash path.
type TrashItem struct {
	ID           string    // Unique identifier
	OriginalPath string    // Original location
	Size         int64     // Size in bytes
	DeletedAt    time.Time // Deletion timestamp
	TrashPath    string    // Current location in trash
}

// ErrPermissionDenied indicates insufficient permissions to access or modify a path.
//
// This error is returned when the user lacks the necessary permissions to
// read, write, or delete a file or directory.
type ErrPermissionDenied struct {
	Path string // The path that caused the permission error
}

// Error implements the error interface.
func (e ErrPermissionDenied) Error() string {
	return "permission denied: " + e.Path
}

// ErrPathNotFound indicates a non-existent path.
//
// This error is returned when attempting to scan, clean, or restore a path
// that does not exist in the filesystem.
type ErrPathNotFound struct {
	Path string // The path that was not found
}

// Error implements the error interface.
func (e ErrPathNotFound) Error() string {
	return "path not found: " + e.Path
}

// ErrTrashFull indicates the trash directory has exceeded its size limit.
//
// This error is returned when attempting to move items to trash would exceed
// the configured maximum trash size.
type ErrTrashFull struct {
	CurrentSize int64 // Current trash directory size in bytes
	MaxSize     int64 // Maximum allowed trash size in bytes
}

// Error implements the error interface.
func (e ErrTrashFull) Error() string {
	return "trash directory is full"
}

// ErrPluginLoadFailed indicates a plugin failed to load.
//
// This error is returned when a plugin cannot be loaded due to missing files,
// invalid format, or runtime errors during initialization.
type ErrPluginLoadFailed struct {
	PluginName string // Name of the plugin that failed to load
	Reason     error  // Underlying error that caused the failure
}

// Error implements the error interface.
func (e ErrPluginLoadFailed) Error() string {
	if e.Reason != nil {
		return "failed to load plugin '" + e.PluginName + "': " + e.Reason.Error()
	}
	return "failed to load plugin '" + e.PluginName + "'"
}

// Unwrap returns the underlying error for error chain inspection.
func (e ErrPluginLoadFailed) Unwrap() error {
	return e.Reason
}
