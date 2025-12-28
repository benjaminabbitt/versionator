// versionator-plugin-patch-csproj patches version in .csproj files.
package main

import (
	"regexp"
	"strings"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// CsprojPatch implements the PatchPlugin interface for .NET project files.
type CsprojPatch struct{}

func (p *CsprojPatch) Name() string {
	return "patch-csproj"
}

func (p *CsprojPatch) FilePattern() string {
	return "*.csproj"
}

func (p *CsprojPatch) Description() string {
	return ".NET project file (*.csproj)"
}

func (p *CsprojPatch) Patch(content, version string) (string, error) {
	// Match <Version>...</Version> or <version>...</version>
	reVersion := regexp.MustCompile(`<[Vv]ersion>[^<]*</[Vv]ersion>`)
	if !reVersion.MatchString(content) {
		return content, nil
	}

	// Replace only the first match for project version
	replaced := false
	return reVersion.ReplaceAllStringFunc(content, func(match string) string {
		if !replaced {
			replaced = true
			if strings.HasPrefix(match, "<V") {
				return "<Version>" + version + "</Version>"
			}
			return "<version>" + version + "</version>"
		}
		return match
	}), nil
}

func main() {
	sdk.ServePatch(&CsprojPatch{})
}
