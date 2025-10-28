package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// ScanAsync performs an asynchronous scan using a worker pool
// Returns channels for targets and errors
func (s *Scanner) ScanAsync(ctx context.Context, paths []string, opts ScanOptions) (<-chan types.Target, <-chan error) {
	targetChan := make(chan types.Target, 100)
	errorChan := make(chan error, 10)

	go func() {
		defer close(targetChan)
		defer close(errorChan)

		// Determine concurrency level
		concurrency := opts.Concurrency
		if concurrency <= 0 {
			concurrency = runtime.NumCPU() * 2
		}

		// Create worker pool
		pool := newWorkerPool(concurrency, s, opts)

		// Start workers
		pool.start(ctx, targetChan, errorChan)

		// Submit paths to workers
		for _, path := range paths {
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			case pool.jobs <- path:
			}
		}

		// Close jobs channel and wait for workers to finish
		close(pool.jobs)
		pool.wg.Wait()
	}()

	return targetChan, errorChan
}

// workerPool manages concurrent scanning operations
type workerPool struct {
	workers int
	jobs    chan string
	scanner *Scanner
	opts    ScanOptions
	wg      sync.WaitGroup
}

// newWorkerPool creates a new worker pool
func newWorkerPool(workers int, scanner *Scanner, opts ScanOptions) *workerPool {
	return &workerPool{
		workers: workers,
		jobs:    make(chan string, workers*2),
		scanner: scanner,
		opts:    opts,
	}
}

// start launches the worker goroutines
func (p *workerPool) start(ctx context.Context, targetChan chan<- types.Target, errorChan chan<- error) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, targetChan, errorChan)
	}
}

// worker processes jobs from the jobs channel
func (p *workerPool) worker(ctx context.Context, targetChan chan<- types.Target, errorChan chan<- error) {
	defer p.wg.Done()

	for path := range p.jobs {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Scan the path
		targets, err := p.scanner.scanPathAsync(ctx, path, p.opts, targetChan)
		if err != nil {
			select {
			case errorChan <- fmt.Errorf("error scanning %s: %w", path, err):
			case <-ctx.Done():
				return
			}
		}

		// Send targets to channel
		for _, target := range targets {
			select {
			case targetChan <- target:
			case <-ctx.Done():
				return
			}
		}
	}
}

// scanPathAsync scans a single path and sends targets to the channel as they're found
func (s *Scanner) scanPathAsync(ctx context.Context, rootPath string, opts ScanOptions, targetChan chan<- types.Target) ([]types.Target, error) {
	targets := make([]types.Target, 0)
	rootDepth := strings.Count(rootPath, string(os.PathSeparator))

	// First, try to match the root directory itself
	profile, err := s.profileLoader.MatchProfile(rootPath)
	if err == nil && profile != nil {
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
			fmt.Fprintf(os.Stderr, "Warning: error accessing %s: %v\n", path, err)
			return nil
		}

		// Skip the root path itself
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
