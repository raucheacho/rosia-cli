---
title: "Configuration"
description: "Configure Rosia CLI for your workflow"
---

# Configuration

Rosia uses a JSON configuration file to customize its behavior. This guide covers all configuration options and how to manage them.

## Configuration File

The configuration file is located at `~/.rosiarc.json`. If it doesn't exist, Rosia uses default settings.

### Default Configuration

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

## Configuration Options

### trash_retention_days

**Type:** `integer`  
**Default:** `3`  
**Description:** Number of days to keep items in trash before automatic cleanup.

```json
{
  "trash_retention_days": 7
}
```

Set via CLI:

```bash
rosia config set trash_retention_days 7
```

**Recommendations:**
- `1-3 days` - For frequent cleaners with limited disk space
- `7 days` - Balanced approach for most users
- `14+ days` - For cautious users who want extended recovery time

### profiles

**Type:** `array of strings`  
**Default:** `["node", "python", "rust", "flutter", "go"]`  
**Description:** List of enabled profile names. Only enabled profiles are used during scanning.

```json
{
  "profiles": ["node", "python", "rust"]
}
```

Set via CLI:

```bash
rosia config set profiles node,python,rust
```

**Available Profiles:**
- `node` - Node.js projects
- `python` - Python projects
- `rust` - Rust projects
- `flutter` - Flutter projects
- `go` - Go projects

### ignore_paths

**Type:** `array of strings`  
**Default:** `[]`  
**Description:** Absolute paths to exclude from scanning. Useful for system directories or sensitive locations.

```json
{
  "ignore_paths": [
    "/usr/local",
    "/System",
    "/Users/you/important-project"
  ]
}
```

Set via CLI:

```bash
rosia config set ignore_paths /usr/local,/System
```

**Common Ignore Paths:**

**macOS:**
```json
{
  "ignore_paths": [
    "/System",
    "/Library",
    "/usr/local"
  ]
}
```

**Linux:**
```json
{
  "ignore_paths": [
    "/usr",
    "/lib",
    "/bin",
    "/sbin"
  ]
}
```

**Windows:**
```json
{
  "ignore_paths": [
    "C:\\Windows",
    "C:\\Program Files",
    "C:\\Program Files (x86)"
  ]
}
```

### plugins

**Type:** `array of strings`  
**Default:** `[]`  
**Description:** List of enabled plugin names. Plugins extend Rosia's functionality.

```json
{
  "plugins": ["rosia-docker", "rosia-xcode"]
}
```

Set via CLI:

```bash
rosia config set plugins rosia-docker,rosia-xcode
```

See the [Plugins](/plugins/) page for available plugins and how to create your own.

### concurrency

**Type:** `integer`  
**Default:** `0` (auto-detect)  
**Description:** Number of concurrent workers for scanning and cleaning. `0` means auto-detect (NumCPU * 2).

```json
{
  "concurrency": 8
}
```

Set via CLI:

```bash
rosia config set concurrency 8
```

**Recommendations:**
- `0` - Auto-detect (recommended for most users)
- `4-8` - Good for systems with 2-4 CPU cores
- `16+` - For high-end systems with many cores
- Lower values reduce CPU/memory usage but slow down operations

### telemetry_enabled

**Type:** `boolean`  
**Default:** `false`  
**Description:** Enable anonymous usage statistics. When enabled, Rosia sends anonymized data to help improve the tool.

```json
{
  "telemetry_enabled": true
}
```

Set via CLI:

```bash
rosia config set telemetry_enabled true
```

**What's Collected:**
- Number of scans performed
- Total size cleaned
- Profile usage statistics
- Error types (no personal data)

**What's NOT Collected:**
- File paths
- Directory names
- Personal information
- Project details

## Managing Configuration

### View Current Configuration

```bash
rosia config show
```

### Set Individual Values

```bash
rosia config set <key> <value>
```

Examples:

```bash
rosia config set trash_retention_days 7
rosia config set concurrency 8
rosia config set telemetry_enabled true
```

### Reset to Defaults

```bash
rosia config reset
```

This will reset all settings to their default values.

### Manual Editing

You can also edit the configuration file directly:

```bash
# macOS/Linux
nano ~/.rosiarc.json

# Windows
notepad %USERPROFILE%\.rosiarc.json
```

After editing, verify the configuration:

```bash
rosia config show
```

## Profile Configuration

Profiles define cleaning rules for specific technologies. They are stored in the `profiles/` directory.

### Profile Structure

Each profile is a JSON file with the following structure:

```json
{
  "name": "Node.js",
  "version": "1.0.0",
  "patterns": [
    "node_modules",
    "dist",
    "build",
    ".next",
    ".cache",
    "coverage"
  ],
  "detect": [
    "package.json",
    "package-lock.json",
    "yarn.lock"
  ],
  "description": "Cleans Node.js project artifacts",
  "enabled": true
}
```

### Profile Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name of the profile |
| `version` | string | Profile version |
| `patterns` | array | Directory/file names to clean |
| `detect` | array | Files that indicate this technology |
| `description` | string | Human-readable description |
| `enabled` | boolean | Whether the profile is active |

