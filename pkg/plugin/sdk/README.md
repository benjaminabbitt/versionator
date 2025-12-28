# Versionator Plugin SDK

The versionator SDK provides a simple API for creating external language plugins.

## Overview

Versionator uses [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) for external plugin communication via gRPC. This SDK handles all the gRPC boilerplate, allowing you to focus on implementing your language-specific logic.

## Quick Start

### 1. Create a new Go module

```bash
mkdir my-language-plugin
cd my-language-plugin
go mod init github.com/yourname/versionator-plugin-mylang
```

### 2. Add the dependency

```bash
go get github.com/benjaminabbitt/versionator
```

### 3. Implement the plugin

```go
package main

import (
    "github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

type MyLanguagePlugin struct{}

func (p *MyLanguagePlugin) Name() string {
    return "my-language"
}

func (p *MyLanguagePlugin) LanguageName() string {
    return "mylang"
}

func (p *MyLanguagePlugin) GetEmitConfig() *sdk.EmitConfig {
    return &sdk.EmitConfig{
        DefaultOutputPath:  "version.mylang",
        DefaultPackageName: "",
        FileExtension:      ".mylang",
    }
}

func (p *MyLanguagePlugin) GetBuildConfig() *sdk.LinkConfig {
    // Return nil for interpreted languages
    return nil
}

func (p *MyLanguagePlugin) GetPatchConfigs() []sdk.PatchConfig {
    return []sdk.PatchConfig{
        sdk.NewPatchConfig(
            "config.json",           // name
            "config.json",           // file path
            sdk.PatchFormatJSON,     // format
            "version",               // version path
            "Configuration file",    // description
            sdk.PatchJSON(),         // patcher function
        ),
    }
}

func (p *MyLanguagePlugin) Patch(configName, content, version string) (string, error) {
    // Use the built-in patchers or implement custom logic
    return sdk.PatchJSON()(content, version)
}

func main() {
    sdk.Serve(&MyLanguagePlugin{})
}
```

### 4. Build the plugin

```bash
go build -o versionator-plugin-mylang .
```

### 5. Install the plugin

```bash
mkdir -p ~/.config/versionator/plugins
cp versionator-plugin-mylang ~/.config/versionator/plugins/
```

## Interface Reference

### LanguagePlugin

The main interface your plugin must implement:

```go
type LanguagePlugin interface {
    // Name returns the unique plugin name
    Name() string

    // LanguageName returns the language identifier (e.g., "go", "python")
    LanguageName() string

    // GetEmitConfig returns configuration for source file emission
    // Return nil if emit is not supported
    GetEmitConfig() *EmitConfig

    // GetBuildConfig returns configuration for link-time variable injection
    // Return nil for interpreted languages
    GetBuildConfig() *LinkConfig

    // GetPatchConfigs returns configurations for manifest file patching
    // Return nil or empty slice if patching is not supported
    GetPatchConfigs() []PatchConfig

    // Patch performs the actual patching operation
    Patch(configName, content, version string) (string, error)
}
```

### EmitConfig

Configuration for generating version source files:

```go
type EmitConfig struct {
    // DefaultOutputPath is the default path for the generated version file
    DefaultOutputPath string

    // DefaultPackageName is the default package/module name
    DefaultPackageName string

    // FileExtension is the file extension (e.g., ".go", ".py")
    FileExtension string
}
```

### LinkConfig

Configuration for link-time variable injection (compiled languages only):

```go
type LinkConfig struct {
    // VariablePath is the variable to override (e.g., "main.Version")
    VariablePath string

    // FlagTemplate is the compiler flag template
    // Use {{Variable}} and {{Value}} placeholders
    FlagTemplate string
}
```

### PatchConfig

Configuration for patching manifest/config files:

```go
type PatchConfig struct {
    Name        string      // Human-readable name
    FilePath    string      // Default file path
    Format      PatchFormat // File format (json, toml, yaml, xml, etc.)
    VersionPath string      // Path to version field
    Description string      // What this patch target is for
    Patch       PatchFunc   // The patching function
}
```

## Built-in Patchers

The SDK provides common patchers for popular file formats:

| Function | Format | Example Files |
|----------|--------|---------------|
| `PatchJSON()` | JSON | package.json, composer.json |
| `PatchTOML()` | TOML | Cargo.toml, pyproject.toml |
| `PatchYAML()` | YAML | pubspec.yaml |
| `PatchXML()` | XML | pom.xml, *.csproj |
| `PatchGradle()` | Gradle | build.gradle, build.gradle.kts |
| `PatchPythonSetup()` | Python | setup.py |
| `PatchRubyGemspec()` | Ruby | *.gemspec |
| `PatchSwiftPackage()` | Swift | Package.swift |

## Plugin Discovery

Versionator searches for plugins in these directories (in order):

1. `$VERSIONATOR_PLUGIN_DIR` (environment variable)
2. `~/.config/versionator/plugins/`
3. `~/.versionator/plugins/`
4. `/usr/local/lib/versionator/plugins/` (Unix only)
5. `/usr/lib/versionator/plugins/` (Unix only)

## Naming Convention

Plugin binaries must be named with the prefix `versionator-plugin-`:

- `versionator-plugin-lua`
- `versionator-plugin-custom`
- `versionator-plugin-mylang.exe` (Windows)

## Example

See the [Lua plugin example](../../../examples/plugins/lua-plugin/) for a complete working plugin.

## Protocol Details

Communication uses gRPC with the following handshake:

- Protocol Version: 1
- Magic Cookie Key: `VERSIONATOR_PLUGIN`
- Magic Cookie Value: `language`

The protobuf definition is in `pkg/plugin/proto/plugin.proto`.
