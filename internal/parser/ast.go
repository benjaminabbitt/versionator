package parser

import (
	"strings"
)

// VersionFile represents the content of a VERSION file.
// Grammar: version-file = version, [ newline ] ;
type VersionFile struct {
	Version *Version `parser:"@@"`
}

// Version represents a parsed version string with all components.
// Grammar: version = [ prefix ], version-core ;
type Version struct {
	Prefix        string         `parser:"@Prefix?"`
	Core          *VersionCore   `parser:"@@"`
	PreRelease    *PreRelease    `parser:"@@?"`
	BuildMetadata *BuildMetadata `parser:"@@?"`
	Raw           string         // Original input string (set after parsing)
}

// VersionCore represents the numeric parts of a version.
// Grammar: version-core = major [ "." minor [ "." patch [ "." revision ] ] ] ;
type VersionCore struct {
	Major    int  `parser:"@Number"`
	Minor    *int `parser:"( Dot @Number"`
	Patch    *int `parser:"  ( Dot @Number"`
	Revision *int `parser:"    ( Dot @Number )? )? )?"`
}

// PreRelease represents the pre-release portion of a version.
// Grammar: pre-release = "-", identifier, { ".", identifier } ;
// Identifiers can be pure numbers, pure letters, or mixed alphanumeric.
type PreRelease struct {
	Identifiers []*Identifier `parser:"Dash @@ ( Dot @@ )*"`
}

// BuildMetadata represents the build metadata portion of a version.
// Grammar: build-metadata = "+", identifier, { ".", identifier } ;
type BuildMetadata struct {
	Identifiers []*Identifier `parser:"Plus @@ ( Dot @@ )*"`
}

// Identifier represents a single pre-release or build metadata identifier.
// Can be numeric, alphanumeric, mixed (like "5114f85"), or dashes only (like "--").
type Identifier struct {
	Number *string `parser:"  @Number"`
	Ident  *string `parser:"| @Ident"`
	Mixed  *string `parser:"| @Mixed"`
	Dashes *string `parser:"| @Dashes"`
}

// String returns the identifier value as a string.
func (id *Identifier) String() string {
	if id == nil {
		return ""
	}
	if id.Number != nil {
		return *id.Number
	}
	if id.Ident != nil {
		return *id.Ident
	}
	if id.Mixed != nil {
		return *id.Mixed
	}
	if id.Dashes != nil {
		return *id.Dashes
	}
	return ""
}

// IsNumeric returns true if the identifier is purely numeric.
func (id *Identifier) IsNumeric() bool {
	return id != nil && id.Number != nil
}

// identifiersToStrings converts a slice of Identifiers to strings.
func identifiersToStrings(ids []*Identifier) []string {
	if ids == nil {
		return nil
	}
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// PreReleaseString returns the pre-release as a dot-separated string.
func (v *Version) preReleaseIdentifiers() []string {
	if v.PreRelease == nil {
		return nil
	}
	return identifiersToStrings(v.PreRelease.Identifiers)
}

// BuildMetadataIdentifiers returns the build metadata identifiers as strings.
func (v *Version) buildMetadataIdentifiers() []string {
	if v.BuildMetadata == nil {
		return nil
	}
	return identifiersToStrings(v.BuildMetadata.Identifiers)
}

// PreReleaseString returns the pre-release portion as a string.
func (v *Version) PreReleaseString() string {
	ids := v.preReleaseIdentifiers()
	if len(ids) == 0 {
		return ""
	}
	return strings.Join(ids, ".")
}

// BuildMetadataString returns the build metadata as a string.
func (v *Version) BuildMetadataString() string {
	ids := v.buildMetadataIdentifiers()
	if len(ids) == 0 {
		return ""
	}
	return strings.Join(ids, ".")
}
