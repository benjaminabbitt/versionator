// versionator-plugin-patch-maven patches version in pom.xml files.
package main

import (
	"regexp"
	"strings"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// MavenPatch implements the PatchPlugin interface for Maven pom.xml.
type MavenPatch struct{}

func (p *MavenPatch) Name() string {
	return "patch-maven"
}

func (p *MavenPatch) FilePattern() string {
	return "pom.xml"
}

func (p *MavenPatch) Description() string {
	return "Maven project manifest (pom.xml)"
}

func (p *MavenPatch) Patch(content, version string) (string, error) {
	// Match <version>...</version> or <Version>...</Version>
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
	sdk.ServePatch(&MavenPatch{})
}
