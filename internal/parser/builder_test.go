package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CORE FUNCTIONALITY
// =============================================================================
// Happy path tests that verify basic builder operations work correctly.

// TestBuilder_FromScratch_BuildsValidVersion validates that the builder can
// construct a semantic version from individual components.
//
// Why: This is the primary use case for the builder - creating versions
// programmatically without parsing a string. If this fails, the builder
// pattern is fundamentally broken.
//
// What: Creates a version 1.2.3 using the fluent builder API and verifies
// the resulting version string matches expectations.
func TestBuilder_FromScratch_BuildsValidVersion(t *testing.T) {
	// Precondition: Fresh builder instance
	builder := NewBuilder()

	// Action: Set version components and build
	v, err := builder.
		Major(1).
		Minor(2).
		Patch(3).
		Build()

	// Expected: Version 1.2.3 is created successfully
	require.NoError(t, err)
	assert.Equal(t, "1.2.3", v.String())
}

// TestBuilder_Full_BuildsCompleteVersion validates that all version components
// can be combined in a single builder chain.
//
// Why: Real-world versions often include prefix, pre-release, and build
// metadata. This test ensures all components work together without
// interference.
//
// What: Creates a fully-qualified version with prefix, major.minor.patch,
// pre-release tag, and build metadata.
func TestBuilder_Full_BuildsCompleteVersion(t *testing.T) {
	// Precondition: Fresh builder instance
	// Action: Chain all available setters
	v, err := NewBuilder().
		Prefix("v").
		Major(2).
		Minor(1).
		Patch(3).
		PreRelease("beta.2").
		BuildMetadata("sha.abc123").
		Build()

	// Expected: All components are present in output
	require.NoError(t, err)
	assert.Equal(t, "v2.1.3-beta.2+sha.abc123", v.FullString())
}

// TestBuilder_FromVersion_ClonesAndModifies validates that a builder can be
// created from an existing version for modification.
//
// Why: Version bumping workflows often start with the current version and
// apply increments. This must not mutate the original version (immutability).
//
// What: Parses a version, creates a builder from it, modifies the minor
// version and pre-release, then verifies both the new version and that
// the original remains unchanged.
func TestBuilder_FromVersion_ClonesAndModifies(t *testing.T) {
	// Precondition: Parse an existing version
	original, err := Parse("v1.2.3-alpha+build")
	require.NoError(t, err)

	// Action: Create builder from version and modify it
	modified, err := FromVersion(original).
		IncrementMinor().
		PreRelease("beta").
		Build()

	// Expected: New version reflects changes
	require.NoError(t, err)
	assert.Equal(t, "v1.3.0-beta+build", modified.FullString())

	// Expected: Original unchanged (immutability)
	assert.Equal(t, "v1.2.3-alpha+build", original.FullString())
}

// TestBuilder_FromString_ParsesAndModifies validates the convenience method
// for creating a builder directly from a version string.
//
// Why: Many callers have a string representation rather than a Version object.
// FromString provides a shorthand that combines Parse and FromVersion.
//
// What: Creates a builder from "1.2.3", increments patch, and verifies result.
func TestBuilder_FromString_ParsesAndModifies(t *testing.T) {
	// Precondition: Valid version string
	// Action: Create builder from string and modify
	v, err := FromString("1.2.3").
		IncrementPatch().
		Build()

	// Expected: Version incremented to 1.2.4
	require.NoError(t, err)
	assert.Equal(t, "1.2.4", v.String())
}

// =============================================================================
// KEY VARIATIONS
// =============================================================================
// Tests for important alternate flows and optional features.

// TestBuilder_WithPrefix_AddsVersionPrefix validates that version prefixes
// (commonly "v") are preserved in the output.
//
// Why: Many projects use "v" prefixed versions (e.g., v1.0.0). The builder
// must support adding and preserving these prefixes.
//
// What: Creates a version with "v" prefix and verifies both the Prefix field
// and FullString() output.
func TestBuilder_WithPrefix_AddsVersionPrefix(t *testing.T) {
	// Precondition: Fresh builder
	// Action: Set prefix and version components
	v, err := NewBuilder().
		Prefix("v").
		Major(1).
		Minor(0).
		Patch(0).
		Build()

	// Expected: Prefix stored and included in full string
	require.NoError(t, err)
	assert.Equal(t, "v", v.Prefix)
	assert.Equal(t, "v1.0.0", v.FullString())
}

