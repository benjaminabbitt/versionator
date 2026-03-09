package config

import (
	"fmt"
	"os"

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
	Custom           map[string]string      `yaml:"custom,omitempty"`
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

// ReleaseConfig holds release-related configuration
type ReleaseConfig struct {
	// CreateBranch controls whether a release branch is created when tagging
	// Default: true
	CreateBranch bool `yaml:"createBranch"`
	// BranchPrefix is prepended to the tag name to form the branch name
	// Default: "release/" (e.g., tag "v1.2.3" -> branch "release/v1.2.3")
	BranchPrefix string `yaml:"branchPrefix"`
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

// LoggingConfig holds logging-specific configuration
type LoggingConfig struct {
	Output string `yaml:"output"` // console, json, development
}

// ReadConfig reads the configuration from .versionator.yaml file
func ReadConfig() (*Config, error) {
	config := &Config{
		Prefix: "v", // default prefix
		PreRelease: PreReleaseConfig{
			Template: "", // default empty - templates must be explicitly configured
			Stable:   false,                       // default: generated at output time (CD workflow)
		},
		Metadata: MetadataConfig{
			Template: "", // default empty - templates must be explicitly configured
			Stable:   false,           // default: generated at output time
			Git: GitConfig{
				HashLength: 12, // default hash length for MediumHash
			},
		},
		Release: ReleaseConfig{
			CreateBranch: true,       // create release branches by default
			BranchPrefix: "release/", // e.g., "release/v1.2.3"
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
	if c.BranchVersioning.Mode != "" && c.BranchVersioning.Mode != "replace" && c.BranchVersioning.Mode != "append" {
		return fmt.Errorf("branch versioning mode must be 'replace' or 'append', got '%s'", c.BranchVersioning.Mode)
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

// isValidTemplateKey checks if a key is valid for use in templates
func isValidTemplateKey(key string) bool {
	if len(key) == 0 {
		return false
	}

	first := rune(key[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}

	for _, c := range key {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
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
  branchPrefix: "release/"

# Logging configuration
logging:
  # Output format: console, json, development
  output: "console"

# =============================================================================
# AVAILABLE TEMPLATE VARIABLES
# =============================================================================
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
#   {{UncommittedChanges}}           - Count of dirty files
#   {{Dirty}}                        - "dirty" if uncommitted changes
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
