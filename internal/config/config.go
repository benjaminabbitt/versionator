package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/cbroglie/mustache"
	"gopkg.in/yaml.v3"
)

const configFile = ".versionator.yaml"

// Config holds configuration for version metadata behavior
type Config struct {
	Prefix           string                 `yaml:"prefix"`
	PreRelease       PreReleaseConfig       `yaml:"prerelease"`
	Metadata         MetadataConfig         `yaml:"metadata"`
	Release          ReleaseConfig          `yaml:"release"`
	BranchVersioning BranchVersioningConfig `yaml:"branchVersioning"`
	Logging          LoggingConfig          `yaml:"logging"`
	Dirty            DirtyConfig            `yaml:"dirty,omitempty"`
	Custom           map[string]string      `yaml:"custom,omitempty"`
	Updates          []UpdateConfig         `yaml:"updates,omitempty"`
}

// BranchVersioningConfig holds branch-aware versioning configuration
type BranchVersioningConfig struct {
	// Enabled controls whether branch-aware versioning is active
	// Default: false (opt-in)
	Enabled bool `yaml:"enabled"`
	// MainBranches is a list of branch patterns that produce clean versions
	// Supports exact matches and glob patterns (e.g., "release/*")
	// Default: ["main", "master", "release/*"]
	MainBranches []string `yaml:"mainBranches"`
	// PrereleaseTemplate is a Mustache template for the branch pre-release identifier
	// Default: "{{EscapedBranchName}}-{{CommitsSinceTag}}"
	PrereleaseTemplate string `yaml:"prereleaseTemplate"`
	// Mode controls how branch pre-release interacts with existing pre-release
	// "replace" (default): Branch pre-release replaces any existing pre-release
	// "append": Branch pre-release is appended to existing pre-release
	Mode string `yaml:"mode"`
}

// Release template defaults.
//
// The tag default reproduces the historical tag ("v1.2.3") with one deliberate
// exception: build metadata is omitted. Go module versions reject SemVer build
// metadata, so a tag carrying "+" cannot be resolved as a module version. The
// metadata still reaches the tag message, and a project that wants it in the
// tag can append {{MetadataWithPlus}} to its own template.
//
// The branch default is expressed in terms of {{BranchPrefix}} so that the
// existing branchPrefix setting keeps working untouched; a project that sets
// branchTemplate explicitly takes full control of the branch name.
const (
	DefaultTagTemplate     = "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}"
	DefaultBranchTemplate  = "{{BranchPrefix}}{{Tag}}"
	DefaultMessageTemplate = "Release {{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"
)

// ReleaseConfig holds release-related configuration
type ReleaseConfig struct {
	// CreateBranch controls whether a release branch is created when tagging
	// Default: true
	CreateBranch bool `yaml:"createBranch"`
	// BranchPrefix is prepended to the tag name to form the branch name
	// Default: "release/" (e.g., tag "v1.2.3" -> branch "release/v1.2.3")
	// Available to BranchTemplate as {{BranchPrefix}}.
	BranchPrefix string `yaml:"branchPrefix"`
	// TagTemplate is a Mustache template for the tag name.
	// Default: DefaultTagTemplate.
	//
	// Beyond the usual template variables, {{ModulePath}} renders the VERSION
	// file's directory relative to the repository root with a trailing slash
	// (empty at the root), and {{ModuleName}} renders that directory's bare
	// name. {{ModulePath}} is what Go's module system requires of a submodule
	// tag:
	//
	//	tagTemplate: "{{ModulePath}}{{Prefix}}{{MajorMinorPatch}}"
	//	  -> internal/tools/v1.2.3 in internal/tools, v1.2.3 at the root
	//
	// A literal path works equally well when the layout is fixed:
	//
	//	tagTemplate: "internal/tools/{{Prefix}}{{MajorMinorPatch}}"
	TagTemplate string `yaml:"tagTemplate"`
	// BranchTemplate is a Mustache template for the release branch name.
	// Default: DefaultBranchTemplate. In addition to the usual variables,
	// {{Tag}} holds the already-rendered tag name and {{BranchPrefix}} holds
	// the BranchPrefix setting.
	BranchTemplate string `yaml:"branchTemplate"`
	// MessageTemplate is a Mustache template for the annotated tag message.
	// Default: DefaultMessageTemplate. {{Tag}} is available here too.
	// The --message flag overrides this for a single invocation.
	MessageTemplate string `yaml:"messageTemplate"`
}

