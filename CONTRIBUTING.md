# Contributing to Rosia CLI

Thank you for your interest in contributing to Rosia! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Adding New Profiles](#adding-new-profiles)
- [Creating Plugins](#creating-plugins)
- [Code Style and Standards](#code-style-and-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for convenience commands)

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/raucheacho/rosia-cli.git
cd rosia-cli

# Install dependencies
go mod download

# Build the binary
go build -o rosia .

# Run tests
go test ./...

# Run with verbose output
go test -v ./...
```

### Development Workflow

```bash
# Run the CLI during development
go run . scan .

# Build and install locally
go install .

# Run linter (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...
```

## Project Structure

```
rosia-cli/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── scan.go            # Scan command
│   ├── clean.go           # Clean command
│   ├── restore.go         # Restore command
│   ├── config.go          # Config command
│   ├── stats.go           # Stats command
│   ├── plugin.go          # Plugin command
│   └── ui.go              # TUI command
├── internal/              # Internal packages (not exported)
│   ├── scanner/           # Directory scanning engine
│   │   ├── scanner.go     # Core scanner logic
│   │   └── async.go       # Concurrent scanning
│   ├── cleaner/           # File deletion engine
│   │   └── cleaner.go     # Core cleaner logic
│   ├── profiles/          # Profile system
│   │   ├── loader.go      # Profile loading
│   │   └── matcher.go     # Profile matching
│   ├── plugins/           # Plugin system
│   │   ├── plugin.go      # Plugin interface
│   │   ├── loader.go      # Plugin loading
│   │   └── registry.go    # Plugin registry
│   ├── config/            # Configuration management
│   │   └── config.go      # Config loading/saving
│   ├── trash/             # Trash system
│   │   └── trash.go       # Trash operations
│   ├── telemetry/         # Statistics tracking
│   │   └── telemetry.go   # Telemetry store
│   ├── sizecalc/          # Size calculation
│   │   └── sizecalc.go    # Directory size calculation
│   ├── fsutils/           # File system utilities
│   │   ├── fsutils.go     # File operations
│   │   └── paths.go       # Path utilities
│   └── ui/                # TUI components
│       ├── model.go       # Bubble Tea model
│       ├── views.go       # View rendering
│       ├── commands.go    # Commands
│       ├── messages.go    # Messages
│       └── keys.go        # Key bindings
├── pkg/                   # Public packages (exported)
│   ├── types/             # Core types and errors
│   │   └── types.go       # Target, Profile, Config, etc.
│   ├── logger/            # Logging utilities
│   │   └── logger.go      # Color-coded logger
│   └── progress/          # Progress indicators
│       └── progress.go    # Progress bar
├── profiles/              # Built-in profile definitions
│   ├── node.json          # Node.js profile
│   ├── python.json        # Python profile
│   ├── rust.json          # Rust profile
│   ├── flutter.json       # Flutter profile
│   └── go.json            # Go profile
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── go.sum                 # Go module checksums
```

### Architecture Overview

Rosia follows clean architecture principles with clear separation of concerns:

1. **CLI Layer** (`cmd/`): Handles user interaction via Cobra commands
2. **TUI Layer** (`internal/ui/`): Interactive terminal interface using Bubble Tea
3. **Core Engine Layer** (`internal/scanner/`, `internal/cleaner/`): Business logic
4. **Profile & Plugin Management** (`internal/profiles/`, `internal/plugins/`): Extensibility
5. **Storage Layer** (`internal/config/`, `internal/trash/`, `internal/telemetry/`): Persistence

## Adding New Profiles

Profiles define cleaning rules for specific technology stacks. They are JSON files located in the `profiles/` directory.

### Profile Structure

```json
{
  "name": "Technology Name",
  "version": "1.0.0",
  "patterns": [
    "directory_to_clean",
    "another_directory",
    "*.cache"
  ],
  "detect": [
    "indicator_file.json",
    "another_indicator"
  ],
  "description": "Brief description of what this profile cleans",
  "enabled": true
}
```

### Profile Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name of the technology |
| `version` | string | Yes | Profile version (semver) |
| `patterns` | string[] | Yes | Directories/files to clean (supports glob patterns) |
| `detect` | string[] | Yes | Files that indicate this technology is present |
| `description` | string | Yes | Human-readable description |
| `enabled` | bool | No | Whether profile is enabled by default (default: true) |

### Example: Adding a Java/Maven Profile

Create `profiles/java.json`:

```json
{
  "name": "Java",
  "version": "1.0.0",
  "patterns": [
    "target",
    ".gradle",
    "build",
    ".m2/repository"
  ],
  "detect": [
    "pom.xml",
    "build.gradle",
    "build.gradle.kts"
  ],
  "description": "Cleans Java/Maven/Gradle build artifacts",
  "enabled": true
}
```

### Testing Your Profile

```bash
# Create a test project structure
mkdir -p test-project
cd test-project
touch pom.xml
mkdir target

# Run Rosia scan
rosia scan .

# Verify your profile is detected
# You should see "Java" in the profile name column
```

### Profile Best Practices

1. **Be Specific**: Only include directories that are safe to delete
2. **Test Thoroughly**: Test on real projects before submitting
3. **Document Patterns**: Add comments in PR explaining each pattern
4. **Consider Edge Cases**: Think about monorepos and nested projects
5. **Avoid System Directories**: Never target system or user directories

## Creating Plugins

Plugins extend Rosia's functionality beyond built-in profiles. Plugins can be written in Go or any language using JSON-RPC.

### Plugin Interface

All plugins must implement this interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Scan(ctx context.Context) ([]Target, error)
    Clean(ctx context.Context, targets []Target) error
}
```

### Go Plugin Example

Create a new Go module for your plugin:

```bash
mkdir rosia-docker-plugin
cd rosia-docker-plugin
go mod init github.com/yourusername/rosia-docker-plugin
```

Implement the plugin (`main.go`):

```go
package main

import (
    "context"
    "os/exec"
    "strings"
    
    "github.com/raucheacho/rosia-cli/pkg/types"
)

type DockerPlugin struct{}

func (p *DockerPlugin) Name() string {
    return "rosia-docker"
}

func (p *DockerPlugin) Version() string {
    return "1.0.0"
}

func (p *DockerPlugin) Description() string {
    return "Cleans dangling Docker images and containers"
}

func (p *DockerPlugin) Scan(ctx context.Context) ([]types.Target, error) {
    var targets []types.Target
    
    // Find dangling images
    cmd := exec.CommandContext(ctx, "docker", "images", "-f", "dangling=true", "-q")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    imageIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
    for _, id := range imageIDs {
        if id == "" {
            continue
        }
        
        // Get image size
        sizeCmd := exec.CommandContext(ctx, "docker", "image", "inspect", id, "--format", "{{.Size}}")
        sizeOutput, _ := sizeCmd.Output()
        
        targets = append(targets, types.Target{
            Path:        "docker://" + id,
            Size:        parseSize(string(sizeOutput)),
            Type:        "docker-image",
            ProfileName: "docker",
            IsDirectory: false,
        })
    }
    
    return targets, nil
}

func (p *DockerPlugin) Clean(ctx context.Context, targets []types.Target) error {
    for _, target := range targets {
        imageID := strings.TrimPrefix(target.Path, "docker://")
        cmd := exec.CommandContext(ctx, "docker", "rmi", imageID)
        if err := cmd.Run(); err != nil {
            return err
        }
    }
    return nil
}

func parseSize(s string) int64 {
    // Implement size parsing logic
    return 0
}

// Export the plugin
var Plugin DockerPlugin
```

Build as a plugin:

```bash
go build -buildmode=plugin -o rosia-docker.so .
```

Install the plugin:

```bash
mkdir -p ~/.rosia/plugins
cp rosia-docker.so ~/.rosia/plugins/
```

### JSON-RPC Plugin Example

For non-Go languages, implement a JSON-RPC server:

**Plugin Manifest** (`plugin.json`):

```json
{
  "name": "rosia-xcode",
  "version": "1.0.0",
  "description": "Cleans Xcode derived data",
  "executable": "./xcode-plugin",
  "protocol": "jsonrpc"
}
```

**Python Example** (`xcode-plugin`):

```python
#!/usr/bin/env python3
import json
import sys
import os

def scan():
    derived_data = os.path.expanduser("~/Library/Developer/Xcode/DerivedData")
    targets = []
    
    if os.path.exists(derived_data):
        for item in os.listdir(derived_data):
            path = os.path.join(derived_data, item)
            size = get_dir_size(path)
            targets.append({
                "path": path,
                "size": size,
                "type": "xcode-derived-data",
                "profile_name": "xcode",
                "is_directory": True
            })
    
    return targets

def clean(targets):
    import shutil
    for target in targets:
        shutil.rmtree(target["path"])
    return None

def get_dir_size(path):
    total = 0
    for dirpath, dirnames, filenames in os.walk(path):
        for f in filenames:
            fp = os.path.join(dirpath, f)
            total += os.path.getsize(fp)
    return total

# JSON-RPC handler
def handle_request(request):
    method = request.get("method")
    params = request.get("params", {})
    
    if method == "scan":
        result = scan()
    elif method == "clean":
        result = clean(params.get("targets", []))
    else:
        return {"error": "Unknown method"}
    
    return {"result": result}

if __name__ == "__main__":
    for line in sys.stdin:
        request = json.loads(line)
        response = handle_request(request)
        print(json.dumps(response))
        sys.stdout.flush()
```

Make it executable:

```bash
chmod +x xcode-plugin
```

### Plugin Best Practices

1. **Error Handling**: Always handle errors gracefully
2. **Context Cancellation**: Respect context cancellation for long operations
3. **Resource Cleanup**: Clean up resources properly
4. **Testing**: Test plugins independently before integration
5. **Documentation**: Document what your plugin does and any requirements
6. **Permissions**: Check for required permissions before operations
7. **Idempotency**: Ensure clean operations are idempotent

## Code Style and Standards

### Go Style Guide

- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write clear, self-documenting code
- Add comments for exported functions and types

### Naming Conventions

- **Packages**: Short, lowercase, single-word names
- **Files**: Lowercase with underscores (e.g., `scanner_test.go`)
- **Types**: PascalCase (e.g., `ScanOptions`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE for package-level

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to scan directory %s: %w", path, err)
}

// Good: Use custom error types
if errors.Is(err, types.ErrPermissionDenied) {
    logger.Warn("Permission denied: %s", path)
    continue
}

// Bad: Ignore errors
_ = someOperation()

// Bad: Generic error messages
if err != nil {
    return errors.New("error occurred")
}
```

### Logging

```go
// Use the logger package
logger.Info("Scanning directory: %s", path)
logger.Warn("Skipping hidden directory: %s", path)
logger.Error("Failed to delete: %s", err)
logger.Debug("Worker %d processing target", workerID)
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/scanner/...

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

```go
func TestScanner_Scan(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
    
    profiles := []types.Profile{{
        Name:     "test",
        Patterns: []string{"node_modules"},
        Detect:   []string{"package.json"},
    }}
    
    scanner := scanner.NewScanner(profiles)
    
    // Execute
    targets, err := scanner.Scan(context.Background(), []string{tmpDir}, scanner.ScanOptions{})
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, targets, 1)
    assert.Equal(t, "node_modules", filepath.Base(targets[0].Path))
}
```

### Test Coverage Goals

- Aim for 80%+ coverage on core packages
- 100% coverage on critical paths (deletion, trash operations)
- Integration tests for end-to-end flows
- Benchmark tests for performance-critical code

## Submitting Changes

### Pull Request Process

1. **Fork the repository** and create a feature branch
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** following the code style guidelines

3. **Write tests** for new functionality

4. **Run tests and linters**
   ```bash
   go test ./...
   go fmt ./...
   golangci-lint run
   ```

5. **Commit with clear messages**
   ```bash
   git commit -m "feat: add Java profile support"
   ```

6. **Push to your fork**
   ```bash
   git push origin feature/my-new-feature
   ```

7. **Open a Pull Request** with:
   - Clear description of changes
   - Reference to related issues
   - Screenshots/examples if applicable
   - Test results

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(profiles): add Java/Maven profile support

Add profile for cleaning Java build artifacts including Maven
target directories and Gradle build directories.

Closes #123
```

```
fix(scanner): handle symlinks correctly

Previously, the scanner would follow symlinks which could lead
to infinite loops. Now symlinks are detected and skipped.

Fixes #456
```

### Code Review Process

1. Maintainers will review your PR within 3-5 business days
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR
4. Your contribution will be included in the next release

## Questions or Issues?

- 💬 [Discussions](https://github.com/raucheacho/rosia-cli/discussions) - Ask questions
- 🐛 [Issues](https://github.com/raucheacho/rosia-cli/issues) - Report bugs
- 📧 Email: [contact@raucheacho.com]

Thank you for contributing to Rosia! 🎉
