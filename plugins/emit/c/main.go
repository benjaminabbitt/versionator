// versionator-plugin-emit-c generates C version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed c.tmpl
var cTemplate string

// CEmit implements the EmitPlugin interface for C.
type CEmit struct{}

func (p *CEmit) Name() string {
	return "emit-c"
}

func (p *CEmit) Format() string {
	return "c"
}

func (p *CEmit) FileExtension() string {
	return ".c"
}

func (p *CEmit) DefaultOutput() string {
	return "version.c"
}

func (p *CEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(cTemplate, vars)
}

func main() {
	sdk.ServeEmit(&CEmit{})
}
