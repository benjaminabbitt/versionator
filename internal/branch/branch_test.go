package branch

import (
	"testing"
)

func TestIsMainBranch_ExactMatch(t *testing.T) {
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
		result := IsMainBranch(tt.branch, patterns)
		if result != tt.expected {
			t.Errorf("IsMainBranch(%q, %v) = %v, want %v", tt.branch, patterns, result, tt.expected)
		}
	}
}

func TestIsMainBranch_GlobPattern(t *testing.T) {
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
		result := IsMainBranch(tt.branch, patterns)
		if result != tt.expected {
			t.Errorf("IsMainBranch(%q, %v) = %v, want %v", tt.branch, patterns, result, tt.expected)
		}
	}
}

func TestIsMainBranch_EmptyBranch(t *testing.T) {
	patterns := []string{"main", "master"}
	result := IsMainBranch("", patterns)
	if result {
		t.Error("Empty branch should not be considered a main branch")
	}
}

func TestIsMainBranch_EmptyPatterns(t *testing.T) {
	result := IsMainBranch("main", []string{})
	if result {
		t.Error("No patterns should match nothing")
	}
}

func TestSanitizeBranchName_SlashToHyphen(t *testing.T) {
	result := SanitizeBranchName("feature/add-login")
	expected := "feature-add-login"
	if result != expected {
		t.Errorf("SanitizeBranchName('feature/add-login') = %q, want %q", result, expected)
	}
}

func TestSanitizeBranchName_MultipleSlashes(t *testing.T) {
	result := SanitizeBranchName("feature/user/profile")
	expected := "feature-user-profile"
	if result != expected {
		t.Errorf("SanitizeBranchName('feature/user/profile') = %q, want %q", result, expected)
	}
}

func TestSanitizeBranchName_InvalidChars(t *testing.T) {
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
		result := SanitizeBranchName(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeBranchName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeBranchName_CleanupMultipleHyphens(t *testing.T) {
	result := SanitizeBranchName("feature/add--login")
	expected := "feature-add-login"
	if result != expected {
		t.Errorf("SanitizeBranchName('feature/add--login') = %q, want %q", result, expected)
	}
}

func TestSanitizeBranchName_TrimHyphens(t *testing.T) {
	result := SanitizeBranchName("/feature/login/")
	expected := "feature-login"
	if result != expected {
		t.Errorf("SanitizeBranchName('/feature/login/') = %q, want %q", result, expected)
	}
}

func TestSanitizeBranchName_Empty(t *testing.T) {
	result := SanitizeBranchName("")
	if result != "" {
		t.Errorf("SanitizeBranchName('') = %q, want empty string", result)
	}
}

func TestSanitizeBranchName_AlphanumericOnly(t *testing.T) {
	result := SanitizeBranchName("feature123")
	expected := "feature123"
	if result != expected {
		t.Errorf("SanitizeBranchName('feature123') = %q, want %q", result, expected)
	}
}

func TestSanitizeBranchName_PreservesCase(t *testing.T) {
	result := SanitizeBranchName("Feature/AddLogin")
	expected := "Feature-AddLogin"
	if result != expected {
		t.Errorf("SanitizeBranchName('Feature/AddLogin') = %q, want %q", result, expected)
	}
}

func TestDefaultMainBranches(t *testing.T) {
	defaults := DefaultMainBranches()

	// Should contain main and master
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
