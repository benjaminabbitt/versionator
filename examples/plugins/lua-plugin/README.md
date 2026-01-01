# Lua Emit Plugin for Versionator

This is an example external emit plugin that adds Lua support to versionator.

## Features

- Generates `version.lua` files with version constants

## Building

```bash
go build -o versionator-plugin-emit-lua .
```

## Installing

Copy the built binary to one of the plugin directories:

```bash
# User-specific (recommended)
mkdir -p ~/.config/versionator/plugins
cp versionator-plugin-emit-lua ~/.config/versionator/plugins/

# Or use environment variable
export VERSIONATOR_PLUGIN_DIR=/path/to/plugins
cp versionator-plugin-emit-lua $VERSIONATOR_PLUGIN_DIR/
```

## Plugin Search Paths

Versionator searches for plugins in these directories (in order):

1. `$VERSIONATOR_PLUGIN_DIR` (if set)
2. `~/.config/versionator/plugins/`
3. `~/.versionator/plugins/`
4. `/usr/local/lib/versionator/plugins/` (Unix only)
5. `/usr/lib/versionator/plugins/` (Unix only)

## Plugin Naming Convention

External plugin binaries must be named with the appropriate prefix:

| Type | Prefix | Example |
|------|--------|---------|
| Emit | `versionator-plugin-emit-` | `versionator-plugin-emit-lua` |
| Build | `versionator-plugin-build-` | `versionator-plugin-build-rust` |
| Patch | `versionator-plugin-patch-` | `versionator-plugin-patch-npm` |

## Usage

Once installed, the plugin is automatically discovered and loaded:

```bash
# Emit a version.lua file
versionator emit file lua
```

## Development

This plugin uses the versionator SDK. The EmitPlugin interface:

```go
type EmitPlugin interface {
    Name() string                              // Plugin name
    Format() string                            // Format identifier
    FileExtension() string                     // File extension
    DefaultOutput() string                     // Default output path
    Emit(vars map[string]string) (string, error) // Generate content
}
```

See the [plugin SDK documentation](../../../pkg/plugin/sdk/) for more details.
