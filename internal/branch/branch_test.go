package branch

import (
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY - Happy path tests for primary use cases
// =============================================================================

// TestIsMainBranch_ExactMatch_ReturnsTrue validates that branches matching
// exactly with configured patterns are correctly identified as main branches.
//
// Why: This is the primary use case for branch detection - users configure
// exact branch names like "main" or "master" and expect exact matches.
//
// What: Given a list of exact branch patterns, when checking branches that
// match exactly, then IsMainBranch returns true; non-matching branches
// return false.
func TestIsMainBranch_ExactMatch_ReturnsTrue(t *testing.T) {
	// Precondition: Standard main branch patterns configured
	patterns := []string{"main", "master"}

	tests := []struct {
		branch   string
		expected bool
	}{
		{"main", true},
		{"master", true},
		{"develop", false},
		{"feature/foo", false},
	}

	for _, tt := range tests {
		// Action: Check if branch is considered a main branch
		result := IsMainBranch(tt.branch, patterns)

		// Expected: Result matches expected main branch status
		if result != tt.expected {
			t.Errorf("IsMainBranch(%q, %v) = %v, want %v", tt.branch, patterns, result, tt.expected)
		}
	}
}

// TestSanitizeBranchName_SlashToHyphen_ReplacesCorrectly validates the primary
// sanitization behavior of converting slashes to hyphens.
//
// Why: Branch names with slashes (like "feature/add-login") must be converted
// to safe identifiers for use in version strings and filenames.
//
// What: Given a branch name with a slash, when sanitizing, then the slash
// is replaced with a hyphen.
func TestSanitizeBranchName_SlashToHyphen_ReplacesCorrectly(t *testing.T) {
	// Precondition: Branch name with standard feature/ prefix
	input := "feature/add-login"

	// Action: Sanitize the branch name
	result := SanitizeBranchName(input)

	// Expected: Slash converted to hyphen
	expected := "feature-add-login"
	if result != expected {
		t.Errorf("SanitizeBranchName(%q) = %q, want %q", input, result, expected)
	}
}

// TestDefaultMainBranches_ContainsStandardBranches validates that the default
// main branch list includes industry-standard branch names.
//
// Why: Users expect sensible defaults without explicit configuration. Both
// "main" and "master" are widely used as primary branch names.
//
// What: When retrieving default main branches, then both "main" and "master"
// are included in the returned list.
func TestDefaultMainBranches_ContainsStandardBranches(t *testing.T) {
	// Action: Get default main branches
	defaults := DefaultMainBranches()

	// Expected: Contains both standard primary branch names
	hasMain := false
	hasMaster := false
	for _, b := range defaults {
		if b == "main" {
			hasMain = true
		}
		if b == "master" {
			hasMaster = true
		}
	}

	if !hasMain {
		t.Error("DefaultMainBranches should include 'main'")
	}
	if !hasMaster {
		t.Error("DefaultMainBranches should include 'master'")
	}
}

// =============================================================================
// KEY VARIATIONS - Important alternate flows and pattern matching
// =============================================================================

// TestIsMainBranch_GlobPattern_MatchesWildcards validates that glob patterns
// with wildcards correctly match branch name prefixes.
//
// Why: Teams often use branch naming conventions like "release/*" or "hotfix/*"
// that should be treated as main branches for version calculation.
//
// What: Given patterns with wildcards, when checking branches matching those
// patterns, then IsMainBranch returns true; non-matching branches return false.
func TestIsMainBranch_GlobPattern_MatchesWildcards(t *testing.T) {
	// Precondition: Patterns include wildcard expressions
	patterns := []string{"main", "release/*", "hotfix/*"}

	tests := []struct {
		branch   string
		expected bool
	}{
		{"main", true},
		{"release/v1.0", true},
		{"release/v2.0.0", true},
		{"hotfix/urgent", true},
		{"feature/foo", false},
		{"release", false}, // Must have something after /
	}

	for _, tt := range tests {
		// Action: Check if branch matches any pattern
		result := IsMainBranch(tt.branch, patterns)

		// Expected: Result matches expected pattern match status
		if result != tt.expected {
			t.Errorf("IsMainBranch(%q, %v) = %v, want %v", tt.branch, patterns, result, tt.expected)
		}
	}
}

// TestMatchPattern_GlobPattern_MatchesCorrectly validates that the internal
// pattern matching function correctly handles glob patterns.
//
// Why: The matchPattern function is the foundation for IsMainBranch glob
// support; it must correctly evaluate wildcard expressions.
//
// What: Given a valid glob pattern, when matching branches, then branches
// matching the pattern return true; non-matching branches return false.
func TestMatchPattern_GlobPattern_MatchesCorrectly(t *testing.T) {
	// Action & Expected: Valid glob pattern matches appropriately
	if !matchPattern("release/v1.0", "release/*") {
		t.Error("Should match release/* pattern")
	}

	// Action & Expected: Non-matching branch does not match
	if matchPattern("feature/login", "release/*") {
		t.Error("Should not match release/* pattern")
	}
}

// TestSanitizeBranchName_MultipleSlashes_ReplacesAll validates that all slashes
// in a branch name are converted to hyphens.
//
// Why: Some workflows use nested branch structures like "feature/user/profile"
// that contain multiple slashes; all must be sanitized.
//
// What: Given a branch name with multiple slashes, when sanitizing, then all
// slashes are replaced with hyphens.
func TestSanitizeBranchName_MultipleSlashes_ReplacesAll(t *testing.T) {
	// Precondition: Branch name with multiple path segments
	input := "feature/user/profile"

	// Action: Sanitize the branch name
	result := SanitizeBranchName(input)

	// Expected: All slashes converted to hyphens
	expected := "feature-user-profile"
	if result != expected {
		t.Errorf("SanitizeBranchName(%q) = %q, want %q", input, result, expected)
	}
}

// TestSanitizeBranchName_InvalidChars_ReplacesWithHyphens validates that
// various invalid characters are properly sanitized.
//
// Why: Branch names may contain characters invalid for version strings or
// filenames (@, #, _, .); all must be normalized to hyphens.
//
// What: Given branch names with various special characters, when sanitizing,
// then all special characters are replaced with hyphens.
func TestSanitizeBranchName_InvalidChars_ReplacesWithHyphens(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"feature/add@login", "feature-add-login"},
		{"feature/add#123", "feature-add-123"},
		{"feature/add_login", "feature-add-login"},
		{"feature/add.login", "feature-add-login"},
	}

	for _, tt := range tests {
		// Action: Sanitize the branch name
		result := SanitizeBranchName(tt.input)

		// Expected: Special characters replaced with hyphens
		if result != tt.expected {
			t.Errorf("SanitizeBranchName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestSanitizeBranchName_PreservesCase_MaintainsOriginalCasing validates that
// letter casing is preserved during sanitization.
//
// Why: Some teams use mixed-case branch names for readability; sanitization
// should not alter casing.
//
// What: Given a branch name with mixed case, when sanitizing, then the
// original casing is preserved.
func TestSanitizeBranchName_PreservesCase_MaintainsOriginalCasing(t *testing.T) {
	// Precondition: Branch name with mixed case
	input := "Feature/AddLogin"

	// Action: Sanitize the branch name
	result := SanitizeBranchName(input)

	// Expected: Case preserved, only slash converted
	expected := "Feature-AddLogin"
	if result != expected {
		t.Errorf("SanitizeBranchName(%q) = %q, want %q", input, result, expected)
	}
}

// TestSanitizeBranchName_AlphanumericOnly_PassesThrough validates that branch
// names with only valid characters pass through unchanged.
//
// Why: Branch names already containing only valid characters should not be
// modified by sanitization.
//
// What: Given a branch name with only alphanumeric characters, when
// sanitizing, then the name is returned unchanged.
func TestSanitizeBranchName_AlphanumericOnly_PassesThrough(t *testing.T) {
	// Precondition: Branch name with only valid characters
	input := "feature123"

	// Action: Sanitize the branch name
	result := SanitizeBranchName(input)

	// Expected: Name unchanged
	expected := "feature123"
	if result != expected {
		t.Errorf("SanitizeBranchName(%q) = %q, want %q", input, result, expected)
	}
}

// =============================================================================
// ERROR HANDLING - Expected failure modes and fallback behavior
// =============================================================================

// TestMatchPattern_InvalidPattern_FallsBackToExactMatch validates that invalid
// glob patterns gracefully fall back to exact string matching.
//
// Why: Users may accidentally configure invalid patterns; the system should
// handle this gracefully rather than crash or behave unpredictably.
//
// What: Given an invalid glob pattern (unclosed bracket), when matching,
// then the function falls back to exact string comparison.
func TestMatchPattern_InvalidPattern_FallsBackToExactMatch(t *testing.T) {
	// Precondition: An unclosed bracket is an invalid pattern for filepath.Match
	invalidPattern := "release[/*"

	// Action & Expected: Invalid pattern falls back to exact match (returns false for non-exact)
	if matchPattern("release/v1", invalidPattern) {
		t.Error("Invalid pattern should fall back to exact match (not match)")
	}

	// Action & Expected: Exact match should work even with invalid pattern syntax
	if !matchPattern(invalidPattern, invalidPattern) {
		t.Error("Invalid pattern with exact match should return true")
	}
}

// =============================================================================
// EDGE CASES - Boundary conditions and unusual inputs
// =============================================================================

// TestIsMainBranch_EmptyBranch_ReturnsFalse validates that an empty branch
// name is never considered a main branch.
//
// Why: Empty branch names indicate an error condition or uninitialized state;
// they should never match any pattern.
//
// What: Given an empty branch name, when checking against valid patterns,
// then IsMainBranch returns false.
func TestIsMainBranch_EmptyBranch_ReturnsFalse(t *testing.T) {
	// Precondition: Valid patterns configured
	patterns := []string{"main", "master"}

	// Action: Check empty branch name
	result := IsMainBranch("", patterns)

	// Expected: Empty branch is not a main branch
	if result {
		t.Error("Empty branch should not be considered a main branch")
	}
}

// TestIsMainBranch_EmptyPatterns_ReturnsFalse validates that when no patterns
// are configured, no branch matches.
//
// Why: An empty pattern list represents a misconfiguration or edge case;
// nothing should match when there are no patterns to match against.
//
// What: Given an empty pattern list, when checking any branch, then
// IsMainBranch returns false.
func TestIsMainBranch_EmptyPatterns_ReturnsFalse(t *testing.T) {
	// Precondition: No patterns configured
	patterns := []string{}

	// Action: Check a valid branch name
	result := IsMainBranch("main", patterns)

	// Expected: No patterns means no matches
	if result {
		t.Error("No patterns should match nothing")
	}
}

// TestSanitizeBranchName_Empty_ReturnsEmpty validates that sanitizing an empty
// string returns an empty string.
//
// Why: Empty input should produce empty output without errors; this handles
// edge cases where branch name is not available.
//
// What: Given an empty string, when sanitizing, then an empty string is
// returned.
func TestSanitizeBranchName_Empty_ReturnsEmpty(t *testing.T) {
	// Action: Sanitize empty string
	result := SanitizeBranchName("")

	// Expected: Returns empty string
	if result != "" {
		t.Errorf("SanitizeBranchName('') = %q, want empty string", result)
	}
}

// TestSanitizeBranchName_CleanupMultipleHyphens_CollapsesToSingle validates
// that consecutive hyphens are collapsed to a single hyphen.
//
// Why: After replacing multiple adjacent special characters, the result may
// contain consecutive hyphens which look unprofessional in version strings.
//
// What: Given a branch name that would result in multiple consecutive hyphens,
// when sanitizing, then consecutive hyphens are collapsed to one.
func TestSanitizeBranchName_CleanupMultipleHyphens_CollapsesToSingle(t *testing.T) {
	// Precondition: Branch name with consecutive hyphens
	input := "feature/add--login"

	// Action: Sanitize the branch name
	result := SanitizeBranchName(input)

	// Expected: Multiple hyphens collapsed to single hyphen
	expected := "feature-add-login"
	if result != expected {
		t.Errorf("SanitizeBranchName(%q) = %q, want %q", input, result, expected)
	}
}

// TestSanitizeBranchName_TrimHyphens_RemovesLeadingAndTrailing validates that
// leading and trailing hyphens are removed from the result.
//
// Why: Leading/trailing slashes in branch names would result in leading/
// trailing hyphens which are invalid in version identifiers.
//
// What: Given a branch name with leading and trailing slashes, when
// sanitizing, then the resulting hyphens are trimmed.
func TestSanitizeBranchName_TrimHyphens_RemovesLeadingAndTrailing(t *testing.T) {
	// Precondition: Branch name with leading and trailing slashes
	input := "/feature/login/"

	// Action: Sanitize the branch name
	result := SanitizeBranchName(input)

	// Expected: Leading and trailing hyphens trimmed
	expected := "feature-login"
	if result != expected {
		t.Errorf("SanitizeBranchName(%q) = %q, want %q", input, result, expected)
	}
}
