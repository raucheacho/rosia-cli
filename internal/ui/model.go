package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raucheacho/rosia-cli/internal/cleaner"
	"github.com/raucheacho/rosia-cli/internal/scanner"
	"github.com/raucheacho/rosia-cli/pkg/types"
)

// Screen represents the current screen state
type Screen int

const (
	ScreenScanning Screen = iota
	ScreenSelection
	ScreenConfirmation
	ScreenCleaning
	ScreenSummary
)

// TUIModel represents the BubbleTea model for the TUI
type TUIModel struct {
	// Core data
	targets  []types.Target
	selected map[int]bool
	cursor   int

	// State
	screen       Screen
	scanning     bool
	cleaning     bool
	scanProgress float64
	currentDir   string
	err          error

	// Components
	viewport viewport.Model
	progress progress.Model

	// Dependencies
	scanner *scanner.Scanner
	cleaner *cleaner.Cleaner
	ctx     context.Context

	// Results
	cleanReport *types.CleanReport

	// Configuration
	scanPaths []string
	width     int
	height    int
}

// NewTUIModel creates a new TUI model
func NewTUIModel(ctx context.Context, scanner *scanner.Scanner, cleaner *cleaner.Cleaner, scanPaths []string) *TUIModel {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	prog := progress.New(progress.WithDefaultGradient())

	return &TUIModel{
		targets:   make([]types.Target, 0),
		selected:  make(map[int]bool),
		cursor:    0,
		screen:    ScreenScanning,
		scanning:  true,
		viewport:  vp,
		progress:  prog,
		scanner:   scanner,
		cleaner:   cleaner,
		ctx:       ctx,
		scanPaths: scanPaths,
		width:     80,
		height:    24,
	}
}

// Init initializes the model
func (m *TUIModel) Init() tea.Cmd {
	return tea.Batch(
		m.startScan(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case scanProgressMsg:
		m.scanProgress = msg.progress
		m.currentDir = msg.currentDir
		return m, nil

	case scanCompleteMsg:
		m.scanning = false
		m.targets = msg.targets
		m.screen = ScreenSelection
		m.viewport.SetContent(m.renderTargetList())
		return m, nil

	case scanErrorMsg:
		m.err = msg.err
		m.scanning = false
		return m, tea.Quit

	case cleanProgressMsg:
		return m, nil

	case cleanCompleteMsg:
		m.cleaning = false
		m.cleanReport = msg.report
		m.screen = ScreenSummary
		return m, nil

	case cleanErrorMsg:
		m.err = msg.err
		m.cleaning = false
		return m, tea.Quit
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m *TUIModel) View() string {
	switch m.screen {
	case ScreenScanning:
		return m.renderScanningScreen()
	case ScreenSelection:
		return m.renderSelectionScreen()
	case ScreenConfirmation:
		return m.renderConfirmationScreen()
	case ScreenCleaning:
		return m.renderCleaningScreen()
	case ScreenSummary:
		return m.renderSummaryScreen()
	default:
		return "Unknown screen"
	}
}
