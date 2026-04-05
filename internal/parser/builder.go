package parser

import (
	"fmt"
	"strings"
)

// Builder provides a fluent API for constructing and mutating versions.
// It can be seeded from a parsed Version or created from scratch.
//
// Example usage:
//
//	// From scratch
//	v, err := NewBuilder().
//	    Major(1).
//	    Minor(2).
//	    Patch(3).
//	    PreRelease("alpha.1").
//	    Build()
//
//	// From existing version
//	parsed, _ := Parse("v1.2.3-alpha")
//	v, err := FromVersion(parsed).
//	    IncrementMinor().
//	    PreRelease("beta.1").
//	    Build()
type Builder struct {
	prefix        string
	major         int
	minor         int
	patch         int
	revision      *int // nil = not set (3-component), non-nil = 4-component assembly
	preRelease    string
	buildMetadata string
	err           error // captures first error for deferred checking
}

// NewBuilder creates a new version builder starting at 0.0.0.
func NewBuilder() *Builder {
	return &Builder{}
}

// FromVersion creates a builder seeded with values from an existing Version.
// The builder is a copy - modifications don't affect the original.
func FromVersion(v *Version) *Builder {
	if v == nil {
		return NewBuilder()
	}

	b := &Builder{
		prefix:        v.Prefix,
		major:         v.Major(),
		minor:         v.Minor(),
		patch:         v.Patch(),
		preRelease:    v.PreReleaseString(),
		buildMetadata: v.BuildMetadataString(),
	}

	// Preserve revision if it's an assembly version
	if v.IsAssemblyVersion() {
		rev := v.Revision()
		b.revision = &rev
	}

	return b
}

// FromString parses a version string and creates a builder from it.
// If parsing fails, Build() will return the parse error.
func FromString(s string) *Builder {
	v, err := Parse(s)
	if err != nil {
		return &Builder{err: err}
	}
	return FromVersion(v)
}

// --- Setters (return *Builder for chaining) ---

// Prefix sets the version prefix (e.g., "v", "V").
func (b *Builder) Prefix(p string) *Builder {
	b.prefix = p
	return b
}

// WithPrefix is an alias for Prefix.
func (b *Builder) WithPrefix(p string) *Builder {
	return b.Prefix(p)
}

// NoPrefix removes the version prefix.
func (b *Builder) NoPrefix() *Builder {
	b.prefix = ""
	return b
}

// Major sets the major version number.
func (b *Builder) Major(n int) *Builder {
	if n < 0 {
		b.err = fmt.Errorf("%s: major=%d", ErrNegativeVersion, n)
		return b
	}
	b.major = n
	return b
}

// Minor sets the minor version number.
func (b *Builder) Minor(n int) *Builder {
	if n < 0 {
		b.err = fmt.Errorf("%s: minor=%d", ErrNegativeVersion, n)
		return b
	}
	b.minor = n
	return b
}

// Patch sets the patch version number.
func (b *Builder) Patch(n int) *Builder {
	if n < 0 {
		b.err = fmt.Errorf("%s: patch=%d", ErrNegativeVersion, n)
		return b
	}
	b.patch = n
	return b
}

// Revision sets the revision number (4th component for assembly versions).
// Use nil or call ClearRevision() to remove the revision component.
func (b *Builder) Revision(n int) *Builder {
	if n < 0 {
		b.err = fmt.Errorf("%s: revision=%d", ErrNegativeVersion, n)
		return b
	}
	b.revision = &n
	return b
}

// ClearRevision removes the revision component, making this a 3-component version.
func (b *Builder) ClearRevision() *Builder {
	b.revision = nil
	return b
}

// PreRelease sets the pre-release identifier (e.g., "alpha.1", "beta", "rc.2").
// Pass empty string to clear the pre-release.
func (b *Builder) PreRelease(pr string) *Builder {
	b.preRelease = pr
	return b
}

