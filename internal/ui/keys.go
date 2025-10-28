package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress handles keyboard input based on current screen
func (m *TUIModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenScanning:
		return m.handleScanningKeys(msg)
	case ScreenSelection:
		return m.handleSelectionKeys(msg)
	case ScreenConfirmation:
		return m.handleConfirmationKeys(msg)
	case ScreenCleaning:
		return m.handleCleaningKeys(msg)
	case ScreenSummary:
		return m.handleSummaryKeys(msg)
	default:
		return m, nil
	}
}

// handleScanningKeys handles keys during scanning
func (m *TUIModel) handleScanningKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handleSelectionKeys handles keys during target selection
func (m *TUIModel) handleSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.viewport.SetContent(m.renderTargetList())
		}

	case "down", "j":
		if m.cursor < len(m.targets)-1 {
			m.cursor++
			m.viewport.SetContent(m.renderTargetList())
		}

	case " ":
		// Toggle selection
		m.selected[m.cursor] = !m.selected[m.cursor]
		m.viewport.SetContent(m.renderTargetList())

	case "a":
		// Select all
		for i := range m.targets {
			m.selected[i] = true
		}
		m.viewport.SetContent(m.renderTargetList())

	case "n":
		// Deselect all
		m.selected = make(map[int]bool)
		m.viewport.SetContent(m.renderTargetList())

	case "enter":
		// Move to confirmation screen
		if m.hasSelection() {
			m.screen = ScreenConfirmation
		}
	}

	return m, nil
}

// handleConfirmationKeys handles keys during confirmation
func (m *TUIModel) handleConfirmationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "n":
		// Cancel and go back to selection
		m.screen = ScreenSelection
		return m, nil

	case "y", "enter":
		// Confirm and start cleaning
		m.screen = ScreenCleaning
		m.cleaning = true
		return m, m.startClean()
	}

	return m, nil
}

// handleCleaningKeys handles keys during cleaning
func (m *TUIModel) handleCleaningKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handleSummaryKeys handles keys on summary screen
func (m *TUIModel) handleSummaryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "enter", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// hasSelection returns true if any targets are selected
func (m *TUIModel) hasSelection() bool {
	for _, selected := range m.selected {
		if selected {
			return true
		}
	}
	return false
}
