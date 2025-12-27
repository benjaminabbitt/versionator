package emit

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cbroglie/mustache"

	"github.com/benjaminabbitt/versionator/pkg/plugin"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"
)

//go:embed templates/*
var templateFS embed.FS

// Format represents a supported output format
type Format string

const (
	FormatPython    Format = "python"
	FormatJSON      Format = "json"
	FormatYAML      Format = "yaml"
	FormatGo        Format = "go"
	FormatC         Format = "c"
	FormatCHeader   Format = "c-header"
	FormatCPP       Format = "cpp"
	FormatCPPHeader Format = "cpp-header"
	FormatJS        Format = "js"
	FormatTS        Format = "ts"
	FormatJava      Format = "java"
	FormatKotlin    Format = "kotlin"
	FormatCSharp    Format = "csharp"
	FormatPHP       Format = "php"
	FormatSwift     Format = "swift"
	FormatRuby      Format = "ruby"
	FormatRust      Format = "rust"
)

// templateFiles maps formats to their template file names
// Files use double extensions (e.g., .tmpl.py) for IDE syntax highlighting support
var templateFiles = map[Format]string{
	FormatPython:    "templates/python.tmpl",
	FormatJSON:      "templates/json.tmpl",
	FormatYAML:      "templates/yaml.tmpl",
	FormatGo:        "templates/go.tmpl",
	FormatC:         "templates/c.tmpl",
	FormatCHeader:   "templates/c-header.tmpl",
	FormatCPP:       "templates/cpp.tmpl",
	FormatCPPHeader: "templates/cpp-header.tmpl",
	FormatJS:        "templates/js.tmpl",
	FormatTS:        "templates/ts.tmpl",
	FormatJava:      "templates/java.tmpl",
	FormatKotlin:    "templates/kotlin.tmpl",
	FormatCSharp:    "templates/csharp.tmpl",
	FormatPHP:       "templates/php.tmpl",
	FormatSwift:     "templates/swift.tmpl",
	FormatRuby:      "templates/ruby.tmpl",
	FormatRust:      "templates/rust.tmpl",
}

