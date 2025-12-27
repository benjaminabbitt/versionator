// Package pythonsetuptools provides the Python/Setuptools language plugin for versionator.
//
// Legacy Python projects using setuptools may have version in:
//   - setup.cfg: INI-style configuration
//   - setup.py: Python script (not patchable)
//
// setup.cfg example:
//
//	[metadata]
//	name = mypackage
//	version = 1.2.3
//
// For modern Python projects, use the "python" plugin which targets pyproject.toml.
//
// Injection methods:
//   - emit: Generate _version.py with __version__
//   - patch: Update version in setup.cfg
package pythonsetuptools

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
)

// Plugin implements the LanguagePlugin interface for Python/Setuptools
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "python-setuptools"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "python-setuptools"
}

// GetEmitConfig returns configuration for Python source file emission
func (p *Plugin) GetEmitConfig() *plugin.EmitConfig {
	return &plugin.EmitConfig{
		DefaultOutputPath:  "_version.py",
		DefaultPackageName: "",
		FileExtension:      ".py",
	}
}

// GetBuildConfig returns nil - Python doesn't support link-time injection
func (p *Plugin) GetBuildConfig() *plugin.LinkConfig {
	return nil
}

// GetPatchConfigs returns configuration for patching setup.cfg
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return []plugin.PatchConfig{
		{
			Name:        "setup.cfg",
			FilePath:    "setup.cfg",
			Format:      plugin.PatchFormatTOML, // INI-like format
			VersionPath: "metadata.version",
			Description: "Setuptools configuration (legacy)",
		},
	}
}

func init() {
	plugin.Register(&Plugin{})
}