// PreReleaseConfig holds pre-release identifier configuration
// Pre-release follows SemVer 2.0.0: appended with - (dash)
//
// IMPORTANT: The Template is a Mustache template string.
// Use DASHES (-) to separate pre-release identifiers per SemVer 2.0.0.
// Example: "alpha-{{CommitsSinceTag}}" → "alpha-5"
//
// The leading dash (-) is automatically prepended when using {{PreReleaseWithDash}}
// Do NOT include the leading dash in your template.
//
// Stability controls where the pre-release value lives:
//   - Stable=true: Value is written to VERSION file (traditional release workflow)
//   - Stable=false: Value is generated from template at output time (default, CD workflow)
type PreReleaseConfig struct {
	Template string `yaml:"template"` // Mustache template with DASHES as separators: "alpha-{{CommitsSinceTag}}" → "alpha-5"
	Stable   bool   `yaml:"stable"`   // If true, value is written to VERSION file; if false, generated at output time
}

// MetadataConfig holds build metadata configuration
// Metadata follows SemVer 2.0.0: appended with + (plus)
//
// IMPORTANT: The Template is a Mustache template string.
// Use DOTS (.) to separate metadata identifiers per SemVer 2.0.0.
// Example: "{{BuildDateTimeCompact}}.{{ShortHash}}" → "20241211103045.abc1234"
//
// The leading plus (+) is automatically prepended when using {{MetadataWithPlus}}
// Do NOT include the leading plus in your template.
//
// Stability controls where the metadata value lives:
//   - Stable=true: Value is written to VERSION file
//   - Stable=false: Value is generated from template at output time (default)
type MetadataConfig struct {
	Template string    `yaml:"template"` // Mustache template with DOTS as separators: "{{BuildDateTimeCompact}}.{{ShortHash}}" → "20241211.abc1234"
	Stable   bool      `yaml:"stable"`   // If true, value is written to VERSION file; if false, generated at output time
	Git      GitConfig `yaml:"git"`
}

// GitConfig holds git-specific configuration
type GitConfig struct {
	HashLength int `yaml:"hashLength"`
}

// DirtyConfig holds what counts as a dirty working tree for the dirty-derived
// template variables ({{Dirty}}, {{UncommittedChanges}}, {{DateTimeDirty}}) and
// the ci DIRTY variable.
type DirtyConfig struct {
	// IgnoreUntracked leaves untracked files out, so a tree is dirty only for
	// staged or unstaged changes to tracked files (as `git describe --dirty`).
	// Default: false (untracked files count). The --ignore-untracked flag
	// overrides it. The release command's clean-tree check follows it too:
	// untracked-only trees can release, tracked changes still refuse.
	IgnoreUntracked bool `yaml:"ignoreUntracked"`
}

// LoggingConfig holds logging-specific configuration
type LoggingConfig struct {
	Output string `yaml:"output"` // console, json, development
}

// UpdateConfig holds configuration for a single structured file update
// Updates are applied during release to keep manifest files in sync with VERSION
type UpdateConfig struct {
	// File is the path to the file to update (relative to repo root)
	File string `yaml:"file"`
	// Path is the dasel selector for the value to update (e.g., "package.version")
	Path string `yaml:"path"`
	// Template is a Mustache template for the new value (e.g., "{{MajorMinorPatch}}")
	Template string `yaml:"template"`
	// Format explicitly sets the file format (json, yaml, toml). Auto-detected from extension if empty.
	Format string `yaml:"format,omitempty"`
}

