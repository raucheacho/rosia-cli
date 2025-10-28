package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/raucheacho/rosia-cli/internal/cleaner"
	"github.com/raucheacho/rosia-cli/internal/scanner"
)

// Run starts the TUI application
func Run(ctx context.Context, scanner *scanner.Scanner, cleaner *cleaner.Cleaner, scanPaths []string) error {
	model := NewTUIModel(ctx, scanner, cleaner, scanPaths)

	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