// WithPreRelease is an alias for PreRelease.
func (b *Builder) WithPreRelease(pr string) *Builder {
	return b.PreRelease(pr)
}

// ClearPreRelease removes the pre-release identifier.
func (b *Builder) ClearPreRelease() *Builder {
	b.preRelease = ""
	return b
}

// BuildMetadata sets the build metadata (e.g., "build.123", "sha.abc123").
// Pass empty string to clear the metadata.
func (b *Builder) BuildMetadata(m string) *Builder {
	b.buildMetadata = m
	return b
}

// WithBuildMetadata is an alias for BuildMetadata.
func (b *Builder) WithBuildMetadata(m string) *Builder {
	return b.BuildMetadata(m)
}

// Metadata is an alias for BuildMetadata.
func (b *Builder) Metadata(m string) *Builder {
	return b.BuildMetadata(m)
}

// ClearMetadata removes the build metadata.
func (b *Builder) ClearMetadata() *Builder {
	b.buildMetadata = ""
	return b
}

// ClearBuildMetadata is an alias for ClearMetadata.
func (b *Builder) ClearBuildMetadata() *Builder {
	return b.ClearMetadata()
}

// --- Increment/Decrement Operations ---

// resetRevision resets revision to 0 if it was set, preserving 4-component format
func (b *Builder) resetRevision() {
	if b.revision != nil {
		zero := 0
		b.revision = &zero
	}
}

// IncrementMajor increments major, resets minor, patch, and revision to 0, clears pre-release.
func (b *Builder) IncrementMajor() *Builder {
	b.major++
	b.minor = 0
	b.patch = 0
	b.resetRevision()
	b.preRelease = ""
	return b
}

// IncrementMinor increments minor, resets patch and revision, clears pre-release.
func (b *Builder) IncrementMinor() *Builder {
	b.minor++
	b.patch = 0
	b.resetRevision()
	b.preRelease = ""
	return b
}

// IncrementPatch increments patch, resets revision, clears pre-release.
func (b *Builder) IncrementPatch() *Builder {
	b.patch++
	b.resetRevision()
	b.preRelease = ""
	return b
}

// DecrementMajor decrements major, resets minor, patch, and revision.
// Sets error if major would go below 0.
func (b *Builder) DecrementMajor() *Builder {
	if b.major == 0 {
		b.err = fmt.Errorf("cannot decrement major below 0")
		return b
	}
	b.major--
	b.minor = 0
	b.patch = 0
	b.resetRevision()
	return b
}

// DecrementMinor decrements minor, resets patch and revision.
// Sets error if minor would go below 0.
func (b *Builder) DecrementMinor() *Builder {
	if b.minor == 0 {
		b.err = fmt.Errorf("cannot decrement minor below 0")
		return b
	}
	b.minor--
	b.patch = 0
	b.resetRevision()
	return b
}

// DecrementPatch decrements patch, resets revision.
// Sets error if patch would go below 0.
func (b *Builder) DecrementPatch() *Builder {
	if b.patch == 0 {
		b.err = fmt.Errorf("cannot decrement patch below 0")
		return b
	}
	b.patch--
	b.resetRevision()
	return b
}

// --- Convenience Methods ---

// Release clears pre-release and metadata, producing a clean release version.
func (b *Builder) Release() *Builder {
	b.preRelease = ""
	b.buildMetadata = ""
	return b
}

// Alpha sets pre-release to "alpha" or "alpha.N".
func (b *Builder) Alpha(n ...int) *Builder {
	if len(n) > 0 && n[0] > 0 {
		b.preRelease = fmt.Sprintf("alpha.%d", n[0])
	} else {
		b.preRelease = "alpha"
	}
	return b
}

// Beta sets pre-release to "beta" or "beta.N".
func (b *Builder) Beta(n ...int) *Builder {
	if len(n) > 0 && n[0] > 0 {
		b.preRelease = fmt.Sprintf("beta.%d", n[0])
	} else {
		b.preRelease = "beta"
	}
	return b
}

