package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ToVersionData converts a parser.Version to the data needed for the version package.
// Returns: prefix, major, minor, patch, revision, preRelease, buildMetadata, raw
func (v *Version) ToVersionData() (prefix string, major, minor, patch int, revision *int, preRelease, buildMetadata, raw string) {
	if v == nil {
		return "", 0, 0, 0, nil, "", "", ""
	}

	prefix = v.Prefix
	major = v.Major()
	minor = v.Minor()
	patch = v.Patch()
	if v.IsAssemblyVersion() {
		rev := v.Revision()
		revision = &rev
	}
	preRelease = v.PreReleaseString()
	buildMetadata = v.BuildMetadataString()
	raw = v.Raw

	return
}

// FromVersionData creates a parser.Version from version data components.
// This is used when constructing versions programmatically.
func FromVersionData(prefix string, major, minor, patch int, revision *int, preRelease, buildMetadata string) *Version {
	// Build the raw string
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString(fmt.Sprintf("%d.%d.%d", major, minor, patch))
	if revision != nil {
		sb.WriteString(fmt.Sprintf(".%d", *revision))
	}
	if preRelease != "" {
		sb.WriteByte('-')
		sb.WriteString(preRelease)
	}
	if buildMetadata != "" {
		sb.WriteByte('+')
		sb.WriteString(buildMetadata)
	}
	raw := sb.String()

	// Build the version structure
	minorPtr := &minor
	patchPtr := &patch

	v := &Version{
		Prefix: prefix,
		Core: &VersionCore{
			Major:    major,
			Minor:    minorPtr,
			Patch:    patchPtr,
			Revision: revision,
		},
		Raw: raw,
	}

	// Parse pre-release identifiers
	if preRelease != "" {
		v.PreRelease = &PreRelease{
			Identifiers: parseIdentifiers(preRelease),
		}
	}

	// Parse build metadata identifiers
	if buildMetadata != "" {
		v.BuildMetadata = &BuildMetadata{
			Identifiers: parseIdentifiers(buildMetadata),
		}
	}

	return v
}

// parseIdentifiers splits a dot-separated string into Identifier structs.
func parseIdentifiers(s string) []*Identifier {
	parts := strings.Split(s, ".")
	result := make([]*Identifier, len(parts))
	for i, part := range parts {
		part := part // Capture loop variable to avoid pointer aliasing
		id := &Identifier{}
		// Check if it's purely numeric
		if _, err := strconv.Atoi(part); err == nil {
			id.Number = &part
		} else if containsDigit(part) {
			// Mixed alphanumeric
			id.Mixed = &part
		} else {
			// Pure letters
			id.Ident = &part
		}
		result[i] = id
	}
	return result
}

// containsDigit returns true if the string contains at least one digit.
func containsDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the Version.
func (v *Version) Clone() *Version {
	if v == nil {
		return nil
	}

	clone := &Version{
		Prefix: v.Prefix,
		Raw:    v.Raw,
	}

	if v.Core != nil {
		clone.Core = &VersionCore{
			Major: v.Core.Major,
		}
		if v.Core.Minor != nil {
			minor := *v.Core.Minor
			clone.Core.Minor = &minor
		}
		if v.Core.Patch != nil {
			patch := *v.Core.Patch
			clone.Core.Patch = &patch
		}
		if v.Core.Revision != nil {
			revision := *v.Core.Revision
			clone.Core.Revision = &revision
		}
	}

	if v.PreRelease != nil && len(v.PreRelease.Identifiers) > 0 {
		clone.PreRelease = &PreRelease{
			Identifiers: make([]*Identifier, len(v.PreRelease.Identifiers)),
		}
		for i, id := range v.PreRelease.Identifiers {
			clone.PreRelease.Identifiers[i] = cloneIdentifier(id)
		}
	}

	if v.BuildMetadata != nil && len(v.BuildMetadata.Identifiers) > 0 {
		clone.BuildMetadata = &BuildMetadata{
			Identifiers: make([]*Identifier, len(v.BuildMetadata.Identifiers)),
		}
		for i, id := range v.BuildMetadata.Identifiers {
			clone.BuildMetadata.Identifiers[i] = cloneIdentifier(id)
		}
	}

	return clone
}