// TemplateData holds the data passed to templates
type TemplateData struct {
	// Version components
	Major                    string // Major version number (e.g., "1")
	Minor                    string // Minor version number (e.g., "2")
	Patch                    string // Patch version number (e.g., "3")
	Revision                 string // Revision version number (e.g., "4", primarily for .NET)
	MajorMinorPatch          string // Core version: Major.Minor.Patch (e.g., "1.2.3")
	MajorMinorPatchRevision  string // Full .NET version: Major.Minor.Patch.Revision (e.g., "1.2.3.4")
	MajorMinor               string // Major.Minor (e.g., "1.2")
	Prefix                   string // Version prefix (e.g., "v")
	AssemblyVersion          string // .NET assembly version: Major.Minor.Patch.Revision (e.g., "1.2.3.0")

	// Rendered pre-release (from template config, dash-separated items)
	PreRelease         string // Rendered pre-release (e.g., "alpha-5")
	PreReleaseWithDash string // With leading dash (e.g., "-alpha-5")
	PreReleaseLabel    string // Just the label part (e.g., "alpha" from "alpha.5")
	PreReleaseNumber   string // Just the number part (e.g., "5" from "alpha.5")

	// Rendered metadata (from template config, dot-separated items)
	Metadata         string // Rendered metadata (e.g., "20241211103045.4846bcd2e133")
	MetadataWithPlus string // With leading plus (e.g., "+20241211103045.4846bcd2e133")

	// VCS/Git info
	Hash               string // Full commit hash (40 chars for git)
	ShortHash          string // Short commit hash (7 chars)
	MediumHash         string // Medium commit hash (12 chars)
	ShortHashWithDot   string // Short hash with leading dot (e.g., ".abc1234") - for metadata
	MediumHashWithDot  string // Medium hash with leading dot (e.g., ".abc1234def5") - for metadata
	ShortHashWithDash  string // Short hash with leading dash (e.g., "-abc1234") - for Go prerelease
	MediumHashWithDash string // Medium hash with leading dash (e.g., "-abc1234def5") - for Go prerelease
	BranchName         string // Current branch name (e.g., "feature/foo")
	EscapedBranchName  string // Branch name with slashes replaced (e.g., "feature-foo")
	CommitsSinceTag    string // Commits since last tag (e.g., "12")
	BuildNumber        string // Alias for CommitsSinceTag (GitVersion compatibility)
	BuildNumberPadded  string // Padded commits since tag, 4 digits (e.g., "0012")
	UncommittedChanges string // Count of uncommitted changes (e.g., "3")
	Dirty              string // "dirty" if uncommitted changes > 0, empty otherwise
	DirtyWithDot       string // ".dirty" if uncommitted changes > 0 - for metadata
	DirtyWithDash      string // "-dirty" if uncommitted changes > 0 - for Go prerelease
	VersionSourceHash  string // Hash of the commit the last tag points to

	// Commit author info
	CommitAuthor      string // Name of the commit author
	CommitAuthorEmail string // Email of the commit author

	// Commit timestamps (all in UTC)
	CommitDate        string // ISO 8601 format: 2024-01-15T10:30:00Z
	CommitDateCompact string // Compact format: 20240115103045 (YYYYMMDDHHmmss)
	CommitDateShort   string // Date only: 2024-01-15
	CommitYear        string // Year: 2024
	CommitMonth       string // Month: 01 (zero-padded)
	CommitDay         string // Day: 15 (zero-padded)

	// Build timestamps (all in UTC)
	BuildDateTimeUTC             string // ISO 8601 format: 2024-01-15T10:30:00Z
	BuildDateTimeCompact         string // Compact format: 20240115103045 (YYYYMMDDHHmmss)
	BuildDateTimeCompactWithDot  string // With leading dot (e.g., ".20240115103045") - for metadata
	BuildDateTimeCompactWithDash string // With leading dash (e.g., "-20240115103045") - for Go prerelease
	BuildDateUTC                 string // Date only: 2024-01-15
	BuildYear                    string // Year: 2024
	BuildMonth                   string // Month: 01 (zero-padded)
	BuildDay                     string // Day: 15 (zero-padded)

	// Custom holds arbitrary key-value pairs from config and --set flags
	Custom map[string]string

	// PluginVariables holds plugin-specific template variables (e.g., GitShortHash, ShaShortHash from git plugin)
	PluginVariables map[string]string
}

// SupportedFormats returns a list of supported format names
func SupportedFormats() []string {
	return []string{
		string(FormatPython),
		string(FormatJSON),
		string(FormatYAML),
		string(FormatGo),
		string(FormatC),
		string(FormatCHeader),
		string(FormatCPP),
		string(FormatCPPHeader),
		string(FormatJS),
		string(FormatTS),
		string(FormatJava),
		string(FormatKotlin),
		string(FormatCSharp),
		string(FormatPHP),
		string(FormatSwift),
		string(FormatRuby),
		string(FormatRust),
	}
}

// IsValidFormat checks if the given format is supported
func IsValidFormat(format string) bool {
	_, ok := templateFiles[Format(format)]
	return ok
}

// getTemplate loads a template from the embedded filesystem
func getTemplate(format Format) (string, error) {
	filename, ok := templateFiles[format]
	if !ok {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	content, err := templateFS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", filename, err)
	}

	return string(content), nil
}

// Render generates the version output for the given format
func Render(format Format, version string) (string, error) {
	tmplStr, err := getTemplate(format)
	if err != nil {
		return "", err
	}

	return RenderTemplate(tmplStr, version)
}

// dirtyFlag returns "dirty" if uncommittedChanges > 0, empty string otherwise
func dirtyFlag(uncommittedChanges int) string {
	if uncommittedChanges > 0 {
		return "dirty"
	}
	return ""
}

// withDotPrefix returns the string with a leading dot, or empty if input is empty
// Useful for metadata suffixes like ".abc1234" or ".dirty"
func withDotPrefix(s string) string {
	if s == "" {
		return ""
	}
	return "." + s
}

// withDashPrefix returns the string with a leading dash, or empty if input is empty
// Useful for Go-style prerelease suffixes like "-abc1234" or "-dirty"
func withDashPrefix(s string) string {
	if s == "" {
		return ""
	}
	return "-" + s
}

