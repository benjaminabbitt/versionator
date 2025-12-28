// versionator-plugin-emit-yaml generates YAML version files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed yaml.tmpl
var yamlTemplate string

// YAMLEmit implements the EmitPlugin interface for YAML.
type YAMLEmit struct{}

func (p *YAMLEmit) Name() string {
	return "emit-yaml"
}

func (p *YAMLEmit) Format() string {
	return "yaml"
}

func (p *YAMLEmit) FileExtension() string {
	return ".yaml"
}

func (p *YAMLEmit) DefaultOutput() string {
	return "version.yaml"
}

func (p *YAMLEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(yamlTemplate, vars)
}

func main() {
	sdk.ServeEmit(&YAMLEmit{})
}
