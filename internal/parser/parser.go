package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

// Error constants for parser validation.
const (
	ErrEmptyVersion         = "version string cannot be empty"
	ErrLeadingZero          = "numeric identifier cannot have leading zeros"
	ErrEmptyIdentifier      = "identifier cannot be empty"
	ErrMajorRequired        = "major version is required"
	ErrNegativeVersion      = "version component cannot be negative"
	ErrInvalidPreRelease    = "invalid pre-release identifier"
	ErrInvalidBuildMetadata = "invalid build metadata"
)

// Parser is the version string parser built from the grammar.
var Parser *participle.Parser[VersionFile]

func init() {
	var err error
	Parser, err = participle.Build[VersionFile](
		participle.Lexer(VersionLexer),
		participle.UseLookahead(2),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to build version parser: %v", err))
	}
}

// Parse parses a version string and returns the parsed Version.
// Returns an error if the string is not a valid version.
// Only v/V prefixes are supported per SemVer convention.
func Parse(input string) (*Version, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New(ErrEmptyVersion)
	}

	file, err := Parser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if file.Version == nil {
		return nil, errors.New(ErrMajorRequired)
	}

	// Store original input
	file.Version.Raw = input

	// Validate the parsed version
	if err := file.Version.Validate(); err != nil {
		return nil, err
	}

	return file.Version, nil
}

// ParseLenient parses a version string, returning a best-effort Version.
// Unlike Parse, it does not return an error for minor validation issues.
// Use this for reading existing VERSION files that may have legacy formats.
func ParseLenient(input string) *Version {
	v, err := Parse(input)
	if err != nil {
		// Return a zero version with the raw input preserved
		return &Version{Raw: input, Core: &VersionCore{}}
	}
	return v
}

// Validate checks if the parsed version is valid according to SemVer 2.0.0.
func (v *Version) Validate() error {
	if v.Core == nil {
		return errors.New(ErrMajorRequired)
	}

	// Validate major (required, non-negative)
	if v.Core.Major < 0 {
		return errors.New(ErrNegativeVersion)
	}

	// Check for leading zeros in numeric components
	if err := v.validateLeadingZeros(); err != nil {
		return err
	}

	// Validate pre-release identifiers
	if v.PreRelease != nil {
		for _, id := range v.PreRelease.Identifiers {
			if err := validatePreReleaseIdentifier(id); err != nil {
				return fmt.Errorf("%s: %w", ErrInvalidPreRelease, err)
			}
		}
	}

	// Validate build metadata identifiers
	if v.BuildMetadata != nil {
		for _, id := range v.BuildMetadata.Identifiers {
			if err := validateBuildMetadataIdentifier(id); err != nil {
				return fmt.Errorf("%s: %w", ErrInvalidBuildMetadata, err)
			}
		}
	}

	return nil
}

// validateLeadingZeros checks that numeric version components don't have leading zeros.
// Per SemVer 2.0.0: "A normal version number MUST NOT contain leading zeroes."
func (v *Version) validateLeadingZeros() error {
	// Extract numeric parts from raw input and check for leading zeros
	raw := v.Raw
	if v.Prefix != "" {
		raw = strings.TrimPrefix(raw, v.Prefix)
	}

	// Find the core version part (before - or +)
	corePart := raw
	if idx := strings.IndexAny(raw, "-+"); idx >= 0 {
		corePart = raw[:idx]
	}

	// Split by dots and check each numeric component
	parts := strings.Split(corePart, ".")
	for _, part := range parts {
		if len(part) > 1 && part[0] == '0' {
			// Check if it's all digits (numeric identifier)
			allDigits := true
			for _, c := range part {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return fmt.Errorf("%s: %s", ErrLeadingZero, part)
			}
		}
	}

	return nil
}

// validatePreReleaseIdentifier validates a pre-release identifier.
// Numeric identifiers must not have leading zeros per SemVer 2.0.0.
func validatePreReleaseIdentifier(id *Identifier) error {
	if id == nil {
		return errors.New(ErrEmptyIdentifier)
	}

	val := id.String()
	if val == "" {
		return errors.New(ErrEmptyIdentifier)
	}

	// For numeric identifiers, check for leading zeros
	if id.IsNumeric() && len(val) > 1 && val[0] == '0' {
		return fmt.Errorf("%s: %s", ErrLeadingZero, val)
	}

	return nil
}

