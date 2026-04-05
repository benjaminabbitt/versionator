package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// intPtr returns a pointer to the given int value.
func intPtr(n int) *int { return &n }

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose: parsing valid SemVer version strings
// and converting them to structured Version objects.
// =============================================================================

// TestParse_SemVerCore_ValidVersions validates that the parser correctly handles
// standard semantic version strings (MAJOR.MINOR.PATCH format).
//
// Why: Version parsing is the foundational capability of this package. If basic
// version parsing fails, all downstream functionality (bumping, comparing, etc.)
// will be broken.
//
// What: Given valid version strings in M.M.P format, the parser should correctly
// extract major, minor, and patch components into a Version struct.
func TestParse_SemVerCore_ValidVersions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		major    int
		minor    int
		patch    int
		hasError bool
	}{
		{name: "full version", input: "1.2.3", major: 1, minor: 2, patch: 3},
		{name: "zero version", input: "0.0.0", major: 0, minor: 0, patch: 0},
		{name: "large numbers", input: "100.200.300", major: 100, minor: 200, patch: 300},
		{name: "partial major only", input: "1", major: 1, minor: 0, patch: 0},
		{name: "partial major.minor", input: "1.2", major: 1, minor: 2, patch: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Input is a valid SemVer core version string
			// Action: Parse the version string
			v, err := Parse(tt.input)

			// Expected: Successfully parse and extract correct components
			if tt.hasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.major, v.Major())
			assert.Equal(t, tt.minor, v.Minor())
			assert.Equal(t, tt.patch, v.Patch())
		})
	}
}

// TestVersion_String_OutputFormat validates that Version.String() returns the
// canonical SemVer format without prefix.
//
// Why: Consistent string output is essential for writing VERSION files, generating
// tags, and interoperating with other SemVer-aware tools.
//
// What: Given a parsed version, String() should return M.M.P[-prerelease][+metadata]
// format, normalizing partial versions to full three-component form.
func TestVersion_String_OutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "core only", input: "1.2.3", expected: "1.2.3"},
		{name: "with prefix", input: "v1.2.3", expected: "1.2.3"},
		{name: "with pre-release", input: "1.0.0-alpha", expected: "1.0.0-alpha"},
		{name: "with metadata", input: "1.0.0+build", expected: "1.0.0+build"},
		{name: "full", input: "v1.0.0-alpha+build", expected: "1.0.0-alpha+build"},
		{name: "partial", input: "1", expected: "1.0.0"},
		{name: "partial major.minor", input: "1.2", expected: "1.2.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string can be parsed
			v, err := Parse(tt.input)
			require.NoError(t, err)

			// Action: Convert to string representation
			// Expected: Canonical SemVer format without prefix
			assert.Equal(t, tt.expected, v.String())
		})
	}
}

// TestVersion_FullString_WithPrefix validates that FullString() preserves the
// original prefix in the output.
//
// Why: Many projects use "v1.0.0" format for git tags. Preserving the prefix
// ensures round-trip compatibility with existing versioning conventions.
//
// What: Given a version with prefix, FullString() should include the prefix
// in the output while String() should exclude it.
func TestVersion_FullString_WithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "without prefix", input: "1.2.3", expected: "1.2.3"},
		{name: "with v prefix", input: "v1.2.3", expected: "v1.2.3"},
		{name: "with V prefix", input: "V1.2.3", expected: "V1.2.3"},
		{name: "with pre-release", input: "v1.0.0-alpha", expected: "v1.0.0-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string can be parsed
			v, err := Parse(tt.input)
			require.NoError(t, err)

			// Action: Convert to full string representation
			// Expected: Output includes prefix if present in input
			assert.Equal(t, tt.expected, v.FullString())
		})
	}
}

// TestVersion_CoreVersion_StripsExtras validates that CoreVersion() returns only
// the M.M.P portion without pre-release or metadata.
//
// Why: Some version comparisons and file formats require only the core version
// without any extensions. This supports those use cases.
//
// What: Given any valid version, CoreVersion() should return only Major.Minor.Patch.
func TestVersion_CoreVersion_StripsExtras(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "core only", input: "1.2.3", expected: "1.2.3"},
		{name: "with pre-release", input: "1.0.0-alpha", expected: "1.0.0"},
		{name: "with metadata", input: "1.0.0+build", expected: "1.0.0"},
		{name: "full", input: "v1.0.0-alpha+build", expected: "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string can be parsed
			v, err := Parse(tt.input)
			require.NoError(t, err)

			// Action: Extract core version
			// Expected: Only M.M.P, no prefix/prerelease/metadata
			assert.Equal(t, tt.expected, v.CoreVersion())
		})
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests covering important alternate flows: prefixes, pre-release tags,
// build metadata, combined formats, and assembly versions.
// =============================================================================

// TestParse_WithPrefix_VPrefixHandling validates parsing of versions with v/V prefix.
//
// Why: The "v" prefix is ubiquitous in git tags (e.g., v1.0.0). The parser must
// correctly extract and preserve this prefix for tag generation and comparison.
//
// What: Given versions with v or V prefix, parser should extract the prefix
// separately from version components, and HasPrefix() should return true.
func TestParse_WithPrefix_VPrefixHandling(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prefix string
		major  int
		minor  int
		patch  int
	}{
		{name: "lowercase v", input: "v1.2.3", prefix: "v", major: 1, minor: 2, patch: 3},
		{name: "uppercase V", input: "V1.2.3", prefix: "V", major: 1, minor: 2, patch: 3},
		{name: "v with zeros", input: "v0.0.0", prefix: "v", major: 0, minor: 0, patch: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string has v/V prefix
			// Action: Parse the prefixed version
			v, err := Parse(tt.input)

			// Expected: Prefix extracted, version components correct, HasPrefix true
			require.NoError(t, err)
			assert.Equal(t, tt.prefix, v.Prefix)
			assert.Equal(t, tt.major, v.Major())
			assert.Equal(t, tt.minor, v.Minor())
			assert.Equal(t, tt.patch, v.Patch())
			assert.True(t, v.HasPrefix())
		})
	}
}

