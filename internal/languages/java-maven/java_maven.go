// Package javamaven provides the Java/Maven language plugin for versionator.
//
// Maven projects use pom.xml for project configuration including version.
// This plugin configures versionator for Maven-based Java projects:
//
//	<project>
//	  <groupId>com.example</groupId>
//	  <artifactId>myapp</artifactId>
//	  <version>1.2.3</version>
//	</project>
//
// Injection methods:
//   - emit: Generate Version.java with constants
//   - patch: Update version in pom.xml
package javamaven

import (
	"github.com/benjaminabbitt/versionator/internal/plugin"
)

// Plugin implements the LanguagePlugin interface for Java/Maven
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "java-maven"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "java-maven"
}

// GetEmitConfig returns configuration for Java source file emission
func (p *Plugin) GetEmitConfig() *plugin.EmitConfig {
	return &plugin.EmitConfig{
		DefaultOutputPath:  "src/main/java/version/Version.java",
		DefaultPackageName: "version",
		FileExtension:      ".java",
	}
}

// GetBuildConfig returns nil - Java doesn't support link-time injection
func (p *Plugin) GetBuildConfig() *plugin.LinkConfig {
	return nil
}

// GetPatchConfigs returns configuration for patching pom.xml
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return []plugin.PatchConfig{
		{
			Name:        "pom.xml",
			FilePath:    "pom.xml",
			Format:      plugin.PatchFormatXML,
			VersionPath: "project.version",
			Description: "Maven project manifest",
			Patch:       plugin.PatchXML(),
		},
	}
}

func init() {
	plugin.Register(&Plugin{})
}
