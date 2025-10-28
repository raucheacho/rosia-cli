---
title: "Plugins"
description: "Extend Rosia with custom plugins"
---

# Plugins

Rosia's plugin system allows you to extend its functionality with custom cleaning logic for technologies and tools not covered by built-in profiles.

## Overview

Plugins enable you to:

- Add support for new technologies (Docker, Xcode, Android, etc.)
- Implement custom scanning logic
- Create specialized cleaning workflows
- Integrate with external tools

## Plugin Types

Rosia supports two types of plugins:

1. **Go Plugins** - Native Go plugins using Go's plugin system
2. **JSON-RPC Plugins** - External executables communicating via JSON-RPC (any language)

## Using Plugins

### Installing Plugins

1. Download or build the plugin
2. Place it in `~/.rosia/plugins/`
3. Enable it in your configuration

```bash
# Create plugins directory
mkdir -p ~/.rosia/plugins

# Copy plugin
cp rosia-docker.so ~/.rosia/plugins/

# Enable plugin
rosia config set plugins rosia-docker
```

### Listing Plugins

View all loaded plugins:

```bash
rosia plugin list
```

### Plugin Information

Get detailed information about a plugin:

```bash
rosia plugin info rosia-docker
```

## Available Plugins

### rosia-docker

Clean Docker images, volumes, and containers.

**Features:**
- Scan for dangling images
- Clean unused volumes
- Remove stopped containers

**Installation:**
```bash
go install github.com/raucheacho/rosia-docker@latest
cp $(go env GOPATH)/bin/rosia-docker ~/.rosia/plugins/
rosia config set plugins rosia-docker
```

### rosia-xcode

Clean Xcode derived data and archives.

**Features:**
- Clean DerivedData directory
- Remove old archives
- Clean module cache

**Installation:**
```bash
brew install rosia-xcode
rosia config set plugins rosia-xcode
```

## Creating Go Plugins

### Plugin Interface

All Go plugins must implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Scan(ctx context.Context) ([]Target, error)
    Clean(ctx context.Context, targets []Target) error
}
```

### Basic Plugin Example

Create a file `myplugin.go`:

```go
package main

