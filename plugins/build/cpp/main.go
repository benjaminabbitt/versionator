// versionator-plugin-build-cpp generates C++ preprocessor flags for version injection.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed flags.tmpl
var flagsTemplate string

// CPPBuild implements the BuildPlugin interface for C++.
type CPPBuild struct{}

func (p *CPPBuild) Name() string {
	return "build-cpp"
}

func (p *CPPBuild) Format() string {
	return "cpp"
}

func (p *CPPBuild) GenerateFlags(vars map[string]string) (string, error) {
	return mustache.Render(flagsTemplate, vars)
}

func main() {
	sdk.ServeBuild(&CPPBuild{})
}
