package cmd

import (
	"fmt"
	"strings"

	versioncmd "github.com/benjaminabbitt/versionator/cmd/version"
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"
	"github.com/spf13/cobra"
)

var (
	versionTemplate    string
	prefixOverride     string
	prereleaseTemplate string
	metadataTemplate   string
	setVars            []string
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"ver"},
	Short:   "Show or manage version",
	Long: `Show the current version or manage version components.

When called without subcommands, shows the current version from VERSION file.
Use subcommands to increment/decrement version components.

Subcommands:
  major     - Manage major version component
  minor     - Manage minor version component
  patch     - Manage patch version component
  revision  - Manage revision version component (4th component for .NET)

By default, outputs the full SemVer version (Major.Minor.Patch[-PreRelease][+Metadata]).

Use --template to customize the output format with Mustache syntax.

FLAGS WITH OPTIONAL VALUES (use = syntax for values, e.g., --prefix=value):
  --prefix, -p            Enable prefix (default "v" if no value given)
  --prefix="release-"     Use custom prefix value
  --prerelease            Enable pre-release with config defaults
  --prerelease="..."      Use custom template (YOU provide dash separators)
  --metadata              Enable metadata with config defaults
  --metadata="..."        Use custom template (YOU provide dot separators)

EXAMPLES:
  # Show current version
  versionator version                              # Output: 1.2.3-alpha+build.1

  # Increment major version
  versionator version major inc

  # Increment minor version
  versionator version minor inc

  # Increment patch version
  versionator version patch inc

  # With template
  versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
                                                   # Output: v1.2.3`,
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
	rootCmd.AddCommand(versionCmd)

	// Add version subcommands from subpackage
	versionCmd.AddCommand(versioncmd.MajorCmd)
	versionCmd.AddCommand(versioncmd.MinorCmd)
	versionCmd.AddCommand(versioncmd.PatchCmd)
	versionCmd.AddCommand(versioncmd.RevisionCmd)
	versionCmd.AddCommand(versioncmd.RenderCmd)
	versionCmd.AddCommand(versioncmd.SyncCmd)
	versionCmd.AddCommand(versioncmd.SetCmd)

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

	// Add custom variable flag
	versionCmd.Flags().StringArrayVar(&setVars, "set", nil, "Set custom variable (key=value), can be repeated")
}
