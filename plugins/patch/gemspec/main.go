// versionator-plugin-patch-gemspec patches version in *.gemspec files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// GemspecPatch implements the PatchPlugin interface for Ruby gemspec files.
type GemspecPatch struct{}

func (p *GemspecPatch) Name() string {
	return "patch-gemspec"
}

func (p *GemspecPatch) FilePattern() string {
	return "*.gemspec"
}

func (p *GemspecPatch) Description() string {
	return "Ruby gem specification (*.gemspec)"
}

func (p *GemspecPatch) Patch(content, version string) (string, error) {
	// Match spec.version = "..." or .version = "..."
	reDouble := regexp.MustCompile(`(\.version\s*=\s*)"[^"]*"`)
	reSingle := regexp.MustCompile(`(\.version\s*=\s*)'[^']*'`)

	if reDouble.MatchString(content) {
		return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
	}
	if reSingle.MatchString(content) {
		return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
	}
	return content, nil
}

func main() {
	sdk.ServePatch(&GemspecPatch{})
}
