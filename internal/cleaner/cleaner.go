package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/raucheacho/rosia-cli/internal/telemetry"
	"github.com/raucheacho/rosia-cli/internal/trash"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/types"
)

// Cleaner handles safe deletion of targets with trash backup
type Cleaner struct {
	trashSystem    *trash.System
	telemetryStore telemetry.TelemetryStore
}

// CleanOptions configures the cleaning operation
type CleanOptions struct {
	SkipConfirmation bool
	UseTrash         bool
	Concurrency      int
}

// CleanProgress reports progress during async cleaning
type CleanProgress struct {
	Current int
	Total   int
	Target  types.Target
	Error   error
}

// New creates a new Cleaner with the specified trash system
func New(trashSystem *trash.System) *Cleaner {
	return &Cleaner{
		trashSystem:    trashSystem,
		telemetryStore: nil,
	}
}

// SetTelemetryStore sets the telemetry store for the cleaner
func (c *Cleaner) SetTelemetryStore(store telemetry.TelemetryStore) {
	c.telemetryStore = store
}

// Clean safely deletes targets with confirmation and trash backup
func (c *Cleaner) Clean(ctx context.Context, targets []types.Target, opts CleanOptions) (*types.CleanReport, error) {
	startTime := time.Now()
	logger.Debug("Starting clean operation for %d targets", len(targets))

	report := &types.CleanReport{
		TotalSize:    0,
		FilesDeleted: 0,
		Errors:       []types.CleanError{},
		TrashedItems: []string{},
	}

	// Process each target
	for _, target := range targets {
		// Check context cancellation
		select {
		case <-ctx.Done():
			logger.Debug("Clean operation cancelled by context: %v", ctx.Err())
			return report, ctx.Err()
		default:
		}

		logger.Debug("Cleaning target: %s", target.Path)

		// Check permissions before deletion
		if err := c.canDelete(target.Path); err != nil {
			logger.Error("Permission check failed for %s: %v", target.Path, err)
			report.Errors = append(report.Errors, types.CleanError{
				Target: target,
				Error:  err,
			})
			continue
		}

		// Move to trash if enabled, otherwise delete directly
		if opts.UseTrash {
			// Move to trash (this also removes the file from original location)
			id, err := c.trashSystem.Move(target)
			if err != nil {
				logger.Error("Failed to move %s to trash: %v", target.Path, err)
				report.Errors = append(report.Errors, types.CleanError{
					Target: target,
					Error:  fmt.Errorf("failed to move to trash: %w", err),
				})
				continue
			}
			logger.Debug("Moved %s to trash with ID: %s", target.Path, id)
			report.TrashedItems = append(report.TrashedItems, id)
		} else {
			// Delete directly without trash backup
			if err := os.RemoveAll(target.Path); err != nil {
				logger.Error("Failed to delete %s: %v", target.Path, err)
				report.Errors = append(report.Errors, types.CleanError{
					Target: target,
					Error:  fmt.Errorf("failed to delete: %w", err),
				})
				continue
			}
			logger.Debug("Deleted %s", target.Path)
		}

		// Update report
		report.TotalSize += target.Size
		report.FilesDeleted++
	}

	report.Duration = time.Since(startTime)
	logger.Info("Clean operation completed: %d files deleted, %d errors", report.FilesDeleted, len(report.Errors))

	// Record clean events in telemetry
	if c.telemetryStore != nil {
		c.recordCleanEvents(targets, report)
	}

	return report, nil
}

// recordCleanEvents records clean events in telemetry for each profile type
func (c *Cleaner) recordCleanEvents(targets []types.Target, report *types.CleanReport) {
	// Group targets by profile to record aggregate events
	profileSizes := make(map[string]int64)
	for _, target := range targets {
		// Only count successfully cleaned targets
		wasError := false
		for _, cleanErr := range report.Errors {
			if cleanErr.Target.Path == target.Path {
				wasError = true
				break
			}
		}
		if !wasError {
			profileSizes[target.ProfileName] += target.Size
		}
	}

	// Record an event for each profile type
	for profileName, size := range profileSizes {
		event := telemetry.TelemetryEvent{
			Type:      "clean",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"size":     size,
				"profile":  profileName,
				"duration": report.Duration.Seconds(),
			},
		}

		if err := c.telemetryStore.Record(event); err != nil {
			logger.Warn("Failed to record clean telemetry for profile %s: %v", profileName, err)
		}
	}
}

