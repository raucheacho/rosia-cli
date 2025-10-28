---
title: "Commands"
description: "Complete reference for all Rosia CLI commands"
---

# Commands Reference

Complete guide to all Rosia CLI commands and their options.

## Global Flags

These flags work with all commands:

- `--verbose, -v` - Enable verbose logging
- `--config, -c <path>` - Specify custom config file path
- `--help, -h` - Show help for any command

## rosia scan

Scan directories to identify cleanable targets.

### Usage

```bash
rosia scan [paths...] [flags]
```

### Examples

```bash
# Scan current directory
rosia scan .

# Scan multiple directories
rosia scan ~/projects ~/workspace

# Scan with depth limit
rosia scan . --depth 5

# Include hidden directories
rosia scan . --include-hidden

# Dry-run (preview without changes)
rosia scan . --dry-run

# Verbose output
rosia scan . --verbose
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--depth` | `-d` | int | unlimited | Maximum directory depth to scan |
| `--include-hidden` | | bool | false | Include hidden directories in scan |
| `--dry-run` | | bool | false | Show what would be cleaned without making changes |

### Output

The scan command displays a table of detected targets:

```
Found 15 targets:
┌─────────────────────────────────────────────┬──────────┬─────────┐
│ Path                                        │ Size     │ Profile │
├─────────────────────────────────────────────┼──────────┼─────────┤
│ /Users/you/projects/app/node_modules        │ 450 MB   │ node    │
│ /Users/you/projects/api/target              │ 1.2 GB   │ rust    │
│ /Users/you/projects/web/dist                │ 25 MB    │ node    │
└─────────────────────────────────────────────┴──────────┴─────────┘

Total: 1.7 GB across 15 targets
```

---

## rosia clean

Clean detected targets with confirmation.

### Usage

```bash
rosia clean [targets...] [flags]
```

### Examples

```bash
# Clean with confirmation prompt
rosia clean

# Clean without confirmation (use with caution)
rosia clean --yes

# Clean specific targets
rosia clean ~/projects/app/node_modules ~/projects/api/target

# Skip trash system (permanent deletion)
rosia clean --no-trash --yes
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--yes` | `-y` | bool | false | Skip confirmation prompt |
| `--no-trash` | | bool | false | Skip trash system and delete permanently |

### Confirmation Prompt

By default, Rosia asks for confirmation before cleaning:

```
About to clean 15 targets (1.7 GB)

Targets:
  - /Users/you/projects/app/node_modules (450 MB)
  - /Users/you/projects/api/target (1.2 GB)
  - /Users/you/projects/web/dist (25 MB)
  ...

Continue? [y/N]: 
```

### Output

After cleaning, Rosia displays a summary report:

```
Cleaning...
✓ Cleaned /Users/you/projects/app/node_modules (450 MB)
✓ Cleaned /Users/you/projects/api/target (1.2 GB)
✓ Cleaned /Users/you/projects/web/dist (25 MB)

Summary:
  Total Size: 1.7 GB
  Files Deleted: 15
  Duration: 5.2s
  Trashed Items: 15

All items moved to trash. Use 'rosia restore --list' to view.
```

---

## rosia ui

Launch interactive terminal UI for visual selection.

### Usage

```bash
rosia ui [path] [flags]
```

### Examples

```bash
# Launch TUI for current directory
rosia ui .

# Launch TUI for specific path
rosia ui ~/projects
```

### Keyboard Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Space` | Toggle selection |
| `a` | Select all |
| `n` | Deselect all |
| `Enter` | Confirm and clean selected targets |
| `q` / `Esc` | Quit without cleaning |

### Interface

The TUI displays an interactive list of targets:

```
┌─ Rosia CLI ─────────────────────────────────────────────────────┐
│                                                                   │
│  Scan Results: 15 targets found (1.7 GB)                        │
│                                                                   │
│  [x] node_modules                    450 MB    node              │
│  [x] target                          1.2 GB    rust              │
│  [ ] dist                            25 MB     node              │
│  [x] build                           120 MB    flutter           │
│  [ ] .dart_tool                      45 MB     flutter           │
│                                                                   │
│  Selected: 3 targets (1.77 GB)                                  │
│                                                                   │
│  [Space] Toggle  [a] Select All  [n] Deselect All  [Enter] Clean│
└───────────────────────────────────────────────────────────────────┘
```

---

## rosia restore

Restore previously deleted targets from trash.

### Usage

```bash
rosia restore [id] [flags]
```

### Examples