// cloneIdentifier creates a copy of an Identifier.
func cloneIdentifier(id *Identifier) *Identifier {
	if id == nil {
		return nil
	}
	clone := &Identifier{}
	if id.Number != nil {
		s := *id.Number
		clone.Number = &s
	}
	if id.Ident != nil {
		s := *id.Ident
		clone.Ident = &s
	}
	if id.Mixed != nil {
		s := *id.Mixed
		clone.Mixed = &s
	}
	return clone
}

// SetMajor updates the major version and resets minor, patch, and pre-release.
func (v *Version) SetMajor(major int) {
	if v.Core == nil {
		v.Core = &VersionCore{}
	}
	v.Core.Major = major
	zero := 0
	v.Core.Minor = &zero
	v.Core.Patch = &zero
	v.PreRelease = nil
	v.updateRaw()
}

// SetMinor updates the minor version and resets patch and pre-release.
func (v *Version) SetMinor(minor int) {
	if v.Core == nil {
		v.Core = &VersionCore{}
	}
	v.Core.Minor = &minor
	zero := 0
	v.Core.Patch = &zero
	v.PreRelease = nil
	v.updateRaw()
}

// SetPatch updates the patch version and clears pre-release.
func (v *Version) SetPatch(patch int) {
	if v.Core == nil {
		v.Core = &VersionCore{}
	}
	v.Core.Patch = &patch
	v.PreRelease = nil
	v.updateRaw()
}

// SetPreRelease updates the pre-release identifier.
func (v *Version) SetPreRelease(preRelease string) {
	if preRelease == "" {
		v.PreRelease = nil
	} else {
		v.PreRelease = &PreRelease{
			Identifiers: parseIdentifiers(preRelease),
		}
	}
	v.updateRaw()
}

// SetBuildMetadata updates the build metadata.
func (v *Version) SetBuildMetadata(metadata string) {
	if metadata == "" {
		v.BuildMetadata = nil
	} else {
		v.BuildMetadata = &BuildMetadata{
			Identifiers: parseIdentifiers(metadata),
		}
	}
	v.updateRaw()
}

// SetPrefix updates the version prefix.
func (v *Version) SetPrefix(prefix string) {
	v.Prefix = prefix
	v.updateRaw()
}

// updateRaw rebuilds the Raw string from current values.
func (v *Version) updateRaw() {
	v.Raw = v.FullString()
}

// resetRevision resets revision to 0 if it was set, preserving 4-component format
func (v *Version) resetRevision() {
	if v.Core != nil && v.Core.Revision != nil {
		zero := 0
		v.Core.Revision = &zero
	}
}

// IncrementMajor increments the major version, resets minor, patch, and revision.
func (v *Version) IncrementMajor() {
	if v.Core == nil {
		v.Core = &VersionCore{}
	}
	v.Core.Major++
	zero := 0
	v.Core.Minor = &zero
	v.Core.Patch = &zero
	v.resetRevision()
	v.PreRelease = nil
	v.updateRaw()
}

// IncrementMinor increments the minor version, resets patch and revision.
func (v *Version) IncrementMinor() {
	if v.Core == nil {
		v.Core = &VersionCore{}
	}
	if v.Core.Minor == nil {
		one := 1
		v.Core.Minor = &one
	} else {
		*v.Core.Minor++
	}
	zero := 0
	v.Core.Patch = &zero
	v.resetRevision()
	v.PreRelease = nil
	v.updateRaw()
}

// IncrementPatch increments the patch version, resets revision.
func (v *Version) IncrementPatch() {
	if v.Core == nil {
		v.Core = &VersionCore{}
	}
	if v.Core.Patch == nil {
		one := 1
		v.Core.Patch = &one
	} else {
		*v.Core.Patch++
	}
	v.resetRevision()
	v.PreRelease = nil
	v.updateRaw()
}