// formatCommitDateTime formats a commit datetime as ISO 8601, or empty string if zero
func formatCommitDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// formatPreReleaseNumber formats the pre-release number as string, empty if -1
func formatPreReleaseNumber(n int) string {
	if n < 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// VCSInfo holds all VCS-related information
type VCSInfo struct {
	Identifier         string
	IdentifierShort    string // 7 chars
	IdentifierMedium   string // 12 chars
	BranchName         string
	CommitDate         time.Time
	CommitsSinceTag    int
	UncommittedChanges int
	VersionSourceHash  string
	CommitAuthor       string
	CommitAuthorEmail  string
}

// formattedVCSFields holds pre-formatted VCS fields for template rendering
type formattedVCSFields struct {
	CommitsSinceTag    string
	BuildNumberPadded  string
	UncommittedChanges string
	Dirty              string
	DirtyWithDot       string // ".dirty" if uncommitted changes > 0 - for metadata
	DirtyWithDash      string // "-dirty" if uncommitted changes > 0 - for Go prerelease
	CommitDate         string
	CommitDateCompact  string
	CommitDateShort    string
	CommitYear         string
	CommitMonth        string
	CommitDay          string
}

// formatVCSFields converts VCSInfo to formatted string fields for templates
func formatVCSFields(info VCSInfo) formattedVCSFields {
	dirty := dirtyFlag(info.UncommittedChanges)
	f := formattedVCSFields{
		UncommittedChanges: strconv.Itoa(info.UncommittedChanges),
		Dirty:              dirty,
		DirtyWithDot:       withDotPrefix(dirty),
		DirtyWithDash:      withDashPrefix(dirty),
	}

	// Format commits since tag
	if info.CommitsSinceTag >= 0 {
		f.CommitsSinceTag = strconv.Itoa(info.CommitsSinceTag)
		f.BuildNumberPadded = fmt.Sprintf("%04d", info.CommitsSinceTag)
	}

	// Format commit date fields
	if !info.CommitDate.IsZero() {
		f.CommitDate = info.CommitDate.Format(time.RFC3339)
		f.CommitDateCompact = info.CommitDate.Format("20060102150405")
		f.CommitDateShort = info.CommitDate.Format("2006-01-02")
		f.CommitYear = info.CommitDate.Format("2006")
		f.CommitMonth = info.CommitDate.Format("01")
		f.CommitDay = info.CommitDate.Format("02")
	}

	return f
}

// formattedBuildTime holds pre-formatted build time fields for template rendering
type formattedBuildTime struct {
	DateTime            string
	DateTimeCompact     string
	DateTimeCompactDot  string // With leading dot for metadata (.YYYYMMDDHHmmss)
	DateTimeCompactDash string // With leading dash for Go prerelease (-YYYYMMDDHHmmss)
	DateOnly            string
	Year                string
	Month               string
	Day                 string
}

// formatBuildTime creates formatted build time fields from current UTC time
func formatBuildTime() formattedBuildTime {
	now := time.Now().UTC()
	dateTimeCompact := now.Format("20060102150405")
	return formattedBuildTime{
		DateTime:            now.Format(time.RFC3339),
		DateTimeCompact:     dateTimeCompact,
		DateTimeCompactDot:  "." + dateTimeCompact,
		DateTimeCompactDash: "-" + dateTimeCompact,
		DateOnly:            now.Format("2006-01-02"),
		Year:                now.Format("2006"),
		Month:               now.Format("01"),
		Day:                 now.Format("02"),
	}
}

// getVCSInfo retrieves all VCS information
// Returns empty/zero values if not in a VCS repository
func getVCSInfo() VCSInfo {
	info := VCSInfo{CommitsSinceTag: -1} // -1 indicates no tags

	activeVCS := vcs.GetActiveVCS()
	if activeVCS == nil {
		return info
	}

	// Get full identifier (40 chars for git)
	if id, err := activeVCS.GetVCSIdentifier(40); err == nil {
		info.Identifier = id
	}

	// Get short identifier (7 chars is common default)
	if id, err := activeVCS.GetVCSIdentifier(7); err == nil {
		info.IdentifierShort = id
	}

	// Get medium identifier (12 chars for Go-like versions)
	if id, err := activeVCS.GetVCSIdentifier(12); err == nil {
		info.IdentifierMedium = id
	}

	// Get branch name
	if branch, err := activeVCS.GetBranchName(); err == nil {
		info.BranchName = branch
	}

	// Get commit date
	if date, err := activeVCS.GetCommitDate(); err == nil {
		info.CommitDate = date
	}

	// Get commits since tag
	if count, err := activeVCS.GetCommitsSinceTag(); err == nil {
		info.CommitsSinceTag = count
	}

	// Get uncommitted changes count
	if count, err := activeVCS.GetUncommittedChanges(); err == nil {
		info.UncommittedChanges = count
	}

	// Get version source hash (last tag's commit)
	if hash, err := activeVCS.GetLastTagCommit(); err == nil {
		info.VersionSourceHash = hash
	}

	// Get commit author
	if author, err := activeVCS.GetCommitAuthor(); err == nil {
		info.CommitAuthor = author
	}

	// Get commit author email
	if email, err := activeVCS.GetCommitAuthorEmail(); err == nil {
		info.CommitAuthorEmail = email
	}

	return info
}

// RenderTemplate renders a custom Mustache template with the given version
func RenderTemplate(tmplStr string, versionStr string) (string, error) {
	// Parse the version
	sv := version.Parse(versionStr)

	// Get VCS information and format fields
	vcsInfo := getVCSInfo()
	vcsFields := formatVCSFields(vcsInfo)
	buildTime := formatBuildTime()

	data := TemplateData{
		// Version components
		Major:                   sv.MajorString(),
		Minor:                   sv.MinorString(),
		Patch:                   sv.PatchString(),
		Revision:                sv.RevisionString(),
		MajorMinorPatch:         sv.CoreVersion(),
		MajorMinorPatchRevision: sv.CoreVersionWithRevision(),
		MajorMinor:              sv.MajorMinor(),
		Prefix:                  sv.Prefix,
		AssemblyVersion:         sv.AssemblyVersion(),

		// Pre-release components (from parsed version)
		PreReleaseLabel:  sv.PreReleaseLabel(),
		PreReleaseNumber: formatPreReleaseNumber(sv.PreReleaseNumber()),

		// VCS/Git info
		Hash:               vcsInfo.Identifier,
		ShortHash:          vcsInfo.IdentifierShort,
		MediumHash:         vcsInfo.IdentifierMedium,
		ShortHashWithDot:   withDotPrefix(vcsInfo.IdentifierShort),
		MediumHashWithDot:  withDotPrefix(vcsInfo.IdentifierMedium),
		ShortHashWithDash:  withDashPrefix(vcsInfo.IdentifierShort),
		MediumHashWithDash: withDashPrefix(vcsInfo.IdentifierMedium),
		BranchName:         vcsInfo.BranchName,
		EscapedBranchName:  version.EscapedBranchName(vcsInfo.BranchName),
		CommitsSinceTag:    vcsFields.CommitsSinceTag,
		BuildNumber:        vcsFields.CommitsSinceTag,
		BuildNumberPadded:  vcsFields.BuildNumberPadded,
		UncommittedChanges: vcsFields.UncommittedChanges,
		Dirty:              vcsFields.Dirty,
		DirtyWithDot:       vcsFields.DirtyWithDot,
		DirtyWithDash:      vcsFields.DirtyWithDash,
		VersionSourceHash:  vcsInfo.VersionSourceHash,

		// Commit author info
		CommitAuthor:      vcsInfo.CommitAuthor,
		CommitAuthorEmail: vcsInfo.CommitAuthorEmail,

		// Commit timestamps
		CommitDate:        vcsFields.CommitDate,
		CommitDateCompact: vcsFields.CommitDateCompact,
		CommitDateShort:   vcsFields.CommitDateShort,
		CommitYear:        vcsFields.CommitYear,
		CommitMonth:       vcsFields.CommitMonth,
		CommitDay:         vcsFields.CommitDay,

		// Build timestamps
		BuildDateTimeUTC:             buildTime.DateTime,
		BuildDateTimeCompact:         buildTime.DateTimeCompact,
		BuildDateTimeCompactWithDot:  buildTime.DateTimeCompactDot,
		BuildDateTimeCompactWithDash: buildTime.DateTimeCompactDash,
		BuildDateUTC:                 buildTime.DateOnly,
		BuildYear:                    buildTime.Year,
		BuildMonth:                   buildTime.Month,
		BuildDay:                     buildTime.Day,
	}

	result, err := mustache.Render(tmplStr, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return result, nil
}

// ValidateOutputPath checks if an output file path is valid
// Returns an error if:
// - The path is empty
// - The path points to an existing directory
// - The parent directory doesn't exist and can't be determined
func ValidateOutputPath(filepath string) error {
	if filepath == "" {
		return fmt.Errorf(ErrOutputPathEmpty)
	}

	// Check if path points to an existing directory
	info, err := os.Stat(filepath)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("output path '%s' %s", filepath, ErrOutputPathIsDirectory)
		}
		// File exists and is a file - that's OK (will be overwritten)
		return nil
	}

	// File doesn't exist - check if parent directory exists
	if !os.IsNotExist(err) {
		return fmt.Errorf("cannot access path '%s': %w", filepath, err)
	}

	// Check parent directory
	dir := filepath[:len(filepath)-len(filepath[strings.LastIndex(filepath, "/")+1:])]
	if dir == "" {
		dir = "."
	}
	if strings.Contains(filepath, "/") || strings.Contains(filepath, "\\") {
		// Has a directory component - check if parent exists
		parentInfo, parentErr := os.Stat(dir)
		if parentErr != nil {
			if os.IsNotExist(parentErr) {
				return fmt.Errorf("parent directory '%s' %s", dir, ErrParentDirNotExist)
			}
			return fmt.Errorf("cannot access parent directory '%s': %w", dir, parentErr)
		}
		if !parentInfo.IsDir() {
			return fmt.Errorf("parent path '%s' %s", dir, ErrParentNotDirectory)
		}
	}

	return nil
}