```bash
# List all trashed items
rosia restore --list

# Restore specific item by ID
rosia restore 20250428_143022_node_modules

# Restore with verbose output
rosia restore 20250428_143022_node_modules --verbose
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--list` | `-l` | bool | false | List all trashed items |

### List Output

```bash
rosia restore --list
```

```
Trashed Items:
┌──────────────────────────────────┬─────────────────────────────┬──────────┬─────────────────────┐
│ ID                               │ Original Path               │ Size     │ Deleted At          │
├──────────────────────────────────┼─────────────────────────────┼──────────┼─────────────────────┤
│ 20250428_143022_node_modules     │ /Users/you/app/node_modules │ 450 MB   │ 2025-04-28 14:30:22 │
│ 20250428_143045_target           │ /Users/you/api/target       │ 1.2 GB   │ 2025-04-28 14:30:45 │
│ 20250427_091530_dist             │ /Users/you/web/dist         │ 25 MB    │ 2025-04-27 09:15:30 │
└──────────────────────────────────┴─────────────────────────────┴──────────┴─────────────────────┘

Total: 3 items (1.7 GB)
Retention: Items older than 3 days will be auto-deleted
```

### Restore Output

```bash
rosia restore 20250428_143022_node_modules
```

```
Restoring 20250428_143022_node_modules...
✓ Restored to /Users/you/app/node_modules (450 MB)
```

---

## rosia config

Manage configuration settings.

### Usage

```bash
rosia config <subcommand> [args] [flags]
```

### Subcommands

#### show

Display current configuration:

```bash
rosia config show
```

Output:

```json
{
  "trash_retention_days": 3,
  "profiles": ["node", "python", "rust", "flutter", "go"],
  "ignore_paths": [],
  "plugins": [],
  "concurrency": 0,
  "telemetry_enabled": false
}
```

#### set

Set a configuration value:

```bash
rosia config set <key> <value>
```

Examples:

```bash
# Set trash retention to 7 days
rosia config set trash_retention_days 7

# Set concurrency to 8 workers
rosia config set concurrency 8

# Enable telemetry
rosia config set telemetry_enabled true

# Add ignore path
rosia config set ignore_paths /usr/local,/System
```

#### reset

Reset configuration to defaults:

```bash
rosia config reset
```

---

## rosia stats

Display cleaning statistics and history.

### Usage

```bash
rosia stats [flags]
```

### Examples

```bash
# Show statistics
rosia stats

# Show with verbose details
rosia stats --verbose
```

### Output

```
Rosia Statistics
─────────────────────────────────────────────────────

Total Scans:           42
Total Cleaned:         15.7 GB
Total Files Deleted:   1,234

Average Size by Type:
  node_modules:        500 MB
  target:              100 MB
  dist:                10 MB
  build:               50 MB
  __pycache__:         5 MB

Last Scan:             2025-04-28 14:30:22
Last Clean:            2025-04-28 14:35:10

Disk Space Saved:      15.7 GB
```

---

## rosia plugin

Manage plugins.

### Usage

```bash
rosia plugin <subcommand> [args] [flags]
```

### Subcommands

#### list

List all loaded plugins:

```bash
rosia plugin list
```

Output:

```
Loaded Plugins:
┌─────────────────┬─────────┬──────────────────────────────────┐
│ Name            │ Version │ Description                      │
├─────────────────┼─────────┼──────────────────────────────────┤
│ rosia-docker    │ 1.0.0   │ Clean Docker images and volumes  │
│ rosia-xcode     │ 1.2.0   │ Clean Xcode derived data         │
└─────────────────┴─────────┴──────────────────────────────────┘
```

#### info

Show detailed information about a plugin:

```bash
rosia plugin info <plugin-name>
```

Example:

```bash
rosia plugin info rosia-docker
```

Output:

```
Plugin: rosia-docker
Version: 1.0.0
Description: Clean Docker images and volumes

Capabilities:
  - Scan for dangling Docker images
  - Clean unused Docker volumes
  - Remove stopped containers

Author: raucheacho
License: MIT
```

---

## rosia version

Display version information.

### Usage

```bash
rosia version
```

### Output

```
Rosia CLI v0.1.0
Built with Go 1.21+
Commit: a1b2c3d
Build Date: 2025-04-28
```

---

## Exit Codes

Rosia uses standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Permission denied |
| 4 | Path not found |

## Environment Variables

Rosia respects these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ROSIA_CONFIG` | Path to config file | `~/.rosiarc.json` |
| `ROSIA_TRASH_DIR` | Path to trash directory | `~/.rosia/trash` |
| `ROSIA_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

Example:

```bash
ROSIA_LOG_LEVEL=debug rosia scan .
```