// TestParse_PreRelease_Identifiers validates parsing of pre-release suffixes.
//
// Why: Pre-release tags (alpha, beta, rc.1) are critical for release workflows.
// They affect version precedence and identify unstable releases.
//
// What: Given versions with various pre-release formats, parser should correctly
// extract the pre-release string and IsPreRelease() should return true.
func TestParse_PreRelease_Identifiers(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		preRelease string
	}{
		{name: "alpha", input: "1.0.0-alpha", preRelease: "alpha"},
		{name: "alpha.1", input: "1.0.0-alpha.1", preRelease: "alpha.1"},
		{name: "numeric", input: "1.0.0-0.3.7", preRelease: "0.3.7"},
		{name: "complex", input: "1.0.0-x.7.z.92", preRelease: "x.7.z.92"},
		{name: "with dashes", input: "1.0.0-x-y-z", preRelease: "x-y-z"},
		{name: "rc.1", input: "1.0.0-rc.1", preRelease: "rc.1"},
		{name: "beta.2", input: "1.0.0-beta.2", preRelease: "beta.2"},
		{name: "beta.11", input: "1.0.0-beta.11", preRelease: "beta.11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string has pre-release suffix
			// Action: Parse the version
			v, err := Parse(tt.input)

			// Expected: Pre-release extracted, IsPreRelease() true
			require.NoError(t, err)
			assert.True(t, v.IsPreRelease())
			assert.Equal(t, tt.preRelease, v.PreReleaseString())
		})
	}
}

// TestParse_BuildMetadata_Identifiers validates parsing of build metadata suffixes.
//
// Why: Build metadata (commit SHA, build number, timestamp) provides traceability
// without affecting version precedence per SemVer spec.
//
// What: Given versions with +metadata suffix, parser should extract metadata
// correctly and HasBuildMetadata() should return true.
func TestParse_BuildMetadata_Identifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		metadata string
	}{
		{name: "simple", input: "1.0.0+build", metadata: "build"},
		{name: "with number", input: "1.0.0+build.123", metadata: "build.123"},
		{name: "timestamp", input: "1.0.0+20130313144700", metadata: "20130313144700"},
		{name: "sha", input: "1.0.0+exp.sha.5114f85", metadata: "exp.sha.5114f85"},
		{name: "complex", input: "1.0.0+21AF26D3", metadata: "21AF26D3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string has build metadata suffix
			// Action: Parse the version
			v, err := Parse(tt.input)

			// Expected: Metadata extracted, HasBuildMetadata() true
			require.NoError(t, err)
			assert.True(t, v.HasBuildMetadata())
			assert.Equal(t, tt.metadata, v.BuildMetadataString())
		})
	}
}

// TestParse_Combined_AllComponents validates parsing versions with all components.
//
// Why: Real-world versions often combine prefix, pre-release, and metadata
// (e.g., v1.0.0-rc.1+build.123). Parser must handle full complexity.
//
// What: Given versions with all optional components, parser should correctly
// extract every part into appropriate Version fields.
func TestParse_Combined_AllComponents(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		prefix     string
		major      int
		minor      int
		patch      int
		preRelease string
		metadata   string
	}{
		{
			name:       "full with prefix",
			input:      "v1.2.3-alpha.1+build.456",
			prefix:     "v",
			major:      1,
			minor:      2,
			patch:      3,
			preRelease: "alpha.1",
			metadata:   "build.456",
		},
		{
			name:       "pre-release and metadata",
			input:      "1.0.0-beta+exp.sha.5114f85",
			major:      1,
			minor:      0,
			patch:      0,
			preRelease: "beta",
			metadata:   "exp.sha.5114f85",
		},
		{
			name:       "SemVer spec example",
			input:      "1.0.0-alpha+001",
			major:      1,
			minor:      0,
			patch:      0,
			preRelease: "alpha",
			metadata:   "001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version has prefix, pre-release, and/or metadata
			// Action: Parse the full version string
			v, err := Parse(tt.input)

			// Expected: All components correctly extracted
			require.NoError(t, err)
			assert.Equal(t, tt.prefix, v.Prefix)
			assert.Equal(t, tt.major, v.Major())
			assert.Equal(t, tt.minor, v.Minor())
			assert.Equal(t, tt.patch, v.Patch())
			assert.Equal(t, tt.preRelease, v.PreReleaseString())
			assert.Equal(t, tt.metadata, v.BuildMetadataString())
		})
	}
}

// TestParse_AssemblyVersion_FourComponents validates parsing .NET assembly versions.
//
// Why: .NET projects use 4-component versions (M.m.p.revision). Supporting this
// format enables versioning for .NET libraries and applications.
//
// What: Given a 4-component version string, parser should extract all four
// components and IsAssemblyVersion() should return true.
func TestParse_AssemblyVersion_FourComponents(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		major    int
		minor    int
		patch    int
		revision int
	}{
		{name: "full assembly", input: "1.2.3.4", major: 1, minor: 2, patch: 3, revision: 4},
		{name: "zero assembly", input: "0.0.0.0", major: 0, minor: 0, patch: 0, revision: 0},
		{name: "large revision", input: "1.0.0.65534", major: 1, minor: 0, patch: 0, revision: 65534},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string has 4 components
			// Action: Parse the assembly version
			v, err := Parse(tt.input)

			// Expected: All 4 components extracted, IsAssemblyVersion() true
			require.NoError(t, err)
			assert.True(t, v.IsAssemblyVersion())
			assert.Equal(t, tt.major, v.Major())
			assert.Equal(t, tt.minor, v.Minor())
			assert.Equal(t, tt.patch, v.Patch())
			assert.Equal(t, tt.revision, v.Revision())
		})
	}
}

