# Changelog

All notable changes to Rosia CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-10-28

### Added

#### Core Features
- **Scanner Engine**: Fast concurrent directory scanning with configurable worker pools
- **Cleaner Engine**: Safe file deletion with trash system and restoration capability
- **Profile System**: Built-in support for Node.js, Python, Rust, Flutter, and Go projects
- **Trash System**: Temporary storage for deleted files with configurable retention period
- **Configuration Management**: User preferences stored in ~/.rosiarc.json
- **Telemetry System**: Local statistics tracking (opt-in for cloud sync)
- **Plugin System**: Extensible architecture for custom cleaning logic

#### CLI Commands
- `rosia scan`: Scan directories for cleanable files and caches
- `rosia clean`: Clean detected targets with confirmation and trash backup
- `rosia restore`: Restore trashed items to original locations
- `rosia ui`: Interactive TUI for visual selection and cleaning
- `rosia config`: Manage configuration (show, set, reset)
- `rosia stats`: Display usage statistics
- `rosia plugin`: Manage plugins (list, info)
- `rosia version`: Display version information

#### User Interface
- **CLI**: Command-line interface with rich help text and examples
- **TUI**: Interactive terminal UI with keyboard navigation
  - Arrow keys for navigation
  - Space for selection toggle
  - Batch operations (select all, deselect all)
  - Real-time progress indicators
  - Confirmation dialogs

#### Technology Support
- **Node.js**: node_modules, dist, build, .next, .cache, coverage
- **Python**: venv, __pycache__, .pytest_cache, .tox, dist, build
- **Rust**: target/
- **Flutter**: build/, .dart_tool/
- **Go**: vendor/, bin/

#### Safety Features
- Confirmation prompts before deletion
- Files moved to trash by default (not permanently deleted)
- Trash retention period (default: 3 days)
- Permission checks before deletion
- Dry-run mode for testing
- Error isolation (continue on individual failures)

#### Performance
- Concurrent scanning with worker pools
- Configurable concurrency (auto-detect based on CPU cores)
- Efficient size calculation
- Progress indicators for long operations
- Optimized for large directory trees

#### Cross-Platform Support
- Linux (amd64, arm64, arm)
- macOS (amd64, arm64)
- Windows (amd64)
- Platform-specific configuration paths
- Platform-appropriate file operations

#### Documentation
- Comprehensive README with installation and usage instructions
- Detailed help text for all commands
- Contributing guidelines
- Release guide
- Plugin development guide

#### Testing
- Unit tests for all core components
- Integration tests for end-to-end flows
- Performance benchmarks
- Test coverage > 80%

#### Build & Distribution
- GoReleaser configuration for automated releases
- GitHub Actions CI/CD pipeline
- Homebrew formula for macOS/Linux
- Scoop manifest for Windows
- Installation scripts (install.sh, install.ps1)
- Docker support (optional)

### Technical Details

#### Architecture
- Clean architecture with separation of concerns
- Modular design for easy extension
- Interface-based abstractions
- Dependency injection for testability

#### Dependencies
- cobra: CLI framework
- bubbletea: TUI framework
- lipgloss: Terminal styling
- testify: Testing utilities

#### Configuration
- Default config location: ~/.rosiarc.json
- Configurable trash retention period
- Configurable concurrency
- Profile enable/disable
- Ignore paths
- Plugin management
- Telemetry opt-in

### Known Limitations

- Plugin system requires Go plugins (.so files) or JSON-RPC
- Trash system uses local storage (no cloud backup)
- Telemetry is local-only by default
- No GUI (terminal-only)

### Future Enhancements

See [GitHub Issues](https://github.com/raucheacho/rosia-cli/issues) for planned features.

[Unreleased]: https://github.com/raucheacho/rosia-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/raucheacho/rosia-cli/releases/tag/v0.1.0
