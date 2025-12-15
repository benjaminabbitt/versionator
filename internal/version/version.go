package version

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/benjaminabbitt/versionator/internal/vcs"
)

const versionFile = "VERSION"

// Version represents a semantic version with all components
// This is the unified struct replacing both SemVer and VersionData
type Version struct {
	Prefix        string // Optional prefix (e.g., "v", "release-")
	Major         int    // Major version
	Minor         int    // Minor version
	Patch         int    // Patch version
	PreRelease    string // Pre-release identifier (e.g., "alpha.1")
	BuildMetadata string // Build metadata (e.g., "build.123")
	Raw           string // Original parsed string
}

// VersionLevel represents the semantic version component to modify
type VersionLevel int

const (
	MajorLevel VersionLevel = iota
	MinorLevel
	PatchLevel
)

// semverRegex matches semantic versions with optional pre-release and build metadata
// Based on SemVer 2.0.0 specification (see resources/semver-2.md)
// Note: prefix is extracted separately before applying this regex
var semverRegex = regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*))?(?:\+([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*))?$`)

// Parse parses a version string into a Version struct
// Extracts prefix as all characters (letters, dashes) before the first digit
// Returns a Version with zero values for unparseable input
func Parse(version string) Version {
	v := Version{Raw: version}

	// Find first digit - everything before it is the prefix
	firstDigit := -1
	for i, c := range version {
		if c >= '0' && c <= '9' {
			firstDigit = i
			break
		}
	}

	// No digit found - unparseable
	if firstDigit == -1 {
		return v
	}

	// Extract prefix (everything before first digit)
	v.Prefix = version[:firstDigit]
	semverPart := version[firstDigit:]

	matches := semverRegex.FindStringSubmatch(semverPart)
	if matches == nil {
		return v
	}

	// Capture groups: [0]=full match, [1]=major, [2]=minor, [3]=patch, [4]=prerelease, [5]=metadata

	// Major version (required)
	if matches[1] != "" {
		major, err := strconv.Atoi(matches[1])
		if err != nil {
			return Version{Raw: version}
		}
		v.Major = major
	}

	// Minor version (optional, defaults to 0)
	if matches[2] != "" {
		minor, err := strconv.Atoi(matches[2])
		if err != nil {
			return Version{Raw: version}
		}
		v.Minor = minor
	}

	// Patch version (optional, defaults to 0)
	if matches[3] != "" {
		patch, err := strconv.Atoi(matches[3])
		if err != nil {
			return Version{Raw: version}
		}
		v.Patch = patch
	}

	v.PreRelease = matches[4]
	v.BuildMetadata = matches[5]

	return v
}

// getVersionPath returns the path to the VERSION file
func getVersionPath() (string, error) {
	activeVCS := vcs.GetActiveVCS()
	if activeVCS != nil {
		root, err := activeVCS.GetRepositoryRoot()
		if err == nil {
			return filepath.Join(root, versionFile), nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, versionFile), nil
}

// Load reads the VERSION file and returns the parsed Version
// If VERSION doesn't exist, creates a default 0.0.0 (using config prefix if set)
// VERSION file content is the source of truth - it takes priority over config
func Load() (*Version, error) {
	logger := logging.GetLogger()

	path, err := getVersionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err == nil {
		// VERSION file exists - parse it directly
		// The VERSION file is the source of truth
		versionStr := strings.TrimSpace(string(data))
		v := Parse(versionStr)

		logger.Debug(LogVersionLoaded,
			zap.String("path", path),
			zap.String("version", v.String()))
		return &v, nil
	}

	// VERSION doesn't exist, create default
	if os.IsNotExist(err) {
		// Use config prefix as default for new files only
		cfg, _ := config.ReadConfig()
		defaultPrefix := ""
		if cfg != nil {
			defaultPrefix = cfg.Prefix
		}

		v := &Version{Major: 0, Minor: 0, Patch: 0, Prefix: defaultPrefix}
		logger.Info(LogVersionCreated,
			zap.String("path", path),
			zap.String("version", v.String()))
		if saveErr := Save(v); saveErr != nil {
			return nil, fmt.Errorf("failed to create VERSION: %w", saveErr)
		}
		return v, nil
	}

	logger.Error(LogFileReadError, zap.String("path", path), zap.Error(err))
	return nil, fmt.Errorf("failed to read VERSION: %w", err)
}

// Save writes the version to the VERSION file
func Save(v *Version) error {
	logger := logging.GetLogger()

	if err := v.Validate(); err != nil {
		logger.Error(LogVersionParseError, zap.String("version", v.String()), zap.Error(err))
		return fmt.Errorf("invalid version: %w", err)
	}

	path, err := getVersionPath()
	if err != nil {
		return err
	}

	// Write full version string with newline
	content := v.FullString() + "\n"

	if err := os.WriteFile(path, []byte(content), FilePermission); err != nil {
		logger.Error(LogFileWriteError, zap.String("path", path), zap.Error(err))
		return fmt.Errorf("failed to write VERSION: %w", err)
	}

	logger.Debug(LogVersionSaved,
		zap.String("path", path),
		zap.String("version", v.String()))
	return nil
}

// String returns the SemVer 2.0.0 compliant string (no prefix)
// Format: Major.Minor.Patch[-PreRelease][+BuildMetadata]
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		result += "-" + v.PreRelease
	}
	if v.BuildMetadata != "" {
		result += "+" + v.BuildMetadata
	}
	return result
}

// CoreVersion returns just Major.Minor.Patch
func (v *Version) CoreVersion() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// FullString returns the full version string with prefix
// Format: [Prefix]Major.Minor.Patch[-PreRelease][+BuildMetadata]
func (v *Version) FullString() string {
	return v.Prefix + v.String()
}

// MajorMinor returns the Major.Minor string
func (v *Version) MajorMinor() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// MajorString returns the major version as a string
func (v *Version) MajorString() string {
	return strconv.Itoa(v.Major)
}

// MinorString returns the minor version as a string
func (v *Version) MinorString() string {
	return strconv.Itoa(v.Minor)
}

// PatchString returns the patch version as a string
func (v *Version) PatchString() string {
	return strconv.Itoa(v.Patch)
}

// SemVer returns Major.Minor.Patch[-PreRelease] (no metadata)
func (v *Version) SemVer() string {
	result := v.CoreVersion()
	if v.PreRelease != "" {
		result += "-" + v.PreRelease
	}
	return result
}

// FullSemVer returns Major.Minor.Patch[-PreRelease][+BuildMetadata]
func (v *Version) FullSemVer() string {
	return v.String()
}

// PreReleaseWithDash returns the pre-release with a leading dash, or empty if none
func (v *Version) PreReleaseWithDash() string {
	if v.PreRelease != "" {
		return "-" + v.PreRelease
	}
	return ""
}

// BuildMetadataWithPlus returns the build metadata with a leading plus, or empty if none
func (v *Version) BuildMetadataWithPlus() string {
	if v.BuildMetadata != "" {
		return "+" + v.BuildMetadata
	}
	return ""
}

// HasPrefix returns true if the version has a prefix
func (v *Version) HasPrefix() bool {
	return v.Prefix != ""
}

// IsPreRelease returns true if this version has a pre-release tag
func (v *Version) IsPreRelease() bool {
	return v.PreRelease != ""
}

// HasBuildMetadata returns true if this version has build metadata
func (v *Version) HasBuildMetadata() bool {
	return v.BuildMetadata != ""
}

// PreReleaseLabel returns just the label portion of the pre-release (e.g., "alpha" from "alpha.5")
func (v *Version) PreReleaseLabel() string {
	if v.PreRelease == "" {
		return ""
	}
	parts := strings.Split(v.PreRelease, ".")
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return part
		}
	}
	return v.PreRelease
}

// PreReleaseNumber returns the numeric portion of the pre-release (e.g., 5 from "alpha.5")
// Returns -1 if no numeric portion exists
func (v *Version) PreReleaseNumber() int {
	if v.PreRelease == "" {
		return -1
	}
	parts := strings.Split(v.PreRelease, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if num, err := strconv.Atoi(parts[i]); err == nil {
			return num
		}
	}
	return -1
}

// PreReleaseLabelWithDash returns the pre-release label with a leading dash, or empty if none
func (v *Version) PreReleaseLabelWithDash() string {
	label := v.PreReleaseLabel()
	if label != "" {
		return "-" + label
	}
	return ""
}

// AssemblyVersion returns an assembly-compatible version (Major.Minor.Patch.0)
func (v *Version) AssemblyVersion() string {
	return v.CoreVersion() + ".0"
}

// PrefixedString returns the version with prefix (Major.Minor.Patch)
func (v *Version) PrefixedString() string {
	return v.Prefix + v.CoreVersion()
}

// SemVerString returns Major.Minor.Patch[-PreRelease] format without prefix
func (v *Version) SemVerString() string {
	return v.SemVer()
}

// PrefixedSemVerString returns Major.Minor.Patch[-PreRelease] format with the original prefix
func (v *Version) PrefixedSemVerString() string {
	return v.Prefix + v.SemVer()
}

// PrefixedFullSemVer returns Major.Minor.Patch[-PreRelease][+BuildMetadata] with the original prefix
func (v *Version) PrefixedFullSemVer() string {
	return v.Prefix + v.FullSemVer()
}

// OriginalString returns the version in the same format as the input
func (v *Version) OriginalString() string {
	if v.HasPrefix() {
		return v.PrefixedString()
	}
	return v.CoreVersion()
}

// OriginalSemVerString returns the semver string preserving the original prefix style
func (v *Version) OriginalSemVerString() string {
	if v.HasPrefix() {
		return v.PrefixedSemVerString()
	}
	return v.SemVerString()
}

// OriginalFullSemVer returns the full semver string preserving the original prefix style
func (v *Version) OriginalFullSemVer() string {
	if v.HasPrefix() {
		return v.PrefixedFullSemVer()
	}
	return v.FullSemVer()
}

// Validate checks if the version is valid according to SemVer 2.0.0
func (v *Version) Validate() error {
	if v.Major < 0 {
		return errors.New(ErrMajorVersionNegative)
	}
	if v.Minor < 0 {
		return errors.New(ErrMinorVersionNegative)
	}
	if v.Patch < 0 {
		return errors.New(ErrPatchVersionNegative)
	}

	if v.PreRelease != "" {
		if err := validateIdentifier(v.PreRelease); err != nil {
			return fmt.Errorf("%s: %w", ErrInvalidPreRelease, err)
		}
	}

	if v.BuildMetadata != "" {
		if err := validateIdentifier(v.BuildMetadata); err != nil {
			return fmt.Errorf("%s: %w", ErrInvalidMetadata, err)
		}
	}

	return nil
}

// validateIdentifier checks if an identifier is valid per SemVer 2.0.0
func validateIdentifier(id string) error {
	for _, part := range strings.Split(id, ".") {
		if part == "" {
			return errors.New(ErrEmptyIdentifierPart)
		}
		for _, c := range part {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '-') {
				return fmt.Errorf("%s: '%c'", ErrInvalidIdentifierChar, c)
			}
		}
	}
	return nil
}

// IncrementMajor increments the major version, resets minor and patch
func (v *Version) IncrementMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	v.PreRelease = ""
}

// IncrementMinor increments the minor version, resets patch
func (v *Version) IncrementMinor() {
	v.Minor++
	v.Patch = 0
	v.PreRelease = ""
}

// IncrementPatch increments the patch version
func (v *Version) IncrementPatch() {
	v.Patch++
	v.PreRelease = ""
}

// DecrementMajor decrements the major version, returns error if already 0
func (v *Version) DecrementMajor() error {
	if v.Major == 0 {
		return errors.New(ErrCannotDecrementMajor)
	}
	v.Major--
	v.Minor = 0
	v.Patch = 0
	return nil
}

// DecrementMinor decrements the minor version, returns error if already 0
func (v *Version) DecrementMinor() error {
	if v.Minor == 0 {
		return errors.New(ErrCannotDecrementMinor)
	}
	v.Minor--
	v.Patch = 0
	return nil
}

// DecrementPatch decrements the patch version, returns error if already 0
func (v *Version) DecrementPatch() error {
	if v.Patch == 0 {
		return errors.New(ErrCannotDecrementPatch)
	}
	v.Patch--
	return nil
}

// --- Package-level convenience functions ---

// GetCurrentVersion reads the current version from VERSION file
func GetCurrentVersion() (string, error) {
	v, err := Load()
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

// Increment increments the specified version level
func Increment(level VersionLevel) error {
	logger := logging.GetLogger()

	v, err := Load()
	if err != nil {
		return err
	}

	oldVersion := v.String()

	switch level {
	case MajorLevel:
		v.IncrementMajor()
	case MinorLevel:
		v.IncrementMinor()
	case PatchLevel:
		v.IncrementPatch()
	default:
		return fmt.Errorf("%s: %d", ErrInvalidVersionLevel, level)
	}

	logger.Info(LogVersionIncremented,
		zap.String("level", levelString(level)),
		zap.String("from", oldVersion),
		zap.String("to", v.String()))

	return Save(v)
}

// Decrement decrements the specified version level
func Decrement(level VersionLevel) error {
	logger := logging.GetLogger()

	v, err := Load()
	if err != nil {
		return err
	}

	oldVersion := v.String()

	var decrementErr error
	switch level {
	case MajorLevel:
		decrementErr = v.DecrementMajor()
	case MinorLevel:
		decrementErr = v.DecrementMinor()
	case PatchLevel:
		decrementErr = v.DecrementPatch()
	default:
		return fmt.Errorf("%s: %d", ErrInvalidVersionLevel, level)
	}

	if decrementErr != nil {
		return decrementErr
	}

	logger.Info(LogVersionDecremented,
		zap.String("level", levelString(level)),
		zap.String("from", oldVersion),
		zap.String("to", v.String()))

	return Save(v)
}

func levelString(level VersionLevel) string {
	switch level {
	case MajorLevel:
		return "major"
	case MinorLevel:
		return "minor"
	case PatchLevel:
		return "patch"
	default:
		return "unknown"
	}
}

// SetPrefix sets the version prefix in the VERSION file
// The VERSION file is the source of truth - config.Prefix is only used as default for new projects
func SetPrefix(prefix string) error {
	logger := logging.GetLogger()

	v, err := Load()
	if err != nil {
		return err
	}

	oldPrefix := v.Prefix
	v.Prefix = prefix

	logger.Debug(LogPrefixSet, zap.String("from", oldPrefix), zap.String("to", prefix))
	return Save(v)
}

// GetPrefix returns the current version prefix
func GetPrefix() (string, error) {
	v, err := Load()
	if err != nil {
		return "", err
	}
	return v.Prefix, nil
}

// SetPreRelease sets the pre-release tag
func SetPreRelease(preRelease string) error {
	v, err := Load()
	if err != nil {
		return err
	}
	v.PreRelease = preRelease
	return Save(v)
}

// SetMetadata sets the build metadata
func SetMetadata(metadata string) error {
	v, err := Load()
	if err != nil {
		return err
	}
	v.BuildMetadata = metadata
	return Save(v)
}

// StripPrefix removes the 'v' or 'V' prefix from a version string if present
func StripPrefix(version string) string {
	return strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V")
}

// EscapedBranchName takes a branch name and escapes slashes to dashes
func EscapedBranchName(branchName string) string {
	return strings.ReplaceAll(branchName, "/", "-")
}

