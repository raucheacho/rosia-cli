package types

import "time"

// Target represents a cleanable file or directory
type Target struct {
	Path         string
	Size         int64
	Type         string
	ProfileName  string
	LastAccessed time.Time
	IsDirectory  bool
}

// Profile defines cleaning rules for a technology
type Profile struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Patterns    []string `json:"patterns"`
	Detect      []string `json:"detect"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
}

// Config represents user configuration
type Config struct {
	TrashRetentionDays int      `json:"trash_retention_days"`
	Profiles           []string `json:"profiles"`
	IgnorePaths        []string `json:"ignore_paths"`
	Plugins            []string `json:"plugins"`
	Concurrency        int      `json:"concurrency"`
	TelemetryEnabled   bool     `json:"telemetry_enabled"`
}

// CleanReport summarizes a cleaning operation
type CleanReport struct {
	TotalSize    int64
	FilesDeleted int
	Errors       []CleanError
	Duration     time.Duration
	TrashedItems []string
}

// CleanError represents an error during cleaning
type CleanError struct {
	Target Target
	Error  error
}

// TrashMetadata stores information about trashed items
type TrashMetadata struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	Size         int64     `json:"size"`
	DeletedAt    time.Time `json:"deleted_at"`
	ProfileName  string    `json:"profile_name"`
}

// TrashItem represents a trashed item with its metadata
type TrashItem struct {
	ID           string
	OriginalPath string
	Size         int64
	DeletedAt    time.Time
	TrashPath    string
}