import (
    "context"
    "os/exec"
    "strings"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Description() string {
    return "Custom cleaning logic for my tools"
}

func (p *MyPlugin) Scan(ctx context.Context) ([]Target, error) {
    var targets []Target
    
    // Example: Find cache directories
    cmd := exec.CommandContext(ctx, "find", "/tmp", "-name", "my-cache-*")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    paths := strings.Split(string(output), "\n")
    for _, path := range paths {
        if path == "" {
            continue
        }
        
        // Calculate size
        size := calculateSize(path)
        
        targets = append(targets, Target{
            Path:        path,
            Size:        size,
            Type:        "cache",
            ProfileName: "my-plugin",
        })
    }
    
    return targets, nil
}

func (p *MyPlugin) Clean(ctx context.Context, targets []Target) error {
    for _, target := range targets {
        // Implement cleaning logic
        cmd := exec.CommandContext(ctx, "rm", "-rf", target.Path)
        if err := cmd.Run(); err != nil {
            return err
        }
    }
    return nil
}

// Export the plugin
var Plugin MyPlugin

func calculateSize(path string) int64 {
    // Implement size calculation
    return 0
}
```

### Building the Plugin

```bash
go build -buildmode=plugin -o myplugin.so myplugin.go
```

### Installing the Plugin

```bash
cp myplugin.so ~/.rosia/plugins/
rosia config set plugins my-plugin
```

### Testing the Plugin

```bash
rosia plugin list
rosia plugin info my-plugin
rosia scan .
```

## Creating JSON-RPC Plugins

JSON-RPC plugins allow you to write plugins in any language.

### Plugin Manifest

Create a manifest file `plugin.json`:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Custom cleaning logic",
  "executable": "./bin/plugin",
  "protocol": "jsonrpc",
  "methods": {
    "scan": "scan",
    "clean": "clean"
  }
}
```

### Python Example

Create `plugin.py`:

```python
#!/usr/bin/env python3
import json
import sys
import os

def scan():
    """Scan for cleanable targets"""
    targets = []
    
    # Example: Find Python cache directories
    for root, dirs, files in os.walk('/tmp'):
        if '__pycache__' in dirs:
            path = os.path.join(root, '__pycache__')
            size = get_dir_size(path)
            
            targets.append({
                'path': path,
                'size': size,
                'type': 'cache',
                'profile_name': 'my-plugin'
            })
    
    return targets

def clean(targets):
    """Clean specified targets"""
    import shutil
    
    for target in targets:
        try:
            shutil.rmtree(target['path'])
        except Exception as e:
            print(f"Error cleaning {target['path']}: {e}", file=sys.stderr)
    
    return {'success': True}

def get_dir_size(path):
    """Calculate directory size"""
    total = 0
    for dirpath, dirnames, filenames in os.walk(path):
        for f in filenames:
            fp = os.path.join(dirpath, f)
            if os.path.exists(fp):
                total += os.path.getsize(fp)
    return total

def handle_request(request):
    """Handle JSON-RPC request"""
    method = request.get('method')
    params = request.get('params', {})
    
    if method == 'scan':
        result = scan()
    elif method == 'clean':
        result = clean(params.get('targets', []))
    else:
        return {'error': f'Unknown method: {method}'}
    
    return {'result': result}

if __name__ == '__main__':
    # Read JSON-RPC request from stdin
    for line in sys.stdin:
        request = json.loads(line)
        response = handle_request(request)
        print(json.dumps(response))
        sys.stdout.flush()
```

### Node.js Example

Create `plugin.js`:

```javascript
#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const readline = require('readline');

async function scan() {
  const targets = [];
  
  // Example: Find node_modules in /tmp
  const findNodeModules = (dir) => {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const fullPath = path.join(dir, entry.name);
        
        if (entry.name === 'node_modules') {
          const size = getDirSize(fullPath);
          targets.push({
            path: fullPath,
            size: size,
            type: 'dependencies',
            profile_name: 'my-plugin'
          });
        }
      }
    }
  };
  
  findNodeModules('/tmp');
  return targets;
}

async function clean(targets) {
  for (const target of targets) {
    try {
      fs.rmSync(target.path, { recursive: true, force: true });
    } catch (err) {
      console.error(`Error cleaning ${target.path}:`, err);
    }
  }
  return { success: true };
}

function getDirSize(dirPath) {
  let size = 0;
  const files = fs.readdirSync(dirPath, { withFileTypes: true });
  
  for (const file of files) {
    const filePath = path.join(dirPath, file.name);
    if (file.isDirectory()) {
      size += getDirSize(filePath);
    } else {
      size += fs.statSync(filePath).size;
    }
  }
  
  return size;
}

async function handleRequest(request) {
  const { method, params = {} } = request;
  
  let result;
  if (method === 'scan') {
    result = await scan();
  } else if (method === 'clean') {
    result = await clean(params.targets || []);
  } else {
    return { error: `Unknown method: ${method}` };
  }
  
  return { result };
}

// Read JSON-RPC requests from stdin
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', async (line) => {
  const request = JSON.parse(line);
  const response = await handleRequest(request);
  console.log(JSON.stringify(response));
});
```

### Installing JSON-RPC Plugin

```bash
# Create plugin directory
mkdir -p ~/.rosia/plugins/my-plugin

# Copy files
cp plugin.json ~/.rosia/plugins/my-plugin/
cp plugin.py ~/.rosia/plugins/my-plugin/bin/plugin
chmod +x ~/.rosia/plugins/my-plugin/bin/plugin

# Enable plugin
rosia config set plugins my-plugin
```

## Plugin Best Practices

### Error Handling

Always handle errors gracefully:

```go
func (p *MyPlugin) Scan(ctx context.Context) ([]Target, error) {
    targets, err := p.scanInternal(ctx)
    if err != nil {
        // Log error but don't crash
        log.Printf("Plugin scan error: %v", err)
        return []Target{}, nil // Return empty list
    }
    return targets, nil
}
```

### Context Cancellation

Respect context cancellation:

```go
func (p *MyPlugin) Scan(ctx context.Context) ([]Target, error) {
    for _, path := range paths {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            // Continue scanning
        }
    }
    return targets, nil
}
```

### Performance

- Use concurrent operations for large scans
- Cache results when possible
- Limit memory usage
- Provide progress updates

### Security

- Validate all inputs
- Don't execute arbitrary commands
- Check permissions before deletion
- Use safe file operations

## Plugin Development Workflow

1. **Design** - Define what your plugin will clean
2. **Implement** - Write the plugin code
3. **Test** - Test with sample data
4. **Build** - Compile (Go) or package (JSON-RPC)
5. **Install** - Copy to plugins directory
6. **Enable** - Add to configuration
7. **Verify** - Test with Rosia commands

## Debugging Plugins

### Enable Verbose Logging

```bash
rosia scan . --verbose
```

### Check Plugin Loading

```bash
rosia plugin list
```

### Test Plugin Directly

For Go plugins:

```bash
go run -buildmode=plugin myplugin.go
```

For JSON-RPC plugins:

```bash
echo '{"method":"scan","params":{}}' | ./plugin.py
```

## Plugin Examples

### Docker Plugin

Clean Docker resources:

```go
func (p *DockerPlugin) Scan(ctx context.Context) ([]Target, error) {
    // Find dangling images
    cmd := exec.CommandContext(ctx, "docker", "images", "-f", "dangling=true", "-q")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    imageIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
    var targets []Target
    
    for _, id := range imageIDs {
        if id == "" {
            continue
        }
        
        // Get image size
        cmd := exec.CommandContext(ctx, "docker", "image", "inspect", id, "--format", "{{.Size}}")
        output, _ := cmd.Output()
        size, _ := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
        
        targets = append(targets, Target{
            Path:        id,
            Size:        size,
            Type:        "docker-image",
            ProfileName: "docker",
        })
    }
    
    return targets, nil
}
```

### Xcode Plugin

Clean Xcode derived data:

```go
func (p *XcodePlugin) Scan(ctx context.Context) ([]Target, error) {
    homeDir, _ := os.UserHomeDir()
    derivedDataPath := filepath.Join(homeDir, "Library", "Developer", "Xcode", "DerivedData")
    
    var targets []Target
    entries, err := os.ReadDir(derivedDataPath)
    if err != nil {
        return nil, err
    }
    
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        
        path := filepath.Join(derivedDataPath, entry.Name())
        size := calculateSize(path)
        
        targets = append(targets, Target{
            Path:        path,
            Size:        size,
            Type:        "derived-data",
            ProfileName: "xcode",
        })
    }
    
    return targets, nil
}
```

## Contributing Plugins

Want to share your plugin with the community?

1. Create a GitHub repository
2. Add documentation
3. Submit to the [plugin registry](https://github.com/raucheacho/rosia-plugins)
4. Share on social media

## Plugin Registry

Browse available plugins at: https://github.com/raucheacho/rosia-plugins

## Support

Need help with plugin development?

- 📖 [Plugin API Documentation](https://pkg.go.dev/github.com/raucheacho/rosia-cli/internal/plugins)
- 💬 [GitHub Discussions](https://github.com/raucheacho/rosia-cli/discussions)
- 🐛 [Report Issues](https://github.com/raucheacho/rosia-cli/issues)
