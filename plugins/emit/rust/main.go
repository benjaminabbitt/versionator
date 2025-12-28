// versionator-plugin-emit-rust generates Rust version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed rust.tmpl
var rustTemplate string

// RustEmit implements the EmitPlugin interface for Rust.
type RustEmit struct{}

func (p *RustEmit) Name() string {
	return "emit-rust"
}

func (p *RustEmit) Format() string {
	return "rust"
}

func (p *RustEmit) FileExtension() string {
	return ".rs"
}

func (p *RustEmit) DefaultOutput() string {
	return "src/version.rs"
}

func (p *RustEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(rustTemplate, vars)
}

func main() {
	sdk.ServeEmit(&RustEmit{})
}
