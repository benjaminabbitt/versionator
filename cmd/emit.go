package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"
	"github.com/spf13/cobra"
)

var (
	emitOutput             string
	emitTemplate           string
	emitTemplateFile       string
	dumpOutput             string
	emitPrereleaseTemplate string
	emitMetadataTemplate   string
	emitPrefixOverride     string
)

var emitCmd = &cobra.Command{
	Use:   "emit [format]",
	Short: "Emit version in various formats",
	Long: `Emit the current version in various programming language formats.

Supported formats: ` + strings.Join(emit.SupportedFormats(), ", ") + `

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

TEMPLATE VARIABLES (Mustache syntax):

  Version Components:
    {{Major}}                - Major version (e.g., "1")
    {{Minor}}                - Minor version (e.g., "2")
    {{Patch}}                - Patch version (e.g., "3")
    {{MajorMinorPatch}}      - Core version: Major.Minor.Patch (e.g., "1.2.3")
    {{MajorMinor}}           - Major.Minor (e.g., "1.2")
    {{Prefix}}               - Version prefix (e.g., "v")

  Pre-release (rendered from --prerelease template):
    {{PreRelease}}           - Rendered pre-release (e.g., "alpha-5")
    {{PreReleaseWithDash}}   - With dash prefix (e.g., "-alpha-5") or empty

  Metadata (rendered from --metadata template):
    {{Metadata}}             - Rendered metadata (e.g., "20241211.abc1234")
    {{MetadataWithPlus}}     - With plus prefix (e.g., "+20241211.abc1234")

  VCS/Git Information:
    {{Hash}}                 - Full commit hash (40 chars for git)
    {{ShortHash}}            - Short commit hash (7 chars)
    {{MediumHash}}           - Medium commit hash (12 chars)
    {{BranchName}}           - Current branch (e.g., "feature/foo")
    {{EscapedBranchName}}    - Branch with slashes replaced (e.g., "feature-foo")
    {{CommitsSinceTag}}      - Commits since last tag (e.g., "42")
    {{BuildNumber}}          - Alias for CommitsSinceTag (GitVersion compatibility)
    {{BuildNumberPadded}}    - Padded to 4 digits (e.g., "0042")
    {{UncommittedChanges}}   - Count of dirty files (e.g., "3")
    {{Dirty}}                - "dirty" if uncommitted changes > 0, empty otherwise
    {{VersionSourceHash}}    - Hash of commit the last tag points to

  Commit Author:
    {{CommitAuthor}}         - Name of the commit author
    {{CommitAuthorEmail}}    - Email of the commit author

  Commit Timestamp (UTC):
    {{CommitDate}}           - ISO 8601: 2024-01-15T10:30:00Z
    {{CommitDateCompact}}    - Compact: 20240115103045 (YYYYMMDDHHmmss)
    {{CommitDateShort}}      - Date only: 2024-01-15

  Build Timestamp (UTC):
    {{BuildDateTimeUTC}}     - ISO 8601: 2024-01-15T10:30:00Z
    {{BuildDateTimeCompact}} - Compact: 20240115103045 (YYYYMMDDHHmmss)
    {{BuildDateUTC}}         - Date only: 2024-01-15
    {{BuildYear}}            - Year: 2024
    {{BuildMonth}}           - Month: 01 (zero-padded)
    {{BuildDay}}             - Day: 15 (zero-padded)

Use 'versionator vars' to see all template variables and their current values.

EXAMPLES:
  # Print Python version to stdout
  versionator emit python

  # With pre-release and metadata
  versionator emit python --prerelease "alpha" --metadata "{{ShortSha}}"

  # Use config defaults for prerelease/metadata
  versionator emit python --prerelease --metadata

  # Use custom template string
  versionator emit --template '{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}' \
    --prerelease "rc-1" --metadata "{{BuildDateTimeCompact}}"

  # Write to file
  versionator emit python --output mypackage/_version.py

  # Use template file
  versionator emit --template-file _version.tmpl.py --output _version.py

  # Dump a template for customization
  versionator emit dump python --output _version.tmpl.py`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load version data
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		// Handle prefix override
		var prefix string
		if cmd.Flags().Changed("prefix") {
			if emitPrefixOverride == useDefaultMarker {
				// Flag provided without value - use default "v"
				prefix = "v"
			} else {
				prefix = emitPrefixOverride
			}
		} else {
			// Use prefix from VERSION file
			prefix = vd.Prefix
		}

		// Handle prerelease template
		var prereleaseResult string
		if cmd.Flags().Changed("prerelease") {
			baseData := emit.BuildTemplateDataFromVersion(vd)
			if emitPrereleaseTemplate == useDefaultMarker {
				// Flag provided without value - use defaults from config
				template, _ := versionator.GetPreReleaseTemplate()
				if template != "" {
					prereleaseResult, err = emit.RenderTemplateWithData(template, baseData)
					if err != nil {
						return fmt.Errorf("error rendering prerelease template: %w", err)
					}
					prereleaseResult = strings.TrimSpace(prereleaseResult)
				}
			} else {
				// Render the provided template
				prereleaseResult, err = emit.RenderTemplateWithData(emitPrereleaseTemplate, baseData)
				if err != nil {
					return fmt.Errorf("error rendering prerelease template: %w", err)
				}
				prereleaseResult = strings.TrimSpace(prereleaseResult)
			}
		}

		// Handle metadata template
		var metadataResult string
		if cmd.Flags().Changed("metadata") {
			baseData := emit.BuildTemplateDataFromVersion(vd)
			if emitMetadataTemplate == useDefaultMarker {
				// Flag provided without value - use defaults from config
				template, _ := versionator.GetMetadataTemplate()
				if template != "" {
					metadataResult, err = emit.RenderTemplateWithData(template, baseData)
					if err != nil {
						return fmt.Errorf("error rendering metadata template: %w", err)
					}
					metadataResult = strings.TrimSpace(metadataResult)
				}
			} else {
				// Render the provided template
				metadataResult, err = emit.RenderTemplateWithData(emitMetadataTemplate, baseData)
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

		var content string
		var templateStr string

		// Check if using template file
		if emitTemplateFile != "" {
			data, err := os.ReadFile(emitTemplateFile)
			if err != nil {
				return fmt.Errorf("error reading template file: %w", err)
			}
			templateStr = string(data)
		} else if emitTemplate != "" {
			templateStr = emitTemplate
		}

		// Render content
		if templateStr != "" {
			content, err = emit.RenderTemplateWithData(templateStr, templateData)
			if err != nil {
				return fmt.Errorf("error rendering template: %w", err)
			}
		} else {
			// Require format argument if no template
			if len(args) == 0 {
				return fmt.Errorf("format argument required (or use --template/--template-file)\nSupported formats: %s", strings.Join(emit.SupportedFormats(), ", "))
			}

			format := emit.Format(args[0])
			if !emit.IsValidFormat(string(format)) {
				return fmt.Errorf("unsupported format '%s'\nSupported formats: %s", format, strings.Join(emit.SupportedFormats(), ", "))
			}

			// For built-in formats, use RenderTemplateWithData for consistency
			tmplStr, err := emit.GetEmbeddedTemplate(format)
			if err != nil {
				return fmt.Errorf("error getting template: %w", err)
			}
			content, err = emit.RenderTemplateWithData(tmplStr, templateData)
			if err != nil {
				return fmt.Errorf("error rendering format: %w", err)
			}
		}

		// Output to file or stdout
		if emitOutput != "" {
			if err := emit.WriteToFile(content, emitOutput); err != nil {
				return fmt.Errorf("error writing to file: %w", err)
			}
			fmt.Printf("Version %s written to %s\n", vd.CoreVersion(), emitOutput)
		} else {
			fmt.Print(content)
		}
		return nil
	},
}

var emitDumpCmd = &cobra.Command{
	Use:   "dump [format]",
	Short: "Dump embedded template to filesystem for customization",
	Long: `Dump the embedded template for a format to the filesystem.

This allows you to customize the template and use it with --template-file.

Supported formats: ` + strings.Join(emit.SupportedFormats(), ", ") + `

See 'versionator emit --help' for the full list of template variables.

Examples:
  # Print Python template to stdout
  versionator emit dump python

  # Save Python template to file for editing
  versionator emit dump python --output _version.tmpl.py

  # Then use your customized template
  versionator emit --template-file _version.tmpl.py --output _version.py`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := emit.Format(args[0])
		if !emit.IsValidFormat(string(format)) {
			return fmt.Errorf("unsupported format '%s'\nSupported formats: %s", format, strings.Join(emit.SupportedFormats(), ", "))
		}

		template, err := emit.GetEmbeddedTemplate(format)
		if err != nil {
			return fmt.Errorf("error getting template: %w", err)
		}

		// Output to file or stdout
		if dumpOutput != "" {
			if err := emit.WriteToFile(template, dumpOutput); err != nil {
				return fmt.Errorf("error writing to file: %w", err)
			}
			fmt.Printf("Template for '%s' written to %s\n", format, dumpOutput)
		} else {
			fmt.Print(template)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(emitCmd)
	emitCmd.AddCommand(emitDumpCmd)

	emitCmd.Flags().StringVarP(&emitOutput, "output", "o", "", "Output file path (default: stdout)")
	emitCmd.Flags().StringVarP(&emitTemplate, "template", "t", "", "Custom Mustache template string")
	emitCmd.Flags().StringVarP(&emitTemplateFile, "template-file", "f", "", "Path to template file")

	// Add prefix flag - optional value, defaults to "v" if no value provided
	emitCmd.Flags().StringVarP(&emitPrefixOverride, "prefix", "p", "", "Version prefix (default 'v' if flag provided without value)")
	emitCmd.Flag("prefix").NoOptDefVal = useDefaultMarker

	// Add prerelease flag - optional value, uses config defaults if no value provided
	emitCmd.Flags().StringVar(&emitPrereleaseTemplate, "prerelease", "", "Pre-release template (uses config default if flag provided without value)")
	emitCmd.Flag("prerelease").NoOptDefVal = useDefaultMarker

	// Add metadata flag - optional value, uses config defaults if no value provided
	emitCmd.Flags().StringVar(&emitMetadataTemplate, "metadata", "", "Metadata template (uses config default if flag provided without value)")
	emitCmd.Flag("metadata").NoOptDefVal = useDefaultMarker

	emitDumpCmd.Flags().StringVarP(&dumpOutput, "output", "o", "", "Output file path (default: stdout)")
}
