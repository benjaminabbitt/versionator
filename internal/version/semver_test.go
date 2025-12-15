package version

import "testing"

func TestParse_BasicVersions(t *testing.T) {
	tests := []struct {
		input    string
		major    int
		minor    int
		patch    int
		preRel   string
		metadata string
	}{
		{"1.2.3", 1, 2, 3, "", ""},
		{"v1.2.3", 1, 2, 3, "", ""},
		{"0.0.1", 0, 0, 1, "", ""},
		{"10.20.30", 10, 20, 30, "", ""},
		{"1.0.0", 1, 0, 0, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.Major != tt.major {
				t.Errorf("Major: got %d, want %d", sv.Major, tt.major)
			}
			if sv.Minor != tt.minor {
				t.Errorf("Minor: got %d, want %d", sv.Minor, tt.minor)
			}
			if sv.Patch != tt.patch {
				t.Errorf("Patch: got %d, want %d", sv.Patch, tt.patch)
			}
			if sv.PreRelease != tt.preRel {
				t.Errorf("PreRelease: got %q, want %q", sv.PreRelease, tt.preRel)
			}
			if sv.BuildMetadata != tt.metadata {
				t.Errorf("BuildMetadata: got %q, want %q", sv.BuildMetadata, tt.metadata)
			}
		})
	}
}

func TestParse_PreRelease(t *testing.T) {
	tests := []struct {
		input  string
		preRel string
	}{
		{"1.2.3-alpha", "alpha"},
		{"1.2.3-beta", "beta"},
		{"1.2.3-alpha.1", "alpha.1"},
		{"1.2.3-beta.2", "beta.2"},
		{"1.2.3-rc.1", "rc.1"},
		{"1.0.0-0.3.7", "0.3.7"},
		{"1.0.0-x.7.z.92", "x.7.z.92"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreRelease != tt.preRel {
				t.Errorf("PreRelease: got %q, want %q", sv.PreRelease, tt.preRel)
			}
		})
	}
}

func TestParse_BuildMetadata(t *testing.T) {
	tests := []struct {
		input    string
		metadata string
	}{
		{"1.2.3+build", "build"},
		{"1.2.3+build.123", "build.123"},
		{"1.2.3+20230101", "20230101"},
		{"1.2.3-alpha+build.123", "build.123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.BuildMetadata != tt.metadata {
				t.Errorf("BuildMetadata: got %q, want %q", sv.BuildMetadata, tt.metadata)
			}
		})
	}
}

func TestParse_PartialVersions(t *testing.T) {
	tests := []struct {
		input string
		major int
		minor int
		patch int
	}{
		{"1", 1, 0, 0},
		{"1.2", 1, 2, 0},
		{"v1", 1, 0, 0},
		{"v1.2", 1, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.Major != tt.major {
				t.Errorf("Major: got %d, want %d", sv.Major, tt.major)
			}
			if sv.Minor != tt.minor {
				t.Errorf("Minor: got %d, want %d", sv.Minor, tt.minor)
			}
			if sv.Patch != tt.patch {
				t.Errorf("Patch: got %d, want %d", sv.Patch, tt.patch)
			}
		})
	}
}

func TestSemVer_String(t *testing.T) {
	// String() returns full semver including prerelease and metadata
	sv := Parse("1.2.3-alpha+build")
	if sv.String() != "1.2.3-alpha+build" {
		t.Errorf("String(): got %q, want %q", sv.String(), "1.2.3-alpha+build")
	}
}

func TestSemVer_SemVerString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"1.2.3-alpha", "1.2.3-alpha"},
		{"1.2.3-alpha+build", "1.2.3-alpha"},
		{"1.2.3+build", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.SemVerString() != tt.expected {
				t.Errorf("SemVerString(): got %q, want %q", sv.SemVerString(), tt.expected)
			}
		})
	}
}

