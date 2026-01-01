// versionator-plugin-patch-setuppy patches version in setup.py files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// SetuppyPatch implements the PatchPlugin interface for setup.py.
type SetuppyPatch struct{}

func (p *SetuppyPatch) Name() string {
	return "patch-setuppy"
}

func (p *SetuppyPatch) FilePattern() string {
	return "setup.py"
}

func (p *SetuppyPatch) Description() string {
	return "Python setuptools setup.py"
}

func (p *SetuppyPatch) Patch(content, version string) (string, error) {
	// Match version="..." or version='...' (with optional spaces around =)
	reDouble := regexp.MustCompile(`(\bversion\s*=\s*)"[^"]*"`)
	reSingle := regexp.MustCompile(`(\bversion\s*=\s*)'[^']*'`)

	if reDouble.MatchString(content) {
		return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
	}
	if reSingle.MatchString(content) {
		return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
	}
	return content, nil
}

func main() {
	sdk.ServePatch(&SetuppyPatch{})
}
