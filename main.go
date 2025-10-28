// Rosia CLI is a universal, fast, and secure command-line tool for cleaning
// development dependencies, builds, and caches across multiple project types.
//
// Rosia helps developers reclaim disk space by safely removing cleanable files
// like node_modules, target/, build/, and other technology-specific artifacts.
//
// Features:
//   - Fast concurrent scanning with configurable worker pools
//   - Multi-technology support (Node.js, Python, Rust, Flutter, Go, etc.)
//   - Safe deletion with trash system and restoration capability
//   - Interactive TUI for visual selection
//   - Extensible plugin system
//   - Cross-platform support (Linux, macOS, Windows)
//
// Usage:
//
//	rosia scan .              # Scan current directory
//	rosia clean --yes         # Clean without confirmation
//	rosia ui ~/projects       # Launch interactive UI
//	rosia restore <id>        # Restore from trash
//	rosia stats               # View statistics
//
// For more information, visit: https://github.com/raucheacho/rosia-cli
package main

import (
	"os"

	"github.com/raucheacho/rosia-cli/cmd"
)

// main is the entry point for the Rosia CLI application.
func main() {
	exitCode := cmd.ExecuteWithExitCode()
	os.Exit(exitCode)
}
