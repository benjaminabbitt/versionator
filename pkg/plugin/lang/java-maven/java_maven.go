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
//
// # SNAPSHOT Versions
//
// Maven SNAPSHOT versions (e.g., 1.2.3-SNAPSHOT) are supported via the
// prerelease command:
//
//	versionator prerelease set SNAPSHOT    # Sets version to 1.2.3-SNAPSHOT
//	versionator prerelease clear           # Removes SNAPSHOT suffix
//
// Maven handles SNAPSHOT timestamp resolution (e.g., 1.2.3-20231215.143022-1)
// internally during deployment via "mvn deploy". Versionator does not generate
// Maven-style timestamps as this is Maven's responsibility during the artifact
// deployment process.
//
// Typical workflow:
//  1. Development: 1.2.3-SNAPSHOT (set via prerelease)
//  2. Release: 1.2.3 (clear prerelease, increment as needed)
//  3. Next dev cycle: 1.2.4-SNAPSHOT
package javamaven

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
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
