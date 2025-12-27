// Package javagradle provides the Java/Gradle language plugin for versionator.
//
// Gradle projects use build.gradle or build.gradle.kts for configuration.
// This plugin configures versionator for Gradle-based Java projects:
//
// Groovy DSL (build.gradle):
//
//	version = '1.2.3'
//
// Kotlin DSL (build.gradle.kts):
//
//	version = "1.2.3"
//
// Injection methods:
//   - emit: Generate Version.java with constants
//   - patch: Update version in build.gradle
//
// Note: Gradle's DSL is not a standard format. The patch implementation
// uses simple text replacement for the version assignment.
package javagradle

import (
	"github.com/benjaminabbitt/versionator/internal/plugin"
)

// Plugin implements the LanguagePlugin interface for Java/Gradle
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "java-gradle"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "java-gradle"
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

// GetPatchConfigs returns configuration for patching Gradle build files
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return []plugin.PatchConfig{
		{
			Name:        "build.gradle",
			FilePath:    "build.gradle",
			Format:      plugin.PatchFormatTOML,
			VersionPath: "version",
			Description: "Gradle build script (Groovy DSL)",
			Patch:       plugin.PatchGradle(),
		},
		{
			Name:        "build.gradle.kts",
			FilePath:    "build.gradle.kts",
			Format:      plugin.PatchFormatTOML,
			VersionPath: "version",
			Description: "Gradle build script (Kotlin DSL)",
			Patch:       plugin.PatchGradle(),
		},
	}
}

func init() {
	plugin.Register(&Plugin{})
}
