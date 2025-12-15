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
	Prefix     string            `yaml:"prefix"`
	PreRelease PreReleaseConfig  `yaml:"prerelease"`
	Metadata   MetadataConfig    `yaml:"metadata"`
	Logging    LoggingConfig     `yaml:"logging"`
	Custom     map[string]string `yaml:"custom,omitempty"`
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
// The VERSION file is the source of truth for current pre-release value.
// This template is stored for use with 'prerelease enable' and '--prerelease' flag.
type PreReleaseConfig struct {
	Template string `yaml:"template"` // Mustache template with DASHES as separators: "alpha-{{CommitsSinceTag}}" → "alpha-5"
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
// The VERSION file is the source of truth for current metadata value.
// This template is stored for use with 'metadata enable' and '--metadata' flag.
type MetadataConfig struct {
	Template string    `yaml:"template"` // Mustache template with DOTS as separators: "{{BuildDateTimeCompact}}.{{ShortHash}}" → "20241211.abc1234"
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
			Template: "", // empty by default, user must configure
		},
		Metadata: MetadataConfig{
			Template: "", // empty by default, user must configure
			Git: GitConfig{
				HashLength: 12, // default hash length for MediumHash
			},
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
# Example output: 1.2.3-alpha-5
prerelease:
  # Set to true to enable pre-release in version output
  enabled: false

  # Mustache template string for pre-release identifier
  # IMPORTANT: Use DASHES (-) to separate identifiers per SemVer 2.0.0
  # Example: "alpha-{{CommitsSinceTag}}" → "alpha-5"
  # The leading dash is added automatically - do NOT include it here
  template: "alpha-{{CommitsSinceTag}}"

# Build metadata configuration
# Metadata follows SemVer 2.0.0: appended with plus (+)
# Example output: 1.2.3+20241211103045.abc1234def5
metadata:
  # Set to true to enable metadata in version output
  enabled: false

  # Metadata type (currently only "git" supported)
  type: "git"

  # Mustache template string for build metadata
  # IMPORTANT: Use DOTS (.) to separate identifiers per SemVer 2.0.0
  # Example: "{{BuildDateTimeCompact}}.{{MediumHash}}" → "20241211103045.abc1234def5"
  # The leading plus is added automatically - do NOT include it here
  template: "{{BuildDateTimeCompact}}.{{MediumHash}}"

  # Git-specific configuration
  git:
    # Length of commit hash for MediumHash variable
    hashLength: 12

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
# Pre-release (from VERSION file):
#   {{PreRelease}}                   - Full pre-release string
#   {{PreReleaseWithDash}}           - With dash prefix (-alpha.5)
#   {{PreReleaseLabel}}              - Label only (alpha from alpha.5)
#   {{PreReleaseNumber}}             - Number only (5 from alpha.5)
#
# Metadata (from VERSION file):
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
