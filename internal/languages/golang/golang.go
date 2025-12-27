// Package golang provides the Go language plugin for versionator.
//
// Go projects typically use semantic versioning with a "v" prefix (e.g., v1.2.3).
// This plugin configures versionator with Go-specific defaults including:
//   - Version prefix: "v"
//   - Output path: internal/version/version.go
//   - Package name: version
//   - Pre-release template compatible with Go pseudo-versions
//
// The generated version file contains constants for use in Go applications:
//
//	package version
//
//	const (
//	    Version     = "1.2.3"
//	    Major       = 1
//	    Minor       = 2
//	    Patch       = 3
//	    PreRelease  = ""
//	    Metadata    = ""
//	    GitHash     = "abc1234"
//	    GitBranch   = "main"
//	    BuildDate   = "2024-01-15"
//	)
package golang

import (
	"github.com/benjaminabbitt/versionator/internal/plugin"
)

// Plugin implements the LanguagePlugin interface for Go
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "golang"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "go"
}

// GetEmitConfig returns configuration for Go source file emission
func (p *Plugin) GetEmitConfig() *plugin.EmitConfig {
	return &plugin.EmitConfig{
		DefaultOutputPath:  "internal/version/version.go",
		DefaultPackageName: "version",
		FileExtension:      ".go",
	}
}

// GetBuildConfig returns configuration for Go linker flag injection
func (p *Plugin) GetBuildConfig() *plugin.LinkConfig {
	return &plugin.LinkConfig{
		VariablePath: "main.Version",
		FlagTemplate: "-X {{Variable}}={{Value}}",
	}
}

// GetPatchConfigs returns nil - Go uses git tags, not manifest files
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return nil
}

func init() {
	plugin.Register(&Plugin{})
}
