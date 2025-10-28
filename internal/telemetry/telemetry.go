// Package telemetry provides statistics tracking and reporting functionality.
//
// The telemetry system records scan and clean operations locally in ~/.rosia/stats.json,
// enabling users to track disk space savings over time. All data is stored locally
// unless the user explicitly opts in to cloud telemetry.
//
// Example usage:
//
//	store := telemetry.NewStore("~/.rosia/stats.json")
//	store.Record(telemetry.TelemetryEvent{
//	    Type: "scan",
//	    Data: map[string]interface{}{"targets_found": 42},
//	})
//	stats, _ := store.GetStats()
package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TelemetryEvent represents a single telemetry event.
//
// Events are recorded for scan and clean operations with associated metadata.
type TelemetryEvent struct {
	Type      string                 `json:"type"`      // Event type (e.g., "scan", "clean")
	Timestamp time.Time              `json:"timestamp"` // When the event occurred
	Data      map[string]interface{} `json:"data"`      // Event-specific data
}

// Stats represents aggregated telemetry statistics.
//
// Stats are computed from recorded events and provide insights into
// cleaning history and disk space savings.
type Stats struct {
	TotalScans        int              `json:"total_scans"`          // Total number of scans performed
	TotalCleaned      int64            `json:"total_cleaned"`        // Total bytes cleaned
	AverageSizeByType map[string]int64 `json:"average_size_by_type"` // Average size per target type
	LastScan          time.Time        `json:"last_scan"`            // Timestamp of last scan
	Events            []TelemetryEvent `json:"events"`               // All recorded events
}

// TelemetryStore defines the interface for telemetry operations.
//
// Implementations handle recording events and computing statistics.
type TelemetryStore interface {
	Record(event TelemetryEvent) error
	GetStats() (*Stats, error)
	Export() ([]byte, error)
}

// FileStore implements TelemetryStore using a JSON file
type FileStore struct {
	filePath string
	mu       sync.RWMutex
}

// NewFileStore creates a new FileStore instance
func NewFileStore(filePath string) (*FileStore, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create telemetry directory %s: %w", dir, err)
	}

	store := &FileStore{
		filePath: filePath,
	}

	// Initialize file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		initialStats := &Stats{
			TotalScans:        0,
			TotalCleaned:      0,
			AverageSizeByType: make(map[string]int64),
			Events:            []TelemetryEvent{},
		}
		if err := store.save(initialStats); err != nil {
			return nil, fmt.Errorf("failed to initialize telemetry file: %w", err)
		}
	}

	return store, nil
}

// Record appends a new telemetry event to the store
func (fs *FileStore) Record(event TelemetryEvent) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	stats, err := fs.load()
	if err != nil {
		return fmt.Errorf("failed to load telemetry stats: %w", err)
	}

	// Update aggregated statistics based on event type BEFORE adding to events list
	switch event.Type {
	case "scan":
		stats.TotalScans++
		if timestamp, ok := event.Data["timestamp"].(time.Time); ok {
			stats.LastScan = timestamp
		} else if timestampStr, ok := event.Data["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				stats.LastScan = t
			}
		}
	case "clean":
		if size, ok := event.Data["size"].(float64); ok {
			stats.TotalCleaned += int64(size)
		} else if size, ok := event.Data["size"].(int64); ok {
			stats.TotalCleaned += size
		}

		// Update average size by type (before adding event to list)
		if profileName, ok := event.Data["profile"].(string); ok {
			if size, ok := event.Data["size"].(float64); ok {
				fs.updateAverageSize(stats, profileName, int64(size))
			} else if size, ok := event.Data["size"].(int64); ok {
				fs.updateAverageSize(stats, profileName, size)
			}
		}
	}

	// Add event to the list AFTER updating aggregates
	stats.Events = append(stats.Events, event)

	return fs.save(stats)
}

// updateAverageSize updates the running average for a profile type
func (fs *FileStore) updateAverageSize(stats *Stats, profileName string, size int64) {
	if stats.AverageSizeByType == nil {
		stats.AverageSizeByType = make(map[string]int64)
	}

	// Simple running average calculation
	currentAvg := stats.AverageSizeByType[profileName]
	count := fs.countCleanEventsByProfile(stats, profileName)

	if count == 0 {
		stats.AverageSizeByType[profileName] = size
	} else {
		newAvg := ((currentAvg * int64(count)) + size) / int64(count+1)
		stats.AverageSizeByType[profileName] = newAvg
	}
}

// countCleanEventsByProfile counts clean events for a specific profile
func (fs *FileStore) countCleanEventsByProfile(stats *Stats, profileName string) int {
	count := 0
	for _, event := range stats.Events {
		if event.Type == "clean" {
			if profile, ok := event.Data["profile"].(string); ok && profile == profileName {
				count++
			}
		}
	}
	return count
}

// GetStats returns the current aggregated statistics
func (fs *FileStore) GetStats() (*Stats, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.load()
}

// Export returns the raw JSON data
func (fs *FileStore) Export() ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	stats, err := fs.load()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(stats, "", "  ")
}

// load reads the stats from the file
func (fs *FileStore) load() (*Stats, error) {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read telemetry file %s: %w", fs.filePath, err)
	}

	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse telemetry file %s: %w", fs.filePath, err)
	}

	// Initialize map if nil
	if stats.AverageSizeByType == nil {
		stats.AverageSizeByType = make(map[string]int64)
	}

	return &stats, nil
}

// save writes the stats to the file
func (fs *FileStore) save(stats *Stats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry stats: %w", err)
	}

	if err := os.WriteFile(fs.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write telemetry file %s: %w", fs.filePath, err)
	}

	return nil
}

// GetDefaultStatsPath returns the default path for the stats file
// Uses platform-specific paths (XDG on Linux, ~/Library on macOS, %LOCALAPPDATA% on Windows)
func GetDefaultStatsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// For backward compatibility, keep stats file in ~/.rosia
	// In the future, this could use fsutils.GetStatsFilePath() for platform-specific paths
	return filepath.Join(homeDir, ".rosia", "stats.json"), nil
}
