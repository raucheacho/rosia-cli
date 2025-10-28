package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/raucheacho/rosia-cli/internal/cleaner"
	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/pkg/types"
)

// startScan initiates the scanning process
func (m *TUIModel) startScan() tea.Cmd {
	return func() tea.Msg {
		opts := scanner.ScanOptions{
			MaxDepth:      10,
			IncludeHidden: false,
			Concurrency:   0, // Use default
		}

		targetsChan, errChan := m.scanner.ScanAsync(m.ctx, m.scanPaths, opts)

		var targets []types.Target
		var scanErr error

		// Collect targets from channel
		for {
			select {
			case target, ok := <-targetsChan:
				if !ok {
					// Channel closed, scanning complete
					if scanErr != nil {
						return scanErrorMsg{err: scanErr}
					}
					return scanCompleteMsg{targets: targets}
				}
				targets = append(targets, target)

			case err, ok := <-errChan:
				if ok && err != nil {
					scanErr = err
				}
			}
		}
	}
}

// startClean initiates the cleaning process
func (m *TUIModel) startClean() tea.Cmd {
	return func() tea.Msg {
		// Get selected targets
		selectedTargets := make([]types.Target, 0)
		for i, target := range m.targets {
			if m.selected[i] {
				selectedTargets = append(selectedTargets, target)
			}
		}

		if len(selectedTargets) == 0 {
			return cleanErrorMsg{err: nil} // No targets selected
		}

		// Clean targets
		opts := cleaner.CleanOptions{
			SkipConfirmation: true,
			UseTrash:         true,
			Concurrency:      0,
		}
		report, err := m.cleaner.Clean(m.ctx, selectedTargets, opts)
		if err != nil {
			return cleanErrorMsg{err: err}
		}

		return cleanCompleteMsg{report: report}
	}
}
