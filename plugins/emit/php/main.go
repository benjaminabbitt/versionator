// versionator-plugin-emit-php generates PHP version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed php.tmpl
var phpTemplate string

// PHPEmit implements the EmitPlugin interface for PHP.
type PHPEmit struct{}

func (p *PHPEmit) Name() string {
	return "emit-php"
}

func (p *PHPEmit) Format() string {
	return "php"
}

func (p *PHPEmit) FileExtension() string {
	return ".php"
}

func (p *PHPEmit) DefaultOutput() string {
	return "src/Version.php"
}

func (p *PHPEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(phpTemplate, vars)
}

func main() {
	sdk.ServeEmit(&PHPEmit{})
}