// TestVersion_AssemblyVersionOutput_Conversion validates converting SemVer to assembly format.
//
// Why: When outputting to .NET projects, SemVer versions must be converted to
// 4-component format with revision defaulting to 0.
//
// What: Given any version, AssemblyVersion() should return M.M.P.R format with
// revision defaulting to 0 for standard SemVer inputs.
func TestVersion_AssemblyVersionOutput_Conversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "semver to assembly", input: "1.2.3", expected: "1.2.3.0"},
		{name: "assembly passthrough", input: "1.2.3.4", expected: "1.2.3.4"},
		{name: "partial", input: "1", expected: "1.0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string can be parsed
			v, err := Parse(tt.input)
			require.NoError(t, err)

			// Action: Convert to assembly version format
			// Expected: 4-component output with revision (defaults to 0)
			assert.Equal(t, tt.expected, v.AssemblyVersion())
		})
	}
}

// TestVersion_Predicates_BooleanQueries validates all version predicate methods.
//
// Why: Predicates enable conditional logic in version manipulation workflows
// (e.g., "if pre-release, handle differently").
//
// What: Each predicate should return true/false based on version properties.
func TestVersion_Predicates_BooleanQueries(t *testing.T) {
	t.Run("HasPrefix", func(t *testing.T) {
		// Precondition: Two versions, one with prefix, one without
		v1, _ := Parse("v1.0.0")
		v2, _ := Parse("1.0.0")

		// Action/Expected: HasPrefix() reflects presence of prefix
		assert.True(t, v1.HasPrefix())
		assert.False(t, v2.HasPrefix())
	})

	t.Run("IsPreRelease", func(t *testing.T) {
		// Precondition: Two versions, one pre-release, one stable
		v1, _ := Parse("1.0.0-alpha")
		v2, _ := Parse("1.0.0")

		// Action/Expected: IsPreRelease() reflects pre-release status
		assert.True(t, v1.IsPreRelease())
		assert.False(t, v2.IsPreRelease())
	})

	t.Run("HasBuildMetadata", func(t *testing.T) {
		// Precondition: Two versions, one with metadata, one without
		v1, _ := Parse("1.0.0+build")
		v2, _ := Parse("1.0.0")

		// Action/Expected: HasBuildMetadata() reflects metadata presence
		assert.True(t, v1.HasBuildMetadata())
		assert.False(t, v2.HasBuildMetadata())
	})

	t.Run("IsAssemblyVersion", func(t *testing.T) {
		// Precondition: Two versions, one 4-component, one 3-component
		v1, _ := Parse("1.2.3.4")
		v2, _ := Parse("1.2.3")

		// Action/Expected: IsAssemblyVersion() reflects component count
		assert.True(t, v1.IsAssemblyVersion())
		assert.False(t, v2.IsAssemblyVersion())
	})

	t.Run("IsPartial", func(t *testing.T) {
		// Precondition: Versions with varying component counts
		v1, _ := Parse("1")
		v2, _ := Parse("1.2")
		v3, _ := Parse("1.2.3")

		// Action/Expected: IsPartial() true when missing minor or patch
		assert.True(t, v1.IsPartial())
		assert.True(t, v2.IsPartial())
		assert.False(t, v3.IsPartial())
	})
}

// =============================================================================
// ERROR HANDLING
// Tests validating expected failure modes and error messages.
// =============================================================================

// TestParse_Invalid_LeadingZeros validates rejection of leading zeros in version components.
//
// Why: SemVer 2.0.0 explicitly forbids leading zeros in numeric identifiers
// (e.g., "01.2.3" is invalid). This prevents ambiguous version sorting.
//
// What: Given version strings with leading zeros, parser should return error
// containing ErrLeadingZero message.
func TestParse_Invalid_LeadingZeros(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "major leading zero", input: "01.2.3"},
		{name: "minor leading zero", input: "1.02.3"},
		{name: "patch leading zero", input: "1.2.03"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version string has leading zeros
			// Action: Attempt to parse
			_, err := Parse(tt.input)

			// Expected: Parse fails with leading zero error
			require.Error(t, err)
			assert.Contains(t, err.Error(), ErrLeadingZero)
		})
	}
}

// TestParse_Invalid_EmptyString validates rejection of empty version strings.
//
// Why: Empty string is not a valid version and should fail fast with clear error.
//
// What: Given empty string input, parser should return ErrEmptyVersion error.
func TestParse_Invalid_EmptyString(t *testing.T) {
	// Precondition: Empty string input
	// Action: Attempt to parse
	_, err := Parse("")

	// Expected: Parse fails with empty version error
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrEmptyVersion)
}

// TestParse_Invalid_WhitespaceOnly validates rejection of whitespace-only strings.
//
// Why: Whitespace-only input is effectively empty and should not parse
// to a valid version.
//
// What: Given whitespace-only input, parser should return error.
func TestParse_Invalid_WhitespaceOnly(t *testing.T) {
	// Precondition: Whitespace-only input
	// Action: Attempt to parse
	_, err := Parse("   ")

	// Expected: Parse fails (whitespace is trimmed, resulting in empty)
	require.Error(t, err)
}

