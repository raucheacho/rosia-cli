package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))
)

// renderScanningScreen renders the scanning progress screen
func (m *TUIModel) renderScanningScreen() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🔍 Scanning for cleanable targets..."))
	b.WriteString("\n\n")

	if m.currentDir != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Current: %s", m.currentDir)))
		b.WriteString("\n\n")
	}

	// Progress bar
	b.WriteString(m.progress.ViewAs(m.scanProgress))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Press q to quit"))

	return b.String()
}

// renderSelectionScreen renders the target selection screen
func (m *TUIModel) renderSelectionScreen() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("📦 Found %d cleanable targets", len(m.targets))))
	b.WriteString("\n\n")

	if len(m.targets) == 0 {
		b.WriteString(infoStyle.Render("No targets found. Press q to quit."))
		return b.String()
	}

	// Render viewport with target list
	b.WriteString(m.viewport.View())
	b.WriteString("\n\n")

	// Show selection count and total size
	selectedCount := 0
	var totalSize int64
	for i, target := range m.targets {
		if m.selected[i] {
			selectedCount++
			totalSize += target.Size
		}
	}

	if selectedCount > 0 {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Selected: %d targets (%s)", selectedCount, formatSize(totalSize))))
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate • space: select • a: select all • n: deselect all • enter: confirm • q: quit"))

	return b.String()
}

// renderTargetList renders the list of targets for the viewport
func (m *TUIModel) renderTargetList() string {
	var b strings.Builder

	for i, target := range m.targets {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = selectedStyle.Render("[✓]")
		}

		line := fmt.Sprintf("%s%s %s (%s) - %s",
			cursor,
			checkbox,
			target.Path,
			formatSize(target.Size),
			target.ProfileName,
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// renderConfirmationScreen renders the confirmation dialog
func (m *TUIModel) renderConfirmationScreen() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⚠️  Confirm Deletion"))
	b.WriteString("\n\n")

	// Calculate totals
	selectedCount := 0
	var totalSize int64
	for i, target := range m.targets {
		if m.selected[i] {
			selectedCount++
			totalSize += target.Size
		}
	}

	b.WriteString(fmt.Sprintf("You are about to clean %s targets, freeing up %s\n\n",
		successStyle.Render(fmt.Sprintf("%d", selectedCount)),
		successStyle.Render(formatSize(totalSize)),
	))

	b.WriteString(infoStyle.Render("Files will be moved to trash and can be restored later."))
	b.WriteString("\n\n")

	b.WriteString("Do you want to proceed?\n\n")
	b.WriteString(helpStyle.Render("y/enter: confirm • n/q: cancel"))

	return b.String()
}

// renderCleaningScreen renders the cleaning progress screen
func (m *TUIModel) renderCleaningScreen() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🧹 Cleaning targets..."))
	b.WriteString("\n\n")

	b.WriteString(infoStyle.Render("Please wait while files are being moved to trash..."))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Press q to quit (cleaning will continue)"))

	return b.String()
}

// renderSummaryScreen renders the post-clean summary
func (m *TUIModel) renderSummaryScreen() string {
	var b strings.Builder

	if m.cleanReport == nil {
		b.WriteString(errorStyle.Render("❌ Cleaning failed"))
		b.WriteString("\n\n")
		if m.err != nil {
			b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		}
		b.WriteString(helpStyle.Render("Press q to quit"))
		return b.String()
	}

	b.WriteString(titleStyle.Render("✨ Cleaning Complete!"))
	b.WriteString("\n\n")

	// Success summary
	b.WriteString(successStyle.Render(fmt.Sprintf("✓ Cleaned %d files", m.cleanReport.FilesDeleted)))
	b.WriteString("\n")
	b.WriteString(successStyle.Render(fmt.Sprintf("✓ Freed up %s", formatSize(m.cleanReport.TotalSize))))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("✓ Duration: %s", m.cleanReport.Duration)))
	b.WriteString("\n\n")

	// Errors if any
	if len(m.cleanReport.Errors) > 0 {
		b.WriteString(errorStyle.Render(fmt.Sprintf("⚠ %d errors occurred:", len(m.cleanReport.Errors))))
		b.WriteString("\n")
		for i, cleanErr := range m.cleanReport.Errors {
			if i >= 5 {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(m.cleanReport.Errors)-5))
				break
			}
			b.WriteString(fmt.Sprintf("  • %s: %v\n", cleanErr.Target.Path, cleanErr.Error))
		}
		b.WriteString("\n")
	}

	// Trash info
	if len(m.cleanReport.TrashedItems) > 0 {
		b.WriteString(infoStyle.Render("Files moved to trash. Use 'rosia restore <id>' to restore."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press q or enter to quit"))

	return b.String()
}

// formatSize formats bytes into human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
