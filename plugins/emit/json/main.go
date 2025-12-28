// versionator-plugin-emit-json generates JSON version files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed json.tmpl
var jsonTemplate string

// JSONEmit implements the EmitPlugin interface for JSON.
type JSONEmit struct{}

func (p *JSONEmit) Name() string {
	return "emit-json"
}

func (p *JSONEmit) Format() string {
	return "json"
}

func (p *JSONEmit) FileExtension() string {
	return ".json"
}

func (p *JSONEmit) DefaultOutput() string {
	return "version.json"
}

func (p *JSONEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(jsonTemplate, vars)
}

func main() {
	sdk.ServeEmit(&JSONEmit{})
}
