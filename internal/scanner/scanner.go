package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/raucheacho/rosia-cli/internal/profiles"
	"github.com/raucheacho/rosia-cli/internal/sizecalc"
	"github.com/raucheacho/rosia-cli/internal/telemetry"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/raucheacho/rosia-cli/pkg/types"
)

// Scanner handles directory scanning and target detection
type Scanner struct {
	profileLoader  *profiles.Loader
	sizeCalc       *sizecalc.SizeCalc
	telemetryStore telemetry.TelemetryStore
}

// ScanOptions configures the scanning behavior
type ScanOptions struct {
	MaxDepth      int
	IncludeHidden bool
	IgnorePaths   []string
	DryRun        bool
	Concurrency   int
}

// NewScanner creates a new scanner with the given profile loader
func NewScanner(loader *profiles.Loader) *Scanner {
	return &Scanner{
		profileLoader:  loader,
		sizeCalc:       sizecalc.NewSizeCalc(0), // 0 means auto-detect concurrency
		telemetryStore: nil,
	}
}

// NewScannerWithSizeCalc creates a new scanner with a custom size calculator
func NewScannerWithSizeCalc(loader *profiles.Loader, sizeCalc *sizecalc.SizeCalc) *Scanner {
	return &Scanner{
		profileLoader:  loader,
		sizeCalc:       sizeCalc,
		telemetryStore: nil,
	}
}

// SetTelemetryStore sets the telemetry store for the scanner
func (s *Scanner) SetTelemetryStore(store telemetry.TelemetryStore) {
	s.telemetryStore = store
}

// Scan performs a synchronous scan of the given paths
func (s *Scanner) Scan(ctx context.Context, paths []string, opts ScanOptions) ([]types.Target, error) {
	targets := make([]types.Target, 0)

	for _, path := range paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			logger.Debug("Scan cancelled by context: %v", ctx.Err())
			return targets, ctx.Err()
		default:
		}

		// Scan this path
		logger.Debug("Scanning path: %s", path)
		pathTargets, err := s.scanPath(ctx, path, opts)
		if err != nil {
			logger.Error("Failed to scan path %s: %v", path, err)
			return targets, fmt.Errorf("failed to scan path %s: %w", path, err)
		}

		logger.Debug("Found %d targets in path: %s", len(pathTargets), path)
		targets = append(targets, pathTargets...)
	}

	// Calculate sizes for all targets
	if len(targets) > 0 {
		logger.Debug("Calculating sizes for %d targets", len(targets))
		targets, err := s.sizeCalc.CalculateTargets(ctx, targets)
		if err != nil {
			logger.Error("Failed to calculate sizes: %v", err)
			return targets, fmt.Errorf("failed to calculate sizes: %w", err)
		}

		// Record scan event in telemetry
		if s.telemetryStore != nil {
			s.recordScanEvent(len(targets))
		}

		return targets, nil
	}

	logger.Debug("No targets found")

	// Record scan event even if no targets found
	if s.telemetryStore != nil {
		s.recordScanEvent(0)
	}

	return targets, nil
}

// recordScanEvent records a scan event in telemetry
func (s *Scanner) recordScanEvent(targetsFound int) {
	event := telemetry.TelemetryEvent{
		Type:      "scan",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"timestamp":     time.Now(),
			"targets_found": targetsFound,
		},
	}

	if err := s.telemetryStore.Record(event); err != nil {
		logger.Warn("Failed to record scan telemetry: %v", err)
	}
}

// scanPath scans a single path recursively
func (s *Scanner) scanPath(ctx context.Context, rootPath string, opts ScanOptions) ([]types.Target, error) {
	targets := make([]types.Target, 0)
	rootDepth := strings.Count(rootPath, string(os.PathSeparator))

	// First, try to match the root directory itself
	profile, err := s.profileLoader.MatchProfile(rootPath)
	if err == nil && profile != nil {
		// Check if root path matches any patterns
		baseName := filepath.Base(rootPath)
		if s.profileLoader.MatchesPattern(baseName, profile) {
			target, err := s.createTarget(rootPath, profile)
			if err == nil {
				targets = append(targets, target)
			}
		}
	}

	// Walk the directory tree
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Log error but continue walking
			logger.Warn("Error accessing path %s: %v", path, err)
			return nil
		}

		// Skip the root path itself (already checked above)
		if path == rootPath {
			return nil
		}

		// Check depth limit
		if opts.MaxDepth > 0 {
			currentDepth := strings.Count(path, string(os.PathSeparator))
			if currentDepth-rootDepth > opts.MaxDepth {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}

		// Skip hidden files/directories unless IncludeHidden is true
		if !opts.IncludeHidden && isHidden(d.Name()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Check if path should be ignored
		if s.shouldIgnore(path, opts.IgnorePaths) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Only process directories for profile matching
		if !d.IsDir() {
			return nil
		}

		// Get the parent directory for profile matching
		parentDir := filepath.Dir(path)
		profile, err := s.profileLoader.MatchProfile(parentDir)
		if err != nil {
			// Continue on error
			return nil
		}

		// If no profile matched the parent, try matching the current directory
		if profile == nil {
			profile, err = s.profileLoader.MatchProfile(path)
			if err != nil {
				return nil
			}
		}

		// If we have a profile, check if this directory matches any patterns
		if profile != nil {
			baseName := d.Name()
			if s.profileLoader.MatchesPattern(baseName, profile) {
				target, err := s.createTarget(path, profile)
				if err == nil {
					targets = append(targets, target)
					// Skip descending into matched directories
					return fs.SkipDir
				}
			}
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		return targets, fmt.Errorf("error walking directory: %w", err)
	}

	return targets, nil
}

// createTarget creates a Target from a path and profile
func (s *Scanner) createTarget(path string, profile *types.Profile) (types.Target, error) {
	info, err := os.Stat(path)
	if err != nil {
		return types.Target{}, err
	}

	target := types.Target{
		Path:         path,
		Type:         profile.Name,
		ProfileName:  profile.Name,
		IsDirectory:  info.IsDir(),
		LastAccessed: getLastAccessTime(info),
		Size:         0, // Will be calculated later by SizeCalc
	}

	return target, nil
}

// shouldIgnore checks if a path should be ignored based on ignore patterns
func (s *Scanner) shouldIgnore(path string, ignorePaths []string) bool {
	for _, ignorePath := range ignorePaths {
		// Check for exact match or prefix match
		if path == ignorePath || strings.HasPrefix(path, ignorePath+string(os.PathSeparator)) {
			return true
		}

		// Check for glob pattern match
		matched, err := filepath.Match(ignorePath, path)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// isHidden checks if a file or directory name is hidden
func isHidden(name string) bool {
	// On Unix-like systems, hidden files start with a dot
	return len(name) > 0 && name[0] == '.'
}

// getLastAccessTime extracts the last access time from FileInfo
func getLastAccessTime(info os.FileInfo) time.Time {
	// Use ModTime as a fallback since access time is platform-specific
	return info.ModTime()
}
