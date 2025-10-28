---
title: "Rosia CLI"
description: "Universal, fast, and secure CLI tool for cleaning development dependencies and caches"
featured_image: ""
---

# Rosia CLI

**Reclaim disk space from development dependencies and caches**

Rosia is a universal, fast, and secure command-line tool that helps developers reclaim disk space by cleaning dependencies, builds, and caches across multiple project types. Whether you're working with Node.js, Python, Rust, Flutter, or Go projects, Rosia intelligently detects and safely removes cleanable artifacts.

## Key Features

- 🚀 **Fast Scanning** - Concurrent directory traversal with configurable worker pools
- 🎯 **Multi-Technology Support** - Built-in profiles for Node.js, Python, Rust, Flutter, Go, and more
- 🛡️ **Safe Deletion** - Trash system with restoration capability before permanent removal
- 🎨 **Interactive TUI** - Beautiful terminal interface for visual selection and progress tracking
- 🔌 **Extensible** - Plugin system for custom cleaning logic
- ⚙️ **Configurable** - JSON-based configuration for profiles, ignore paths, and preferences
- 📊 **Statistics** - Track cleaning history and disk space savings over time
- 🌍 **Cross-Platform** - Works on Linux, macOS, and Windows

## Quick Start

Install Rosia:

```bash
# Homebrew (macOS/Linux)
brew install rosia

# Scoop (Windows)
scoop install rosia

# Go Install
go install github.com/raucheacho/rosia-cli@latest
```

Scan and clean your projects:

```bash
# Scan current directory
rosia scan .

# Launch interactive TUI
rosia ui .

# Clean with confirmation
rosia clean
```

## Why Rosia?

Development projects accumulate gigabytes of dependencies, build artifacts, and caches over time. Manually finding and cleaning these files across different project types is tedious and error-prone. Rosia automates this process with:

- **Intelligent Detection** - Automatically identifies project types and their cleanable artifacts
- **Safety First** - Trash system ensures you can restore accidentally deleted files
- **Speed** - Concurrent operations make cleaning large codebases fast
- **Flexibility** - Extensible through plugins and configurable profiles

## Get Started

Ready to reclaim your disk space? Check out the [Getting Started](/getting-started/) guide to install and configure Rosia for your workflow.

## Community & Support

- 🐛 [Report a bug](https://github.com/raucheacho/rosia-cli/issues/new)
- 💡 [Request a feature](https://github.com/raucheacho/rosia-cli/issues/new)
- 📖 [View on GitHub](https://github.com/raucheacho/rosia-cli)

---

Made with ❤️ by developers, for developers.
