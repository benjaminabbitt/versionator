// versionator-plugin-build-csharp generates .NET MSBuild property flags for version injection.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed flags.tmpl
var flagsTemplate string

// CSharpBuild implements the BuildPlugin interface for C#.
type CSharpBuild struct{}

func (p *CSharpBuild) Name() string {
	return "build-csharp"
}

func (p *CSharpBuild) Format() string {
	return "csharp"
}

func (p *CSharpBuild) GenerateFlags(vars map[string]string) (string, error) {
	return mustache.Render(flagsTemplate, vars)
}

func main() {
	sdk.ServeBuild(&CSharpBuild{})
}
