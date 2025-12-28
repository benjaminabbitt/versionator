// versionator-plugin-emit-cpp generates C++ version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed cpp.tmpl
var cppTemplate string

// CPPEmit implements the EmitPlugin interface for C++.
type CPPEmit struct{}

func (p *CPPEmit) Name() string {
	return "emit-cpp"
}

func (p *CPPEmit) Format() string {
	return "cpp"
}

func (p *CPPEmit) FileExtension() string {
	return ".cpp"
}

func (p *CPPEmit) DefaultOutput() string {
	return "version.cpp"
}

func (p *CPPEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(cppTemplate, vars)
}

func main() {
	sdk.ServeEmit(&CPPEmit{})
}
