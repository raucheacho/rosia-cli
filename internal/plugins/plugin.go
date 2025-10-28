// Package plugins provides the plugin system for extending Rosia's functionality.
//
// Plugins allow third-party extensions to add custom scanning and cleaning logic
// beyond the built-in profiles. Plugins can be written in Go (using Go's plugin system)
// or in any language via JSON-RPC.
//
// Example Go plugin:
//
//	type MyPlugin struct{}
//
//	func (p *MyPlugin) Name() string { return "my-plugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//	func (p *MyPlugin) Description() string { return "Custom cleaning" }
//	func (p *MyPlugin) Scan(ctx context.Context) ([]types.Target, error) { ... }
//	func (p *MyPlugin) Clean(ctx context.Context, targets []types.Target) error { ... }
//
//	var Plugin MyPlugin
package plugins

import (
	"context"

	"github.com/raucheacho/rosia-cli/pkg/types"
)

// Plugin defines the interface that all plugins must implement.
//
// Plugins extend Rosia's functionality by providing custom scanning and cleaning
// logic. They are loaded from ~/.rosia/plugins/ at startup.
type Plugin interface {
	// Name returns the unique identifier for the plugin
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a human-readable description of the plugin
	Description() string

	// Scan performs scanning and returns detected targets
	Scan(ctx context.Context) ([]types.Target, error)

	// Clean performs cleaning operations on the given targets
	Clean(ctx context.Context, targets []types.Target) error
}
