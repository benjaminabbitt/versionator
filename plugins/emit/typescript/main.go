// versionator-plugin-emit-typescript generates TypeScript version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed ts.tmpl
var tsTemplate string

// TypeScriptEmit implements the EmitPlugin interface for TypeScript.
type TypeScriptEmit struct{}

func (p *TypeScriptEmit) Name() string {
	return "emit-typescript"
}

func (p *TypeScriptEmit) Format() string {
	return "ts"
}

func (p *TypeScriptEmit) FileExtension() string {
	return ".ts"
}

func (p *TypeScriptEmit) DefaultOutput() string {
	return "version.ts"
}

func (p *TypeScriptEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(tsTemplate, vars)
}

func main() {
	sdk.ServeEmit(&TypeScriptEmit{})
}
