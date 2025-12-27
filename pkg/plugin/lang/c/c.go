// Package c provides the C language plugin for versionator.
//
// C projects can use version information in two ways:
//
// 1. Emit (header file):
//
//	#ifndef VERSION_H
//	#define VERSION_H
//
//	#define VERSION "1.2.3"
//	#define VERSION_MAJOR 1
//	#define VERSION_MINOR 2
//	#define VERSION_PATCH 3
//
//	#endif
//
// 2. Link (preprocessor defines):
//
//	gcc -DVERSION="1.2.3" -DVERSION_MAJOR=1 ...
//
// Both methods allow compile-time access to version information.
package c

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
)

// Plugin implements the LanguagePlugin interface for C
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "c"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "c"
}

// GetEmitConfig returns configuration for C header file emission
func (p *Plugin) GetEmitConfig() *plugin.EmitConfig {
	return &plugin.EmitConfig{
		DefaultOutputPath:  "version.h",
		DefaultPackageName: "",
		FileExtension:      ".h",
	}
}

// GetBuildConfig returns configuration for C preprocessor define injection
func (p *Plugin) GetBuildConfig() *plugin.LinkConfig {
	return &plugin.LinkConfig{
		VariablePath: "VERSION",
		FlagTemplate: `-D{{Variable}}="{{Value}}"`,
	}
}

// GetPatchConfigs returns nil - C has no standard package manifest
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return nil
}

func init() {
	plugin.Register(&Plugin{})
}
