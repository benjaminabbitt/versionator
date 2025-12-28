// versionator-plugin-build-rust generates Rust environment variable flags for version injection.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed flags.tmpl
var flagsTemplate string

// RustBuild implements the BuildPlugin interface for Rust.
type RustBuild struct{}

func (p *RustBuild) Name() string {
	return "build-rust"
}

func (p *RustBuild) Format() string {
	return "rust"
}

func (p *RustBuild) GenerateFlags(vars map[string]string) (string, error) {
	return mustache.Render(flagsTemplate, vars)
}

func main() {
	sdk.ServeBuild(&RustBuild{})
}
