// versionator-plugin-patch-npm patches version in package.json files.
package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// NPMPatch implements the PatchPlugin interface for npm package.json.
type NPMPatch struct{}

func (p *NPMPatch) Name() string {
	return "patch-npm"
}

func (p *NPMPatch) FilePattern() string {
	return "package.json"
}

func (p *NPMPatch) Description() string {
	return "npm/Node.js package manifest (package.json)"
}

func (p *NPMPatch) Patch(content, version string) (string, error) {
	// Validate JSON
	var js interface{}
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Match "version": "..." at top level
	re := regexp.MustCompile(`"version"\s*:\s*"[^"]*"`)
	if !re.MatchString(content) {
		return content, nil
	}
	patched := re.ReplaceAllString(content, `"version": "`+version+`"`)

	// Validate patched JSON
	if err := json.Unmarshal([]byte(patched), &js); err != nil {
		return "", fmt.Errorf("patched JSON is invalid: %w", err)
	}
	return patched, nil
}

func main() {
	sdk.ServePatch(&NPMPatch{})
}
