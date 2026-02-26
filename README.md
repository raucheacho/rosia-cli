# Rosia CLI

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)

Rosia is a fast command-line tool that helps developers reclaim disk space by cleaning dependencies, builds, and caches across multiple project types.

## Features

- 🚀 **Fast Scanning**: Concurrent directory traversal
- 🎯 **Multi-Technology Support**: Node.js, Python, Rust, Flutter, Go
- 🛡️ **Safe Deletion**: Trash system with restoration capability
- 🎨 **Interactive TUI**: Visual selection and progress tracking
- ⚙️ **Configurable**: JSON-based configuration
- 🌍 **Cross-Platform**: Linux, macOS, and Windows

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap raucheacho/rosia
brew install rosia
```

### Scoop (Windows)

```powershell
scoop bucket add rosia https://github.com/raucheacho/scoop-rosia
scoop install rosia
```

### Go Install

```bash
go install github.com/raucheacho/rosia-cli@latest
```

### Manual Installation

Download the latest binary from the [releases page](https://github.com/raucheacho/rosia-cli/releases).

## Usage

### Quick Start

```bash
# Scan for cleanable files
rosia scan .

# Clean with confirmation
rosia clean --yes

# Interactive TUI
rosia ui .
```

### Commands

| Command | Description |
|---------|-------------|
| `rosia scan [paths...]` | Scan directories for cleanable targets |
| `rosia clean [paths...]` | Clean detected targets |
| `rosia ui [path]` | Launch interactive TUI |
| `rosia restore <id>` | Restore from trash |
| `rosia config` | Manage configuration |
| `rosia version` | Display version |

### Scan Options

```bash
rosia scan . --depth 5 --include-hidden
```

- `--depth <n>`: Maximum directory depth
- `--include-hidden`: Include hidden directories

### Clean Options

```bash
rosia clean . --yes --no-trash
```

- `--yes, -y`: Skip confirmation
- `--no-trash`: Delete permanently (not recommended)

### Restore

```bash
# List trashed items
rosia restore --list

# Restore specific item
rosia restore 20250226_143022_node_modules_FROM_app
```

## Configuration

Config file: `~/.rosiarc.json`

```json
{
  "trash_retention_days": 3,
  "ignore_paths": ["/System", "/usr/local"]
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `trash_retention_days` | 3 | Days to keep items in trash |
| `ignore_paths` | [] | Paths to exclude from scanning |

## Supported Technologies

- **Node.js**: `node_modules`, `dist`, `build`, `.next`
- **Python**: `venv`, `__pycache__`, `.tox`
- **Rust**: `target/`
- **Flutter**: `build/`, `.dart_tool/`
- **Go**: `vendor/`

## Examples

```bash
# Scan and clean Node.js projects
rosia scan ~/projects
rosia clean --yes

# Interactive cleaning
rosia ui ~/workspace

# Scan with custom depth
rosia scan ~/projects --depth 3
```

## License

MIT License - see [LICENSE](LICENSE) file.