// TestBuilder_WithPreRelease_SetsPreReleaseTag validates that pre-release
// identifiers are correctly attached to the version.
//
// Why: Pre-release versions (alpha, beta, rc) are critical for release
// management. They must be properly formatted and detectable.
//
// What: Creates a version with pre-release "alpha.1" and verifies the
// output format and IsPreRelease() detection.
func TestBuilder_WithPreRelease_SetsPreReleaseTag(t *testing.T) {
	// Precondition: Fresh builder
	// Action: Set pre-release identifier
	v, err := NewBuilder().
		Major(1).
		Minor(0).
		Patch(0).
		PreRelease("alpha.1").
		Build()

	// Expected: Pre-release in output and flag set
	require.NoError(t, err)
	assert.Equal(t, "1.0.0-alpha.1", v.String())
	assert.True(t, v.IsPreRelease())
}

// TestBuilder_WithBuildMetadata_SetsBuildInfo validates that build metadata
// is correctly attached to the version.
//
// Why: Build metadata (CI build numbers, git SHA) is used for traceability
// but does not affect version precedence per SemVer spec.
//
// What: Creates a version with build metadata and verifies format and detection.
func TestBuilder_WithBuildMetadata_SetsBuildInfo(t *testing.T) {
	// Precondition: Fresh builder
	// Action: Set build metadata
	v, err := NewBuilder().
		Major(1).
		Minor(0).
		Patch(0).
		BuildMetadata("build.123").
		Build()

	// Expected: Metadata in output and flag set
	require.NoError(t, err)
	assert.Equal(t, "1.0.0+build.123", v.String())
	assert.True(t, v.HasBuildMetadata())
}

// TestBuilder_IncrementMajor_ResetsLowerComponents validates that incrementing
// major version resets minor, patch, and clears pre-release per SemVer rules.
//
// Why: SemVer spec requires that when major is bumped, minor and patch reset
// to zero. Pre-release must also be cleared (moving to stable).
//
// What: Starts with 1.5.3-alpha, increments major, verifies 2.0.0 with no
// pre-release.
func TestBuilder_IncrementMajor_ResetsLowerComponents(t *testing.T) {
	// Precondition: Version with minor, patch, and pre-release set
	// Action: Increment major
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		PreRelease("alpha").
		IncrementMajor().
		Build()

	// Expected: Major incremented, minor/patch reset, pre-release cleared
	require.NoError(t, err)
	assert.Equal(t, 2, v.Major())
	assert.Equal(t, 0, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestBuilder_IncrementMinor_ResetsPatchAndClearsPreRelease validates that
// incrementing minor version resets patch and clears pre-release.
//
// Why: SemVer spec requires patch reset when minor is bumped. Pre-release
// is cleared because the version is now stable at the new minor.
//
// What: Starts with 1.5.3-alpha, increments minor, verifies 1.6.0 stable.
func TestBuilder_IncrementMinor_ResetsPatchAndClearsPreRelease(t *testing.T) {
	// Precondition: Version with patch and pre-release set
	// Action: Increment minor
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		PreRelease("alpha").
		IncrementMinor().
		Build()

	// Expected: Minor incremented, patch reset, pre-release cleared
	require.NoError(t, err)
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 6, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestBuilder_IncrementPatch_ClearsPreRelease validates that incrementing
// patch version clears the pre-release identifier.
//
// Why: After a patch bump, the version should be stable. Pre-release must
// be explicitly re-applied if needed.
//
// What: Starts with 1.5.3-alpha, increments patch, verifies 1.5.4 stable.
func TestBuilder_IncrementPatch_ClearsPreRelease(t *testing.T) {
	// Precondition: Version with pre-release set
	// Action: Increment patch
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		PreRelease("alpha").
		IncrementPatch().
		Build()

	// Expected: Patch incremented, pre-release cleared
	require.NoError(t, err)
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 5, v.Minor())
	assert.Equal(t, 4, v.Patch())
	assert.False(t, v.IsPreRelease())
}

// TestBuilder_DecrementMajor_ResetsLowerComponents validates that decrementing
// major version resets minor and patch to zero.
//
// Why: When rolling back major version, lower components should reset to
// avoid confusion about which minor/patch it represents.
//
// What: Starts with 2.5.3, decrements major, verifies 1.0.0.
func TestBuilder_DecrementMajor_ResetsLowerComponents(t *testing.T) {
	// Precondition: Version with minor and patch set
	// Action: Decrement major
	v, err := NewBuilder().
		Major(2).Minor(5).Patch(3).
		DecrementMajor().
		Build()

	// Expected: Major decremented, minor/patch reset
	require.NoError(t, err)
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 0, v.Minor())
	assert.Equal(t, 0, v.Patch())
}

// TestBuilder_ConveniencePreRelease_CreatesCorrectTags validates the shorthand
// methods for common pre-release types (Alpha, Beta, RC, Dev, Snapshot).
//
// Why: Most projects use standard pre-release naming. Convenience methods
// reduce boilerplate and ensure consistent formatting.
//
// What: Tests each convenience method with and without numeric suffix.
func TestBuilder_ConveniencePreRelease_CreatesCorrectTags(t *testing.T) {
	tests := []struct {
		name     string
		builder  *Builder
		expected string
	}{
		{"alpha", NewBuilder().Major(1).Alpha(), "1.0.0-alpha"},
		{"alpha.1", NewBuilder().Major(1).Alpha(1), "1.0.0-alpha.1"},
		{"beta", NewBuilder().Major(1).Beta(), "1.0.0-beta"},
		{"beta.2", NewBuilder().Major(1).Beta(2), "1.0.0-beta.2"},
		{"rc.1", NewBuilder().Major(1).RC(1), "1.0.0-rc.1"},
		{"dev", NewBuilder().Major(1).Dev(), "1.0.0-dev"},
		{"dev.5", NewBuilder().Major(1).Dev(5), "1.0.0-dev.5"},
		{"snapshot", NewBuilder().Major(1).Snapshot(), "1.0.0-SNAPSHOT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Builder with convenience method applied
			// Action: Build version
			v, err := tt.builder.Build()

			// Expected: Correct pre-release format
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.String())
		})
	}
}