// TestValidatePreReleaseIdentifier_InvalidCases validates pre-release identifier validation.
//
// Why: Pre-release identifiers have specific rules (no leading zeros in numeric
// identifiers, non-empty). Validation prevents invalid versions.
//
// What: Various invalid identifier scenarios should return appropriate errors.
func TestValidatePreReleaseIdentifier_InvalidCases(t *testing.T) {
	t.Run("nil identifier returns error", func(t *testing.T) {
		// Precondition: nil identifier
		// Action: Validate
		err := validatePreReleaseIdentifier(nil)

		// Expected: Error for empty identifier
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrEmptyIdentifier)
	})

	t.Run("empty string identifier returns error", func(t *testing.T) {
		// Precondition: Identifier with all nil fields
		id := &Identifier{}

		// Action: Validate
		err := validatePreReleaseIdentifier(id)

		// Expected: Error for empty identifier
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrEmptyIdentifier)
	})

	t.Run("numeric with leading zero returns error", func(t *testing.T) {
		// Precondition: Numeric identifier with leading zero
		val := "007"
		id := &Identifier{Number: &val}

		// Action: Validate
		err := validatePreReleaseIdentifier(id)

		// Expected: Error for leading zero
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrLeadingZero)
	})
}

// TestValidatePreReleaseIdentifier_ValidCases validates acceptance of valid identifiers.
//
// Why: Ensures valid identifiers pass validation without errors.
//
// What: Valid numeric and alphanumeric identifiers should pass validation.
func TestValidatePreReleaseIdentifier_ValidCases(t *testing.T) {
	t.Run("valid numeric passes", func(t *testing.T) {
		// Precondition: Valid numeric identifier without leading zero
		val := "42"
		id := &Identifier{Number: &val}

		// Action: Validate
		err := validatePreReleaseIdentifier(id)

		// Expected: No error
		require.NoError(t, err)
	})

	t.Run("valid alphanumeric passes", func(t *testing.T) {
		// Precondition: Valid alphanumeric identifier
		val := "alpha"
		id := &Identifier{Ident: &val}

		// Action: Validate
		err := validatePreReleaseIdentifier(id)

		// Expected: No error
		require.NoError(t, err)
	})
}

// TestValidateBuildMetadataIdentifier_Cases validates build metadata identifier rules.
//
// Why: Build metadata has different rules than pre-release (leading zeros allowed).
// Tests ensure correct differentiation.
//
// What: Invalid cases should error; leading zeros should be allowed.
func TestValidateBuildMetadataIdentifier_Cases(t *testing.T) {
	t.Run("nil identifier returns error", func(t *testing.T) {
		// Precondition: nil identifier
		// Action: Validate
		err := validateBuildMetadataIdentifier(nil)

		// Expected: Error for empty identifier
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrEmptyIdentifier)
	})

	t.Run("empty string returns error", func(t *testing.T) {
		// Precondition: Identifier with all nil fields
		id := &Identifier{}

		// Action: Validate
		err := validateBuildMetadataIdentifier(id)

		// Expected: Error for empty identifier
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrEmptyIdentifier)
	})

	t.Run("leading zero allowed in build metadata", func(t *testing.T) {
		// Precondition: Numeric identifier with leading zero
		val := "007"
		id := &Identifier{Number: &val}

		// Action: Validate build metadata (not pre-release)
		err := validateBuildMetadataIdentifier(id)

		// Expected: No error - build metadata allows leading zeros
		require.NoError(t, err)
	})
}

// TestBuilderMethod_InvalidPrefix validates rejection of invalid prefixes.
//
// Why: Only v/V prefixes are valid per SemVer convention. Other prefixes
// should be rejected during Build().
//
// What: Given an invalid prefix like "ver", Build() should return error.
func TestBuilderMethod_InvalidPrefix(t *testing.T) {
	// Precondition: Builder with invalid prefix
	// Action: Attempt to build
	_, err := NewBuilder().
		Prefix("ver").
		Major(1).
		Build()

	// Expected: Error for invalid prefix
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prefix")
}

// TestBuilderMethod_NegativeRevision validates rejection of negative revision.
//
// Why: Version components cannot be negative. Builder should validate this.
//
// What: Given negative revision, Build() should return error.
func TestBuilderMethod_NegativeRevision(t *testing.T) {
	// Precondition: Builder with negative revision
	// Action: Attempt to build
	_, err := NewBuilder().
		Major(1).
		Revision(-1).
		Build()

	// Expected: Error for negative version component
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrNegativeVersion)
}

// =============================================================================
// EDGE CASES
// Tests covering boundary conditions and unusual but valid inputs.
// =============================================================================

// TestParseLenient_FallbackBehavior validates lenient parsing for invalid inputs.
//
// Why: When reading existing VERSION files, we may encounter legacy or
// non-standard formats. Lenient parsing provides best-effort handling.
//
// What: Invalid inputs should return zero-version with raw preserved;
// valid inputs should parse normally.
func TestParseLenient_FallbackBehavior(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "valid version", input: "1.2.3", expected: "1.2.3"},
		{name: "invalid version", input: "not-a-version", expected: ""},
		{name: "empty string", input: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Various input formats
			// Action: Parse leniently
			v := ParseLenient(tt.input)

			// Expected: Non-nil version returned; valid inputs parsed correctly
			assert.NotNil(t, v)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, v.String())
			}
		})
	}
}

