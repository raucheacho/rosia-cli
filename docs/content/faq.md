---
title: "FAQ"
description: "Frequently Asked Questions about Rosia CLI"
---

# Frequently Asked Questions

## General

### What is Rosia?

Rosia is a command-line tool that helps developers reclaim disk space by cleaning dependencies, builds, and caches across multiple project types. It supports Node.js, Python, Rust, Flutter, Go, and more.

### Is Rosia safe to use?

Yes! Rosia includes several safety features:
- Confirmation prompts before deletion
- Trash system for recovery
- Permission checks before deletion
- Dry-run mode to preview changes

### How much disk space can I save?

It depends on your projects, but users typically save:
- 500MB - 2GB per Node.js project (node_modules)
- 100MB - 1GB per Rust project (target/)
- 50MB - 500MB per Python project (venv, __pycache__)

### Does Rosia delete source code?

No! Rosia only targets build artifacts, dependencies, and caches. It never touches your source code files.

## Installation

### Which platforms are supported?

Rosia works on:
- macOS (Intel and Apple Silicon)
- Linux (x86_64, ARM64)
- Windows (x86_64)

### How do I update Rosia?

**Homebrew:**
```bash
brew upgrade rosia
```

**Scoop:**
```bash
scoop update rosia
```

**Go Install:**
```bash
go install github.com/raucheacho/rosia-cli@latest
```

### Can I install Rosia without admin privileges?

Yes! Use Go install or download the binary to a user directory:

```bash
# Download to user bin
mkdir -p ~/bin
curl -L https://github.com/raucheacho/rosia-cli/releases/latest/download/rosia_$(uname -s)_$(uname -m).tar.gz | tar xz -C ~/bin
export PATH="$HOME/bin:$PATH"
```

## Usage

### How do I scan multiple directories?

```bash
rosia scan ~/projects ~/workspace ~/repos
```

### Can I exclude certain directories?

Yes! Add them to your configuration:

```bash
rosia config set ignore_paths /path/to/exclude,/another/path
```

Or use the ignore_paths in `~/.rosiarc.json`:

```json
{
  "ignore_paths": [
    "/usr/local",
    "/System",
    "/Users/you/important-project"
  ]
}
```

### How do I clean without confirmation?

Use the `--yes` flag:

```bash
rosia clean --yes
```

**Warning:** This skips the confirmation prompt. Use with caution!

### Can I restore deleted files?

Yes! Files are moved to trash before deletion:

```bash
# List trashed items
rosia restore --list

# Restore specific item
rosia restore 20250428_143022_node_modules
```

### How long are files kept in trash?

By default, 3 days. You can change this:

```bash
rosia config set trash_retention_days 7
```

## Performance

### Why is scanning slow?

Large directory trees take time to traverse. You can:

1. Limit scan depth:
```bash
rosia scan . --depth 5
```

2. Increase concurrency:
```bash
rosia config set concurrency 16
```

3. Add ignore paths for large directories

### How many workers should I use?

By default, Rosia uses `NumCPU * 2`. For most systems:
- 2-4 cores: 4-8 workers
- 4-8 cores: 8-16 workers
- 8+ cores: 16-32 workers

Set manually:
```bash
rosia config set concurrency 16
```

### Does Rosia use a lot of memory?

No. Rosia streams results and uses minimal memory even for large scans. Typical usage is under 100MB.

## Configuration

### Where is the config file?

`~/.rosiarc.json`

### How do I reset configuration?

```bash
rosia config reset
```

### Can I use multiple config files?

Yes! Use the `--config` flag:

```bash
rosia scan . --config /path/to/custom-config.json
```

Or set the environment variable:

```bash
export ROSIA_CONFIG=/path/to/custom-config.json
rosia scan .
```

### How do I disable a profile?

Edit your config to remove it from the profiles list:

```bash
rosia config set profiles node,python,rust
```

## Profiles

### What profiles are available?