// TestBuilder_AssemblyVersion_SupportsFourComponents validates that the builder
// supports .NET-style assembly versions with a fourth (revision) component.
//
// Why: .NET assemblies use 4-component versions (Major.Minor.Build.Revision).
// The builder must support this while still outputting valid SemVer.
//
// What: Creates 1.2.3.456 and verifies both SemVer and assembly outputs.
func TestBuilder_AssemblyVersion_SupportsFourComponents(t *testing.T) {
	// Precondition: Fresh builder
	// Action: Set all four components including revision
	v, err := NewBuilder().
		Major(1).Minor(2).Patch(3).Revision(456).
		Build()

	// Expected: IsAssemblyVersion true, both formats correct
	require.NoError(t, err)
	assert.True(t, v.IsAssemblyVersion())
	assert.Equal(t, "1.2.3", v.String())              // SemVer format (3 components)
	assert.Equal(t, "1.2.3.456", v.AssemblyVersion()) // Assembly format (4 components)
}

// TestBuilder_Release_ClearsPreReleaseAndMetadata validates that the Release()
// method removes both pre-release and build metadata.
//
// Why: When promoting a release candidate to stable, both pre-release tag
// and build metadata should be stripped for a clean release version.
//
// What: Starts with 1.0.0-rc.1+build.123, calls Release(), verifies 1.0.0.
func TestBuilder_Release_ClearsPreReleaseAndMetadata(t *testing.T) {
	// Precondition: Version with both pre-release and metadata
	// Action: Call Release() to promote to stable
	v, err := NewBuilder().
		Major(1).Minor(0).Patch(0).
		PreRelease("rc.1").
		BuildMetadata("build.123").
		Release().
		Build()

	// Expected: Clean release version with no tags
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.String())
	assert.False(t, v.IsPreRelease())
	assert.False(t, v.HasBuildMetadata())
}

// TestBuilder_ClearRevision_RemovesFourthComponent validates that ClearRevision()
// removes the assembly version fourth component.
//
// Why: When converting from assembly version back to pure SemVer, the
// revision must be removable.
//
// What: Creates 1.2.3.456, clears revision, verifies 1.2.3 only.
func TestBuilder_ClearRevision_RemovesFourthComponent(t *testing.T) {
	// Precondition: Assembly version with revision
	// Action: Clear the revision
	v, err := NewBuilder().
		Major(1).Minor(2).Patch(3).Revision(456).
		ClearRevision().
		Build()

	// Expected: No longer assembly version, just 1.2.3
	require.NoError(t, err)
	assert.False(t, v.IsAssemblyVersion())
	assert.Equal(t, "1.2.3", v.String())
}

