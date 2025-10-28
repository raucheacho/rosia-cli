// Package cmd provides the command-line interface for Rosia CLI.
//
// This package implements all CLI commands using the Cobra framework, including:
//   - scan: Scan directories for cleanable targets
//   - clean: Delete detected targets with confirmation
//   - restore: Restore items from trash
//   - ui: Launch interactive terminal UI
//   - config: Manage configuration settings
//   - stats: Display cleaning statistics
//   - plugin: Manage plugins
//   - version: Display version information
//
// Each command is implemented in its own file (e.g., scan.go, clean.go) and
// registered with the root command in root.go.
//
// Example usage:
//
//	// Scan current directory
//	rosia scan .
//
//	// Clean with confirmation
//	rosia clean
//
//	// Launch interactive UI
//	rosia ui ~/projects
//
//	// View statistics
//	rosia stats
package cmd
