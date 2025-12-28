// versionator-plugin-emit-javascript generates JavaScript version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed js.tmpl
var jsTemplate string

// JavaScriptEmit implements the EmitPlugin interface for JavaScript.
type JavaScriptEmit struct{}

func (p *JavaScriptEmit) Name() string {
	return "emit-javascript"
}

func (p *JavaScriptEmit) Format() string {
	return "js"
}

func (p *JavaScriptEmit) FileExtension() string {
	return ".js"
}

func (p *JavaScriptEmit) DefaultOutput() string {
	return "version.js"
}

func (p *JavaScriptEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(jsTemplate, vars)
}

func main() {
	sdk.ServeEmit(&JavaScriptEmit{})
}
