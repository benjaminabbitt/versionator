// versionator-plugin-emit-java generates Java version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed java.tmpl
var javaTemplate string

// JavaEmit implements the EmitPlugin interface for Java.
type JavaEmit struct{}

func (p *JavaEmit) Name() string {
	return "emit-java"
}

func (p *JavaEmit) Format() string {
	return "java"
}

func (p *JavaEmit) FileExtension() string {
	return ".java"
}

func (p *JavaEmit) DefaultOutput() string {
	return "src/main/java/version/Version.java"
}

func (p *JavaEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(javaTemplate, vars)
}

func main() {
	sdk.ServeEmit(&JavaEmit{})
}
