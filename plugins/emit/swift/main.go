// versionator-plugin-emit-swift generates Swift version source files.
package main

import (
	_ "embed"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
	"github.com/cbroglie/mustache"
)

//go:embed swift.tmpl
var swiftTemplate string

// SwiftEmit implements the EmitPlugin interface for Swift.
type SwiftEmit struct{}

func (p *SwiftEmit) Name() string {
	return "emit-swift"
}

func (p *SwiftEmit) Format() string {
	return "swift"
}

func (p *SwiftEmit) FileExtension() string {
	return ".swift"
}

func (p *SwiftEmit) DefaultOutput() string {
	return "Sources/Version.swift"
}

func (p *SwiftEmit) Emit(vars map[string]string) (string, error) {
	return mustache.Render(swiftTemplate, vars)
}

func main() {
	sdk.ServeEmit(&SwiftEmit{})
}
