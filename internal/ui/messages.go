package ui

import "github.com/raucheacho/rosia-cli/pkg/types"

// scanProgressMsg represents scan progress updates
type scanProgressMsg struct {
	progress   float64
	currentDir string
}

// scanCompleteMsg represents scan completion
type scanCompleteMsg struct {
	targets []types.Target
}

// scanErrorMsg represents scan errors
type scanErrorMsg struct {
	err error
}

// cleanProgressMsg represents clean progress updates
type cleanProgressMsg struct {
	current int
	total   int
}

// cleanCompleteMsg represents clean completion
type cleanCompleteMsg struct {
	report *types.CleanReport
}

// cleanErrorMsg represents clean errors
type cleanErrorMsg struct {
	err error
}
