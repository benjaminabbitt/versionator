package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
)

// renderReleaseTemplate renders one of the release templates, falling back to
// the documented default when the project has not configured it. The kind is
// used only to name the offending template in errors.
func renderReleaseTemplate(kind, tmpl, fallback string, data emit.TemplateData) (string, error) {
	if tmpl == "" {
		tmpl = fallback
	}
	rendered, err := emit.RenderTemplateWithData(tmpl, data)
	if err != nil {
		return "", fmt.Errorf("error rendering release %s template: %w", kind, err)
	}
	return strings.TrimSpace(rendered), nil
}

// badRefCharacters are the characters git refuses anywhere in a ref name.
// Backslash and the ASCII space are included; control characters are checked
// separately by code point.
const badRefCharacters = " ~^:?*[\\"

// validateTagName rejects names git cannot store as a tag.
//
// Release mutates the repository before it tags — it may commit the VERSION
// file, apply file updates and create a branch — so an unusable tag name
// discovered at tag time leaves a half-finished release behind. The rules
// mirror git-check-ref-format for the single-level name a tag occupies.
func validateTagName(name string) error {
	invalid := func(reason string) error {
		return fmt.Errorf("invalid tag name %q: %s (check the release tag template)", name, reason)
	}

	if name == "" {
		return invalid("renders empty")
	}
	if strings.ContainsAny(name, badRefCharacters) {
		return invalid("contains a character git does not allow in a ref name")
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return invalid("contains a control character")
		}
	}
	if strings.Contains(name, "..") || strings.Contains(name, "@{") || name == "@" {
		return invalid("contains a sequence git reserves")
	}
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") || strings.Contains(name, "//") {
		return invalid("has an empty path component")
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, ".lock") {
		return invalid("ends with '.' or '.lock'")
	}
	// A leading dash makes the tag indistinguishable from a flag to git and to
	// every script that passes it along.
	if strings.HasPrefix(name, "-") {
		return invalid("starts with '-'")
	}
	return nil
}