// TestVersion_NilCore_SafeAccessors validates safe handling of nil Version.Core.
//
// Why: Defensive programming - methods should not panic on nil/zero values.
//
// What: All accessor methods should return sensible defaults for nil Core.
func TestVersion_NilCore_SafeAccessors(t *testing.T) {
	t.Run("Major returns 0 for nil core", func(t *testing.T) {
		// Precondition: Version with nil Core
		v := &Version{Core: nil}

		// Action/Expected: Major() returns 0, not panic
		assert.Equal(t, 0, v.Major())
	})

	t.Run("IsPartial returns true for nil core", func(t *testing.T) {
		// Precondition: Version with nil Core
		v := &Version{Core: nil}

		// Action/Expected: IsPartial() returns true (missing all components)
		assert.True(t, v.IsPartial())
	})

	t.Run("IsPartial returns true for missing minor", func(t *testing.T) {
		// Precondition: Version with only Major set
		v := &Version{Core: &VersionCore{Major: 1}}

		// Action/Expected: IsPartial() returns true
		assert.True(t, v.IsPartial())
	})

	t.Run("CoreVersion returns 0.0.0 for nil core", func(t *testing.T) {
		// Precondition: Version with nil Core
		v := &Version{Core: nil}

		// Action/Expected: CoreVersion() returns valid default
		assert.Equal(t, "0.0.0", v.CoreVersion())
	})

	t.Run("AssemblyVersion returns 0.0.0.0 for nil core", func(t *testing.T) {
		// Precondition: Version with nil Core
		v := &Version{Core: nil}

		// Action/Expected: AssemblyVersion() returns valid default
		assert.Equal(t, "0.0.0.0", v.AssemblyVersion())
	})
}

// TestToVersionData_NilVersion validates safe handling of nil Version pointer.
//
// Why: ToVersionData() may be called on nil Version; it should return zeros.
//
// What: All returned values should be zero/empty for nil Version.
func TestToVersionData_NilVersion(t *testing.T) {
	// Precondition: nil Version pointer
	var v *Version = nil

	// Action: Extract version data
	prefix, major, minor, patch, revision, preRel, metadata, raw := v.ToVersionData()

	// Expected: All values are zero/empty/nil
	assert.Empty(t, prefix)
	assert.Equal(t, 0, major)
	assert.Equal(t, 0, minor)
	assert.Equal(t, 0, patch)
	assert.Nil(t, revision)
	assert.Empty(t, preRel)
	assert.Empty(t, metadata)
	assert.Empty(t, raw)
}

// TestClone_NilVersion validates that cloning nil returns nil.
//
// Why: Clone() should not panic on nil receiver.
//
// What: Cloning nil Version should return nil, not panic.
func TestClone_NilVersion(t *testing.T) {
	// Precondition: nil Version pointer
	var v *Version = nil

	// Action: Clone
	clone := v.Clone()

	// Expected: nil returned, no panic
	assert.Nil(t, clone)
}

// TestIdentifier_String_AllTypes validates String() for all identifier types.
//
// Why: Identifiers can be numbers, alphanumeric, dashes, or mixed. String()
// must handle all types correctly.
//
// What: Each identifier type should return its value from String().
func TestIdentifier_String_AllTypes(t *testing.T) {
	t.Run("dashes identifier", func(t *testing.T) {
		// Precondition: Identifier with dashes
		val := "--"
		id := &Identifier{Dashes: &val}

		// Action/Expected: String() returns the dashes
		assert.Equal(t, "--", id.String())
	})

	t.Run("mixed identifier", func(t *testing.T) {
		// Precondition: Identifier with mixed alphanumeric
		val := "abc123"
		id := &Identifier{Mixed: &val}

		// Action/Expected: String() returns the mixed value
		assert.Equal(t, "abc123", id.String())
	})

	t.Run("nil identifier", func(t *testing.T) {
		// Precondition: nil Identifier pointer
		var id *Identifier = nil

		// Action/Expected: String() returns empty string, no panic
		assert.Equal(t, "", id.String())
	})
}

// =============================================================================
// MINUTIAE
// Tests covering obscure scenarios, version mutation, builder patterns,
// and data conversion roundtrips.
// =============================================================================

// TestEBNF_GrammarAvailable validates that EBNF grammar is accessible.
//
// Why: EBNF grammar exposure enables debugging and documentation generation.
//
// What: EBNF() should return non-empty string containing grammar rules.
func TestEBNF_GrammarAvailable(t *testing.T) {
	// Precondition: Parser is initialized
	// Action: Get EBNF grammar
	ebnf := EBNF()

	// Expected: Non-empty string containing key grammar rules
	assert.NotEmpty(t, ebnf)
	assert.Contains(t, ebnf, "VersionFile")
}

// TestToVersionData_RoundTrip validates conversion to and from version data.
//
// Why: ToVersionData() enables integration with version package. It must
// correctly extract all fields for reconstruction.
//
// What: Parsed versions should have all fields extractable via ToVersionData().
func TestToVersionData_RoundTrip(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedPrefix   string
		expectedMajor    int
		expectedMinor    int
		expectedPatch    int
		expectedPreRel   string
		expectedMetadata string
	}{
		{
			name:          "simple version",
			input:         "1.2.3",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
		},
		{
			name:           "with prefix",
			input:          "v1.2.3",
			expectedPrefix: "v",
			expectedMajor:  1,
			expectedMinor:  2,
			expectedPatch:  3,
		},
		{
			name:           "with pre-release",
			input:          "1.0.0-alpha.1",
			expectedMajor:  1,
			expectedPreRel: "alpha.1",
		},
		{
			name:             "with build metadata",
			input:            "1.0.0+build.123",
			expectedMajor:    1,
			expectedMetadata: "build.123",
		},
		{
			name:             "full version",
			input:            "v2.3.4-beta.2+sha.abc123",
			expectedPrefix:   "v",
			expectedMajor:    2,
			expectedMinor:    3,
			expectedPatch:    4,
			expectedPreRel:   "beta.2",
			expectedMetadata: "sha.abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Valid version string
			v, err := Parse(tt.input)
			require.NoError(t, err)

			// Action: Extract version data tuple
			prefix, major, minor, patch, _, preRel, metadata, raw := v.ToVersionData()

			// Expected: All components correctly extracted
			assert.Equal(t, tt.expectedPrefix, prefix)
			assert.Equal(t, tt.expectedMajor, major)
			assert.Equal(t, tt.expectedMinor, minor)
			assert.Equal(t, tt.expectedPatch, patch)
			assert.Equal(t, tt.expectedPreRel, preRel)
			assert.Equal(t, tt.expectedMetadata, metadata)
			assert.NotEmpty(t, raw)
		})
	}
}

