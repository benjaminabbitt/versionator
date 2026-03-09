package version

import "testing"

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of semantic version parsing and
// string representation. These validate the fundamental parsing contract.
// =============================================================================

// TestParse_BasicVersions_ParsesAllComponentsCorrectly validates that the parser
// correctly extracts all semantic version components from standard version strings.
//
// Why: Parsing is the foundational operation - if basic version strings cannot be
// parsed correctly, all downstream operations (bumping, comparing, formatting) fail.
//
// What: Given valid semver strings in various formats (with/without prefix),
// the parser should extract correct major, minor, patch, pre-release, and metadata values.
func TestParse_BasicVersions_ParsesAllComponentsCorrectly(t *testing.T) {
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
			// Precondition: Valid semver string in standard or prefixed format
			// Action: Parse the version string
			sv := Parse(tt.input)

			// Expected: All components extracted correctly
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

// TestSemVer_String_ReturnsFullSemVerFormat validates that String() returns
// the complete SemVer 2.0.0 compliant representation including all optional parts.
//
// Why: String() is the primary output method used when writing versions to files,
// displaying to users, or passing to external tools. Incorrect output breaks integrations.
//
// What: Given a parsed version with pre-release and metadata, String() should
// return the complete semver string in the format Major.Minor.Patch-PreRelease+Metadata.
func TestSemVer_String_ReturnsFullSemVerFormat(t *testing.T) {
	// Precondition: A version string with all optional components
	// Action: Parse and call String()
	sv := Parse("1.2.3-alpha+build")

	// Expected: Full semver string without prefix
	if sv.String() != "1.2.3-alpha+build" {
		t.Errorf("String(): got %q, want %q", sv.String(), "1.2.3-alpha+build")
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows - different version formats, prefixes,
// and output formatting options that are commonly used.
// =============================================================================

// TestParse_PreRelease_ExtractsPreReleaseIdentifier validates that pre-release
// identifiers are correctly extracted from version strings.
//
// Why: Pre-release versions are essential for release candidates, alpha/beta testing,
// and CI/CD workflows. Incorrect parsing breaks version ordering and release workflows.
//
// What: Given version strings with various pre-release formats (alpha, beta, rc,
// numeric identifiers), the parser should extract the complete pre-release string.
func TestParse_PreRelease_ExtractsPreReleaseIdentifier(t *testing.T) {
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
			// Precondition: Version string with pre-release identifier
			// Action: Parse and extract pre-release
			sv := Parse(tt.input)

			// Expected: Pre-release portion correctly extracted
			if sv.PreRelease != tt.preRel {
				t.Errorf("PreRelease: got %q, want %q", sv.PreRelease, tt.preRel)
			}
		})
	}
}

// TestParse_BuildMetadata_ExtractsMetadataIdentifier validates that build metadata
// is correctly extracted from version strings.
//
// Why: Build metadata carries important information like commit SHA, build number,
// or timestamps. This data is used in CI/CD for traceability.
//
// What: Given version strings with build metadata (with or without pre-release),
// the parser should extract the metadata string after the '+' delimiter.
func TestParse_BuildMetadata_ExtractsMetadataIdentifier(t *testing.T) {
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
			// Precondition: Version string with build metadata
			// Action: Parse and extract metadata
			sv := Parse(tt.input)

			// Expected: Build metadata correctly extracted
			if sv.BuildMetadata != tt.metadata {
				t.Errorf("BuildMetadata: got %q, want %q", sv.BuildMetadata, tt.metadata)
			}
		})
	}
}

// TestParse_Prefix_ExtractsAndPreservesPrefix validates that version prefixes
// (v or V) are correctly detected and preserved.
//
// Why: Many projects use 'v' prefixes in tags (v1.0.0). The parser must preserve
// the original prefix case to maintain consistency with existing project conventions.
//
// What: Given version strings with various prefix styles, the parser should
// correctly identify the presence and case of the prefix.
func TestParse_Prefix_ExtractsAndPreservesPrefix(t *testing.T) {
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
			// Precondition: Version string with or without prefix
			// Action: Parse and check prefix fields
			sv := Parse(tt.input)

			// Expected: Prefix correctly extracted and HasPrefix() returns correct boolean
			if sv.Prefix != tt.prefix {
				t.Errorf("Prefix: got %q, want %q", sv.Prefix, tt.prefix)
			}
			if sv.HasPrefix() != tt.hasPrefix {
				t.Errorf("HasPrefix(): got %v, want %v", sv.HasPrefix(), tt.hasPrefix)
			}
		})
	}
}

