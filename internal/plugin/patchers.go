package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Common patcher functions that plugins can use.
// Each returns a PatchFunc that can be assigned to PatchConfig.Patch

// PatchJSON returns a patcher for JSON files with a top-level "version" field.
func PatchJSON() PatchFunc {
	return func(content, version string) (string, error) {
		// Validate JSON
		var js interface{}
		if err := json.Unmarshal([]byte(content), &js); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}

		// Match "version": "..." at top level
		re := regexp.MustCompile(`"version"\s*:\s*"[^"]*"`)
		if !re.MatchString(content) {
			return "", nil
		}
		patched := re.ReplaceAllString(content, `"version": "`+version+`"`)

		// Validate patched JSON
		if err := json.Unmarshal([]byte(patched), &js); err != nil {
			return "", fmt.Errorf("patched JSON is invalid: %w", err)
		}
		return patched, nil
	}
}

// PatchTOML returns a patcher for TOML files with a version field.
// Works for pyproject.toml, Cargo.toml, and similar files.
func PatchTOML() PatchFunc {
	return func(content, version string) (string, error) {
		// Match version = "..." or version = '...' at start of line
		reDouble := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)"[^"]*"`)
		reSingle := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)'[^']*'`)

		if reDouble.MatchString(content) {
			return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reSingle.MatchString(content) {
			return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return "", nil
	}
}

// PatchXML returns a patcher for XML files with a <version> or <Version> element.
// Works for pom.xml, *.csproj, and similar files.
func PatchXML() PatchFunc {
	return func(content, version string) (string, error) {
		// Match <version>...</version> or <Version>...</Version>
		reVersion := regexp.MustCompile(`<[Vv]ersion>[^<]*</[Vv]ersion>`)
		if !reVersion.MatchString(content) {
			return "", nil
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
}

// PatchYAML returns a patcher for YAML files with a top-level version field.
// Works for pubspec.yaml and similar files.
func PatchYAML() PatchFunc {
	return func(content, version string) (string, error) {
		re := regexp.MustCompile(`(?m)^(version:\s*)["']?[^"'\n]*["']?`)
		if !re.MatchString(content) {
			return "", nil
		}
		return re.ReplaceAllString(content, `${1}`+version), nil
	}
}

// PatchGradle returns a patcher for Gradle files (build.gradle, build.gradle.kts).
// Matches version = "..." or version = '...'
func PatchGradle() PatchFunc {
	return func(content, version string) (string, error) {
		// Match version = "..." or version = '...'
		reDouble := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)"[^"]*"`)
		reSingle := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)'[^']*'`)

		if reDouble.MatchString(content) {
			return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reSingle.MatchString(content) {
			return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return "", nil
	}
}

// PatchPythonSetup returns a patcher for setup.py files.
// Matches version="..." or version='...' in setup() call.
func PatchPythonSetup() PatchFunc {
	return func(content, version string) (string, error) {
		reDouble := regexp.MustCompile(`(version\s*=\s*)"[^"]*"`)
		reSingle := regexp.MustCompile(`(version\s*=\s*)'[^']*'`)

		if reDouble.MatchString(content) {
			return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reSingle.MatchString(content) {
			return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return "", nil
	}
}

// PatchRubyGemspec returns a patcher for *.gemspec files.
// Matches spec.version = "..." or .version = "..."
func PatchRubyGemspec() PatchFunc {
	return func(content, version string) (string, error) {
		reDouble := regexp.MustCompile(`(\.version\s*=\s*)"[^"]*"`)
		reSingle := regexp.MustCompile(`(\.version\s*=\s*)'[^']*'`)

		if reDouble.MatchString(content) {
			return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
		}
		if reSingle.MatchString(content) {
			return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
		}
		return "", nil
	}
}

// PatchSwiftPackage returns a patcher for Package.swift files.
// Matches // VERSION: x.y.z comments.
func PatchSwiftPackage() PatchFunc {
	return func(content, version string) (string, error) {
		re := regexp.MustCompile(`(//\s*VERSION:\s*)[^\n]*`)
		if !re.MatchString(content) {
			return "", nil
		}
		return re.ReplaceAllString(content, `${1}`+version), nil
	}
}