// WriteToFile writes the rendered output to a file
// Validates the path before writing
func WriteToFile(content, filepath string) error {
	if err := ValidateOutputPath(filepath); err != nil {
		return err
	}
	if err := os.WriteFile(filepath, []byte(content), FilePermission); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}
	return nil
}

// EmitToFile renders and writes the version to a file
func EmitToFile(format Format, version, filepath string) error {
	content, err := Render(format, version)
	if err != nil {
		return fmt.Errorf("failed to render version for %s: %w", format, err)
	}
	return WriteToFile(content, filepath)
}

// EmitTemplateToFile renders a custom template and writes to a file
func EmitTemplateToFile(tmplStr, version, filepath string) error {
	content, err := RenderTemplate(tmplStr, version)
	if err != nil {
		return err
	}
	return WriteToFile(content, filepath)
}

// GetEmbeddedTemplate returns the embedded template content for a given format.
// This is useful for users who want to see the default template for customization.
func GetEmbeddedTemplate(format Format) (string, error) {
	return getTemplate(format)
}

// BuildTemplateDataFromVersion creates TemplateData from Version
// This allows rendering templates directly from VERSION data
func BuildTemplateDataFromVersion(v *version.Version) TemplateData {
	// Get VCS information and format fields
	vcsInfo := getVCSInfo()
	vcsFields := formatVCSFields(vcsInfo)
	buildTime := formatBuildTime()

	return TemplateData{
		// Version components
		Major:                   strconv.Itoa(v.Major),
		Minor:                   strconv.Itoa(v.Minor),
		Patch:                   strconv.Itoa(v.Patch),
		Revision:                strconv.Itoa(v.Revision),
		MajorMinorPatch:         v.CoreVersion(),
		MajorMinorPatchRevision: v.CoreVersionWithRevision(),
		MajorMinor:              fmt.Sprintf("%d.%d", v.Major, v.Minor),
		Prefix:                  v.Prefix,
		AssemblyVersion:         v.AssemblyVersion(),

		// Pre-release components
		PreReleaseLabel:  v.PreReleaseLabel(),
		PreReleaseNumber: formatPreReleaseNumber(v.PreReleaseNumber()),

		// VCS/Git info
		Hash:               vcsInfo.Identifier,
		ShortHash:          vcsInfo.IdentifierShort,
		MediumHash:         vcsInfo.IdentifierMedium,
		ShortHashWithDot:   withDotPrefix(vcsInfo.IdentifierShort),
		MediumHashWithDot:  withDotPrefix(vcsInfo.IdentifierMedium),
		ShortHashWithDash:  withDashPrefix(vcsInfo.IdentifierShort),
		MediumHashWithDash: withDashPrefix(vcsInfo.IdentifierMedium),
		BranchName:         vcsInfo.BranchName,
		EscapedBranchName:  version.EscapedBranchName(vcsInfo.BranchName),
		CommitsSinceTag:    vcsFields.CommitsSinceTag,
		BuildNumber:        vcsFields.CommitsSinceTag,
		BuildNumberPadded:  vcsFields.BuildNumberPadded,
		UncommittedChanges: vcsFields.UncommittedChanges,
		Dirty:              vcsFields.Dirty,
		DirtyWithDot:       vcsFields.DirtyWithDot,
		DirtyWithDash:      vcsFields.DirtyWithDash,
		VersionSourceHash:  vcsInfo.VersionSourceHash,

		// Commit author info
		CommitAuthor:      vcsInfo.CommitAuthor,
		CommitAuthorEmail: vcsInfo.CommitAuthorEmail,

		// Commit timestamps
		CommitDate:        vcsFields.CommitDate,
		CommitDateCompact: vcsFields.CommitDateCompact,
		CommitDateShort:   vcsFields.CommitDateShort,
		CommitYear:        vcsFields.CommitYear,
		CommitMonth:       vcsFields.CommitMonth,
		CommitDay:         vcsFields.CommitDay,

		// Build timestamps
		BuildDateTimeUTC:             buildTime.DateTime,
		BuildDateTimeCompact:         buildTime.DateTimeCompact,
		BuildDateTimeCompactWithDot:  buildTime.DateTimeCompactDot,
		BuildDateTimeCompactWithDash: buildTime.DateTimeCompactDash,
		BuildDateUTC:                 buildTime.DateOnly,
		BuildYear:                    buildTime.Year,
		BuildMonth:                   buildTime.Month,
		BuildDay:                     buildTime.Day,
	}
}