// TestBuilder_ClearPreRelease_RemovesPreReleaseTag validates that ClearPreRelease()
// removes the pre-release identifier while keeping other components.
//
// Why: Workflows may need to remove just the pre-release without affecting
// build metadata or triggering a full Release().
//
// What: Creates 1.0.0-alpha, clears pre-release, verifies 1.0.0.
func TestBuilder_ClearPreRelease_RemovesPreReleaseTag(t *testing.T) {
	// Precondition: Version with pre-release
	// Action: Clear just the pre-release
	v, err := NewBuilder().
		Major(1).PreRelease("alpha").
		ClearPreRelease().
		Build()

	// Expected: Pre-release removed, version stable
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.String())
	assert.False(t, v.IsPreRelease())
}

// TestBuilder_NoPrefix_RemovesExistingPrefix validates that NoPrefix() can
// strip a prefix from a version.
//
// Why: Some workflows require removing the "v" prefix (e.g., for npm packages
// that don't use prefixes).
//
// What: Starts with v1.2.3, calls NoPrefix(), verifies 1.2.3 in FullString().
func TestBuilder_NoPrefix_RemovesExistingPrefix(t *testing.T) {
	// Precondition: Version with prefix
	// Action: Remove the prefix
	v, err := FromString("v1.2.3").
		NoPrefix().
		Build()

	// Expected: Prefix empty, FullString without prefix
	require.NoError(t, err)
	assert.Equal(t, "", v.Prefix)
	assert.Equal(t, "1.2.3", v.FullString())
}

// TestBuilder_Chaining_SupportsComplexWorkflows validates that multiple
// builder operations can be chained together correctly.
//
// Why: Real workflows may involve multiple increments and modifications
// in a single chain. The builder must handle complex chains correctly.
//
// What: Chains multiple increments and modifications, verifies final result.
func TestBuilder_Chaining_SupportsComplexWorkflows(t *testing.T) {
	// Precondition: Starting version string
	// Action: Apply complex chain of operations
	v, err := FromString("v1.0.0").
		IncrementMinor().
		IncrementMinor().
		IncrementPatch().
		Alpha(1).
		Metadata("ci.123").
		Build()

	// Expected: All operations applied correctly
	require.NoError(t, err)
	assert.Equal(t, "v1.2.1-alpha.1+ci.123", v.FullString())
}

// =============================================================================
// ERROR HANDLING
// =============================================================================
// Tests for expected failure modes and error conditions.

// TestBuilder_FromString_InvalidString_ReturnsError validates that FromString
// with an invalid version string produces an error at Build time.
//
// Why: Invalid input must be rejected gracefully. The error should surface
// when Build() is called (deferred error pattern).
//
// What: Attempts to build from "not-a-version" and expects an error.
func TestBuilder_FromString_InvalidString_ReturnsError(t *testing.T) {
	// Precondition: Invalid version string
	// Action: Attempt to build
	_, err := FromString("not-a-version").Build()

	// Expected: Error returned
	require.Error(t, err)
}

// TestBuilder_DecrementMajor_BelowZero_ReturnsError validates that decrementing
// major below zero produces an error.
//
// Why: SemVer versions cannot have negative components. This boundary must
// be enforced.
//
// What: Starts with major=0, attempts decrement, expects specific error.
func TestBuilder_DecrementMajor_BelowZero_ReturnsError(t *testing.T) {
	// Precondition: Version at major=0
	// Action: Attempt to decrement major
	_, err := NewBuilder().
		Major(0).
		DecrementMajor().
		Build()

	// Expected: Error about decrementing below zero
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decrement major below 0")
}

// TestBuilder_DecrementMinor_BelowZero_ReturnsError validates that decrementing
// minor below zero produces an error.
//
// Why: SemVer versions cannot have negative components. This boundary must
// be enforced independently for each component.
//
// What: Starts with minor=0, attempts decrement, expects specific error.
func TestBuilder_DecrementMinor_BelowZero_ReturnsError(t *testing.T) {
	// Precondition: Version at minor=0
	// Action: Attempt to decrement minor
	_, err := NewBuilder().
		Major(1).Minor(0).
		DecrementMinor().
		Build()

	// Expected: Error about decrementing below zero
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decrement minor below 0")
}