Built-in profiles:
- `node` - Node.js (node_modules, dist, build, .next, .cache, coverage)
- `python` - Python (venv, __pycache__, .pytest_cache, .tox, dist, build)
- `rust` - Rust (target/)
- `flutter` - Flutter (build/, .dart_tool/)
- `go` - Go (vendor/, bin/)

### Can I create custom profiles?

Yes! Create a JSON file in `~/.rosia/profiles/`:

```json
{
  "name": "Custom",
  "version": "1.0.0",
  "patterns": ["my-cache", "temp-files"],
  "detect": ["my-project.config"],
  "description": "Custom cleaning rules",
  "enabled": true
}
```

Then enable it:
```bash
rosia config set profiles node,python,custom
```

### Why isn't my project detected?

Make sure:
1. The profile is enabled in your config
2. The detection files exist (e.g., package.json for Node.js)
3. You're scanning the correct directory

Test with verbose logging:
```bash
rosia scan . --verbose
```

## Plugins

### What are plugins?

Plugins extend Rosia with custom cleaning logic for tools not covered by built-in profiles (Docker, Xcode, Android, etc.).

### How do I install plugins?

1. Download or build the plugin
2. Place it in `~/.rosia/plugins/`
3. Enable it in your config:

```bash
rosia config set plugins rosia-docker
```

### Can I write plugins in other languages?

Yes! Rosia supports JSON-RPC plugins, which can be written in any language (Python, Node.js, Ruby, etc.).

See the [Plugins](/plugins/) page for details.

### Where can I find plugins?

Browse the [plugin registry](https://github.com/raucheacho/rosia-plugins) for community plugins.

## Troubleshooting

### "Permission denied" errors

Rosia skips directories it cannot access. This is normal for system directories. To avoid these warnings, add them to ignore_paths:

```bash
rosia config set ignore_paths /usr/local,/System
```

### "No targets found"

This means Rosia didn't detect any cleanable directories. Possible reasons:
- No projects in the scanned directory
- Profiles not enabled
- Detection files missing

Try:
```bash
rosia scan . --verbose
```

### Trash directory is full

Reduce retention period or manually clean trash:

```bash
rosia config set trash_retention_days 1
# or
rm -rf ~/.rosia/trash/*
```

### Configuration not loading

Verify JSON syntax:

```bash
cat ~/.rosiarc.json | python -m json.tool
```

If invalid, reset:

```bash
rosia config reset
```

### Hugo site not building

Make sure you have Hugo Extended installed:

```bash
hugo version
```

Should show "extended" in the version string.

## Privacy & Telemetry

### Does Rosia collect data?

By default, no. Telemetry is opt-in only.

### What data is collected if I enable telemetry?

Only anonymized statistics:
- Number of scans
- Total size cleaned
- Profile usage
- Error types (no personal data)

**Never collected:**
- File paths
- Directory names
- Personal information
- Project details

### How do I enable/disable telemetry?

```bash
# Enable
rosia config set telemetry_enabled true

# Disable
rosia config set telemetry_enabled false
```

## Contributing

### How can I contribute?

- Report bugs on [GitHub Issues](https://github.com/raucheacho/rosia-cli/issues)
- Submit pull requests
- Create plugins
- Improve documentation
- Share Rosia with others

### How do I report a bug?

1. Check [existing issues](https://github.com/raucheacho/rosia-cli/issues)
2. Create a new issue with:
   - Rosia version (`rosia version`)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior

### Can I request features?

Yes! Open a [feature request](https://github.com/raucheacho/rosia-cli/issues/new) on GitHub.

## Support

### Where can I get help?

- 📖 [Documentation](https://rosia.raucheacho.com)
- 💬 [GitHub Discussions](https://github.com/raucheacho/rosia-cli/discussions)
- 🐛 [GitHub Issues](https://github.com/raucheacho/rosia-cli/issues)

### Is there a community?

Join the discussion on [GitHub Discussions](https://github.com/raucheacho/rosia-cli/discussions)!

## License

### What license is Rosia under?

MIT License. See [LICENSE](https://github.com/raucheacho/rosia-cli/blob/main/LICENSE) for details.

### Can I use Rosia commercially?

Yes! The MIT license allows commercial use.
