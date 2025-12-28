// Package sdk provides a simple API for creating versionator plugins.
//
// There are three types of plugins:
//   - EmitPlugin: generates version source files
//   - BuildPlugin: generates build/linker flags
//   - PatchPlugin: patches version in config/manifest files
//
// Example emit plugin:
//
//	package main
//
//	import "github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
//
//	type GoEmit struct{}
//
//	func (p *GoEmit) Name() string          { return "emit-go" }
//	func (p *GoEmit) Format() string        { return "go" }
//	func (p *GoEmit) FileExtension() string { return ".go" }
//	func (p *GoEmit) DefaultOutput() string { return "version/version.go" }
//	func (p *GoEmit) Emit(vars map[string]string) (string, error) {
//	    return fmt.Sprintf(`const Version = "%s"`, vars["Version"]), nil
//	}
//
//	func main() { sdk.ServeEmit(&GoEmit{}) }
package sdk

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
	goplugin "github.com/hashicorp/go-plugin"
)

// --- Emit Plugin ---

// EmitPlugin generates version source files
type EmitPlugin = plugin.EmitPluginInterface

// ServeEmit starts an emit plugin server
func ServeEmit(impl EmitPlugin) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins: map[string]goplugin.Plugin{
			plugin.PluginTypeEmit: &plugin.EmitGRPCPlugin{Impl: impl},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}

// --- Build Plugin ---

// BuildPlugin generates build/linker flags
type BuildPlugin = plugin.BuildPluginInterface

// ServeBuild starts a build plugin server
func ServeBuild(impl BuildPlugin) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins: map[string]goplugin.Plugin{
			plugin.PluginTypeBuild: &plugin.BuildGRPCPlugin{Impl: impl},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}

// --- Patch Plugin ---

// PatchPlugin patches version in config/manifest files
type PatchPlugin = plugin.PatchPluginInterface

// ServePatch starts a patch plugin server
func ServePatch(impl PatchPlugin) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins: map[string]goplugin.Plugin{
			plugin.PluginTypePatch: &plugin.PatchGRPCPlugin{Impl: impl},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