// TestBuilder_DecrementPatch_BelowZero_ReturnsError validates that decrementing
// patch below zero produces an error.
//
// Why: SemVer versions cannot have negative components. Patch is the most
// commonly decremented, so this boundary is important.
//
// What: Starts with patch=0, attempts decrement, expects specific error.
func TestBuilder_DecrementPatch_BelowZero_ReturnsError(t *testing.T) {
	// Precondition: Version at patch=0
	// Action: Attempt to decrement patch
	_, err := NewBuilder().
		Major(1).Minor(2).Patch(0).
		DecrementPatch().
		Build()

	// Expected: Error about decrementing below zero
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decrement patch below 0")
}

// TestBuilder_NegativeVersion_ReturnsError validates that setting negative
// values for any version component produces an error.
//
// Why: SemVer prohibits negative version numbers. This must be caught
// at build time with clear error messages.
//
// What: Tests negative major, minor, and patch separately.
func TestBuilder_NegativeVersion_ReturnsError(t *testing.T) {
	t.Run("negative major", func(t *testing.T) {
		// Precondition: Set negative major
		// Action: Build
		_, err := NewBuilder().Major(-1).Build()

		// Expected: Error returned
		require.Error(t, err)
	})

	t.Run("negative minor", func(t *testing.T) {
		// Precondition: Set negative minor
		// Action: Build
		_, err := NewBuilder().Major(1).Minor(-1).Build()

		// Expected: Error returned
		require.Error(t, err)
	})

	t.Run("negative patch", func(t *testing.T) {
		// Precondition: Set negative patch
		// Action: Build
		_, err := NewBuilder().Major(1).Minor(0).Patch(-1).Build()

		// Expected: Error returned
		require.Error(t, err)
	})
}

// TestBuilder_ErrorState_DetectableViaHasError validates that error state
// can be checked before calling Build().
//
// Why: The deferred error pattern allows callers to detect errors early
// using HasError() and Error() without calling Build().
//
// What: Creates invalid builder, checks HasError() returns true.
func TestBuilder_ErrorState_DetectableViaHasError(t *testing.T) {
	// Precondition: Builder with invalid input
	b := NewBuilder().Major(-1)

	// Action: Check error state
	// Expected: HasError true, Error non-nil
	assert.True(t, b.HasError())
	assert.Error(t, b.Error())
}

// TestBuilder_MustBuild_Panic_OnInvalidInput validates that MustBuild()
// panics when the builder contains errors.
//
// Why: MustBuild() is a convenience for cases where errors are "impossible".
// It must panic loudly rather than silently returning invalid data.
//
// What: Creates invalid builder, calls MustBuild(), expects panic.
func TestBuilder_MustBuild_Panic_OnInvalidInput(t *testing.T) {
	// Precondition: Builder with invalid input
	// Action: Call MustBuild
	// Expected: Panic occurs
	assert.Panics(t, func() {
		NewBuilder().Major(-1).MustBuild()
	})
}

