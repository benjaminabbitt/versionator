// versionator-plugin-emit-csharp generates C# version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed csharp.tmpl
var csharpTemplate string

// CSharpEmit implements the EmitPlugin interface for C#.
type CSharpEmit struct{}

func (p *CSharpEmit) Name() string {
	return "emit-csharp"
}

func (p *CSharpEmit) Format() string {
	return "csharp"
}

func (p *CSharpEmit) FileExtension() string {
	return ".cs"
}

func (p *CSharpEmit) DefaultOutput() string {
	return "Version/VersionInfo.cs"
}

func (p *CSharpEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(csharpTemplate, vars)
}

func main() {
	sdk.ServeEmit(&CSharpEmit{})
}