// TestFromVersionData_Construction validates building versions from data tuples.
//
// Why: FromVersionData() enables constructing versions programmatically without
// parsing strings. It must produce correct output.
//
// What: Given version components, FromVersionData() should construct version
// with correct FullString() output.
func TestFromVersionData_Construction(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		major         int
		minor         int
		patch         int
		revision      *int
		preRelease    string
		buildMetadata string
		expected      string
	}{
		{
			name:     "simple version",
			major:    1,
			minor:    2,
			patch:    3,
			expected: "1.2.3",
		},
		{
			name:     "with prefix",
			prefix:   "v",
			major:    1,
			minor:    0,
			patch:    0,
			expected: "v1.0.0",
		},
		{
			name:       "with pre-release",
			major:      1,
			preRelease: "alpha",
			expected:   "1.0.0-alpha",
		},
		{
			name:          "with build metadata",
			major:         1,
			buildMetadata: "build.123",
			expected:      "1.0.0+build.123",
		},
		{
			name:          "full version",
			prefix:        "v",
			major:         2,
			minor:         3,
			patch:         4,
			preRelease:    "rc.1",
			buildMetadata: "sha.def456",
			expected:      "v2.3.4-rc.1+sha.def456",
		},
		{
			name:     "with revision",
			major:    1,
			minor:    2,
			patch:    3,
			revision: intPtr(4),
			expected: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Version data components
			// Action: Construct version from data
			v := FromVersionData(tt.prefix, tt.major, tt.minor, tt.patch, tt.revision, tt.preRelease, tt.buildMetadata)

			// Expected: Version constructed with correct output
			require.NotNil(t, v)
			assert.Equal(t, tt.expected, v.FullString())
		})
	}
}

// TestClone_Independence validates that cloned versions are independent copies.
//
// Why: Clone() must create deep copies so mutations don't affect the original.
//
// What: Modifying a cloned version should not affect the original.
func TestClone_Independence(t *testing.T) {
	t.Run("simple version independence", func(t *testing.T) {
		// Precondition: Parse a version
		v, err := Parse("1.2.3")
		require.NoError(t, err)

		// Action: Clone and modify
		clone := v.Clone()
		require.NotNil(t, clone)
		assert.Equal(t, v.String(), clone.String())

		clone.SetMajor(9)

		// Expected: Original unchanged
		assert.Equal(t, 1, v.Major())
		assert.Equal(t, 9, clone.Major())
	})

	t.Run("with pre-release", func(t *testing.T) {
		// Precondition: Version with pre-release
		v, err := Parse("1.0.0-alpha.1")
		require.NoError(t, err)

		// Action: Clone
		clone := v.Clone()

		// Expected: Pre-release preserved
		assert.Equal(t, "alpha.1", clone.PreReleaseString())
	})

	t.Run("with build metadata", func(t *testing.T) {
		// Precondition: Version with metadata
		v, err := Parse("1.0.0+build.123")
		require.NoError(t, err)

		// Action: Clone
		clone := v.Clone()

		// Expected: Metadata preserved
		assert.Equal(t, "build.123", clone.BuildMetadataString())
	})

	t.Run("assembly version revision", func(t *testing.T) {
		// Precondition: 4-component assembly version
		v, err := Parse("1.2.3.4")
		require.NoError(t, err)

		// Action: Clone
		clone := v.Clone()

		// Expected: Revision preserved
		assert.Equal(t, 4, clone.Revision())
	})

	t.Run("full version all fields", func(t *testing.T) {
		// Precondition: Version with all optional fields
		v, err := Parse("v1.2.3-alpha.1+build.456")
		require.NoError(t, err)

		// Action: Clone
		clone := v.Clone()

		// Expected: All fields preserved
		assert.Equal(t, "v", clone.Prefix)
		assert.Equal(t, 1, clone.Major())
		assert.Equal(t, 2, clone.Minor())
		assert.Equal(t, 3, clone.Patch())
		assert.Equal(t, "alpha.1", clone.PreReleaseString())
		assert.Equal(t, "build.456", clone.BuildMetadataString())
	})
}