### Built-in Profiles

#### Node.js (`node.json`)

```json
{
  "name": "Node.js",
  "version": "1.0.0",
  "patterns": [
    "node_modules",
    "dist",
    "build",
    ".next",
    ".cache",
    "coverage"
  ],
  "detect": [
    "package.json",
    "package-lock.json",
    "yarn.lock"
  ],
  "description": "Cleans Node.js project artifacts",
  "enabled": true
}
```

#### Python (`python.json`)

```json
{
  "name": "Python",
  "version": "1.0.0",
  "patterns": [
    "venv",
    "__pycache__",
    ".pytest_cache",
    ".tox",
    "dist",
    "build",
    "*.egg-info"
  ],
  "detect": [
    "requirements.txt",
    "setup.py",
    "pyproject.toml"
  ],
  "description": "Cleans Python project artifacts",
  "enabled": true
}
```

#### Rust (`rust.json`)

```json
{
  "name": "Rust",
  "version": "1.0.0",
  "patterns": [
    "target"
  ],
  "detect": [
    "Cargo.toml",
    "Cargo.lock"
  ],
  "description": "Cleans Rust project artifacts",
  "enabled": true
}
```

#### Flutter (`flutter.json`)

```json
{
  "name": "Flutter",
  "version": "1.0.0",
  "patterns": [
    "build",
    ".dart_tool"
  ],
  "detect": [
    "pubspec.yaml",
    "pubspec.lock"
  ],
  "description": "Cleans Flutter project artifacts",
  "enabled": true
}
```

#### Go (`go.json`)

```json
{
  "name": "Go",
  "version": "1.0.0",
  "patterns": [
    "vendor",
    "bin"
  ],
  "detect": [
    "go.mod",
    "go.sum"
  ],
  "description": "Cleans Go project artifacts",
  "enabled": true
}
```

### Custom Profiles

You can create custom profiles by adding JSON files to the `profiles/` directory:

1. Create a new JSON file in `~/.rosia/profiles/`:

```bash
mkdir -p ~/.rosia/profiles
nano ~/.rosia/profiles/custom.json
```

2. Define your profile:

```json
{
  "name": "Custom",
  "version": "1.0.0",
  "patterns": [
    "my-cache",
    "temp-files"
  ],
  "detect": [
    "my-project.config"
  ],
  "description": "Custom cleaning rules",
  "enabled": true
}
```

3. Enable the profile:

```bash
rosia config set profiles node,python,custom
```

## Environment Variables

Rosia respects these environment variables:

### ROSIA_CONFIG

Override the default config file location:

```bash
export ROSIA_CONFIG=/path/to/custom/config.json
rosia scan .
```

### ROSIA_TRASH_DIR

Override the default trash directory:

```bash
export ROSIA_TRASH_DIR=/path/to/custom/trash
rosia clean
```

### ROSIA_LOG_LEVEL

Set the log level:

```bash
export ROSIA_LOG_LEVEL=debug
rosia scan .
```

Valid values: `debug`, `info`, `warn`, `error`

## Configuration Examples

### Minimal Configuration

For users who want fast cleaning with minimal safety:

```json
{
  "trash_retention_days": 1,
  "profiles": ["node"],
  "concurrency": 16,
  "telemetry_enabled": false
}
```

### Cautious Configuration

For users who want maximum safety:

```json
{
  "trash_retention_days": 14,
  "profiles": ["node", "python", "rust", "flutter", "go"],
  "ignore_paths": [
    "/usr/local",
    "/System",
    "/Users/you/important-projects"
  ],
  "concurrency": 4,
  "telemetry_enabled": false
}
```

### Performance Configuration

For users with large codebases:

```json
{
  "trash_retention_days": 3,
  "profiles": ["node", "python", "rust"],
  "concurrency": 32,
  "telemetry_enabled": true
}
```

### Team Configuration

For teams with shared standards:

```json
{
  "trash_retention_days": 7,
  "profiles": ["node", "python"],
  "ignore_paths": [
    "/usr/local",
    "/System"
  ],
  "plugins": ["rosia-docker"],
  "concurrency": 0,
  "telemetry_enabled": false
}
```

## Troubleshooting

### Configuration Not Loading

If your configuration isn't being applied:

1. Verify the file exists:
```bash
ls -la ~/.rosiarc.json
```

2. Check JSON syntax:
```bash
cat ~/.rosiarc.json | python -m json.tool
```

3. View current config:
```bash
rosia config show
```

### Invalid Configuration

If Rosia reports invalid configuration:

1. Reset to defaults:
```bash
rosia config reset
```

2. Manually fix the JSON file
3. Verify with `rosia config show`

### Profiles Not Working

If profiles aren't detecting projects:

1. Verify profiles are enabled:
```bash
rosia config show
```

2. Check profile files exist:
```bash
ls -la ~/.rosia/profiles/
```

3. Test with verbose logging:
```bash
rosia scan . --verbose
```
