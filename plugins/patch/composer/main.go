// versionator-plugin-patch-composer patches version in composer.json files.
package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// ComposerPatch implements the PatchPlugin interface for PHP Composer.
type ComposerPatch struct{}

func (p *ComposerPatch) Name() string {
	return "patch-composer"
}

func (p *ComposerPatch) FilePattern() string {
	return "composer.json"
}

func (p *ComposerPatch) Description() string {
	return "Composer package manifest (PHP)"
}

func (p *ComposerPatch) Patch(content, version string) (string, error) {
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
	sdk.ServePatch(&ComposerPatch{})
}
