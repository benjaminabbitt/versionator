// Package cpp provides the C++ language plugin for versionator.
//
// C++ projects can use version information in two ways:
//
// 1. Emit (header file):
//
//	#ifndef VERSION_HPP
//	#define VERSION_HPP
//
//	namespace version {
//	    constexpr const char* VERSION = "1.2.3";
//	    constexpr int MAJOR = 1;
//	    constexpr int MINOR = 2;
//	    constexpr int PATCH = 3;
//	}
//
//	#endif
//
// 2. Link (preprocessor defines):
//
//	g++ -DVERSION="1.2.3" -DVERSION_MAJOR=1 ...
//
// Both methods allow compile-time access to version information.
package cpp

import (
	"github.com/benjaminabbitt/versionator/internal/plugin"
)

// Plugin implements the LanguagePlugin interface for C++
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "cpp"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "cpp"
}

// GetEmitConfig returns configuration for C++ header file emission
func (p *Plugin) GetEmitConfig() *plugin.EmitConfig {
	return &plugin.EmitConfig{
		DefaultOutputPath:  "version.hpp",
		DefaultPackageName: "",
		FileExtension:      ".hpp",
	}
}

// GetBuildConfig returns configuration for C++ preprocessor define injection
func (p *Plugin) GetBuildConfig() *plugin.LinkConfig {
	return &plugin.LinkConfig{
		VariablePath: "VERSION",
		FlagTemplate: `-D{{Variable}}="{{Value}}"`,
	}
}

// GetPatchConfigs returns nil - C++ has no standard package manifest
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return nil
}

func init() {
	plugin.Register(&Plugin{})
}
