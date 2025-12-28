// versionator-plugin-emit-python generates Python version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed python.tmpl
var pythonTemplate string

// PythonEmit implements the EmitPlugin interface for Python.
type PythonEmit struct{}

func (p *PythonEmit) Name() string {
	return "emit-python"
}

func (p *PythonEmit) Format() string {
	return "python"
}

func (p *PythonEmit) FileExtension() string {
	return ".py"
}

func (p *PythonEmit) DefaultOutput() string {
	return "_version.py"
}

func (p *PythonEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(pythonTemplate, vars)
}

func main() {
	sdk.ServeEmit(&PythonEmit{})
}