// RenderTemplateWithData renders a Mustache template with TemplateData
func RenderTemplateWithData(tmplStr string, data TemplateData) (string, error) {
	// Convert to map to support custom variables
	dataMap := templateDataToMap(data)
	result, err := mustache.Render(tmplStr, dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	return result, nil
}

// templateDataToMap converts TemplateData to a map for Mustache rendering
// This allows custom variables to be used alongside built-in variables
func templateDataToMap(data TemplateData) map[string]interface{} {
	m := map[string]interface{}{
		// Version components
		"Major":                   data.Major,
		"Minor":                   data.Minor,
		"Patch":                   data.Patch,
		"Revision":                data.Revision,
		"MajorMinorPatch":         data.MajorMinorPatch,
		"MajorMinorPatchRevision": data.MajorMinorPatchRevision,
		"MajorMinor":              data.MajorMinor,
		"Prefix":                  data.Prefix,
		"AssemblyVersion":         data.AssemblyVersion,

		// Pre-release
		"PreRelease":         data.PreRelease,
		"PreReleaseWithDash": data.PreReleaseWithDash,
		"PreReleaseLabel":    data.PreReleaseLabel,
		"PreReleaseNumber":   data.PreReleaseNumber,

		// Metadata
		"Metadata":         data.Metadata,
		"MetadataWithPlus": data.MetadataWithPlus,

		// VCS/Git info
		"Hash":               data.Hash,
		"ShortHash":          data.ShortHash,
		"MediumHash":         data.MediumHash,
		"ShortHashWithDot":   data.ShortHashWithDot,
		"MediumHashWithDot":  data.MediumHashWithDot,
		"ShortHashWithDash":  data.ShortHashWithDash,
		"MediumHashWithDash": data.MediumHashWithDash,
		"BranchName":         data.BranchName,
		"EscapedBranchName":  data.EscapedBranchName,
		"CommitsSinceTag":    data.CommitsSinceTag,
		"BuildNumber":        data.BuildNumber,
		"BuildNumberPadded":  data.BuildNumberPadded,
		"UncommittedChanges": data.UncommittedChanges,
		"Dirty":              data.Dirty,
		"DirtyWithDot":       data.DirtyWithDot,
		"DirtyWithDash":      data.DirtyWithDash,
		"VersionSourceHash":  data.VersionSourceHash,

		// Commit author
		"CommitAuthor":      data.CommitAuthor,
		"CommitUser":        data.CommitAuthor, // Alias for CommitAuthor
		"CommitAuthorEmail": data.CommitAuthorEmail,
		"CommitUserEmail":   data.CommitAuthorEmail, // Alias for CommitAuthorEmail

		// Commit timestamps
		"CommitDate":          data.CommitDate,
		"CommitDateTime":      data.CommitDate, // Alias for CommitDate
		"CommitDateCompact":   data.CommitDateCompact,
		"CommitDateTimeCompact": data.CommitDateCompact, // Alias for CommitDateCompact
		"CommitDateShort":     data.CommitDateShort,
		"CommitYear":          data.CommitYear,
		"CommitMonth":         data.CommitMonth,
		"CommitDay":           data.CommitDay,

		// Build timestamps
		"BuildDateTimeUTC":             data.BuildDateTimeUTC,
		"BuildDateTimeCompact":         data.BuildDateTimeCompact,
		"BuildDateTimeCompactWithDot":  data.BuildDateTimeCompactWithDot,
		"BuildDateTimeCompactWithDash": data.BuildDateTimeCompactWithDash,
		"BuildDateUTC":                 data.BuildDateUTC,
		"BuildYear":                    data.BuildYear,
		"BuildMonth":                   data.BuildMonth,
		"BuildDay":                     data.BuildDay,
	}

	// Merge custom variables (they can override built-ins if desired)
	for k, v := range data.Custom {
		m[k] = v
	}

	// Merge plugin-provided variables
	// Pass ShortHash as context so plugins can create prefixed variants
	pluginVars := plugin.GetAllTemplateVariables(map[string]string{
		"ShortHash":  data.ShortHash,
		"MediumHash": data.MediumHash,
		"Hash":       data.Hash,
	})
	for k, v := range pluginVars {
		m[k] = v
	}

	// Also merge any explicitly set plugin variables from the data struct
	for k, v := range data.PluginVariables {
		m[k] = v
	}

	return m
}

// RenderTemplateList renders a list of template strings and joins with separator
func RenderTemplateList(templates []string, data TemplateData, sep string) string {
	var parts []string
	for _, tmpl := range templates {
		result, err := RenderTemplateWithData(tmpl, data)
		if err == nil {
			part := strings.TrimSpace(result)
			if part != "" {
				parts = append(parts, part)
			}
		}
	}
	return strings.Join(parts, sep)
}

// MergeCustomVars merges additional custom variables into TemplateData
// Command-line values override config values
func MergeCustomVars(data *TemplateData, extraVars map[string]string) {
	if data.Custom == nil {
		data.Custom = make(map[string]string)
	}
	for k, v := range extraVars {
		data.Custom[k] = v
	}
}

// BuildCompleteTemplateData builds TemplateData with PreRelease and Metadata populated
// prereleaseTemplate: Mustache template for PreRelease (use DASHES as separators)
// metadataTemplate: Mustache template for Metadata (use DOTS as separators)
func BuildCompleteTemplateData(v *version.Version, prereleaseTemplate, metadataTemplate string) TemplateData {
	// Build base template data
	data := BuildTemplateDataFromVersion(v)

	// Render PreRelease from template
	// IMPORTANT: The template should use DASHES (-) to separate identifiers per SemVer 2.0.0
	if prereleaseTemplate != "" {
		prerelease, err := RenderTemplateWithData(prereleaseTemplate, data)
		if err == nil {
			prerelease = strings.TrimSpace(prerelease)
			data.PreRelease = prerelease
			if prerelease != "" {
				data.PreReleaseWithDash = "-" + prerelease
			}
		}
	}

	// Render Metadata from template
	// IMPORTANT: The template should use DOTS (.) to separate identifiers per SemVer 2.0.0
	if metadataTemplate != "" {
		metadata, err := RenderTemplateWithData(metadataTemplate, data)
		if err == nil {
			metadata = strings.TrimSpace(metadata)
			data.Metadata = metadata
			if metadata != "" {
				data.MetadataWithPlus = "+" + metadata
			}
		}
	}

	return data
}
