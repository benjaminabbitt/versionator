package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Common patcher functions that plugins can use.
// Each returns a PatchFunc that can be assigned to PatchConfig.Patch

// Pre-compiled regex patterns for better performance
var (
	// JSON version pattern
	reJSONVersion = regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)

	// TOML version patterns (double and single quotes)
	reTOMLVersionDouble = regexp.MustCompile(`(?m)^(\s*version\s*=\s*)"[^"]*"`)
	reTOMLVersionSingle = regexp.MustCompile(`(?m)^(\s*version\s*=\s*)'[^']*'`)

	// XML version pattern
	reXMLVersion = regexp.MustCompile(`<[Vv]ersion>[^<]*</[Vv]ersion>`)

	// YAML version pattern
	reYAMLVersion = regexp.MustCompile(`(?m)^(version:\s*)["']?[^"'\n]*["']?`)

	// Gradle version patterns - reuse TOML patterns since Gradle uses
	// identical syntax: version = "x.y.z" or version = 'x.y.z'
	reGradleVersionDouble = reTOMLVersionDouble
	reGradleVersionSingle = reTOMLVersionSingle

	// Python setup.py version patterns
	rePythonVersionDouble = regexp.MustCompile(`(version\s*=\s*)"[^"]*"`)
	rePythonVersionSingle = regexp.MustCompile(`(version\s*=\s*)'[^']*'`)

	// Ruby gemspec version patterns
	reRubyVersionDouble = regexp.MustCompile(`(\.version\s*=\s*)"[^"]*"`)
	reRubyVersionSingle = regexp.MustCompile(`(\.version\s*=\s*)'[^']*'`)

	// Swift Package.swift version pattern
	reSwiftVersion = regexp.MustCompile(`(//\s*VERSION:\s*)[^\n]*`)
)

// PatchJSON returns a patcher for JSON files with a top-level "version" field.
func PatchJSON() PatchFunc {
	return func(content, version string) (string, error) {
		// Parse JSON to verify it's valid and check for top-level version
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(content), &obj); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}

		// Check if top-level version exists
		if _, ok := obj["version"]; !ok {
			return content, nil
		}

		// Replace only the first match (top-level version comes first in well-formed package.json)
		replaced := false
		patched := reJSONVersion.ReplaceAllStringFunc(content, func(match string) string {
			if replaced {
				return match
			}
			replaced = true
			// Preserve the "version": prefix with its spacing
			prefixEnd := strings.Index(match, ":") + 1
			for prefixEnd < len(match) && (match[prefixEnd] == ' ' || match[prefixEnd] == '\t') {
				prefixEnd++
			}
			return match[:prefixEnd] + `"` + version + `"`
		})

		// Validate patched JSON
		if err := json.Unmarshal([]byte(patched), &obj); err != nil {
			return "", fmt.Errorf("patched JSON is invalid: %w", err)
		}
		return patched, nil
	}
}

// PatchTOML returns a patcher for TOML files with a version field.
// Works for pyproject.toml, Cargo.toml, and similar files.
func PatchTOML() PatchFunc {
	return func(content, version string) (string, error) {
		if reTOMLVersionDouble.MatchString(content) {
			return reTOMLVersionDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reTOMLVersionSingle.MatchString(content) {
			return reTOMLVersionSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return content, nil
	}
}

// PatchXML returns a patcher for XML files with a <version> or <Version> element.
// Works for pom.xml, *.csproj, and similar files.
func PatchXML() PatchFunc {
	return func(content, version string) (string, error) {
		if !reXMLVersion.MatchString(content) {
			return content, nil
		}

		// Replace only the first match for project version
		replaced := false
		return reXMLVersion.ReplaceAllStringFunc(content, func(match string) string {
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
}

// PatchYAML returns a patcher for YAML files with a top-level version field.
// Works for pubspec.yaml and similar files.
func PatchYAML() PatchFunc {
	return func(content, version string) (string, error) {
		if !reYAMLVersion.MatchString(content) {
			return content, nil
		}
		return reYAMLVersion.ReplaceAllString(content, `${1}`+version), nil
	}
}

// PatchGradle returns a patcher for Gradle files (build.gradle, build.gradle.kts).
// Matches version = "..." or version = '...'
func PatchGradle() PatchFunc {
	return func(content, version string) (string, error) {
		if reGradleVersionDouble.MatchString(content) {
			return reGradleVersionDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reGradleVersionSingle.MatchString(content) {
			return reGradleVersionSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return content, nil
	}
}

// PatchPythonSetup returns a patcher for setup.py files.
// Matches version="..." or version='...' in setup() call.
func PatchPythonSetup() PatchFunc {
	return func(content, version string) (string, error) {
		if rePythonVersionDouble.MatchString(content) {
			return rePythonVersionDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if rePythonVersionSingle.MatchString(content) {
			return rePythonVersionSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return content, nil
	}
}

// PatchRubyGemspec returns a patcher for *.gemspec files.
// Matches spec.version = "..." or .version = "..."
func PatchRubyGemspec() PatchFunc {
	return func(content, version string) (string, error) {
		if reRubyVersionDouble.MatchString(content) {
			return reRubyVersionDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reRubyVersionSingle.MatchString(content) {
			return reRubyVersionSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return content, nil
	}
}

// PatchSwiftPackage returns a patcher for Package.swift files.
// Matches // VERSION: x.y.z comments.
func PatchSwiftPackage() PatchFunc {
	return func(content, version string) (string, error) {
		if !reSwiftVersion.MatchString(content) {
			return content, nil
		}
		return reSwiftVersion.ReplaceAllString(content, `${1}`+version), nil
	}
}
