package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_FromScratch(t *testing.T) {
	v, err := NewBuilder().
		Major(1).
		Minor(2).
		Patch(3).
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.2.3", v.String())
}

func TestBuilder_WithPrefix(t *testing.T) {
	v, err := NewBuilder().
		Prefix("v").
		Major(1).
		Minor(0).
		Patch(0).
		Build()

	require.NoError(t, err)
	assert.Equal(t, "v", v.Prefix)
	assert.Equal(t, "v1.0.0", v.FullString())
}

func TestBuilder_WithPreRelease(t *testing.T) {
	v, err := NewBuilder().
		Major(1).
		Minor(0).
		Patch(0).
		PreRelease("alpha.1").
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.0.0-alpha.1", v.String())
	assert.True(t, v.IsPreRelease())
}

func TestBuilder_WithBuildMetadata(t *testing.T) {
	v, err := NewBuilder().
		Major(1).
		Minor(0).
		Patch(0).
		BuildMetadata("build.123").
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.0.0+build.123", v.String())
	assert.True(t, v.HasBuildMetadata())
}

func TestBuilder_Full(t *testing.T) {
	v, err := NewBuilder().
		Prefix("v").
		Major(2).
		Minor(1).
		Patch(3).
		PreRelease("beta.2").
		BuildMetadata("sha.abc123").
		Build()

	require.NoError(t, err)
	assert.Equal(t, "v2.1.3-beta.2+sha.abc123", v.FullString())
}

func TestBuilder_FromVersion(t *testing.T) {
	original, err := Parse("v1.2.3-alpha+build")
	require.NoError(t, err)

	modified, err := FromVersion(original).
		IncrementMinor().
		PreRelease("beta").
		Build()

	require.NoError(t, err)
	assert.Equal(t, "v1.3.0-beta+build", modified.FullString())

	// Original unchanged
	assert.Equal(t, "v1.2.3-alpha+build", original.FullString())
}

func TestBuilder_FromString(t *testing.T) {
	v, err := FromString("1.2.3").
		IncrementPatch().
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.2.4", v.String())
}

func TestBuilder_FromString_Invalid(t *testing.T) {
	_, err := FromString("not-a-version").Build()
	require.Error(t, err)
}

func TestBuilder_IncrementMajor(t *testing.T) {
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		PreRelease("alpha").
		IncrementMajor().
		Build()

	require.NoError(t, err)
	assert.Equal(t, 2, v.Major())
	assert.Equal(t, 0, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease()) // Pre-release cleared
}

func TestBuilder_IncrementMinor(t *testing.T) {
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		PreRelease("alpha").
		IncrementMinor().
		Build()

	require.NoError(t, err)
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 6, v.Minor())
	assert.Equal(t, 0, v.Patch())
	assert.False(t, v.IsPreRelease())
}

func TestBuilder_IncrementPatch(t *testing.T) {
	v, err := NewBuilder().
		Major(1).Minor(5).Patch(3).
		PreRelease("alpha").
		IncrementPatch().
		Build()

	require.NoError(t, err)
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 5, v.Minor())
	assert.Equal(t, 4, v.Patch())
	assert.False(t, v.IsPreRelease())
}

func TestBuilder_DecrementMajor(t *testing.T) {
	v, err := NewBuilder().
		Major(2).Minor(5).Patch(3).
		DecrementMajor().
		Build()

	require.NoError(t, err)
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 0, v.Minor())
	assert.Equal(t, 0, v.Patch())
}

func TestBuilder_DecrementMajor_BelowZero(t *testing.T) {
	_, err := NewBuilder().
		Major(0).
		DecrementMajor().
		Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decrement major below 0")
}

func TestBuilder_DecrementMinor_BelowZero(t *testing.T) {
	_, err := NewBuilder().
		Major(1).Minor(0).
		DecrementMinor().
		Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decrement minor below 0")
}

func TestBuilder_DecrementPatch_BelowZero(t *testing.T) {
	_, err := NewBuilder().
		Major(1).Minor(2).Patch(0).
		DecrementPatch().
		Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decrement patch below 0")
}

func TestBuilder_NegativeVersion(t *testing.T) {
	t.Run("negative major", func(t *testing.T) {
		_, err := NewBuilder().Major(-1).Build()
		require.Error(t, err)
	})

	t.Run("negative minor", func(t *testing.T) {
		_, err := NewBuilder().Major(1).Minor(-1).Build()
		require.Error(t, err)
	})

	t.Run("negative patch", func(t *testing.T) {
		_, err := NewBuilder().Major(1).Minor(0).Patch(-1).Build()
		require.Error(t, err)
	})
}

func TestBuilder_Release(t *testing.T) {
	v, err := NewBuilder().
		Major(1).Minor(0).Patch(0).
		PreRelease("rc.1").
		BuildMetadata("build.123").
		Release().
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.String())
	assert.False(t, v.IsPreRelease())
	assert.False(t, v.HasBuildMetadata())
}

func TestBuilder_ConveniencePreRelease(t *testing.T) {
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
			v, err := tt.builder.Build()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.String())
		})
	}
}