// TestVersion_SetMajor_ResetsLowerComponents validates SemVer reset behavior.
//
// Why: Per SemVer, incrementing major resets minor, patch, and pre-release.
// SetMajor() should follow this convention.
//
// What: After SetMajor(), minor and patch should be 0, pre-release cleared.
func TestVersion_SetMajor_ResetsLowerComponents(t *testing.T) {
	// Precondition: Version with all components set
	v, err := Parse("1.2.3-alpha")
	require.NoError(t, err)

	// Action: Set major version
	v.SetMajor(5)

	// Expected: Major updated, lower components reset
	assert.Equal(t, 5, v.Major())
	assert.Equal(t, 0, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestVersion_SetMinor_ResetsLowerComponents validates SemVer reset behavior.
//
// Why: Per SemVer, incrementing minor resets patch and pre-release.
//
// What: After SetMinor(), patch should be 0, pre-release cleared.
func TestVersion_SetMinor_ResetsLowerComponents(t *testing.T) {
	// Precondition: Version with all components set
	v, err := Parse("1.2.3-alpha")
	require.NoError(t, err)

	// Action: Set minor version
	v.SetMinor(7)

	// Expected: Minor updated, major preserved, lower components reset
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 7, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestVersion_SetPatch_ClearsPreRelease validates SemVer reset behavior.
//
// Why: Per SemVer, incrementing patch clears pre-release (becomes stable).
//
// What: After SetPatch(), pre-release should be cleared.
func TestVersion_SetPatch_ClearsPreRelease(t *testing.T) {
	// Precondition: Version with pre-release
	v, err := Parse("1.2.3-alpha")
	require.NoError(t, err)

	// Action: Set patch version
	v.SetPatch(9)

	// Expected: Patch updated, pre-release cleared
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 2, v.Minor())
	assert.Equal(t, 9, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestVersion_SetPreRelease_Manipulation validates pre-release modification.
//
// Why: Pre-release tags can be set or cleared during release workflows.
//
// What: SetPreRelease() should add/update/clear pre-release as specified.
func TestVersion_SetPreRelease_Manipulation(t *testing.T) {
	t.Run("set pre-release", func(t *testing.T) {
		// Precondition: Stable version
		v, err := Parse("1.0.0")
		require.NoError(t, err)

		// Action: Set pre-release
		v.SetPreRelease("beta.2")

		// Expected: Pre-release added
		assert.True(t, v.IsPreRelease())
		assert.Equal(t, "beta.2", v.PreReleaseString())
	})

	t.Run("clear pre-release", func(t *testing.T) {
		// Precondition: Pre-release version
		v, err := Parse("1.0.0-alpha")
		require.NoError(t, err)

		// Action: Clear pre-release with empty string
		v.SetPreRelease("")

		// Expected: Pre-release cleared
		assert.False(t, v.IsPreRelease())
	})
}

// TestVersion_SetBuildMetadata_Manipulation validates metadata modification.
//
// Why: Build metadata can be set or cleared during build/release processes.
//
// What: SetBuildMetadata() should add/update/clear metadata as specified.
func TestVersion_SetBuildMetadata_Manipulation(t *testing.T) {
	t.Run("set metadata", func(t *testing.T) {
		// Precondition: Version without metadata
		v, err := Parse("1.0.0")
		require.NoError(t, err)

		// Action: Set build metadata
		v.SetBuildMetadata("sha.abc123")

		// Expected: Metadata added
		assert.True(t, v.HasBuildMetadata())
		assert.Equal(t, "sha.abc123", v.BuildMetadataString())
	})

	t.Run("clear metadata", func(t *testing.T) {
		// Precondition: Version with metadata
		v, err := Parse("1.0.0+build")
		require.NoError(t, err)

		// Action: Clear metadata with empty string
		v.SetBuildMetadata("")

		// Expected: Metadata cleared
		assert.False(t, v.HasBuildMetadata())
	})
}

// TestVersion_SetPrefix_AddPrefix validates adding prefix to version.
//
// Why: Some workflows need to add/change prefix (e.g., for git tagging).
//
// What: SetPrefix() should update the prefix and affect FullString() output.
func TestVersion_SetPrefix_AddPrefix(t *testing.T) {
	// Precondition: Version without prefix
	v, err := Parse("1.0.0")
	require.NoError(t, err)

	// Action: Add prefix
	v.SetPrefix("v")

	// Expected: Prefix added, affects FullString()
	assert.True(t, v.HasPrefix())
	assert.Equal(t, "v", v.Prefix)
	assert.Equal(t, "v1.0.0", v.FullString())
}

// TestVersion_IncrementMajor_Behavior validates major increment on Version.
//
// Why: Direct increment methods are convenience for version bumping workflows.
//
// What: IncrementMajor() should bump major, reset minor/patch/pre-release.
func TestVersion_IncrementMajor_Behavior(t *testing.T) {
	// Precondition: Version with all components
	v, err := Parse("1.2.3-alpha")
	require.NoError(t, err)

	// Action: Increment major
	v.IncrementMajor()

	// Expected: Major bumped, lower components reset
	assert.Equal(t, 2, v.Major())
	assert.Equal(t, 0, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestVersion_IncrementMinor_Behavior validates minor increment on Version.
//
// Why: Minor bumps are common for feature releases.
//
// What: IncrementMinor() should bump minor, reset patch/pre-release.
func TestVersion_IncrementMinor_Behavior(t *testing.T) {
	t.Run("with existing minor", func(t *testing.T) {
		// Precondition: Version with minor set
		v, err := Parse("1.2.3-alpha")
		require.NoError(t, err)

		// Action: Increment minor
		v.IncrementMinor()

		// Expected: Minor bumped, patch reset, pre-release cleared
		assert.Equal(t, 1, v.Major())
		assert.Equal(t, 3, v.Minor())
		assert.Equal(t, 0, v.Patch())
		assert.False(t, v.IsPreRelease())
	})

	t.Run("with nil minor initializes to 1", func(t *testing.T) {
		// Precondition: Version with nil minor
		v := &Version{Core: &VersionCore{Major: 1}}

		// Action: Increment minor
		v.IncrementMinor()

		// Expected: Minor initialized to 1
		assert.Equal(t, 1, v.Minor())
	})
}

// TestVersion_IncrementPatch_Behavior validates patch increment on Version.
//
// Why: Patch bumps are common for bug fixes.
//
// What: IncrementPatch() should bump patch and clear pre-release.
func TestVersion_IncrementPatch_Behavior(t *testing.T) {
	t.Run("with existing patch", func(t *testing.T) {
		// Precondition: Version with patch set
		v, err := Parse("1.2.3-alpha")
		require.NoError(t, err)

		// Action: Increment patch
		v.IncrementPatch()

		// Expected: Patch bumped, pre-release cleared
		assert.Equal(t, 1, v.Major())
		assert.Equal(t, 2, v.Minor())
		assert.Equal(t, 4, v.Patch())
		assert.False(t, v.IsPreRelease())
	})

	t.Run("with nil patch initializes to 1", func(t *testing.T) {
		// Precondition: Version with nil patch
		minor := 2
		v := &Version{Core: &VersionCore{Major: 1, Minor: &minor}}

		// Action: Increment patch
		v.IncrementPatch()

		// Expected: Patch initialized to 1
		assert.Equal(t, 1, v.Patch())
	})
}

// TestBuilderAlias_WithPrefix validates the WithPrefix builder alias.
//
// Why: Builder provides multiple naming styles; aliases must work correctly.
//
// What: WithPrefix() should behave identically to Prefix().
func TestBuilderAlias_WithPrefix(t *testing.T) {
	// Precondition: New builder
	// Action: Use WithPrefix alias
	b := NewBuilder().WithPrefix("v")

	// Expected: Prefix set correctly
	assert.Equal(t, "v", b.GetPrefix())
}

// TestBuilderAlias_WithPreRelease validates the WithPreRelease builder alias.
//
// Why: Builder provides multiple naming styles for ergonomics.
//
// What: WithPreRelease() should behave identically to PreRelease().
func TestBuilderAlias_WithPreRelease(t *testing.T) {
	// Precondition: New builder
	// Action: Use WithPreRelease alias
	b := NewBuilder().WithPreRelease("alpha.1")

	// Expected: Pre-release set correctly
	assert.Equal(t, "alpha.1", b.GetPreRelease())
}

// TestBuilderAlias_WithBuildMetadata validates the WithBuildMetadata builder alias.
//
// Why: Builder provides multiple naming styles for ergonomics.
//
// What: WithBuildMetadata() should behave identically to BuildMetadata().
func TestBuilderAlias_WithBuildMetadata(t *testing.T) {
	// Precondition: New builder
	// Action: Use WithBuildMetadata alias
	b := NewBuilder().WithBuildMetadata("build.123")

	// Expected: Build metadata set correctly
	assert.Equal(t, "build.123", b.GetBuildMetadata())
}

// TestBuilderMethod_ClearMetadata validates clearing build metadata.
//
// Why: Workflows may need to strip metadata before release.
//
// What: ClearMetadata() should remove build metadata from builder.
func TestBuilderMethod_ClearMetadata(t *testing.T) {
	// Precondition: Builder with metadata set
	// Action: Clear metadata
	b := NewBuilder().BuildMetadata("build.123").ClearMetadata()

	// Expected: Metadata cleared
	assert.Empty(t, b.GetBuildMetadata())
}

// TestBuilderMethod_ClearBuildMetadata_Alias validates the alias.
//
// Why: Multiple naming styles for API ergonomics.
//
// What: ClearBuildMetadata() should behave identically to ClearMetadata().
func TestBuilderMethod_ClearBuildMetadata_Alias(t *testing.T) {
	// Precondition: Builder with metadata set
	// Action: Clear using alias
	b := NewBuilder().BuildMetadata("build.123").ClearBuildMetadata()

	// Expected: Metadata cleared
	assert.Empty(t, b.GetBuildMetadata())
}

// TestBuilderMethod_CoreString validates extraction of core version from builder.
//
// Why: Some outputs need only M.M.P without extras.
//
// What: CoreString() should return only Major.Minor.Patch.
func TestBuilderMethod_CoreString(t *testing.T) {
	// Precondition: Builder with all optional fields set
	b := NewBuilder().
		Prefix("v").
		Major(1).Minor(2).Patch(3).
		PreRelease("alpha").
		BuildMetadata("build")

	// Action: Get core string
	// Expected: Only M.M.P returned
	assert.Equal(t, "1.2.3", b.CoreString())
}

// TestBuilderMethod_DecrementMinor_Success validates minor decrement.
//
// Why: Some workflows may need to decrement versions (e.g., fixing mistakes).
//
// What: DecrementMinor() should reduce minor and reset patch.
func TestBuilderMethod_DecrementMinor_Success(t *testing.T) {
	// Precondition: Builder with minor > 0
	// Action: Decrement minor and build
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		DecrementMinor().
		Build()

	// Expected: Minor decremented, patch reset
	require.NoError(t, err)
	assert.Equal(t, 4, v.Minor())
	assert.Equal(t, 0, v.Patch())
}

// TestBuilderMethod_DecrementPatch_Success validates patch decrement.
//
// Why: Some workflows may need to decrement versions.
//
// What: DecrementPatch() should reduce patch.
func TestBuilderMethod_DecrementPatch_Success(t *testing.T) {
	// Precondition: Builder with patch > 0
	// Action: Decrement patch and build
	v, err := NewBuilder().
		Major(1).Minor(2).Patch(5).
		DecrementPatch().
		Build()

	// Expected: Patch decremented
	require.NoError(t, err)
	assert.Equal(t, 4, v.Patch())
}

// TestBuilderMethod_FromVersion_AssemblyRevision validates preserving revision.
//
// Why: When building from existing assembly version, revision must be preserved.
//
// What: FromVersion() should copy revision from assembly versions.
func TestBuilderMethod_FromVersion_AssemblyRevision(t *testing.T) {
	// Precondition: Parse assembly version
	v, err := Parse("1.2.3.4")
	require.NoError(t, err)

	// Action: Create builder from version
	b := FromVersion(v)

	// Expected: Revision preserved
	assert.NotNil(t, b.GetRevision())
	assert.Equal(t, 4, *b.GetRevision())
}

// =============================================================================
// BENCHMARKS
// Performance benchmarks for parsing various version formats.
// =============================================================================

// BenchmarkParse measures parsing performance across different version formats.
//
// Why: Parsing performance matters for tools processing many versions
// (e.g., scanning git tags, processing dependency trees).
//
// What: Benchmark parsing of simple to complex version strings.
func BenchmarkParse(b *testing.B) {
	inputs := []string{
		"1.2.3",
		"v1.2.3",
		"1.0.0-alpha.1+build.456",
		"v1.0.0-beta.2+exp.sha.5114f85",
	}

	for _, input := range inputs {
		b.Run(input, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Parse(input)
			}
		})
	}
}
