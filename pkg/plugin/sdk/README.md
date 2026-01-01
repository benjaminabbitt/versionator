# Versionator Plugin SDK

The versionator SDK provides a simple API for creating external plugins.

## Overview

Versionator uses [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) for external plugin communication via gRPC. This SDK handles all the gRPC boilerplate, allowing you to focus on implementing your plugin logic.

## Plugin Types

There are three types of external plugins:

| Type | Purpose | Serve Function |
|------|---------|----------------|
| **Emit** | Generate version source files | `ServeEmit()` |
| **Build** | Generate build/linker flags | `ServeBuild()` |
| **Patch** | Patch version in manifest files | `ServePatch()` |

## Quick Start: Emit Plugin

### 1. Create a new Go module

```bash
mkdir versionator-plugin-emit-mylang
cd versionator-plugin-emit-mylang
go mod init github.com/yourname/versionator-plugin-emit-mylang
```

### 2. Add the dependency

```bash
go get github.com/benjaminabbitt/versionator
```

### 3. Implement the plugin

```go
package main

import (
    "fmt"
    "github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

type MyEmit struct{}

func (p *MyEmit) Name() string          { return "emit-mylang" }
func (p *MyEmit) Format() string        { return "mylang" }
func (p *MyEmit) FileExtension() string { return ".mylang" }
func (p *MyEmit) DefaultOutput() string { return "version.mylang" }

func (p *MyEmit) Emit(vars map[string]string) (string, error) {
    return fmt.Sprintf(`VERSION = "%s"`, vars["Version"]), nil
}

func main() {
    sdk.ServeEmit(&MyEmit{})
}
```

### 4. Build and install

```bash
go build -o versionator-plugin-emit-mylang .
mkdir -p ~/.config/versionator/plugins
cp versionator-plugin-emit-mylang ~/.config/versionator/plugins/
```

## Quick Start: Patch Plugin

```go
package main

import (
    "regexp"
    "github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

type MyPatch struct{}

func (p *MyPatch) Name() string        { return "patch-myconfig" }
func (p *MyPatch) FilePattern() string { return "myconfig.json" }
func (p *MyPatch) Description() string { return "My config file" }

func (p *MyPatch) Patch(content, version string) (string, error) {
    re := regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)
    if !re.MatchString(content) {
        return content, nil
    }
    return re.ReplaceAllString(content, `${1}"`+version+`"`), nil
}

func main() {
    sdk.ServePatch(&MyPatch{})
}
```

## Quick Start: Build Plugin

```go
package main

import (
    "fmt"
    "github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

type MyBuild struct{}

func (p *MyBuild) Name() string   { return "build-mylang" }
func (p *MyBuild) Format() string { return "mylang" }

func (p *MyBuild) GenerateFlags(vars map[string]string) (string, error) {
    return fmt.Sprintf("-DVERSION=%s", vars["Version"]), nil
}

func main() {
    sdk.ServeBuild(&MyBuild{})
}
```

## Interface Reference

### EmitPlugin

Generates version source files:

```go
type EmitPlugin interface {
    Name() string                              // Plugin name (e.g., "emit-go")
    Format() string                            // Format identifier (e.g., "go")
    FileExtension() string                     // File extension (e.g., ".go")
    DefaultOutput() string                     // Default output path
    Emit(vars map[string]string) (string, error) // Generate content
}
```

The `vars` map contains template variables like `Version`, `Major`, `Minor`, `Patch`, `PreRelease`, `Metadata`, `GitHash`, `GitBranch`, `BuildDate`.

### BuildPlugin

Generates build/linker flags for compiled languages:

```go
type BuildPlugin interface {
    Name() string                                    // Plugin name (e.g., "build-go")
    Format() string                                  // Format identifier
    GenerateFlags(vars map[string]string) (string, error) // Generate flags
}
```

### PatchPlugin

Patches version in manifest/config files:

```go
type PatchPlugin interface {
    Name() string                                 // Plugin name (e.g., "patch-npm")
    FilePattern() string                          // File pattern (e.g., "package.json")
    Description() string                          // Human-readable description
    Patch(content, version string) (string, error) // Patch content
}
```

## Plugin Discovery

Versionator searches for plugins in these directories (in order):

1. `$VERSIONATOR_PLUGIN_DIR` (environment variable)
2. `~/.config/versionator/plugins/`
3. `~/.versionator/plugins/`
4. `/usr/local/lib/versionator/plugins/` (Unix only)
5. `/usr/lib/versionator/plugins/` (Unix only)

## Naming Convention

Plugin binaries must be named with the appropriate prefix:

| Type | Prefix | Example |
|------|--------|---------|
| Emit | `versionator-plugin-emit-` | `versionator-plugin-emit-lua` |
| Build | `versionator-plugin-build-` | `versionator-plugin-build-rust` |
| Patch | `versionator-plugin-patch-` | `versionator-plugin-patch-npm` |

## Example

See the [Lua emit plugin example](../../../examples/plugins/lua-plugin/) for a complete working plugin.

## Protocol Details

Communication uses gRPC with the following handshake:

- Protocol Version: 1
- Magic Cookie Key: `VERSIONATOR_PLUGIN`
- Magic Cookie Value: `v1`

The protobuf definition is in `pkg/plugin/proto/plugin.proto`.