// RC sets pre-release to "rc.N".
func (b *Builder) RC(n int) *Builder {
	b.preRelease = fmt.Sprintf("rc.%d", n)
	return b
}

// Snapshot sets pre-release to "SNAPSHOT" (common in Java/Maven).
func (b *Builder) Snapshot() *Builder {
	b.preRelease = "SNAPSHOT"
	return b
}

// Dev sets pre-release to "dev" or "dev.N".
func (b *Builder) Dev(n ...int) *Builder {
	if len(n) > 0 && n[0] > 0 {
		b.preRelease = fmt.Sprintf("dev.%d", n[0])
	} else {
		b.preRelease = "dev"
	}
	return b
}

// --- Build Methods ---

// Build constructs the version string, parses it back through the grammar,
// and validates the result. This round-trip ensures the built version is
// always grammatically correct and parseable.
//
// Returns an error if:
//   - Any setter recorded an error (e.g., negative version number)
//   - The constructed string fails to parse (grammar violation)
//   - The parsed version fails validation (e.g., leading zeros)
func (b *Builder) Build() (*Version, error) {
	// Check for deferred errors from setters
	if b.err != nil {
		return nil, b.err
	}

	// Validate prefix - only v/V allowed per SemVer convention
	if b.prefix != "" && b.prefix != "v" && b.prefix != "V" {
		return nil, fmt.Errorf("invalid prefix %q: only 'v' or 'V' allowed", b.prefix)
	}

	// Build the version string
	raw := b.String()

	// Round-trip through parser: parse and validate
	// This ensures the output is always grammatically correct
	v, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("builder produced invalid version %q: %w", raw, err)
	}

	return v, nil
}

// MustBuild calls Build and panics on error.
// Use only when you're certain the version is valid.
func (b *Builder) MustBuild() *Version {
	v, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("MustBuild failed: %v", err))
	}
	return v
}

// String returns the version string without building/validating.
// Useful for debugging or when you just need the string representation.
func (b *Builder) String() string {
	var sb strings.Builder

	sb.WriteString(b.prefix)
	sb.WriteString(fmt.Sprintf("%d.%d.%d", b.major, b.minor, b.patch))

	if b.revision != nil {
		sb.WriteString(fmt.Sprintf(".%d", *b.revision))
	}

	if b.preRelease != "" {
		sb.WriteByte('-')
		sb.WriteString(b.preRelease)
	}

	if b.buildMetadata != "" {
		sb.WriteByte('+')
		sb.WriteString(b.buildMetadata)
	}

	return sb.String()
}

// CoreString returns just Major.Minor.Patch without prefix, pre-release, or metadata.
func (b *Builder) CoreString() string {
	return fmt.Sprintf("%d.%d.%d", b.major, b.minor, b.patch)
}

// --- Accessor Methods (for inspection during building) ---

// GetMajor returns the current major version.
func (b *Builder) GetMajor() int { return b.major }

// GetMinor returns the current minor version.
func (b *Builder) GetMinor() int { return b.minor }

// GetPatch returns the current patch version.
func (b *Builder) GetPatch() int { return b.patch }

// GetRevision returns the current revision (nil if not set).
func (b *Builder) GetRevision() *int { return b.revision }

// GetPrefix returns the current prefix.
func (b *Builder) GetPrefix() string { return b.prefix }

// GetPreRelease returns the current pre-release string.
func (b *Builder) GetPreRelease() string { return b.preRelease }

// GetBuildMetadata returns the current build metadata.
func (b *Builder) GetBuildMetadata() string { return b.buildMetadata }

// HasError returns true if any operation recorded an error.
func (b *Builder) HasError() bool { return b.err != nil }

// Error returns the first error encountered, or nil.
func (b *Builder) Error() error { return b.err }
