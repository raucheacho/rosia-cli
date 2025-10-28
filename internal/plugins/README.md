# Rosia Plugin System

The Rosia plugin system allows you to extend Rosia's functionality by adding support for additional tools and technologies not covered by default profiles.

## Plugin Interface

All plugins must implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Scan(ctx context.Context) ([]types.Target, error)
    Clean(ctx context.Context, targets []types.Target) error
}
```

## Creating a Plugin

### 1. Create a new Go module for your plugin

```bash
mkdir rosia-docker-plugin
cd rosia-docker-plugin
go mod init github.com/yourusername/rosia-docker-plugin
```

### 2. Implement the Plugin interface

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
    // Execute: docker images --filter "dangling=true" -q
    cmd := exec.CommandContext(ctx, "docker", "images", "--filter", "dangling=true", "-q")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    imageIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
    targets := make([]types.Target, 0, len(imageIDs))
    
    for _, imageID := range imageIDs {
        if imageID == "" {
            continue
        }
        
        targets = append(targets, types.Target{
            Path:        imageID,
            Type:        "docker-image",
            ProfileName: "rosia-docker",
            IsDirectory: false,
        })
    }
    
    return targets, nil
}

func (p *DockerPlugin) Clean(ctx context.Context, targets []types.Target) error {
    for _, target := range targets {
        if target.ProfileName != "rosia-docker" {
            continue
        }
        
        // Execute: docker rmi <image_id>
        cmd := exec.CommandContext(ctx, "docker", "rmi", target.Path)
        if err := cmd.Run(); err != nil {
            return err
        }
    }
    
    return nil
}

// Export the plugin
var Plugin DockerPlugin
```

### 3. Build the plugin as a shared object

```bash
go build -buildmode=plugin -o rosia-docker.so
```

### 4. Install the plugin

```bash
mkdir -p ~/.rosia/plugins
cp rosia-docker.so ~/.rosia/plugins/
```

### 5. Verify the plugin is loaded

```bash
rosia plugin list
```

## Plugin Directory

Plugins are loaded from `~/.rosia/plugins/`. All `.so` files in this directory will be automatically loaded when Rosia starts.

## Plugin Commands

- `rosia plugin list` - List all loaded plugins
- `rosia plugin info <plugin-name>` - Show detailed information about a plugin

## Notes

- Plugins run in the same process as Rosia, so they have full access to the system
- Plugin errors are isolated and won't crash the main application
- Plugins are called during both scan and clean operations
- Plugin targets are merged with core profile targets
