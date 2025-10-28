// Package sizecalc provides efficient directory size calculation.
//
// The sizecalc package computes directory and file sizes with support for
// concurrent processing, symlink handling, and context cancellation. It's
// used by the scanner to determine the size of detected targets.
//
// Example usage:
//
//	calc := sizecalc.NewSizeCalc(4)
//	size, err := calc.CalculateSize(ctx, "/path/to/directory")
package sizecalc

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// SizeCalc computes directory and file sizes.
//
// It uses concurrent workers to efficiently calculate sizes of large
// directory trees while safely handling symlinks and permission errors.
type SizeCalc struct {
	concurrency int
}

// NewSizeCalc creates a new size calculator
func NewSizeCalc(concurrency int) *SizeCalc {
	if concurrency <= 0 {
		concurrency = runtime.NumCPU() * 2
	}
	return &SizeCalc{
		concurrency: concurrency,
	}
}

// Calculate computes the size of a single path
func (sc *SizeCalc) Calculate(path string) (int64, error) {
	info, err := os.Lstat(path) // Use Lstat to not follow symlinks
	if err != nil {
		return 0, fmt.Errorf("failed to stat path: %w", err)
	}

	// If it's a symlink, return 0 (don't follow)
	if info.Mode()&os.ModeSymlink != 0 {
		return 0, nil
	}

	// If it's a regular file, return its size
	if !info.IsDir() {
		return info.Size(), nil
	}

	// For directories, walk and sum all file sizes
	var totalSize int64
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip files we can't access
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Add file size
		if !d.IsDir() {
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return totalSize, fmt.Errorf("error walking directory: %w", err)
	}

	return totalSize, nil
}

// CalculateTargets computes sizes for multiple targets concurrently
func (sc *SizeCalc) CalculateTargets(ctx context.Context, targets []types.Target) ([]types.Target, error) {
	if len(targets) == 0 {
		return targets, nil
	}

	// Create result slice
	results := make([]types.Target, len(targets))
	copy(results, targets)

	// Create worker pool
	jobs := make(chan int, len(targets))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Start workers
	for i := 0; i < sc.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for idx := range jobs {
				// Check context cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Calculate size
				size, err := sc.Calculate(results[idx].Path)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to calculate size for %s: %w", results[idx].Path, err))
					mu.Unlock()
					continue
				}

				// Update target size
				mu.Lock()
				results[idx].Size = size
				mu.Unlock()
			}
		}()
	}

	// Submit jobs
	for i := range targets {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results, ctx.Err()
		case jobs <- i:
		}
	}

	close(jobs)
	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		return results, fmt.Errorf("encountered %d errors during size calculation: %v", len(errors), errors[0])
	}

	return results, nil
}

// CalculateAsync computes sizes for targets and sends results to a channel
func (sc *SizeCalc) CalculateAsync(ctx context.Context, targets <-chan types.Target) (<-chan types.Target, <-chan error) {
	resultChan := make(chan types.Target, 100)
	errorChan := make(chan error, 10)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Create worker pool
		var wg sync.WaitGroup
		jobs := make(chan types.Target, sc.concurrency*2)

		// Start workers
		for i := 0; i < sc.concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for target := range jobs {
					// Check context cancellation
					select {
					case <-ctx.Done():
						return
					default:
					}

					// Calculate size
					size, err := sc.Calculate(target.Path)
					if err != nil {
						select {
						case errorChan <- fmt.Errorf("failed to calculate size for %s: %w", target.Path, err):
						case <-ctx.Done():
							return
						}
						continue
					}

					// Update target and send result
					target.Size = size
					select {
					case resultChan <- target:
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		// Feed targets to workers
		for target := range targets {
			select {
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				return
			case jobs <- target:
			}
		}

		close(jobs)
		wg.Wait()
	}()

	return resultChan, errorChan
}
