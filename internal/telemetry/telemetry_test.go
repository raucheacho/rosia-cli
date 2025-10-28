package telemetry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStore(t *testing.T) {
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats.json")

	store, err := NewFileStore(statsPath)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Verify file was created
	_, err = os.Stat(statsPath)
	assert.NoError(t, err)

	// Verify initial stats
	stats, err := store.GetStats()
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalScans)
	assert.Equal(t, int64(0), stats.TotalCleaned)
	assert.NotNil(t, stats.AverageSizeByType)
	assert.NotNil(t, stats.Events)
}

func TestFileStore_RecordScanEvent(t *testing.T) {
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats.json")

	store, err := NewFileStore(statsPath)
	require.NoError(t, err)

	now := time.Now()
	event := TelemetryEvent{
		Type:      "scan",
		Timestamp: now,
		Data: map[string]interface{}{
			"timestamp":     now,
			"targets_found": 5,
		},
	}

	err = store.Record(event)
	require.NoError(t, err)

	stats, err := store.GetStats()
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalScans)
	assert.Equal(t, 1, len(stats.Events))
	assert.WithinDuration(t, now, stats.LastScan, time.Second)
}

func TestFileStore_RecordCleanEvent(t *testing.T) {
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats.json")

	store, err := NewFileStore(statsPath)
	require.NoError(t, err)

	event := TelemetryEvent{
		Type:      "clean",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"size":    int64(1024000),
			"profile": "node",
		},
	}

	err = store.Record(event)
	require.NoError(t, err)

	stats, err := store.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(1024000), stats.TotalCleaned)
	assert.Equal(t, int64(1024000), stats.AverageSizeByType["node"])
}

func TestFileStore_MultipleCleanEvents(t *testing.T) {
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats.json")

	store, err := NewFileStore(statsPath)
	require.NoError(t, err)

	// Record first clean event
	event1 := TelemetryEvent{
		Type:      "clean",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"size":    int64(1000),
			"profile": "node",
		},
	}
	err = store.Record(event1)
	require.NoError(t, err)

	// Record second clean event
	event2 := TelemetryEvent{
		Type:      "clean",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"size":    int64(2000),
			"profile": "node",
		},
	}
	err = store.Record(event2)
	require.NoError(t, err)

	stats, err := store.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(3000), stats.TotalCleaned)
	// Average should be (1000 + 2000) / 2 = 1500
	assert.Equal(t, int64(1500), stats.AverageSizeByType["node"])
}

func TestFileStore_Export(t *testing.T) {
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats.json")

	store, err := NewFileStore(statsPath)
	require.NoError(t, err)

	event := TelemetryEvent{
		Type:      "scan",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"targets_found": 3,
		},
	}
	err = store.Record(event)
	require.NoError(t, err)

	data, err := store.Export()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "total_scans")
}

func TestGetDefaultStatsPath(t *testing.T) {
	path, err := GetDefaultStatsPath()
	require.NoError(t, err)
	assert.Contains(t, path, ".rosia")
	assert.Contains(t, path, "stats.json")
}
