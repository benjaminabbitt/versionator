package version

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/benjaminabbitt/versionator/internal/parser"
	"github.com/benjaminabbitt/versionator/internal/vcs"
)

const versionFile = "VERSION"

// Version represents a semantic version with all components
// This is the unified struct replacing both SemVer and VersionData
type Version struct {
	Prefix        string // Optional prefix ("v" or "V" only, per SemVer convention)
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

// Parse parses a version string into a Version struct using the grammar-based parser.
// Returns a Version with zero values for unparseable input (lenient parsing).
// For strict parsing with error reporting, use ParseStrict.
func Parse(version string) Version {
	v, err := ParseStrict(version)
	if err != nil {
		// Return a minimal version with the raw input preserved
		return Version{Raw: version}
	}
	return *v
}

// ParseStrict parses a version string with full validation.
// Returns an error if the version string is invalid according to the grammar.
func ParseStrict(version string) (*Version, error) {
	pv, err := parser.Parse(version)
	if err != nil {
		return nil, err
	}
	return fromParserVersion(pv), nil
}

// fromParserVersion converts a parser.Version to a version.Version
func fromParserVersion(pv *parser.Version) *Version {
	if pv == nil {
		return &Version{}
	}
	return &Version{
		Prefix:        pv.Prefix,
		Major:         pv.Major(),
		Minor:         pv.Minor(),
		Patch:         pv.Patch(),
		PreRelease:    pv.PreReleaseString(),
		BuildMetadata: pv.BuildMetadataString(),
		Raw:           pv.Raw,
	}
}

// toBuilder converts a version.Version to a parser.Builder for mutation/validation
func (v *Version) toBuilder() *parser.Builder {
	b := parser.NewBuilder().
		Prefix(v.Prefix).
		Major(v.Major).
		Minor(v.Minor).
		Patch(v.Patch)

	if v.PreRelease != "" {
		b.PreRelease(v.PreRelease)
	}
	if v.BuildMetadata != "" {
		b.BuildMetadata(v.BuildMetadata)
	}

	return b
}

// findVersionFile walks up from startPath looking for a VERSION file
// Stops at stopPath (exclusive) or filesystem root
// Returns empty string if not found
func findVersionFile(startPath, stopPath string) string {
	currentPath := startPath

	for {
		versionPath := filepath.Join(currentPath, versionFile)
		if _, err := os.Stat(versionPath); err == nil {
			return versionPath
		}

		// Stop if we've reached the stop path (VCS root) or filesystem root
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached filesystem root
			break
		}
		if stopPath != "" && currentPath == stopPath {
			// Reached VCS root without finding VERSION
			break
		}
		currentPath = parentPath
	}

	return ""
}

// getVersionPath returns the path to the VERSION file
// Walks up from cwd looking for an existing VERSION file
// If not found, returns path in cwd (for creating new VERSION)
func getVersionPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Determine stop path (VCS root if in a repo, otherwise empty = walk to filesystem root)
	var stopPath string
	activeVCS := vcs.GetActiveVCS()
	if activeVCS != nil {
		if root, err := activeVCS.GetRepositoryRoot(); err == nil {
			stopPath = root
		}
	}

	// Walk up looking for VERSION file
	if found := findVersionFile(cwd, stopPath); found != "" {
		return found, nil
	}

	// Not found - return path in cwd for creating new VERSION
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

// Save writes the version to the VERSION file.
// Validates the version by round-tripping through the parser before writing.
func Save(v *Version) error {
	logger := logging.GetLogger()

	// Validate by building through the parser (round-trip validation)
	validated, err := v.toBuilder().Build()
	if err != nil {
		logger.Error(LogVersionParseError, zap.String("version", v.String()), zap.Error(err))
		return fmt.Errorf("invalid version: %w", err)
	}

	path, err := getVersionPath()
	if err != nil {
		return err
	}

	// Write the validated version string with newline
	content := validated.FullString() + "\n"

	if err := os.WriteFile(path, []byte(content), FilePermission); err != nil {
		logger.Error(LogFileWriteError, zap.String("path", path), zap.Error(err))
		return fmt.Errorf("failed to write VERSION: %w", err)
	}

	logger.Debug(LogVersionSaved,
		zap.String("path", path),
		zap.String("version", validated.String()))
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

// Validate checks if the version is valid according to SemVer 2.0.0.
// Uses the grammar-based parser for validation via round-trip.
func (v *Version) Validate() error {
	_, err := v.toBuilder().Build()
	return err
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

