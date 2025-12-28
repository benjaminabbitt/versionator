// versionator-plugin-patch-pyproject patches version in pyproject.toml files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// PyprojectPatch implements the PatchPlugin interface for pyproject.toml.
type PyprojectPatch struct{}

func (p *PyprojectPatch) Name() string {
	return "patch-pyproject"
}

func (p *PyprojectPatch) FilePattern() string {
	return "pyproject.toml"
}

func (p *PyprojectPatch) Description() string {
	return "Python project manifest (PEP 621)"
}

func (p *PyprojectPatch) Patch(content, version string) (string, error) {
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
	sdk.ServePatch(&PyprojectPatch{})
}
