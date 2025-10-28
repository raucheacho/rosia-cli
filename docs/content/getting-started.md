---
title: "Getting Started"
description: "Install and configure Rosia CLI"
---

# Getting Started

This guide will help you install Rosia and get started cleaning your development projects.

## Installation

### Homebrew (macOS/Linux)

The easiest way to install Rosia on macOS or Linux:

```bash
brew install rosia
```

### Scoop (Windows)

For Windows users with Scoop:

```bash
scoop install rosia
```

### Go Install

If you have Go installed:

```bash
go install github.com/raucheacho/rosia-cli@latest
```

### Manual Installation

Download the latest binary from the [releases page](https://github.com/raucheacho/rosia-cli/releases).

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

## Verify Installation

Check that Rosia is installed correctly:

```bash
rosia version
```

You should see output similar to:

```
Rosia CLI v0.1.0
Built with Go 1.21+
```

## First Scan

Let's scan your current directory to see what can be cleaned:

```bash
rosia scan .
```

Rosia will traverse your directory tree and identify cleanable targets like `node_modules`, `target`, `build`, etc.

Example output:

```
Scanning /Users/you/projects...

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

## Interactive Cleaning

Launch the interactive TUI for visual selection:

```bash
rosia ui .
```

Use keyboard controls to navigate and select targets:

- `↑/↓` - Navigate through targets
- `Space` - Toggle selection
- `a` - Select all
- `n` - Deselect all
- `Enter` - Confirm and clean selected targets
- `q` - Quit without cleaning

## Clean with Confirmation

Clean detected targets with a confirmation prompt:

```bash
rosia clean
```

You'll be asked to confirm before deletion:

```
About to clean 15 targets (1.7 GB)
Continue? [y/N]: y

Cleaning...
✓ Cleaned /Users/you/projects/app/node_modules (450 MB)
✓ Cleaned /Users/you/projects/api/target (1.2 GB)
...

Cleaned 1.7 GB in 15 targets (5.2s)
```

## Restore Deleted Files

If you accidentally delete something, restore it from trash:

```bash
# List trashed items
rosia restore --list

# Restore specific item
rosia restore 20250428_143022_node_modules
```

## Configuration

Create a configuration file at `~/.rosiarc.json` to customize Rosia's behavior:

```json
{
  "trash_retention_days": 7,
  "profiles": ["node", "python", "rust", "flutter", "go"],
  "ignore_paths": [
    "/usr/local",
    "/System"
  ],
  "concurrency": 8,
  "telemetry_enabled": false
}
```

See the [Configuration](/configuration/) page for detailed options.

## Next Steps

- Learn about all available [Commands](/commands/)
- Customize [Configuration](/configuration/) for your workflow
- Explore [Plugins](/plugins/) to extend functionality
- Check [Statistics](/commands/#rosia-stats) to track your disk space savings

## Common Workflows

### Clean all Node.js projects

```bash
rosia scan ~/projects
rosia clean --yes
```

### Scan with custom depth

```bash
rosia scan ~/projects --depth 3
```

### Dry-run to preview changes

```bash
rosia scan ~/projects --dry-run
```

### View cleaning history

```bash
rosia stats
```

## Troubleshooting

### Permission Denied

If you encounter permission errors, ensure you have write access to the directories being cleaned. Rosia will skip directories it cannot access and continue with others.

### Trash Directory Full

By default, trash items are kept for 3 days. If your trash directory grows too large, you can:

1. Reduce retention period: `rosia config set trash_retention_days 1`
2. Manually clean trash: `rm -rf ~/.rosia/trash/*`

### Slow Scanning

For very large directory trees, you can:

1. Limit scan depth: `rosia scan . --depth 5`
2. Increase concurrency: `rosia config set concurrency 16`
3. Add ignore paths to skip large directories

## Getting Help

- Run `rosia --help` for command-line help
- Run `rosia <command> --help` for command-specific help
- Visit [GitHub Issues](https://github.com/raucheacho/rosia-cli/issues) for support
