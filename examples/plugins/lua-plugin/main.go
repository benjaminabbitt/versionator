// Example versionator language plugin for Lua.
//
// This plugin demonstrates how to create an external language plugin
// using the versionator SDK. It supports:
// - Emitting a version.lua file
// - Patching rockspec files
//
// To build:
//
//	go build -o versionator-plugin-lua .
//
// To install:
//
//	cp versionator-plugin-lua ~/.config/versionator/plugins/
//
// Or set VERSIONATOR_PLUGIN_DIR to the directory containing the plugin.
package main

import (
	"fmt"
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// LuaPlugin implements the versionator language plugin interface for Lua.
type LuaPlugin struct{}

// Name returns the plugin name.
func (p *LuaPlugin) Name() string {
	return "lua"
}

// LanguageName returns the language identifier.
func (p *LuaPlugin) LanguageName() string {
	return "lua"
}

// GetEmitConfig returns configuration for source file emission.
func (p *LuaPlugin) GetEmitConfig() *sdk.EmitConfig {
	return &sdk.EmitConfig{
		DefaultOutputPath:  "version.lua",
		DefaultPackageName: "",
		FileExtension:      ".lua",
	}
}

// GetBuildConfig returns nil since Lua doesn't support link-time injection.
func (p *LuaPlugin) GetBuildConfig() *sdk.LinkConfig {
	return nil // Lua is interpreted, no link-time injection
}

// GetPatchConfigs returns configuration for patching rockspec files.
func (p *LuaPlugin) GetPatchConfigs() []sdk.PatchConfig {
	return []sdk.PatchConfig{
		{
			Name:        "rockspec",
			FilePath:    "*.rockspec",
			Format:      "lua", // Custom format
			VersionPath: "version",
			Description: "LuaRocks package specification",
			Patch:       patchRockspec,
		},
	}
}

// Patch performs the patching operation for a given config.
func (p *LuaPlugin) Patch(configName, content, version string) (string, error) {
	switch configName {
	case "rockspec":
		return patchRockspec(content, version)
	default:
		return "", fmt.Errorf("unknown patch config: %s", configName)
	}
}

// patchRockspec patches version in a .rockspec file.
// Matches: version = "1.0.0-1"
func patchRockspec(content, version string) (string, error) {
	// Rockspec version format: "1.0.0-1" (version-revision)
	// We update just the semver part, keeping revision as -1
	re := regexp.MustCompile(`(version\s*=\s*)"[^"]*"`)
	if !re.MatchString(content) {
		return "", nil
	}
	return re.ReplaceAllString(content, fmt.Sprintf(`${1}"%s-1"`, version)), nil
}

func main() {
	sdk.Serve(&LuaPlugin{})
}
