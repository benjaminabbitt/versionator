// Example versionator emit plugin for Lua.
//
// This plugin demonstrates how to create an external emit plugin
// using the versionator SDK. It generates version.lua files.
//
// To build:
//
//	go build -o versionator-plugin-emit-lua .
//
// To install:
//
//	cp versionator-plugin-emit-lua ~/.config/versionator/plugins/
//
// Or set VERSIONATOR_PLUGIN_DIR to the directory containing the plugin.
package main

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// LuaEmit implements the EmitPlugin interface for Lua.
type LuaEmit struct{}

// Name returns the plugin name.
func (p *LuaEmit) Name() string {
	return "emit-lua"
}

// Format returns the format identifier.
func (p *LuaEmit) Format() string {
	return "lua"
}

// FileExtension returns the file extension.
func (p *LuaEmit) FileExtension() string {
	return ".lua"
}

// DefaultOutput returns the default output path.
func (p *LuaEmit) DefaultOutput() string {
	return "version.lua"
}

// Emit generates Lua source code with version information.
func (p *LuaEmit) Emit(vars map[string]string) (string, error) {
	return fmt.Sprintf(`-- Auto-generated version file
local M = {}

M.VERSION = "%s"
M.MAJOR = %s
M.MINOR = %s
M.PATCH = %s
M.PRERELEASE = "%s"
M.METADATA = "%s"
M.GIT_HASH = "%s"
M.GIT_BRANCH = "%s"
M.BUILD_DATE = "%s"

return M
`, sdk.GetVar(vars, "Version", "0.0.0"),
		sdk.GetNumericVar(vars, "Major"),
		sdk.GetNumericVar(vars, "Minor"),
		sdk.GetNumericVar(vars, "Patch"),
		sdk.GetVar(vars, "PreRelease", ""),
		sdk.GetVar(vars, "Metadata", ""),
		sdk.GetVar(vars, "GitHash", ""),
		sdk.GetVar(vars, "GitBranch", ""),
		sdk.GetVar(vars, "BuildDate", "")), nil
}

func main() {
	sdk.ServeEmit(&LuaEmit{})
}
