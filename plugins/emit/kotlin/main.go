// versionator-plugin-emit-kotlin generates Kotlin version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed kotlin.tmpl
var kotlinTemplate string

// KotlinEmit implements the EmitPlugin interface for Kotlin.
type KotlinEmit struct{}

func (p *KotlinEmit) Name() string {
	return "emit-kotlin"
}

func (p *KotlinEmit) Format() string {
	return "kotlin"
}

func (p *KotlinEmit) FileExtension() string {
	return ".kt"
}

func (p *KotlinEmit) DefaultOutput() string {
	return "src/main/kotlin/version/Version.kt"
}

func (p *KotlinEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(kotlinTemplate, vars)
}

func main() {
	sdk.ServeEmit(&KotlinEmit{})
}