// TestBuilder_RoundTrip_RejectsInvalid validates that Build() correctly
// rejects various invalid configurations via round-trip validation.
//
// Why: Build() parses and validates the constructed string. This ensures
// that invalid configurations are caught even if individual setters don't
// validate immediately.
//
// What: Tests multiple invalid configurations and their error messages.
func TestBuilder_RoundTrip_RejectsInvalid(t *testing.T) {
	tests := []struct {
		name        string
		builder     *Builder
		errContains string
	}{
		{
			name:        "negative major",
			builder:     NewBuilder().Major(-1),
			errContains: "negative",
		},
		{
			name:        "negative minor",
			builder:     NewBuilder().Major(1).Minor(-1),
			errContains: "negative",
		},
		{
			name:        "negative patch",
			builder:     NewBuilder().Major(1).Patch(-1),
			errContains: "negative",
		},
		{
			name:        "decrement below zero",
			builder:     NewBuilder().Major(0).DecrementMajor(),
			errContains: "cannot decrement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Invalid builder configuration
			// Action: Attempt to build
			_, err := tt.builder.Build()

			// Expected: Error containing expected message
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// =============================================================================
// EDGE CASES
// =============================================================================
// Boundary conditions and unusual but valid scenarios.

// TestBuilder_FromNilVersion_StartsFromZero validates that FromVersion(nil)
// creates a builder initialized to 0.0.0.
//
// Why: Passing nil should be handled gracefully, not cause a panic. Starting
// from zero is a sensible default.
//
// What: Calls FromVersion(nil), sets major to 1, verifies 1.0.0.
func TestBuilder_FromNilVersion_StartsFromZero(t *testing.T) {
	// Precondition: nil version input
	// Action: Build from nil
	v, err := FromVersion(nil).
		Major(1).
		Build()

	// Expected: Valid version starting from defaults
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.String())
}

// TestBuilder_String_WorksWithoutBuild validates that String() can be called
// on a builder without calling Build() first.
//
// Why: String() is useful for debugging and logging the current builder state
// before the final Build() call.
//
// What: Creates builder, calls String() without Build(), verifies output.
func TestBuilder_String_WorksWithoutBuild(t *testing.T) {
	// Precondition: Builder with all components set
	b := NewBuilder().
		Prefix("v").
		Major(1).Minor(2).Patch(3).
		PreRelease("alpha").
		BuildMetadata("build")

	// Action: Call String() without Build()
	// Expected: Full version string returned
	assert.Equal(t, "v1.2.3-alpha+build", b.String())
}

// TestBuilder_MustBuild_ReturnsVersion validates that MustBuild() returns
// a valid version when no errors are present.
//
// Why: MustBuild() is the panic-on-error variant of Build(). When valid,
// it should return the version directly without error handling.
//
// What: Creates valid builder, calls MustBuild(), verifies version.
func TestBuilder_MustBuild_ReturnsVersion(t *testing.T) {
	// Precondition: Valid builder
	// Action: Call MustBuild
	v := NewBuilder().Major(1).Minor(2).Patch(3).MustBuild()

	// Expected: Valid version returned
	assert.Equal(t, "1.2.3", v.String())
}

// TestBuilder_Getters_ReturnCorrectValues validates that all getter methods
// return the current builder state correctly.
//
// Why: Getters allow inspection of builder state before Build(). They must
// reflect the current state accurately.
//
// What: Sets all components, verifies each getter returns expected value.
func TestBuilder_Getters_ReturnCorrectValues(t *testing.T) {
	// Precondition: Builder with all components set
	b := NewBuilder().
		Prefix("v").
		Major(1).Minor(2).Patch(3).
		PreRelease("alpha").
		BuildMetadata("build")

	// Action/Expected: Each getter returns correct value
	assert.Equal(t, "v", b.GetPrefix())
	assert.Equal(t, 1, b.GetMajor())
	assert.Equal(t, 2, b.GetMinor())
	assert.Equal(t, 3, b.GetPatch())
	assert.Nil(t, b.GetRevision())
	assert.Equal(t, "alpha", b.GetPreRelease())
	assert.Equal(t, "build", b.GetBuildMetadata())
	assert.False(t, b.HasError())
	assert.NoError(t, b.Error())
}

// TestBuilder_RoundTrip_ParsesCorrectly validates that built versions can be
// successfully re-parsed (round-trip validation).
//
// Why: Build() internally parses the constructed string. This ensures the
// builder produces valid, parseable output in all configurations.
//
// What: Tests various version configurations for round-trip correctness.
func TestBuilder_RoundTrip_ParsesCorrectly(t *testing.T) {
	tests := []struct {
		name       string
		builder    *Builder
		wantFull   string // FullString() output (SemVer format with prefix)
		wantAssem  string // AssemblyVersion() output (4 components)
		isAssembly bool
	}{
		{
			name:      "simple version",
			builder:   NewBuilder().Major(1).Minor(2).Patch(3),
			wantFull:  "1.2.3",
			wantAssem: "1.2.3.0",
		},
		{
			name:      "with prefix",
			builder:   NewBuilder().Prefix("v").Major(1).Minor(0).Patch(0),
			wantFull:  "v1.0.0",
			wantAssem: "1.0.0.0",
		},
		{
			name:      "with pre-release",
			builder:   NewBuilder().Major(1).PreRelease("alpha.1"),
			wantFull:  "1.0.0-alpha.1",
			wantAssem: "1.0.0.0",
		},
		{
			name:      "with metadata",
			builder:   NewBuilder().Major(1).BuildMetadata("build.123"),
			wantFull:  "1.0.0+build.123",
			wantAssem: "1.0.0.0",
		},
		{
			name:      "full version",
			builder:   NewBuilder().Prefix("v").Major(2).Minor(1).Patch(3).PreRelease("rc.1").BuildMetadata("sha.abc"),
			wantFull:  "v2.1.3-rc.1+sha.abc",
			wantAssem: "2.1.3.0",
		},
		{
			name:       "assembly version",
			builder:    NewBuilder().Major(1).Minor(2).Patch(3).Revision(456),
			wantFull:   "1.2.3", // SemVer format (no revision)
			wantAssem:  "1.2.3.456",
			isAssembly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Builder configuration
			// Action: Build and verify round-trip
			v, err := tt.builder.Build()
			require.NoError(t, err, "Build() should succeed")

			// Expected: Output matches expectations
			assert.Equal(t, tt.wantFull, v.FullString(), "FullString() mismatch")
			assert.Equal(t, tt.wantAssem, v.AssemblyVersion(), "AssemblyVersion() mismatch")

			if tt.isAssembly {
				assert.True(t, v.IsAssemblyVersion(), "Should be assembly version")
			}

			// Expected: Can re-parse the output (double round-trip)
			v2, err := Parse(v.FullString())
			require.NoError(t, err, "Should be able to re-parse built version")
			assert.Equal(t, v.FullString(), v2.FullString())
		})
	}
}

// TestBuilder_RoundTrip_PreservesComponents validates that all version
// components are preserved through the build process and subsequent
// modifications don't affect the original.
//
// Why: Immutability is critical for version operations. Building and
// modifying must not corrupt the original data.
//
// What: Builds a complex version, modifies via new builder, verifies both.
func TestBuilder_RoundTrip_PreservesComponents(t *testing.T) {
	// Precondition: Build a complex version
	original, err := NewBuilder().
		Prefix("v").
		Major(2).
		Minor(5).
		Patch(3).
		PreRelease("beta.2").
		BuildMetadata("ci.456").
		Build()
	require.NoError(t, err)

	// Action/Expected: All components preserved after build
	assert.Equal(t, "v", original.Prefix)
	assert.Equal(t, 2, original.Major())
	assert.Equal(t, 5, original.Minor())
	assert.Equal(t, 3, original.Patch())
	assert.Equal(t, "beta.2", original.PreReleaseString())
	assert.Equal(t, "ci.456", original.BuildMetadataString())

	// Action: Create new builder and modify
	modified, err := FromVersion(original).
		IncrementMinor().
		RC(1).
		Build()
	require.NoError(t, err)

	// Expected: Modification worked correctly
	assert.Equal(t, "v", modified.Prefix)                     // Preserved
	assert.Equal(t, 2, modified.Major())                      // Preserved
	assert.Equal(t, 6, modified.Minor())                      // Incremented
	assert.Equal(t, 0, modified.Patch())                      // Reset
	assert.Equal(t, "rc.1", modified.PreReleaseString())      // New pre-release
	assert.Equal(t, "ci.456", modified.BuildMetadataString()) // Preserved

	// Expected: Original unchanged (immutability)
	assert.Equal(t, 5, original.Minor())
}

// =============================================================================
// MINUTIAE
// =============================================================================
// Obscure scenarios, performance tests, and implementation details.

// BenchmarkBuilder_Simple measures the performance of basic builder operations.
//
// Why: Version building may be called frequently in CI pipelines. Tracking
// baseline performance helps detect regressions.
//
// What: Benchmarks creating a simple 1.2.3 version.
func BenchmarkBuilder_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewBuilder().Major(1).Minor(2).Patch(3).Build()
	}
}

// BenchmarkBuilder_FromVersion measures the performance of creating a builder
// from an existing version and modifying it.
//
// Why: FromVersion with modifications is a common pattern in bump operations.
// Performance matters for batch processing.
//
// What: Benchmarks parsing, cloning, and incrementing a version.
func BenchmarkBuilder_FromVersion(b *testing.B) {
	v, _ := Parse("v1.2.3-alpha+build")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FromVersion(v).IncrementPatch().Build()
	}
}