func TestSemVer_FullSemVer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"1.2.3-alpha", "1.2.3-alpha"},
		{"1.2.3+build", "1.2.3+build"},
		{"1.2.3-alpha+build", "1.2.3-alpha+build"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.FullSemVer() != tt.expected {
				t.Errorf("FullSemVer(): got %q, want %q", sv.FullSemVer(), tt.expected)
			}
		})
	}
}

func TestSemVer_PreReleaseWithDash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", ""},
		{"1.2.3-alpha", "-alpha"},
		{"1.2.3-alpha.1", "-alpha.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreReleaseWithDash() != tt.expected {
				t.Errorf("PreReleaseWithDash(): got %q, want %q", sv.PreReleaseWithDash(), tt.expected)
			}
		})
	}
}

func TestSemVer_BuildMetadataWithPlus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", ""},
		{"1.2.3+build", "+build"},
		{"1.2.3-alpha+build.123", "+build.123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.BuildMetadataWithPlus() != tt.expected {
				t.Errorf("BuildMetadataWithPlus(): got %q, want %q", sv.BuildMetadataWithPlus(), tt.expected)
			}
		})
	}
}

func TestSemVer_MajorMinor(t *testing.T) {
	sv := Parse("1.2.3")
	if sv.MajorMinor() != "1.2" {
		t.Errorf("MajorMinor(): got %q, want %q", sv.MajorMinor(), "1.2")
	}
}

func TestSemVer_IsPreRelease(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.2.3", false},
		{"1.2.3-alpha", true},
		{"1.2.3+build", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.IsPreRelease() != tt.expected {
				t.Errorf("IsPreRelease(): got %v, want %v", sv.IsPreRelease(), tt.expected)
			}
		})
	}
}

func TestSemVer_HasBuildMetadata(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.2.3", false},
		{"1.2.3-alpha", false},
		{"1.2.3+build", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.HasBuildMetadata() != tt.expected {
				t.Errorf("HasBuildMetadata(): got %v, want %v", sv.HasBuildMetadata(), tt.expected)
			}
		})
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.2.3", "1.2.3"},
		{"V1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{"vv1.2.3", "v1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := StripPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("StripPrefix(%q): got %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParse_Prefix(t *testing.T) {
	tests := []struct {
		input     string
		prefix    string
		hasPrefix bool
	}{
		{"1.2.3", "", false},
		{"v1.2.3", "v", true},
		{"V1.2.3", "V", true},
		{"1.2.3-alpha", "", false},
		{"v1.2.3-alpha", "v", true},
		{"1.2.3+build", "", false},
		{"v1.2.3+build", "v", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.Prefix != tt.prefix {
				t.Errorf("Prefix: got %q, want %q", sv.Prefix, tt.prefix)
			}
			if sv.HasPrefix() != tt.hasPrefix {
				t.Errorf("HasPrefix(): got %v, want %v", sv.HasPrefix(), tt.hasPrefix)
			}
		})
	}
}

