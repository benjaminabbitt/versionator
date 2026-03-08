package branch

import (
	"path/filepath"
	"strings"
)

// IsMainBranch checks if the branch name matches any of the main branch patterns
// Supports exact matches and glob patterns (e.g., "release/*")
func IsMainBranch(branchName string, patterns []string) bool {
	if branchName == "" {
		return false // Detached HEAD is not a main branch
	}

	for _, pattern := range patterns {
		if matchPattern(branchName, pattern) {
			return true
		}
	}
	return false
}

// matchPattern matches a branch name against a glob-like pattern
func matchPattern(name, pattern string) bool {
	// Use filepath.Match for glob support
	// This handles patterns like "release/*", "hotfix/*"
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		// Invalid pattern - fall back to exact match
		return name == pattern
	}
	return matched
}

// SanitizeBranchName ensures branch name is valid for semver pre-release identifier
// According to SemVer 2.0.0, pre-release identifiers must be alphanumeric and hyphens only
func SanitizeBranchName(branchName string) string {
	if branchName == "" {
		return ""
	}

	// Replace / with - (common for feature/foo branches)
	sanitized := strings.ReplaceAll(branchName, "/", "-")

	// Replace any other non-semver characters
	var result strings.Builder
	for _, c := range sanitized {
		if isValidSemverChar(c) {
			result.WriteRune(c)
		} else {
			// Replace invalid chars with hyphen
			result.WriteRune('-')
		}
	}

	// Clean up multiple consecutive hyphens
	cleaned := result.String()
	for strings.Contains(cleaned, "--") {
		cleaned = strings.ReplaceAll(cleaned, "--", "-")
	}

	// Trim leading/trailing hyphens
	cleaned = strings.Trim(cleaned, "-")

	return cleaned
}

// isValidSemverChar returns true if the character is valid in a semver pre-release identifier
func isValidSemverChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-'
}

// DefaultMainBranches returns the default list of main branch patterns
func DefaultMainBranches() []string {
	return []string{
		"main",
		"master",
		"release/*",
	}
}