// canDelete checks if the target can be safely deleted
func (c *Cleaner) canDelete(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Path does not exist: %s", path)
			return fmt.Errorf("path does not exist: %s", path)
		}
		logger.Error("Failed to stat path %s: %v", path, err)
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// Check write permission on parent directory
	parentDir := path
	if !info.IsDir() {
		parentDir = filepath.Dir(path)
	} else {
		parentDir = filepath.Dir(path)
	}

	parentInfo, err := os.Stat(parentDir)
	if err != nil {
		logger.Error("Failed to stat parent directory %s: %v", parentDir, err)
		return fmt.Errorf("failed to stat parent directory: %w", err)
	}

	// Check if parent directory is writable
	if parentInfo.Mode().Perm()&0200 == 0 {
		logger.Error("Permission denied: parent directory is not writable: %s", parentDir)
		return fmt.Errorf("permission denied: parent directory is not writable: %s", parentDir)
	}

	return nil
}

// CleanAsync performs concurrent cleaning with progress reporting
func (c *Cleaner) CleanAsync(ctx context.Context, targets []types.Target, opts CleanOptions) (<-chan CleanProgress, error) {
	progressCh := make(chan CleanProgress, 10)

	// Default concurrency if not specified
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 4 // Default to 4 workers
	}

	go func() {
		defer close(progressCh)

		// Create job channel
		jobs := make(chan struct {
			index  int
			target types.Target
		}, len(targets))

		// Create worker pool
		results := make(chan CleanProgress, len(targets))

		// Start workers
		for w := 0; w < concurrency; w++ {
			go func() {
				for job := range jobs {
					// Check context cancellation
					select {
					case <-ctx.Done():
						results <- CleanProgress{
							Current: job.index,
							Total:   len(targets),
							Target:  job.target,
							Error:   ctx.Err(),
						}
						continue
					default:
					}

					// Check permissions
					if err := c.canDelete(job.target.Path); err != nil {
						logger.Error("Permission check failed for %s: %v", job.target.Path, err)
						results <- CleanProgress{
							Current: job.index,
							Total:   len(targets),
							Target:  job.target,
							Error:   err,
						}
						continue
					}

					// Clean the target
					var cleanErr error
					if opts.UseTrash {
						_, cleanErr = c.trashSystem.Move(job.target)
						if cleanErr != nil {
							logger.Error("Failed to move %s to trash: %v", job.target.Path, cleanErr)
							cleanErr = fmt.Errorf("failed to move to trash: %w", cleanErr)
						} else {
							logger.Debug("Moved %s to trash", job.target.Path)
						}
					} else {
						cleanErr = os.RemoveAll(job.target.Path)
						if cleanErr != nil {
							logger.Error("Failed to delete %s: %v", job.target.Path, cleanErr)
							cleanErr = fmt.Errorf("failed to delete: %w", cleanErr)
						} else {
							logger.Debug("Deleted %s", job.target.Path)
						}
					}

					results <- CleanProgress{
						Current: job.index,
						Total:   len(targets),
						Target:  job.target,
						Error:   cleanErr,
					}
				}
			}()
		}

		// Send jobs
		for i, target := range targets {
			jobs <- struct {
				index  int
				target types.Target
			}{
				index:  i + 1,
				target: target,
			}
		}
		close(jobs)

		// Collect and forward results
		for i := 0; i < len(targets); i++ {
			progress := <-results
			progressCh <- progress
		}
	}()

	return progressCh, nil
}

// GenerateReportFromProgress creates a CleanReport from async progress results
func GenerateReportFromProgress(progressCh <-chan CleanProgress, startTime time.Time) *types.CleanReport {
	report := &types.CleanReport{
		TotalSize:    0,
		FilesDeleted: 0,
		Errors:       []types.CleanError{},
		TrashedItems: []string{},
	}

	for progress := range progressCh {
		if progress.Error != nil {
			report.Errors = append(report.Errors, types.CleanError{
				Target: progress.Target,
				Error:  progress.Error,
			})
		} else {
			report.TotalSize += progress.Target.Size
			report.FilesDeleted++
		}
	}

	report.Duration = time.Since(startTime)
	return report
}
