// versionator-plugin-build-c generates C preprocessor flags for version injection.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed flags.tmpl
var flagsTemplate string

// CBuild implements the BuildPlugin interface for C.
type CBuild struct{}

func (p *CBuild) Name() string {
	return "build-c"
}

func (p *CBuild) Format() string {
	return "c"
}

func (p *CBuild) GenerateFlags(vars map[string]string) (string, error) {
	return mustache.Render(flagsTemplate, vars)
}

func main() {
	sdk.ServeBuild(&CBuild{})
}
