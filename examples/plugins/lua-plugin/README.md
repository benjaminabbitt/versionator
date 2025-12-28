# Lua Language Plugin for Versionator

This is an example external language plugin that adds Lua support to versionator.

## Features

- **Emit**: Generates `version.lua` files with version constants
- **Patch**: Updates version in `.rockspec` files (LuaRocks package spec)

## Building

```bash
go build -o versionator-plugin-lua .
```

## Installing

Copy the built binary to one of the plugin directories:

```bash
# User-specific (recommended)
mkdir -p ~/.config/versionator/plugins
cp versionator-plugin-lua ~/.config/versionator/plugins/

# Or use environment variable
export VERSIONATOR_PLUGIN_DIR=/path/to/plugins
cp versionator-plugin-lua $VERSIONATOR_PLUGIN_DIR/
```

## Plugin Search Paths

Versionator searches for plugins in these directories (in order):

1. `$VERSIONATOR_PLUGIN_DIR` (if set)
2. `~/.config/versionator/plugins/`
3. `~/.versionator/plugins/`
4. `/usr/local/lib/versionator/plugins/` (Unix only)
5. `/usr/lib/versionator/plugins/` (Unix only)

## Plugin Naming Convention

External plugin binaries must be named with the prefix `versionator-plugin-`:

- `versionator-plugin-lua`
- `versionator-plugin-mylang`
- `versionator-plugin-custom.exe` (Windows)

## Usage

Once installed, the plugin is automatically discovered and loaded:

```bash
# Emit a version.lua file
versionator emit file lua

# Patch rockspec files
versionator emit patch
```

## Development

This plugin uses the versionator SDK. Key interfaces to implement:

```go
type LanguagePlugin interface {
    Name() string
    LanguageName() string
    GetEmitConfig() *EmitConfig
    GetBuildConfig() *LinkConfig  // nil for interpreted languages
    GetPatchConfigs() []PatchConfig
    Patch(configName, content, version string) (string, error)
}
```

See the [plugin SDK documentation](../../../pkg/plugin/sdk/) for more details.
