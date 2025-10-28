# Rosia CLI

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)

Rosia is a universal, fast, and secure command-line tool that helps developers reclaim disk space by cleaning dependencies, builds, and caches across multiple project types. Whether you're working with Node.js, Python, Rust, Flutter, or Go projects, Rosia intelligently detects and safely removes cleanable artifacts.

## Features

- 🚀 **Fast Scanning**: Concurrent directory traversal with configurable worker pools
- 🎯 **Multi-Technology Support**: Built-in profiles for Node.js, Python, Rust, Flutter, Go, and more
- 🛡️ **Safe Deletion**: Trash system with restoration capability before permanent removal
- 🎨 **Interactive TUI**: Beautiful terminal interface for visual selection and progress tracking
- 🔌 **Extensible**: Plugin system for custom cleaning logic
- ⚙️ **Configurable**: JSON-based configuration for profiles, ignore paths, and preferences
- 📊 **Statistics**: Track cleaning history and disk space savings over time
- 🌍 **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Homebrew (macOS/Linux)

```bash
brew install rosia
```

### Scoop (Windows)

```bash
scoop install rosia
```

### Go Install

```bash
go install github.com/raucheacho/rosia-cli@latest
```

### Manual Installation

Download the latest binary from the [releases page](https://github.com/raucheacho/rosia-cli/releases) and add it to your PATH.

**Linux/macOS:**
```bash
curl -L https://github.com/raucheacho/rosia-cli/releases/latest/download/rosia_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv rosia /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/raucheacho/rosia-cli/releases/latest/download/rosia_Windows_x86_64.zip" -OutFile "rosia.zip"
Expand-Archive -Path "rosia.zip" -DestinationPath "."
Move-Item -Path "rosia.exe" -Destination "$env:USERPROFILE\bin\"
```

## Usage

### Quick Start

Scan your current directory for cleanable files:

```bash
rosia scan .
```

Launch the interactive TUI to select and clean targets:

```bash
rosia ui .
```

### Commands

#### `rosia scan [paths...]`

Scan directories to identify cleanable targets.

```bash
# Scan current directory
rosia scan .

# Scan multiple directories
rosia scan ~/projects ~/workspace

# Scan with options
rosia scan . --depth 5 --include-hidden --dry-run
```

**Flags:**
- `--depth <n>`: Maximum directory depth to scan (default: unlimited)
- `--include-hidden`: Include hidden directories in scan
- `--dry-run`: Show what would be cleaned without making changes

#### `rosia clean [targets...]`

Clean detected targets with confirmation.

```bash
# Clean with confirmation prompt
rosia clean

# Clean without confirmation (use with caution)
rosia clean --yes

# Clean specific targets
rosia clean ~/projects/app/node_modules ~/projects/api/target
```

**Flags:**
- `--yes, -y`: Skip confirmation prompt
- `--no-trash`: Skip trash system and delete permanently (not recommended)

#### `rosia ui [path]`

Launch interactive terminal UI for visual selection.

```bash
# Launch TUI for current directory
rosia ui .

# Launch TUI for specific path
rosia ui ~/projects
```

**Keyboard Controls:**
- `↑/↓`: Navigate through targets
- `Space`: Toggle selection
- `a`: Select all
- `n`: Deselect all
- `Enter`: Confirm and clean selected targets
- `q`: Quit without cleaning

#### `rosia restore <id>`

Restore a previously deleted target from trash.

```bash
# List trashed items
rosia restore --list

# Restore specific item
rosia restore 20250428_143022_node_modules
```

**Flags:**
- `--list, -l`: List all trashed items

#### `rosia config`

Manage configuration settings.

```bash
# Show current configuration
rosia config show

# Set configuration value
rosia config set trash_retention_days 7
rosia config set concurrency 8

# Reset to defaults
rosia config reset
```

#### `rosia stats`

Display cleaning statistics and history.

```bash
rosia stats
```

#### `rosia plugin`

Manage plugins.

```bash
# List loaded plugins
rosia plugin list

# Show plugin details
rosia plugin info <plugin-name>
```

#### `rosia version`

Display version information.

```bash
rosia version
```

### Global Flags

- `--verbose, -v`: Enable verbose logging
- `--config, -c <path>`: Specify custom config file path

## Configuration

Rosia uses a JSON configuration file located at `~/.rosiarc.json`. If the file doesn't exist, default settings are used.

### Configuration File Structure

```json
{
  "trash_retention_days": 3,
  "profiles": ["node", "python", "rust", "flutter", "go"],
  "ignore_paths": [
    "/usr/local",
    "/System"
  ],
  "plugins": [],
  "concurrency": 0,
  "telemetry_enabled": false
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `trash_retention_days` | int | 3 | Days to keep items in trash before auto-cleanup |
| `profiles` | string[] | All built-in | Enabled profile names |
| `ignore_paths` | string[] | [] | Paths to exclude from scanning |
| `plugins` | string[] | [] | Enabled plugin names |
| `concurrency` | int | 0 | Worker pool size (0 = auto-detect) |
| `telemetry_enabled` | bool | false | Enable anonymous usage statistics |

### Built-in Profiles

Rosia includes profiles for common development technologies:

- **Node.js**: `node_modules`, `dist`, `build`, `.next`, `.cache`, `coverage`
- **Python**: `venv`, `__pycache__`, `.pytest_cache`, `.tox`, `dist`, `build`
- **Rust**: `target/`
- **Flutter**: `build/`, `.dart_tool/`
- **Go**: `vendor/`, `bin/`

## Examples

### Clean all Node.js projects in a directory

```bash
rosia scan ~/projects
rosia clean --yes
```

### Interactive cleaning with TUI

```bash
rosia ui ~/workspace
# Use arrow keys to navigate, space to select, enter to clean
```

### Scan with custom depth and dry-run

```bash
rosia scan ~/projects --depth 3 --dry-run
```

### Restore accidentally deleted files

```bash
rosia restore --list
rosia restore 20250428_143022_node_modules
```

### View cleaning statistics

```bash
rosia stats
```

## Plugin Development

Rosia supports plugins written in Go or any language via JSON-RPC. See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed plugin development guidelines.

### Quick Plugin Example (Go)

```go
package main

type MyPlugin struct{}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Description() string { return "Custom cleaning logic" }

func (p *MyPlugin) Scan(ctx context.Context) ([]Target, error) {
    // Implement scan logic
    return targets, nil
}

func (p *MyPlugin) Clean(ctx context.Context, targets []Target) error {
    // Implement clean logic
    return nil
}

var Plugin MyPlugin
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Setting up the development environment
- Project structure and architecture
- Adding new profiles
- Creating plugins
- Submitting pull requests

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## Support

- 🐛 [Report a bug](https://github.com/raucheacho/rosia-cli/issues/new?template=bug_report.md)
- 💡 [Request a feature](https://github.com/raucheacho/rosia-cli/issues/new?template=feature_request.md)
- 📖 [Documentation](https://github.com/raucheacho/rosia-cli/wiki)

---

Made with ❤️ by developers, for developers.