// ReadConfig reads the configuration from .versionator.yaml file
func ReadConfig() (*Config, error) {
	config := &Config{
		Prefix: "v", // default prefix
		PreRelease: PreReleaseConfig{
			Template: "",    // default empty - templates must be explicitly configured
			Stable:   false, // default: generated at output time (CD workflow)
		},
		Metadata: MetadataConfig{
			Template: "",    // default empty - templates must be explicitly configured
			Stable:   false, // default: generated at output time
			Git: GitConfig{
				HashLength: 12, // default hash length for MediumHash
			},
		},
		Release: ReleaseConfig{
			CreateBranch:    true,       // create release branches by default
			BranchPrefix:    "release/", // e.g., "release/v1.2.3"
			TagTemplate:     DefaultTagTemplate,
			BranchTemplate:  DefaultBranchTemplate,
			MessageTemplate: DefaultMessageTemplate,
		},
		BranchVersioning: BranchVersioningConfig{
			Enabled:            false, // opt-in
			MainBranches:       []string{"main", "master", "release/*"},
			PrereleaseTemplate: "{{EscapedBranchName}}-{{CommitsSinceTag}}",
			Mode:               "replace",
		},
		Logging: LoggingConfig{
			Output: "console", // default to human-readable console output
		},
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return default config
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// ValidateTemplate checks if a Mustache template is syntactically valid
func ValidateTemplate(template string) error {
	if template == "" {
		return nil
	}
	_, err := mustache.ParseString(template)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}
	return nil
}

// validateReleaseTemplates checks the release tag, branch and message templates.
//
// Templates are validated on write rather than at tag time because release
// mutates the repository — committing VERSION, creating a branch — before it
// reaches the tag. A template that only fails then leaves a half-finished
// release behind.
func (c *Config) validateReleaseTemplates() error {
	templates := []struct {
		name  string
		value string
	}{
		{"release tag template", c.Release.TagTemplate},
		{"release branch template", c.Release.BranchTemplate},
		{"release message template", c.Release.MessageTemplate},
	}
	for _, t := range templates {
		if err := ValidateTemplate(t.value); err != nil {
			return fmt.Errorf("%s: %w", t.name, err)
		}
	}

	// An empty tag template renders an empty tag name, which git rejects well
	// after the release is underway. An unset template is fine — the default
	// applies — but one set to whitespace is a mistake worth naming.
	if c.Release.TagTemplate != "" && strings.TrimSpace(c.Release.TagTemplate) == "" {
		return fmt.Errorf("release tag template: must not be blank")
	}
	return nil
}

// Validate checks if the config is valid, including template syntax
func (c *Config) Validate() error {
	if c.PreRelease.Template != "" {
		if err := ValidateTemplate(c.PreRelease.Template); err != nil {
			return fmt.Errorf("prerelease template: %w", err)
		}
	}
	if c.Metadata.Template != "" {
		if err := ValidateTemplate(c.Metadata.Template); err != nil {
			return fmt.Errorf("metadata template: %w", err)
		}
	}
	if c.BranchVersioning.PrereleaseTemplate != "" {
		if err := ValidateTemplate(c.BranchVersioning.PrereleaseTemplate); err != nil {
			return fmt.Errorf("branch versioning prerelease template: %w", err)
		}
	}
	if err := c.validateReleaseTemplates(); err != nil {
		return err
	}
	if c.BranchVersioning.Mode != "" && c.BranchVersioning.Mode != "replace" && c.BranchVersioning.Mode != "append" {
		return fmt.Errorf("branch versioning mode must be 'replace' or 'append', got '%s'", c.BranchVersioning.Mode)
	}
	for i, update := range c.Updates {
		if update.File == "" {
			return fmt.Errorf("updates[%d]: file is required", i)
		}
		if update.Path == "" {
			return fmt.Errorf("updates[%d]: path is required", i)
		}
		if update.Template == "" {
			return fmt.Errorf("updates[%d]: template is required", i)
		}
		if err := ValidateTemplate(update.Template); err != nil {
			return fmt.Errorf("updates[%d] template: %w", i, err)
		}
		if update.Format != "" {
			validFormats := map[string]bool{"json": true, "yaml": true, "toml": true}
			if !validFormats[update.Format] {
				return fmt.Errorf("updates[%d]: format must be 'json', 'yaml', or 'toml', got '%s'", i, update.Format)
			}
		}
	}
	return nil
}

func WriteConfig(config *Config) error {
	// Validate config before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add a comment header
	content := "# Versionator Configuration\n" + string(data)
	return os.WriteFile(configFile, []byte(content), FilePermission)
}

// Snapshot records the config file's exact bytes (or that there is none) and
// returns a function that puts them back. A command that writes the config
// before a version change that may be refused restores it with this, so the
// refusal leaves the config as it was, byte for byte.
func Snapshot() (func() error, error) {
	data, err := os.ReadFile(configFile)
	if os.IsNotExist(err) {
		return func() error {
			if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return func() error { return os.WriteFile(configFile, data, FilePermission) }, nil
}

// SetCustom sets a custom key-value pair in the config
func SetCustom(key, value string) error {
	if key == "" {
		return fmt.Errorf("custom variable key cannot be empty")
	}

	if !isValidTemplateKey(key) {
		return fmt.Errorf("invalid custom variable key '%s': must be alphanumeric (starting with letter, may contain underscores)", key)
	}

	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	if cfg.Custom == nil {
		cfg.Custom = make(map[string]string)
	}
	cfg.Custom[key] = value

	return WriteConfig(cfg)
}

// GetCustom returns a custom value by key
func GetCustom(key string) (string, bool, error) {
	cfg, err := ReadConfig()
	if err != nil {
		return "", false, err
	}

	if cfg.Custom == nil {
		return "", false, nil
	}
	value, ok := cfg.Custom[key]
	return value, ok, nil
}

// GetAllCustom returns all custom key-value pairs
func GetAllCustom() (map[string]string, error) {
	cfg, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	if cfg.Custom == nil {
		return make(map[string]string), nil
	}
	return cfg.Custom, nil
}

// DeleteCustom removes a custom key from the config
func DeleteCustom(key string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	if cfg.Custom != nil {
		delete(cfg.Custom, key)
	}

	return WriteConfig(cfg)
}

// isASCIILetter reports whether r is an unaccented A-Z or a-z. Template keys are
// restricted to ASCII deliberately, so a key resolves identically regardless of
// the host's locale or normalization form.
func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isTemplateKeyChar reports whether r may appear after a template key's first
// character.
func isTemplateKeyChar(r rune) bool {
	return isASCIILetter(r) || (r >= '0' && r <= '9') || r == '_'
}

// isValidTemplateKey checks if a key is valid for use in templates
func isValidTemplateKey(key string) bool {
	if len(key) == 0 {
		return false
	}

	if !isASCIILetter(rune(key[0])) {
		return false
	}

	for _, c := range key {
		if !isTemplateKeyChar(c) {
			return false
		}
	}
	return true
}

// DefaultConfigYAML returns a well-documented default configuration as YAML
func DefaultConfigYAML() string {
	return `# Versionator Configuration
# See https://github.com/benjaminabbitt/versionator for documentation

# Version prefix (e.g., "v" for v1.2.3)
# Set to empty string for no prefix
prefix: "v"

# Pre-release configuration
# Pre-release follows SemVer 2.0.0: appended with dash (-)
# Example output: 1.2.3-build-5
prerelease:
  # Mustache template string for pre-release identifier
  # IMPORTANT: Use DASHES (-) to separate identifiers per SemVer 2.0.0
  # Example: "build-{{CommitsSinceTag}}" → "build-5"
  # The leading dash is added automatically - do NOT include it here
  # Default is empty - set a template to enable dynamic pre-release
  template: ""

  # Stability controls where the pre-release value lives:
  #   stable: true  - Value is written to VERSION file (traditional release workflow)
  #   stable: false - Value is generated from template at output time (default, CD workflow)
  # When stable is false, templates are re-evaluated on every output command.
  stable: false

# Build metadata configuration
# Metadata follows SemVer 2.0.0: appended with plus (+)
# Example output: 1.2.3+abc1234
metadata:
  # Mustache template string for build metadata
  # IMPORTANT: Use DOTS (.) to separate identifiers per SemVer 2.0.0
  # Example: "{{BuildDateTimeCompact}}.{{MediumHash}}" → "20241211103045.abc1234def5"
  # The leading plus is added automatically - do NOT include it here
  # Default is empty - set a template to enable dynamic metadata
  template: ""

  # Stability controls where the metadata value lives:
  #   stable: true  - Value is written to VERSION file
  #   stable: false - Value is generated from template at output time (default)
  # Metadata is almost always dynamic (commit hash, build time), so default is false.
  stable: false

  # Git-specific configuration
  git:
    # Length of commit hash for MediumHash variable
    hashLength: 12

# Release configuration
release:
  # Create a release branch when tagging (default: true)
  createBranch: true

  # Branch name prefix (default: "release/")
  # Tag "v1.2.3" -> Branch "release/v1.2.3"
  # Available to branchTemplate as {{BranchPrefix}}
  branchPrefix: "release/"

  # Mustache template for the tag name.
  # Build metadata is deliberately absent: Go module versions reject SemVer
  # build metadata, so "v1.2.3+build.5" cannot be resolved as a module version.
  # Add {{MetadataWithPlus}} if your project wants it in the tag anyway.
  tagTemplate: "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}"

  # Mustache template for the release branch name.
  # {{Tag}} holds the already-rendered tag name.
  branchTemplate: "{{BranchPrefix}}{{Tag}}"

  # Mustache template for the annotated tag message.
  # The --message flag overrides this for a single invocation.
  messageTemplate: "Release {{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"

  # Go requires a submodule's tag to be the module's directory path relative to
  # the repository root, followed by the version. {{ModulePath}} produces it,
  # and is empty at the repository root, so one template serves both:
  #
  #   tagTemplate: "{{ModulePath}}{{Prefix}}{{MajorMinorPatch}}"
  #     internal/tools -> internal/tools/v1.2.3
  #     repository root -> v1.2.3
  #
  # A literal path works too when the layout is fixed:
  #
  #   tagTemplate: "internal/tools/{{Prefix}}{{MajorMinorPatch}}"

# Logging configuration
logging:
  # Output format: console, json, development
  output: "console"

# What makes the working tree dirty for {{Dirty}}, {{UncommittedChanges}},
# {{DateTimeDirty}} and the ci DIRTY variable
dirty:
  # true: only staged or unstaged changes to tracked files count (the meaning of
  # git describe --dirty); untracked files do not. Default false: they count.
  # The --ignore-untracked flag overrides this. The release command's
  # clean-tree check follows it too.
  ignoreUntracked: false

# File updates: keep manifest versions in sync with VERSION. Applied by every
# command that changes VERSION (set, bump, config prerelease/metadata/prefix,
# release). Only the value at path is rewritten; the rest of the file is kept
# byte for byte. All updates are checked before VERSION is written, so a refused
# one changes nothing. A Cargo.toml version must be SemVer 2.0.0 (no "v" prefix, no
# 4th component): such a value is refused, not stripped.
# updates:
#   - file: Cargo.toml
#     path: workspace.package.version   # or package.version for one crate
#     template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}"

# =============================================================================
# AVAILABLE TEMPLATE VARIABLES
# =============================================================================
#
# Release-only Variables (tag, branch and message templates):
#   {{ModulePath}}                   - VERSION file's directory relative to the
#                                      repository root, with a trailing slash;
#                                      empty at the root (internal/tools/)
#   {{ModuleName}}                   - That directory's bare name (tools)
#   {{Tag}}                          - The rendered tag name; available to the
#                                      branch and message templates
#   {{BranchPrefix}}                 - The release.branchPrefix setting
#
# Version Components:
#   {{Major}}, {{Minor}}, {{Patch}}  - Version numbers
#   {{MajorMinorPatch}}              - Core version (1.2.3)
#   {{MajorMinor}}                   - Major.Minor (1.2)
#   {{Prefix}}                       - Version prefix (v)
#
# Pre-release (rendered from template or VERSION file depending on stable flag):
#   {{PreRelease}}                   - Full pre-release string
#   {{PreReleaseWithDash}}           - With dash prefix (-alpha.5)
#   {{PreReleaseLabel}}              - Label only (alpha from alpha.5)
#   {{PreReleaseNumber}}             - Number only (5 from alpha.5)
#
# Metadata (rendered from template or VERSION file depending on stable flag):
#   {{Metadata}}                     - Full metadata string
#   {{MetadataWithPlus}}             - With plus prefix (+build.123)
#
# VCS/Git Information:
#   {{Hash}}                         - Full commit hash (40 chars for git)
#   {{ShortHash}}                    - Short hash (7 chars)
#   {{MediumHash}}                   - Medium hash (12 chars)
#   {{BranchName}}                   - Current branch name
#   {{EscapedBranchName}}            - Branch with / replaced by -
#   {{CommitsSinceTag}}              - Commits since last tag
#   {{BuildNumber}}                  - Alias for CommitsSinceTag
#   {{BuildNumberPadded}}            - Padded to 4 digits (0042)
#   {{UncommittedChanges}}           - Count of dirty files (see dirty:)
#   {{Dirty}}                        - "dirty" if uncommitted changes (see dirty:)
#   {{VersionSourceHash}}            - Hash of last tag's commit
#
# Commit Author:
#   {{CommitAuthor}}                 - Commit author name
#   {{CommitAuthorEmail}}            - Commit author email
#
# Commit Timestamps (UTC):
#   {{CommitDate}}                   - ISO 8601 (2024-01-15T10:30:00Z)
#   {{CommitDateCompact}}            - Compact (20240115103045)
#   {{CommitDateShort}}              - Date only (2024-01-15)
#   {{CommitYear}}, {{CommitMonth}}, {{CommitDay}}
#
# Build Timestamps (UTC):
#   {{BuildDateTimeUTC}}             - ISO 8601 (2024-01-15T10:30:00Z)
#   {{BuildDateTimeCompact}}         - Compact (20240115103045)
#   {{BuildDateUTC}}                 - Date only (2024-01-15)
#   {{BuildYear}}, {{BuildMonth}}, {{BuildDay}}
#
# Plugin Variables (git plugin):
#   {{GitShortHash}}                 - Prefixed short hash (git.abc1234)
#   {{ShaShortHash}}                 - Prefixed short hash (sha.abc1234)
`
}