func TestSemVer_PrefixedMethods(t *testing.T) {
	// Test with lowercase 'v' prefix
	sv := Parse("v1.2.3-alpha+build")

	// String() returns full semver without prefix
	if sv.String() != "1.2.3-alpha+build" {
		t.Errorf("String(): got %q, want %q", sv.String(), "1.2.3-alpha+build")
	}
	if sv.PrefixedString() != "v1.2.3" {
		t.Errorf("PrefixedString(): got %q, want %q", sv.PrefixedString(), "v1.2.3")
	}
	if sv.PrefixedSemVerString() != "v1.2.3-alpha" {
		t.Errorf("PrefixedSemVerString(): got %q, want %q", sv.PrefixedSemVerString(), "v1.2.3-alpha")
	}
	if sv.PrefixedFullSemVer() != "v1.2.3-alpha+build" {
		t.Errorf("PrefixedFullSemVer(): got %q, want %q", sv.PrefixedFullSemVer(), "v1.2.3-alpha+build")
	}

	// Test with uppercase 'V' prefix - should preserve case
	svUpper := Parse("V1.2.3-alpha+build")

	if svUpper.PrefixedString() != "V1.2.3" {
		t.Errorf("PrefixedString() with V: got %q, want %q", svUpper.PrefixedString(), "V1.2.3")
	}
	if svUpper.PrefixedSemVerString() != "V1.2.3-alpha" {
		t.Errorf("PrefixedSemVerString() with V: got %q, want %q", svUpper.PrefixedSemVerString(), "V1.2.3-alpha")
	}

	// Test without prefix - Prefixed* methods return same as unprefixed
	svNoPrefix := Parse("1.2.3-alpha+build")

	if svNoPrefix.PrefixedString() != "1.2.3" {
		t.Errorf("PrefixedString() no prefix: got %q, want %q", svNoPrefix.PrefixedString(), "1.2.3")
	}
	if svNoPrefix.PrefixedSemVerString() != "1.2.3-alpha" {
		t.Errorf("PrefixedSemVerString() no prefix: got %q, want %q", svNoPrefix.PrefixedSemVerString(), "1.2.3-alpha")
	}
}

func TestSemVer_OriginalMethods(t *testing.T) {
	// Without v prefix in input
	svNoPrefix := Parse("1.2.3-alpha")
	if svNoPrefix.OriginalString() != "1.2.3" {
		t.Errorf("OriginalString() without prefix: got %q, want %q", svNoPrefix.OriginalString(), "1.2.3")
	}
	if svNoPrefix.OriginalSemVerString() != "1.2.3-alpha" {
		t.Errorf("OriginalSemVerString() without prefix: got %q, want %q", svNoPrefix.OriginalSemVerString(), "1.2.3-alpha")
	}

	// With v prefix in input
	svWithPrefix := Parse("v1.2.3-alpha")
	if svWithPrefix.OriginalString() != "v1.2.3" {
		t.Errorf("OriginalString() with prefix: got %q, want %q", svWithPrefix.OriginalString(), "v1.2.3")
	}
	if svWithPrefix.OriginalSemVerString() != "v1.2.3-alpha" {
		t.Errorf("OriginalSemVerString() with prefix: got %q, want %q", svWithPrefix.OriginalSemVerString(), "v1.2.3-alpha")
	}
}

func TestSemVer_AssemblyVersion(t *testing.T) {
	sv := Parse("1.2.3")
	if sv.AssemblyVersion() != "1.2.3.0" {
		t.Errorf("AssemblyVersion(): got %q, want %q", sv.AssemblyVersion(), "1.2.3.0")
	}
}

// TestParse_SemVer2SpecExamples tests all examples from the SemVer 2.0.0 specification
// See: resources/semver-2.md
func TestParse_SemVer2SpecExamples(t *testing.T) {
	// Pre-release examples from spec section 9
	preReleaseTests := []struct {
		input    string
		preRel   string
		metadata string
	}{
		{"1.0.0-alpha", "alpha", ""},
		{"1.0.0-alpha.1", "alpha.1", ""},
		{"1.0.0-0.3.7", "0.3.7", ""},
		{"1.0.0-x.7.z.92", "x.7.z.92", ""},
		{"1.0.0-x-y-z.--", "x-y-z.--", ""}, // Hyphens in identifiers
	}

	for _, tt := range preReleaseTests {
		t.Run("prerelease_"+tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreRelease != tt.preRel {
				t.Errorf("PreRelease: got %q, want %q", sv.PreRelease, tt.preRel)
			}
			if sv.BuildMetadata != tt.metadata {
				t.Errorf("BuildMetadata: got %q, want %q", sv.BuildMetadata, tt.metadata)
			}
		})
	}

	// Build metadata examples from spec section 10
	buildMetadataTests := []struct {
		input    string
		preRel   string
		metadata string
	}{
		{"1.0.0-alpha+001", "alpha", "001"},
		{"1.0.0+20130313144700", "", "20130313144700"},
		{"1.0.0-beta+exp.sha.5114f85", "beta", "exp.sha.5114f85"},
		{"1.0.0+21AF26D3----117B344092BD", "", "21AF26D3----117B344092BD"}, // Multiple consecutive hyphens
	}

	for _, tt := range buildMetadataTests {
		t.Run("metadata_"+tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreRelease != tt.preRel {
				t.Errorf("PreRelease: got %q, want %q", sv.PreRelease, tt.preRel)
			}
			if sv.BuildMetadata != tt.metadata {
				t.Errorf("BuildMetadata: got %q, want %q", sv.BuildMetadata, tt.metadata)
			}
		})
	}
}