// TestSemVer_SemVerString_ExcludesMetadata validates that SemVerString() returns
// the version without build metadata, per SemVer 2.0.0 comparison semantics.
//
// Why: Build metadata should be ignored when determining version precedence.
// SemVerString() provides a comparable version format for ordering.
//
// What: Given versions with various combinations of pre-release and metadata,
// SemVerString() should include pre-release but exclude metadata.
func TestSemVer_SemVerString_ExcludesMetadata(t *testing.T) {
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
			// Precondition: Version with optional pre-release and/or metadata
			// Action: Call SemVerString()
			sv := Parse(tt.input)

			// Expected: Pre-release included, metadata excluded
			if sv.SemVerString() != tt.expected {
				t.Errorf("SemVerString(): got %q, want %q", sv.SemVerString(), tt.expected)
			}
		})
	}
}

// TestSemVer_FullSemVer_IncludesAllComponents validates that FullSemVer() returns
// the complete version string with all components.
//
// Why: FullSemVer() is used when the complete version including metadata is needed,
// such as for display purposes or when metadata carries important build info.
//
// What: Given versions with various combinations of components, FullSemVer()
// should return the complete Major.Minor.Patch[-PreRelease][+Metadata] string.
func TestSemVer_FullSemVer_IncludesAllComponents(t *testing.T) {
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
			// Precondition: Version with various optional components
			// Action: Call FullSemVer()
			sv := Parse(tt.input)

			// Expected: All components present in output
			if sv.FullSemVer() != tt.expected {
				t.Errorf("FullSemVer(): got %q, want %q", sv.FullSemVer(), tt.expected)
			}
		})
	}
}

// TestSemVer_MajorMinor_ReturnsMajorMinorOnly validates that MajorMinor() returns
// just the Major.Minor portion of the version.
//
// Why: Major.Minor is commonly used for compatibility ranges, documentation,
// and release branch naming (e.g., "1.2.x supports feature Y").
//
// What: Given a version, MajorMinor() should return only "Major.Minor".
func TestSemVer_MajorMinor_ReturnsMajorMinorOnly(t *testing.T) {
	// Precondition: A parsed version
	// Action: Call MajorMinor()
	sv := Parse("1.2.3")

	// Expected: Only Major.Minor returned
	if sv.MajorMinor() != "1.2" {
		t.Errorf("MajorMinor(): got %q, want %q", sv.MajorMinor(), "1.2")
	}
}

// TestSemVer_PrefixedMethods_PreservesPrefixCase validates that Prefixed* methods
// correctly include the original prefix in their output.
//
// Why: When outputting versions for git tags or user display, the original prefix
// style (v vs V) should be preserved for consistency.
//
// What: Given versions with lowercase, uppercase, or no prefix, the Prefixed*
// methods should include the correct prefix while String() excludes it.
func TestSemVer_PrefixedMethods_PreservesPrefixCase(t *testing.T) {
	// Precondition: Version with lowercase 'v' prefix
	// Action: Call various output methods
	sv := Parse("v1.2.3-alpha+build")

	// Expected: String() excludes prefix, Prefixed* methods include it
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

	// Precondition: Version with uppercase 'V' prefix
	// Action: Call Prefixed* methods
	svUpper := Parse("V1.2.3-alpha+build")

	// Expected: Uppercase prefix preserved
	if svUpper.PrefixedString() != "V1.2.3" {
		t.Errorf("PrefixedString() with V: got %q, want %q", svUpper.PrefixedString(), "V1.2.3")
	}
	if svUpper.PrefixedSemVerString() != "V1.2.3-alpha" {
		t.Errorf("PrefixedSemVerString() with V: got %q, want %q", svUpper.PrefixedSemVerString(), "V1.2.3-alpha")
	}

	// Precondition: Version without prefix
	// Action: Call Prefixed* methods
	svNoPrefix := Parse("1.2.3-alpha+build")

	// Expected: No prefix in output (Prefixed* returns same as unprefixed)
	if svNoPrefix.PrefixedString() != "1.2.3" {
		t.Errorf("PrefixedString() no prefix: got %q, want %q", svNoPrefix.PrefixedString(), "1.2.3")
	}
	if svNoPrefix.PrefixedSemVerString() != "1.2.3-alpha" {
		t.Errorf("PrefixedSemVerString() no prefix: got %q, want %q", svNoPrefix.PrefixedSemVerString(), "1.2.3-alpha")
	}
}

