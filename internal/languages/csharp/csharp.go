// Package csharp provides the C#/.NET language plugin for versionator.
//
// .NET projects use .csproj files for project configuration including version:
//
//	<Project Sdk="Microsoft.NET.Sdk">
//	  <PropertyGroup>
//	    <Version>1.2.3</Version>
//	  </PropertyGroup>
//	</Project>
//
// The generated Version.cs file contains:
//
//	namespace Version
//	{
//	    public static class VersionInfo
//	    {
//	        public const string Version = "1.2.3";
//	        public const int Major = 1;
//	        public const int Minor = 2;
//	        public const int Patch = 3;
//	    }
//	}
//
// Injection methods:
//   - emit: Generate Version.cs with constants
//   - build: Pass /p:Version=X.Y.Z to dotnet build
//   - patch: Update Version element in .csproj
package csharp

import (
	"github.com/benjaminabbitt/versionator/internal/plugin"
)

// Plugin implements the LanguagePlugin interface for C#/.NET
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "csharp"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeLanguage)
}

// LanguageName returns the language identifier
func (p *Plugin) LanguageName() string {
	return "csharp"
}

// GetEmitConfig returns configuration for C# source file emission
func (p *Plugin) GetEmitConfig() *plugin.EmitConfig {
	return &plugin.EmitConfig{
		DefaultOutputPath:  "Version.cs",
		DefaultPackageName: "Version",
		FileExtension:      ".cs",
	}
}

// GetBuildConfig returns configuration for .NET MSBuild property injection.
// Use with: dotnet build $(versionator emit build csharp)
// In C# code: Assembly.GetExecutingAssembly().GetName().Version
func (p *Plugin) GetBuildConfig() *plugin.LinkConfig {
	return &plugin.LinkConfig{
		VariablePath: "Version",
		FlagTemplate: "/p:{{Variable}}={{Value}}",
	}
}

// GetPatchConfigs returns configuration for patching .csproj files
func (p *Plugin) GetPatchConfigs() []plugin.PatchConfig {
	return []plugin.PatchConfig{
		{
			Name:        "*.csproj",
			FilePath:    "*.csproj", // User specifies actual filename
			Format:      plugin.PatchFormatXML,
			VersionPath: "Project.PropertyGroup.Version",
			Description: ".NET project file",
			Patch:       plugin.PatchXML(),
		},
	}
}

func init() {
	plugin.Register(&Plugin{})
}
