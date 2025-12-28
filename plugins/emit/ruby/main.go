// versionator-plugin-emit-ruby generates Ruby version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed ruby.tmpl
var rubyTemplate string

// RubyEmit implements the EmitPlugin interface for Ruby.
type RubyEmit struct{}

func (p *RubyEmit) Name() string {
	return "emit-ruby"
}

func (p *RubyEmit) Format() string {
	return "ruby"
}

func (p *RubyEmit) FileExtension() string {
	return ".rb"
}

func (p *RubyEmit) DefaultOutput() string {
	return "lib/version.rb"
}

func (p *RubyEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(rubyTemplate, vars)
}

func main() {
	sdk.ServeEmit(&RubyEmit{})
}
