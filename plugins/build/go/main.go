// versionator-plugin-build-go generates Go linker flags for version injection.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed flags.tmpl
var flagsTemplate string

// GoBuild implements the BuildPlugin interface for Go.
type GoBuild struct{}

func (p *GoBuild) Name() string {
	return "build-go"
}

func (p *GoBuild) Format() string {
	return "go"
}

func (p *GoBuild) GenerateFlags(vars map[string]string) (string, error) {
	return mustache.Render(flagsTemplate, vars)
}

func main() {
	sdk.ServeBuild(&GoBuild{})
}