// TestParse_SemVer2SpecPrecedenceExamples tests the precedence examples from spec section 11
func TestParse_SemVer2SpecPrecedenceExamples(t *testing.T) {
	// These should all parse correctly (precedence comparison not implemented, just parsing)
	examples := []string{
		"1.0.0",
		"2.0.0",
		"2.1.0",
		"2.1.1",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
	}

	for _, input := range examples {
		t.Run(input, func(t *testing.T) {
			sv := Parse(input)
			if sv.Major == 0 && sv.Minor == 0 && sv.Patch == 0 && input != "0.0.0" {
				t.Errorf("Failed to parse %q", input)
			}
		})
	}
}

func TestSemVer_PreReleaseLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", ""},                        // No pre-release
		{"1.2.3-alpha", "alpha"},             // Simple label
		{"1.2.3-alpha.5", "alpha"},           // Label with number
		{"1.2.3-beta.2", "beta"},             // Another label with number
		{"1.2.3-rc.1", "rc"},                 // RC label
		{"1.2.3-0.3.7", "0.3.7"},             // All numeric (returns whole thing)
		{"1.2.3-alpha.beta", "alpha"},        // Multiple non-numeric parts
		{"1.2.3-feature-foo", "feature-foo"}, // Hyphenated label
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreReleaseLabel() != tt.expected {
				t.Errorf("PreReleaseLabel(): got %q, want %q", sv.PreReleaseLabel(), tt.expected)
			}
		})
	}
}

func TestSemVer_PreReleaseNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1.2.3", -1},          // No pre-release
		{"1.2.3-alpha", -1},    // No number
		{"1.2.3-alpha.5", 5},   // Number at end
		{"1.2.3-beta.2", 2},    // Number at end
		{"1.2.3-rc.1", 1},      // RC with number
		{"1.2.3-alpha.1.2", 2}, // Multiple numbers, returns last
		{"1.2.3-0.3.7", 7},     // All numeric, returns last
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreReleaseNumber() != tt.expected {
				t.Errorf("PreReleaseNumber(): got %d, want %d", sv.PreReleaseNumber(), tt.expected)
			}
		})
	}
}

func TestSemVer_PreReleaseLabelWithDash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", ""},               // No pre-release
		{"1.2.3-alpha", "-alpha"},   // Simple label
		{"1.2.3-alpha.5", "-alpha"}, // Label with number (dash only for label)
		{"1.2.3-beta.2", "-beta"},   // Another label
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv := Parse(tt.input)
			if sv.PreReleaseLabelWithDash() != tt.expected {
				t.Errorf("PreReleaseLabelWithDash(): got %q, want %q", sv.PreReleaseLabelWithDash(), tt.expected)
			}
		})
	}
}

func TestEscapedBranchName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"main", "main"},
		{"feature/foo", "feature-foo"},
		{"feature/foo/bar", "feature-foo-bar"},
		{"bugfix/JIRA-123", "bugfix-JIRA-123"},
		{"release/v1.0.0", "release-v1.0.0"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := EscapedBranchName(tt.input)
			if result != tt.expected {
				t.Errorf("EscapedBranchName(%q): got %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
