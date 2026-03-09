package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_SemVerCore(t *testing.T) {
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
			v, err := Parse(tt.input)
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

func TestParse_WithPrefix(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.prefix, v.Prefix)
			assert.Equal(t, tt.major, v.Major())
			assert.Equal(t, tt.minor, v.Minor())
			assert.Equal(t, tt.patch, v.Patch())
			assert.True(t, v.HasPrefix())
		})
	}
}

func TestParse_PreRelease(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.True(t, v.IsPreRelease())
			assert.Equal(t, tt.preRelease, v.PreReleaseString())
		})
	}
}

func TestParse_BuildMetadata(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.True(t, v.HasBuildMetadata())
			assert.Equal(t, tt.metadata, v.BuildMetadataString())
		})
	}
}

func TestParse_Combined(t *testing.T) {
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
			v, err := Parse(tt.input)
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

func TestParse_AssemblyVersion(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.True(t, v.IsAssemblyVersion())
			assert.Equal(t, tt.major, v.Major())
			assert.Equal(t, tt.minor, v.Minor())
			assert.Equal(t, tt.patch, v.Patch())
			assert.Equal(t, tt.revision, v.Revision())
		})
	}
}

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
			_, err := Parse(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), ErrLeadingZero)
		})
	}
}

func TestParse_Invalid_Empty(t *testing.T) {
	_, err := Parse("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrEmptyVersion)
}

func TestParse_Invalid_Whitespace(t *testing.T) {
	// Whitespace-only should fail
	_, err := Parse("   ")
	require.Error(t, err)
}

func TestParseLenient(t *testing.T) {
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
			v := ParseLenient(tt.input)
			assert.NotNil(t, v)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, v.String())
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.String())
		})
	}
}

func TestVersion_FullString(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.FullString())
		})
	}
}

func TestVersion_CoreVersion(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.CoreVersion())
		})
	}
}

func TestVersion_AssemblyVersionOutput(t *testing.T) {
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
			v, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.AssemblyVersion())
		})
	}
}

func TestVersion_Predicates(t *testing.T) {
	t.Run("HasPrefix", func(t *testing.T) {
		v1, _ := Parse("v1.0.0")
		v2, _ := Parse("1.0.0")
		assert.True(t, v1.HasPrefix())
		assert.False(t, v2.HasPrefix())
	})

	t.Run("IsPreRelease", func(t *testing.T) {
		v1, _ := Parse("1.0.0-alpha")
		v2, _ := Parse("1.0.0")
		assert.True(t, v1.IsPreRelease())
		assert.False(t, v2.IsPreRelease())
	})

	t.Run("HasBuildMetadata", func(t *testing.T) {
		v1, _ := Parse("1.0.0+build")
		v2, _ := Parse("1.0.0")
		assert.True(t, v1.HasBuildMetadata())
		assert.False(t, v2.HasBuildMetadata())
	})

	t.Run("IsAssemblyVersion", func(t *testing.T) {
		v1, _ := Parse("1.2.3.4")
		v2, _ := Parse("1.2.3")
		assert.True(t, v1.IsAssemblyVersion())
		assert.False(t, v2.IsAssemblyVersion())
	})

	t.Run("IsPartial", func(t *testing.T) {
		v1, _ := Parse("1")
		v2, _ := Parse("1.2")
		v3, _ := Parse("1.2.3")
		assert.True(t, v1.IsPartial())
		assert.True(t, v2.IsPartial())
		assert.False(t, v3.IsPartial())
	})
}

func TestEBNF(t *testing.T) {
	ebnf := EBNF()
	assert.NotEmpty(t, ebnf)
	assert.Contains(t, ebnf, "VersionFile")
}

// Benchmark parsing
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
