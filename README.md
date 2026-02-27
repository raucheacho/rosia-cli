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
brew install raucheacho/tap/rosia
```

### Scoop (Windows)

```powershell
scoop bucket add raucheacho https://github.com/raucheacho/scoop-bucket
scoop install rosia
```

### Go Install

```bash
go install github.com/raucheacho/rosia-cli@latest
```

### Manual Installation

Download the latest binary from the [releases page](https://github.com/raucheacho/rosia-cli/releases).

## Quick Start

```bash
# Scan for cleanable files
rosia scan .

# Clean with confirmation
rosia clean --yes

# Interactive TUI
rosia ui .
```

## Commands

| Command | Description |
|---------|-------------|
| `rosia scan [paths...]` | Scan directories for cleanable targets |
| `rosia clean [paths...]` | Clean detected targets |
| `rosia ui [paths...]` | Launch interactive TUI |
| `rosia trash list` | List trashed items |
| `rosia trash clean [--all]` | Clean old items from trash |
| `rosia restore <id>` | Restore item from trash |
| `rosia config show` | Show configuration |
| `rosia config reset` | Reset configuration to defaults |
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

### Trash & Restore

```bash
# List trashed items
rosia trash list

# Clean old items (respects trash_retention_days)
rosia trash clean

# Clean all items immediately (permanent deletion)
rosia trash clean --all

# Restore specific item
rosia restore 20250226_143022_node_modules
```

## Configuration

### Config File

**Location**: `~/.rosiarc.json`

Created automatically on first run with sensible defaults.

```json
{
  "trash_retention_days": 3,
  "profiles": ["node", "python", "rust", "flutter", "go"],
  "ignore_paths": []
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `trash_retention_days` | 3 | Days to keep items in trash |
| `profiles` | all | Enabled technology profiles |
| `ignore_paths` | [] | Paths to exclude from scanning |

### Profiles File

**Location**: `~/.rosia/profiles.json`

Created automatically on first run. Contains cleaning rules for each technology.

Example profile:
```json
{
  "name": "Node.js",
  "version": "1.0.0",
  "patterns": ["node_modules", "dist", "build"],
  "detect": ["package.json"],
  "enabled": true
}
```

Edit this file to customize patterns or add new technologies.

## Supported Technologies

| Technology | Patterns | Detector |
|------------|----------|----------|
| **Node.js** | `node_modules`, `dist`, `build`, `.next` | `package.json` |
| **Python** | `venv`, `__pycache__`, `.tox` | `requirements.txt`, `pyproject.toml` |
| **Rust** | `target/` | `Cargo.toml` |
| **Flutter** | `build/`, `.dart_tool/` | `pubspec.yaml` |
| **Go** | `vendor/`, `bin/` | `go.mod` |

## Troubleshooting

### "Failed to load profiles"

Delete the profiles file and it will be recreated with defaults:
```bash
rm ~/.rosia/profiles.json
rosia scan .
```

### Rosia doesn't find my projects

Check that your project has a detector file (e.g., `package.json` for Node.js projects).

### Permission denied errors

Rosia needs read access to scan directories and write access to move files to trash.

## License

MIT License - see [LICENSE](LICENSE) file.