func TestBuilder_AssemblyVersion(t *testing.T) {
	v, err := NewBuilder().
		Major(1).Minor(2).Patch(3).Revision(456).
		Build()

	require.NoError(t, err)
	assert.True(t, v.IsAssemblyVersion())
	assert.Equal(t, "1.2.3", v.String())           // SemVer format (3 components)
	assert.Equal(t, "1.2.3.456", v.AssemblyVersion()) // Assembly format (4 components)
}

func TestBuilder_ClearRevision(t *testing.T) {
	v, err := NewBuilder().
		Major(1).Minor(2).Patch(3).Revision(456).
		ClearRevision().
		Build()

	require.NoError(t, err)
	assert.False(t, v.IsAssemblyVersion())
	assert.Equal(t, "1.2.3", v.String())
}

func TestBuilder_ClearPreRelease(t *testing.T) {
	v, err := NewBuilder().
		Major(1).PreRelease("alpha").
		ClearPreRelease().
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.String())
	assert.False(t, v.IsPreRelease())
}

func TestBuilder_NoPrefix(t *testing.T) {
	v, err := FromString("v1.2.3").
		NoPrefix().
		Build()

	require.NoError(t, err)
	assert.Equal(t, "", v.Prefix)
	assert.Equal(t, "1.2.3", v.FullString())
}

func TestBuilder_String(t *testing.T) {
	b := NewBuilder().
		Prefix("v").
		Major(1).Minor(2).Patch(3).
		PreRelease("alpha").
		BuildMetadata("build")

	// String() works without Build()
	assert.Equal(t, "v1.2.3-alpha+build", b.String())
}

func TestBuilder_MustBuild(t *testing.T) {
	v := NewBuilder().Major(1).Minor(2).Patch(3).MustBuild()
	assert.Equal(t, "1.2.3", v.String())
}

func TestBuilder_MustBuild_Panic(t *testing.T) {
	assert.Panics(t, func() {
		NewBuilder().Major(-1).MustBuild()
	})
}

func TestBuilder_Getters(t *testing.T) {
	b := NewBuilder().
		Prefix("v").
		Major(1).Minor(2).Patch(3).
		PreRelease("alpha").
		BuildMetadata("build")

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

func TestBuilder_ErrorState(t *testing.T) {
	b := NewBuilder().Major(-1)
	assert.True(t, b.HasError())
	assert.Error(t, b.Error())
}

func TestBuilder_Chaining(t *testing.T) {
	// Complex chaining scenario
	v, err := FromString("v1.0.0").
		IncrementMinor().
		IncrementMinor().
		IncrementPatch().
		Alpha(1).
		Metadata("ci.123").
		Build()

	require.NoError(t, err)
	assert.Equal(t, "v1.2.1-alpha.1+ci.123", v.FullString())
}

func TestBuilder_FromNilVersion(t *testing.T) {
	v, err := FromVersion(nil).
		Major(1).
		Build()

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.String())
}

// --- Round-trip Validation Tests ---
// These verify that Build() parses and validates the constructed string

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
			// Build and verify round-trip
			v, err := tt.builder.Build()
			require.NoError(t, err, "Build() should succeed")

			// Verify the built version matches expected
			assert.Equal(t, tt.wantFull, v.FullString(), "FullString() mismatch")
			assert.Equal(t, tt.wantAssem, v.AssemblyVersion(), "AssemblyVersion() mismatch")

			if tt.isAssembly {
				assert.True(t, v.IsAssemblyVersion(), "Should be assembly version")
			}

			// Verify we can parse the output again (double round-trip)
			v2, err := Parse(v.FullString())
			require.NoError(t, err, "Should be able to re-parse built version")
			assert.Equal(t, v.FullString(), v2.FullString())
		})
	}
}

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
			_, err := tt.builder.Build()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestBuilder_RoundTrip_PreservesComponents(t *testing.T) {
	// Build a complex version
	original, err := NewBuilder().
		Prefix("v").
		Major(2).
		Minor(5).
		Patch(3).
		PreRelease("beta.2").
		BuildMetadata("ci.456").
		Build()
	require.NoError(t, err)

	// Verify all components are preserved after round-trip
	assert.Equal(t, "v", original.Prefix)
	assert.Equal(t, 2, original.Major())
	assert.Equal(t, 5, original.Minor())
	assert.Equal(t, 3, original.Patch())
	assert.Equal(t, "beta.2", original.PreReleaseString())
	assert.Equal(t, "ci.456", original.BuildMetadataString())

	// Create new builder from this version and modify
	modified, err := FromVersion(original).
		IncrementMinor().
		RC(1).
		Build()
	require.NoError(t, err)

	// Verify modification worked correctly
	assert.Equal(t, "v", modified.Prefix)       // Preserved
	assert.Equal(t, 2, modified.Major())        // Preserved
	assert.Equal(t, 6, modified.Minor())        // Incremented
	assert.Equal(t, 0, modified.Patch())        // Reset
	assert.Equal(t, "rc.1", modified.PreReleaseString()) // New pre-release
	assert.Equal(t, "ci.456", modified.BuildMetadataString()) // Preserved

	// Original unchanged
	assert.Equal(t, 5, original.Minor())
}

// Benchmark builder operations
func BenchmarkBuilder_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewBuilder().Major(1).Minor(2).Patch(3).Build()
	}
}

func BenchmarkBuilder_FromVersion(b *testing.B) {
	v, _ := Parse("v1.2.3-alpha+build")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FromVersion(v).IncrementPatch().Build()
	}
}