// TestSemVer_OriginalMethods_PreservesInputFormat validates that Original* methods
// return the version in the same format style as the input.
//
// Why: When round-tripping versions (read, modify, write), preserving the original
// format prevents unnecessary churn in version files and git tags.
//
// What: Given versions with or without prefixes, Original* methods should
// match the input format.
func TestSemVer_OriginalMethods_PreservesInputFormat(t *testing.T) {
	// Precondition: Version without prefix
	// Action: Call Original* methods
	svNoPrefix := Parse("1.2.3-alpha")

	// Expected: No prefix in output
	if svNoPrefix.OriginalString() != "1.2.3" {
		t.Errorf("OriginalString() without prefix: got %q, want %q", svNoPrefix.OriginalString(), "1.2.3")
	}
	if svNoPrefix.OriginalSemVerString() != "1.2.3-alpha" {
		t.Errorf("OriginalSemVerString() without prefix: got %q, want %q", svNoPrefix.OriginalSemVerString(), "1.2.3-alpha")
	}

	// Precondition: Version with prefix
	// Action: Call Original* methods
	svWithPrefix := Parse("v1.2.3-alpha")

	// Expected: Prefix included in output
	if svWithPrefix.OriginalString() != "v1.2.3" {
		t.Errorf("OriginalString() with prefix: got %q, want %q", svWithPrefix.OriginalString(), "v1.2.3")
	}
	if svWithPrefix.OriginalSemVerString() != "v1.2.3-alpha" {
		t.Errorf("OriginalSemVerString() with prefix: got %q, want %q", svWithPrefix.OriginalSemVerString(), "v1.2.3-alpha")
	}
}

// TestSemVer_IsPreRelease_DetectsPreReleaseVersions validates that IsPreRelease()
// correctly identifies versions with pre-release identifiers.
//
// Why: Pre-release detection is used in CI/CD to determine if a version is stable,
// and in version comparison logic where pre-release < release.
//
// What: Given versions with and without pre-release identifiers, IsPreRelease()
// should return true only when a pre-release identifier is present.
func TestSemVer_IsPreRelease_DetectsPreReleaseVersions(t *testing.T) {
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
			// Precondition: Version with or without pre-release
			// Action: Call IsPreRelease()
			sv := Parse(tt.input)

			// Expected: Correct boolean based on pre-release presence
			if sv.IsPreRelease() != tt.expected {
				t.Errorf("IsPreRelease(): got %v, want %v", sv.IsPreRelease(), tt.expected)
			}
		})
	}
}

// TestSemVer_HasBuildMetadata_DetectsMetadata validates that HasBuildMetadata()
// correctly identifies versions with build metadata.
//
// Why: Metadata detection is used when formatting output or determining if
// additional build information is available.
//
// What: Given versions with and without build metadata, HasBuildMetadata()
// should return true only when metadata is present.
func TestSemVer_HasBuildMetadata_DetectsMetadata(t *testing.T) {
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
			// Precondition: Version with or without metadata
			// Action: Call HasBuildMetadata()
			sv := Parse(tt.input)

			// Expected: Correct boolean based on metadata presence
			if sv.HasBuildMetadata() != tt.expected {
				t.Errorf("HasBuildMetadata(): got %v, want %v", sv.HasBuildMetadata(), tt.expected)
			}
		})
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes. Currently, Parse() is lenient and returns
// zero values for invalid input rather than errors. These tests would be added
// if strict parsing with error returns is implemented.
// =============================================================================

// Note: Parse() uses lenient parsing and does not return errors.
// For strict parsing with error handling, use ParseStrict().
// Error handling tests for ParseStrict would go here.

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and specification compliance.
// =============================================================================

