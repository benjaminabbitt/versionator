package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"

	"github.com/spf13/cobra"
)

var logOutput string
var versionTemplate string
var prereleaseTemplate string
var metadataTemplate string
var prefixOverride string
var setVars []string

// Marker for "flag provided without value" - use defaults
const useDefaultMarker = "\x00DEFAULT\x00"

var rootCmd = &cobra.Command{
	Use:   "versionator",
	Short: "A semantic version management tool",
	Long: `Versionator is a CLI tool for managing semantic versions.
It allows you to increment and decrement major, minor, and patch versions
stored in a VERSION file in the current directory.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// If log format wasn't explicitly set via flag, use config default
		if !cmd.PersistentFlags().Changed("log-format") {
			if cfg, err := config.ReadConfig(); err == nil {
				logOutput = cfg.Logging.Output
			}
		}

		// Initialize logger with the specified output format
		if err := logging.InitLogger(logOutput); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current version",
	Long: `Show the current version from VERSION file.

By default, outputs the full SemVer version (Major.Minor.Patch[-PreRelease][+Metadata]).

Use --template to customize the output format with Mustache syntax.

FLAGS WITH OPTIONAL VALUES (use = syntax for values, e.g., --prefix=value):
  --prefix, -p            Enable prefix (default "v" if no value given)
  --prefix="release-"     Use custom prefix value
  --prerelease            Enable pre-release with config defaults
  --prerelease="..."      Use custom template (YOU provide dash separators)
  --metadata              Enable metadata with config defaults
  --metadata="..."        Use custom template (YOU provide dot separators)

IMPORTANT - SEPARATOR CONVENTIONS (per SemVer 2.0.0):
  Pre-release: Components separated by DASHES (e.g., "alpha-1", "beta-{{CommitsSinceTag}}")
               The leading dash (-) is auto-prepended via {{PreReleaseWithDash}}
  Metadata:    Components separated by DOTS (e.g., "{{BuildDateTimeCompact}}.{{ShortSha}}")
               The leading plus (+) is auto-prepended via {{MetadataWithPlus}}

TEMPLATE VARIABLES:
  Version Components:
    {{Major}}            - Major version number
    {{Minor}}            - Minor version number
    {{Patch}}            - Patch version number
    {{MajorMinorPatch}}  - Major.Minor.Patch
    {{Prefix}}           - Version prefix (e.g., "v")

  Pre-release (rendered from --prerelease template):
    {{PreRelease}}         - Rendered pre-release (e.g., "alpha-5")
    {{PreReleaseWithDash}} - With dash prefix (e.g., "-alpha-5")

  Metadata (rendered from --metadata template):
    {{Metadata}}           - Rendered metadata (e.g., "20241211.abc1234")
    {{MetadataWithPlus}}   - With plus prefix (e.g., "+20241211.abc1234")

  VCS/Git (available in all templates):
    {{ShortHash}}        - Short commit hash (7 chars)
    {{MediumHash}}       - Medium commit hash (12 chars)
    {{Hash}}             - Full commit hash
    {{BranchName}}       - Current branch name
    {{CommitsSinceTag}}  - Commits since last tag
    {{BuildNumber}}      - Alias for CommitsSinceTag
    {{BuildNumberPadded}} - Padded to 4 digits (e.g., "0042")

  Commit Info:
    {{CommitDate}}       - Last commit datetime (ISO 8601)
    {{CommitDateCompact}} - Compact: 20241211103045
    {{CommitAuthor}}     - Commit author name
    {{CommitAuthorEmail}} - Commit author email

  Build Timestamps:
    {{BuildDateTimeCompact}} - Compact: 20241211103045
    {{BuildDateUTC}}         - Date only: 2024-12-11

  Custom Variables:
    Use --set key=value to inject custom variables
    Custom vars from .versionator.yaml config are also available

EXAMPLES:
  # Basic version (includes prerelease/metadata from VERSION file)
  versionator version                              # Output: 1.2.3-alpha+build.1

  # With prefix
  versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
                                                   # Output: v1.2.3

  # Full SemVer with prerelease and metadata
  versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
    --prerelease "alpha-{{CommitsSinceTag}}" \
    --metadata "{{BuildDateTimeCompact}}.{{ShortSha}}"
                                                   # Output: 1.2.3-alpha-5+20241211103045.abc1234

  # Use config defaults for prerelease/metadata
  versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
    --prerelease --metadata

  # With custom variables
  versionator version -t "{{AppName}} v{{MajorMinorPatch}}" --set AppName="My App"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading version: %w", err)
		}

		// Parse --set flags into a map
		extraVars := parseSetFlags(setVars)

		// If no template specified, output full SemVer (including prerelease and metadata from VERSION file)
		if versionTemplate == "" {
			fmt.Fprintln(cmd.OutOrStdout(), vd.String())
			return nil
		}

		// Handle prefix override
		var prefix string
		if cmd.Flags().Changed("prefix") {
			if prefixOverride == useDefaultMarker {
				// Flag provided without value - use default "v"
				prefix = "v"
			} else {
				prefix = prefixOverride
			}
		} else {
			// Use prefix from VERSION file
			prefix = vd.Prefix
		}

		// Handle prerelease
		var prereleaseResult string
		if cmd.Flags().Changed("prerelease") {
			if prereleaseTemplate == useDefaultMarker {
				// Flag provided without value - use config elements
				prereleaseResult, err = versionator.RenderPreRelease()
				if err != nil {
					return fmt.Errorf("error rendering prerelease: %w", err)
				}
			} else {
				// Render the provided template (legacy support)
				templateData := emit.BuildTemplateDataFromVersion(vd)
				prereleaseResult, err = emit.RenderTemplateWithData(prereleaseTemplate, templateData)
				if err != nil {
					return fmt.Errorf("error rendering prerelease template: %w", err)
				}
				prereleaseResult = strings.TrimSpace(prereleaseResult)
			}
		}

		// Handle metadata
		var metadataResult string
		if cmd.Flags().Changed("metadata") {
			if metadataTemplate == useDefaultMarker {
				// Flag provided without value - use config elements
				metadataResult, err = versionator.RenderMetadata()
				if err != nil {
					return fmt.Errorf("error rendering metadata: %w", err)
				}
			} else {
				// Render the provided template (legacy support)
				templateData := emit.BuildTemplateDataFromVersion(vd)
				metadataResult, err = emit.RenderTemplateWithData(metadataTemplate, templateData)
				if err != nil {
					return fmt.Errorf("error rendering metadata template: %w", err)
				}
				metadataResult = strings.TrimSpace(metadataResult)
			}
		}

		// Build template data with rendered prerelease and metadata
		templateData := emit.BuildTemplateDataFromVersion(vd)
		templateData.Prefix = prefix
		templateData.PreRelease = prereleaseResult
		if prereleaseResult != "" {
			templateData.PreReleaseWithDash = "-" + prereleaseResult
		}
		templateData.Metadata = metadataResult
		if metadataResult != "" {
			templateData.MetadataWithPlus = "+" + metadataResult
		}

		// Load custom vars from config
		configCustomVars, err := config.GetAllCustom()
		if err == nil && len(configCustomVars) > 0 {
			emit.MergeCustomVars(&templateData, configCustomVars)
		}

		// Merge command-line custom vars (override config custom vars)
		emit.MergeCustomVars(&templateData, extraVars)

		result, err := emit.RenderTemplateWithData(versionTemplate, templateData)
		if err != nil {
			return fmt.Errorf("error rendering template: %w", err)
		}

		// Trim any trailing newlines that might be added by template rendering
		fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(result))
		return nil
	},
}