// validateBuildMetadataIdentifier validates a build metadata identifier.
// Leading zeros ARE allowed in build metadata per SemVer 2.0.0.
func validateBuildMetadataIdentifier(id *Identifier) error {
	if id == nil {
		return errors.New(ErrEmptyIdentifier)
	}

	val := id.String()
	if val == "" {
		return errors.New(ErrEmptyIdentifier)
	}

	return nil
}

// Major returns the major version number.
func (v *Version) Major() int {
	if v.Core == nil {
		return 0
	}
	return v.Core.Major
}

// Minor returns the minor version number (0 if not specified).
func (v *Version) Minor() int {
	if v.Core == nil || v.Core.Minor == nil {
		return 0
	}
	return *v.Core.Minor
}

// Patch returns the patch version number (0 if not specified).
func (v *Version) Patch() int {
	if v.Core == nil || v.Core.Patch == nil {
		return 0
	}
	return *v.Core.Patch
}

// Revision returns the revision number for assembly versions (0 if not specified).
func (v *Version) Revision() int {
	if v.Core == nil || v.Core.Revision == nil {
		return 0
	}
	return *v.Core.Revision
}

// HasPrefix returns true if the version has a prefix (e.g., "v").
func (v *Version) HasPrefix() bool {
	return v.Prefix != ""
}

// IsPreRelease returns true if the version has a pre-release tag.
func (v *Version) IsPreRelease() bool {
	return v.PreRelease != nil && len(v.PreRelease.Identifiers) > 0
}

// HasBuildMetadata returns true if the version has build metadata.
func (v *Version) HasBuildMetadata() bool {
	return v.BuildMetadata != nil && len(v.BuildMetadata.Identifiers) > 0
}

// IsAssemblyVersion returns true if this is a 4-component assembly version.
func (v *Version) IsAssemblyVersion() bool {
	return v.Core != nil && v.Core.Revision != nil
}

// IsPartial returns true if the version is missing minor or patch components.
func (v *Version) IsPartial() bool {
	if v.Core == nil {
		return true
	}
	return v.Core.Minor == nil || v.Core.Patch == nil
}

// String returns the SemVer 2.0.0 compliant string (without prefix).
// Format: Major.Minor.Patch[-PreRelease][+BuildMetadata]
func (v *Version) String() string {
	if v.Core == nil {
		return "0.0.0"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d", v.Core.Major))
	sb.WriteByte('.')
	sb.WriteString(fmt.Sprintf("%d", v.Minor()))
	sb.WriteByte('.')
	sb.WriteString(fmt.Sprintf("%d", v.Patch()))

	if v.IsPreRelease() {
		sb.WriteByte('-')
		sb.WriteString(v.PreReleaseString())
	}

	if v.HasBuildMetadata() {
		sb.WriteByte('+')
		sb.WriteString(v.BuildMetadataString())
	}

	return sb.String()
}

// CoreVersion returns just Major.Minor.Patch without pre-release or metadata.
func (v *Version) CoreVersion() string {
	if v.Core == nil {
		return "0.0.0"
	}
	return fmt.Sprintf("%d.%d.%d", v.Core.Major, v.Minor(), v.Patch())
}

// FullString returns the version with prefix.
// Format: [Prefix]Major.Minor.Patch[-PreRelease][+BuildMetadata]
func (v *Version) FullString() string {
	return v.Prefix + v.String()
}

// AssemblyVersion returns a 4-component assembly version.
// Format: Major.Minor.Patch.Revision (Revision defaults to 0)
func (v *Version) AssemblyVersion() string {
	if v.Core == nil {
		return "0.0.0.0"
	}
	return fmt.Sprintf("%d.%d.%d.%d", v.Core.Major, v.Minor(), v.Patch(), v.Revision())
}

// EBNF returns the EBNF grammar string from the parser.
func EBNF() string {
	return Parser.String()
}