// TestParse_PartialVersions_DefaultsMissingComponents validates that the parser
// handles incomplete version strings by defaulting missing components to zero.
//
// Why: Users may specify abbreviated versions like "1.2" or just "1". The parser
// should handle these gracefully for convenience while maintaining valid semver output.
//
// What: Given version strings with missing minor and/or patch components,
// the parser should default missing values to 0.
func TestParse_PartialVersions_DefaultsMissingComponents(t *testing.T) {
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
			// Precondition: Partial version string (missing minor and/or patch)
			// Action: Parse the version
			sv := Parse(tt.input)

			// Expected: Missing components default to 0
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

// TestParse_SemVer2SpecExamples_ParsesSpecificationExamples validates parsing
// against all examples from the SemVer 2.0.0 specification.
//
// Why: Compliance with the official specification ensures interoperability with
// other semver implementations and tools. These are the canonical test cases.
//
// What: Given the exact pre-release and build metadata examples from the SemVer
// 2.0.0 spec (sections 9 and 10), all should parse correctly.
//
// See: resources/semver-2.md
func TestParse_SemVer2SpecExamples_ParsesSpecificationExamples(t *testing.T) {
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
			// Precondition: Pre-release example from SemVer 2.0.0 spec
			// Action: Parse the version
			sv := Parse(tt.input)

			// Expected: Components match spec examples exactly
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
			// Precondition: Build metadata example from SemVer 2.0.0 spec
			// Action: Parse the version
			sv := Parse(tt.input)

			// Expected: Components match spec examples exactly
			if sv.PreRelease != tt.preRel {
				t.Errorf("PreRelease: got %q, want %q", sv.PreRelease, tt.preRel)
			}
			if sv.BuildMetadata != tt.metadata {
				t.Errorf("BuildMetadata: got %q, want %q", sv.BuildMetadata, tt.metadata)
			}
		})
	}
}

// TestParse_SemVer2SpecPrecedenceExamples_ParsesAllPrecedenceVersions validates
// parsing of all versions used in the SemVer 2.0.0 precedence examples.
//
// Why: The precedence examples from spec section 11 represent the full range of
// version formats that must be handled for correct version ordering.
//
// What: Given all versions from the precedence example in the spec, each should
// parse without returning zero values (indicating parse failure).
func TestParse_SemVer2SpecPrecedenceExamples_ParsesAllPrecedenceVersions(t *testing.T) {
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
			// Precondition: Version from SemVer 2.0.0 precedence example
			// Action: Parse the version
			sv := Parse(input)

			// Expected: Version parses successfully (not all zeros unless input is 0.0.0)
			if sv.Major == 0 && sv.Minor == 0 && sv.Patch == 0 && input != "0.0.0" {
				t.Errorf("Failed to parse %q", input)
			}
		})
	}
}

// =============================================================================
// MINUTIAE
// Tests for obscure scenarios, utility functions, and detailed component extraction.
// =============================================================================

// TestSemVer_PreReleaseWithDash_ReturnsFormattedPreRelease validates that
// PreReleaseWithDash() returns the pre-release with a leading dash for templating.
//
// Why: Template systems often need pre-release with the dash included to construct
// filenames or paths like "app-1.2.3-alpha.tar.gz".
//
// What: Given versions with and without pre-release, PreReleaseWithDash() should
// return "-prerelease" or empty string respectively.
func TestSemVer_PreReleaseWithDash_ReturnsFormattedPreRelease(t *testing.T) {
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
			// Precondition: Version with or without pre-release
			// Action: Call PreReleaseWithDash()
			sv := Parse(tt.input)

			// Expected: Pre-release with leading dash, or empty string
			if sv.PreReleaseWithDash() != tt.expected {
				t.Errorf("PreReleaseWithDash(): got %q, want %q", sv.PreReleaseWithDash(), tt.expected)
			}
		})
	}
}

// TestSemVer_BuildMetadataWithPlus_ReturnsFormattedMetadata validates that
// BuildMetadataWithPlus() returns the metadata with a leading plus for templating.
//
// Why: Similar to PreReleaseWithDash(), templates may need metadata with the plus
// sign included for constructing version strings.
//
// What: Given versions with and without metadata, BuildMetadataWithPlus() should
// return "+metadata" or empty string respectively.
func TestSemVer_BuildMetadataWithPlus_ReturnsFormattedMetadata(t *testing.T) {
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
			// Precondition: Version with or without metadata
			// Action: Call BuildMetadataWithPlus()
			sv := Parse(tt.input)

			// Expected: Metadata with leading plus, or empty string
			if sv.BuildMetadataWithPlus() != tt.expected {
				t.Errorf("BuildMetadataWithPlus(): got %q, want %q", sv.BuildMetadataWithPlus(), tt.expected)
			}
		})
	}
}

// TestSemVer_PreReleaseLabel_ExtractsLabelPortion validates that PreReleaseLabel()
// extracts just the label portion from a pre-release identifier.
//
// Why: When incrementing pre-release numbers or categorizing releases, the label
// portion (alpha, beta, rc) must be extracted separately from any numeric suffix.
//
// What: Given pre-release identifiers with various formats, PreReleaseLabel()
// should return just the non-numeric label portion.
func TestSemVer_PreReleaseLabel_ExtractsLabelPortion(t *testing.T) {
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
			// Precondition: Version with various pre-release formats
			// Action: Call PreReleaseLabel()
			sv := Parse(tt.input)

			// Expected: Label portion extracted correctly
			if sv.PreReleaseLabel() != tt.expected {
				t.Errorf("PreReleaseLabel(): got %q, want %q", sv.PreReleaseLabel(), tt.expected)
			}
		})
	}
}

