// Package ui provides the interactive terminal user interface (TUI) for Rosia CLI.
//
// The TUI is built using the Bubble Tea framework and provides an interactive
// way to select and clean targets with visual feedback. It includes:
//   - List view of detected targets with sizes
//   - Keyboard navigation and selection
//   - Real-time scan progress
//   - Confirmation dialogs
//   - Post-clean summary screens
//
// The TUI follows the Elm architecture with a Model-View-Update pattern:
//   - model.go: Application state and business logic
//   - views.go: Rendering functions for different screens
//   - commands.go: Side effects and async operations
//   - messages.go: Message types for state updates
//   - keys.go: Keyboard bindings
//
// Example usage:
//
//	tui := ui.NewTUI(scanner, cleaner)
//	if err := tui.Run(); err != nil {
//	    log.Fatal(err)
//	}
package ui