func init() {
	// Add persistent flag for log output format
	rootCmd.PersistentFlags().StringVar(&logOutput, "log-format", "console", "Log output format (console, json, development)")

	// Add template flag to version command
	versionCmd.Flags().StringVarP(&versionTemplate, "template", "t", "", "Template string for version output (Mustache syntax)")

	// Add prefix flag - optional value, defaults to "v" if no value provided
	versionCmd.Flags().StringVarP(&prefixOverride, "prefix", "p", "", "Version prefix (default 'v' if flag provided without value)")
	versionCmd.Flag("prefix").NoOptDefVal = useDefaultMarker

	// Add prerelease flag - optional value, uses config defaults if no value provided
	versionCmd.Flags().StringVar(&prereleaseTemplate, "prerelease", "", "Pre-release template (uses config default if flag provided without value)")
	versionCmd.Flag("prerelease").NoOptDefVal = useDefaultMarker

	// Add metadata flag - optional value, uses config defaults if no value provided
	versionCmd.Flags().StringVar(&metadataTemplate, "metadata", "", "Metadata template (uses config default if flag provided without value)")
	versionCmd.Flag("metadata").NoOptDefVal = useDefaultMarker

	// Add --set flag for custom variables (can be used multiple times)
	versionCmd.Flags().StringArrayVar(&setVars, "set", nil, "Set custom variable (key=value), can be repeated")

	// Add version command to show current version
	rootCmd.AddCommand(versionCmd)
}

// parseSetFlags parses --set key=value flags into a map
func parseSetFlags(setFlags []string) map[string]string {
	result := make(map[string]string)
	for _, s := range setFlags {
		if idx := strings.Index(s, "="); idx > 0 {
			key := s[:idx]
			value := s[idx+1:]
			result[key] = value
		}
	}
	return result
}