// TestSemVer_PreReleaseNumber_ExtractsNumericPortion validates that PreReleaseNumber()
// extracts the numeric portion from a pre-release identifier.
//
// Why: When incrementing pre-release versions (alpha.1 -> alpha.2), the numeric
// portion must be extracted for arithmetic operations.
//
// What: Given pre-release identifiers with various formats, PreReleaseNumber()
// should return the last numeric component, or -1 if none exists.
func TestSemVer_PreReleaseNumber_ExtractsNumericPortion(t *testing.T) {
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
			// Precondition: Version with various pre-release formats
			// Action: Call PreReleaseNumber()
			sv := Parse(tt.input)

			// Expected: Numeric portion extracted, or -1 if none
			if sv.PreReleaseNumber() != tt.expected {
				t.Errorf("PreReleaseNumber(): got %d, want %d", sv.PreReleaseNumber(), tt.expected)
			}
		})
	}
}

// TestSemVer_PreReleaseLabelWithDash_ReturnsFormattedLabel validates that
// PreReleaseLabelWithDash() returns the label with a leading dash.
//
// Why: For template systems that need the label portion with dash for constructing
// strings like "app-alpha" without the numeric suffix.
//
// What: Given versions with various pre-release formats, PreReleaseLabelWithDash()
// should return "-label" or empty string.
func TestSemVer_PreReleaseLabelWithDash_ReturnsFormattedLabel(t *testing.T) {
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
			// Precondition: Version with or without pre-release
			// Action: Call PreReleaseLabelWithDash()
			sv := Parse(tt.input)

			// Expected: Label with leading dash, or empty string
			if sv.PreReleaseLabelWithDash() != tt.expected {
				t.Errorf("PreReleaseLabelWithDash(): got %q, want %q", sv.PreReleaseLabelWithDash(), tt.expected)
			}
		})
	}
}

// TestSemVer_AssemblyVersion_ReturnsWindowsFormat validates that AssemblyVersion()
// returns a Windows assembly-compatible four-part version.
//
// Why: Windows assemblies require four-part versions (Major.Minor.Patch.Build).
// This method provides compatibility with .NET and Windows versioning systems.
//
// What: Given any version, AssemblyVersion() should return "Major.Minor.Patch.0".
func TestSemVer_AssemblyVersion_ReturnsWindowsFormat(t *testing.T) {
	// Precondition: A standard semver version
	// Action: Call AssemblyVersion()
	sv := Parse("1.2.3")

	// Expected: Four-part version with trailing .0
	if sv.AssemblyVersion() != "1.2.3.0" {
		t.Errorf("AssemblyVersion(): got %q, want %q", sv.AssemblyVersion(), "1.2.3.0")
	}
}

// TestStripPrefix_RemovesPrefixFromVersionString validates that StripPrefix()
// removes v/V prefixes from version strings.
//
// Why: Some tools or APIs require unprefixed version strings. StripPrefix()
// provides a consistent way to remove prefixes for interoperability.
//
// What: Given version strings with various prefix styles, StripPrefix() should
// remove only the first v/V prefix.
func TestStripPrefix_RemovesPrefixFromVersionString(t *testing.T) {
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
			// Precondition: Version string with or without prefix
			// Action: Call StripPrefix()
			result := StripPrefix(tt.input)

			// Expected: Single v/V prefix removed
			if result != tt.expected {
				t.Errorf("StripPrefix(%q): got %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEscapedBranchName_EscapesSlashesToDashes validates that EscapedBranchName()
// replaces forward slashes with dashes for safe use in version identifiers.
//
// Why: Git branch names often contain slashes (feature/foo) but slashes are invalid
// in many contexts (filenames, URLs, semver identifiers). This provides safe escaping.
//
// What: Given branch names with various slash patterns, EscapedBranchName() should
// replace all slashes with dashes.
func TestEscapedBranchName_EscapesSlashesToDashes(t *testing.T) {
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
			// Precondition: Branch name with or without slashes
			// Action: Call EscapedBranchName()
			result := EscapedBranchName(tt.input)

			// Expected: All slashes replaced with dashes
			if result != tt.expected {
				t.Errorf("EscapedBranchName(%q): got %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
