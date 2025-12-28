// versionator-plugin-patch-swift patches version in Package.swift files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// SwiftPatch implements the PatchPlugin interface for Swift Package Manager.
type SwiftPatch struct{}

func (p *SwiftPatch) Name() string {
	return "patch-swift"
}

func (p *SwiftPatch) FilePattern() string {
	return "Package.swift"
}

func (p *SwiftPatch) Description() string {
	return "Swift Package Manager manifest (comment-based)"
}

func (p *SwiftPatch) Patch(content, version string) (string, error) {
	// Match // VERSION: x.y.z comments
	re := regexp.MustCompile(`(//\s*VERSION:\s*)[^\n]*`)
	if !re.MatchString(content) {
		return content, nil
	}
	return re.ReplaceAllString(content, `${1}`+version), nil
}

func main() {
	sdk.ServePatch(&SwiftPatch{})
}
