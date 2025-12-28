// versionator-plugin-emit-go generates Go version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed go.tmpl
var goTemplate string

// GoEmit implements the EmitPlugin interface for Go.
type GoEmit struct{}

func (p *GoEmit) Name() string {
	return "emit-go"
}

func (p *GoEmit) Format() string {
	return "go"
}

func (p *GoEmit) FileExtension() string {
	return ".go"
}

func (p *GoEmit) DefaultOutput() string {
	return "internal/version/version.go"
}

func (p *GoEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(goTemplate, vars)
}

func main() {
	sdk.ServeEmit(&GoEmit{})
}
