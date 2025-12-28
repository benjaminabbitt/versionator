// versionator-plugin-patch-cargo patches version in Cargo.toml files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// CargoPatch implements the PatchPlugin interface for Cargo.toml.
type CargoPatch struct{}

func (p *CargoPatch) Name() string {
	return "patch-cargo"
}

func (p *CargoPatch) FilePattern() string {
	return "Cargo.toml"
}

func (p *CargoPatch) Description() string {
	return "Cargo package manifest (Rust)"
}

func (p *CargoPatch) Patch(content, version string) (string, error) {
	// Match version = "..." or version = '...' at start of line
	reDouble := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)"[^"]*"`)
	reSingle := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)'[^']*'`)

	if reDouble.MatchString(content) {
		return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
	}
	if reSingle.MatchString(content) {
		return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
	}
	return content, nil
}

func main() {
	sdk.ServePatch(&CargoPatch{})
}
