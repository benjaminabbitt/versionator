package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var prereleaseCmd = &cobra.Command{
	Use:   "prerelease",
	Short: "Manage version pre-release behavior",
	Long: `Commands to enable or disable pre-release identifiers.

Pre-release follows SemVer 2.0.0 specification:
- Appended with a dash (-) - this is added automatically
- Use DASHES (-) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3-alpha-5`,
}

var prereleaseEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable pre-release identifier",
	Long: `Enable pre-release identifier by rendering the config template and setting it in VERSION file.

If a template is configured in .versionator.yaml, it will be rendered and set as a static value.
If no template is configured, defaults to "alpha".

The VERSION file is the source of truth - this command writes to it directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load current version
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		// Determine prerelease value: use config template if set, else default to "alpha"
		prerelease := "alpha"
		if cfg, err := config.ReadConfig(); err == nil && cfg.PreRelease.Template != "" {
			templateData := emit.BuildTemplateDataFromVersion(vd)
			rendered, err := emit.RenderTemplateWithData(cfg.PreRelease.Template, templateData)
			if err == nil && rendered != "" {
				prerelease = rendered
			}
		}

		// Set prerelease in VERSION file
		if err := version.SetPreRelease(prerelease); err != nil {
			return fmt.Errorf("error setting pre-release: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release enabled with value '%s'\n", prerelease)

		// Show current version
		vd, err = version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prereleaseDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable pre-release identifier",
	Long: `Disable pre-release identifier by clearing it from the VERSION file.

The VERSION file is the source of truth - this command removes the pre-release from it directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Clear prerelease from VERSION file
		if err := version.SetPreRelease(""); err != nil {
			return fmt.Errorf("error clearing pre-release: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Pre-release disabled")

		// Show current version without pre-release
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prereleaseStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pre-release status",
	Long: `Show current pre-release status from VERSION file (source of truth).

Also shows the configured template from .versionator.yaml if set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load version - VERSION file is source of truth
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading version: %w", err)
		}

		if vd.PreRelease != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Pre-release: ENABLED")
			fmt.Fprintf(cmd.OutOrStdout(), "Value: %s\n", vd.PreRelease)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Pre-release: DISABLED")
		}

		// Show config template if set
		if cfg, err := config.ReadConfig(); err == nil && cfg.PreRelease.Template != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Config template: %s\n", cfg.PreRelease.Template)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prereleaseSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set pre-release value",
	Long: `Set a static pre-release value in both config and VERSION file.

This updates:
1. The config file (.versionator.yaml) - so 'prerelease enable' can restore it
2. The VERSION file - the source of truth for the current version

Use 'prerelease template' for dynamic values with variables like {{CommitsSinceTag}}.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dots (e.g., "alpha.1")

Examples:
  versionator prerelease set alpha
  versionator prerelease set beta.1
  versionator prerelease set rc.2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := args[0]

		// Update config with static value as template
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.PreRelease.Template = value
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		// Update VERSION file
		if err := version.SetPreRelease(value); err != nil {
			return fmt.Errorf("error setting pre-release: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release set to: %s\n", value)

		// Show current version with pre-release
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prereleaseClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear pre-release value from VERSION file",
	Long:  "Remove the pre-release identifier from VERSION file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.SetPreRelease(""); err != nil {
			return fmt.Errorf("error clearing pre-release: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Pre-release cleared")

		// Show current version without pre-release
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prereleaseTemplateCmd = &cobra.Command{
	Use:   "template [template-string]",
	Short: "Get or set the pre-release template",
	Long: `Get or set the pre-release template.

When setting a template, it is saved to .versionator.yaml config AND rendered
immediately to set the pre-release value in VERSION file.

IMPORTANT: Use DASHES (-) to separate pre-release identifiers per SemVer 2.0.0.
The leading dash (-) is added automatically - do NOT include it in your template.

The template uses Mustache syntax. Available variables:
  {{ShortHash}}            - Short git commit hash, 7 chars (e.g., "abc1234")
  {{MediumHash}}           - Medium git commit hash, 12 chars (e.g., "abc1234def01")
  {{Hash}}                 - Full git commit hash (40 chars)
  {{BranchName}}           - Current branch name
  {{EscapedBranchName}}    - Branch name with / replaced by -
  {{CommitsSinceTag}}      - Commits since last tag
  {{BuildDateTimeCompact}} - Compact timestamp (20241211103045)
  {{BuildDateUTC}}         - Date only (2024-12-11)
  {{CommitDate}}           - Commit date ISO 8601
  {{CommitDateCompact}}    - Commit date compact (20241211103045)

Examples:
  versionator prerelease template                              # Show current template
  versionator prerelease template "alpha"                      # Static "alpha"
  versionator prerelease template "alpha-{{CommitsSinceTag}}"  # "alpha-5"
  versionator prerelease template "rc-{{CommitsSinceTag}}"     # "rc-5"
  versionator prerelease template "beta-{{EscapedBranchName}}" # "beta-feature-foo"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		// If no argument, show current template
		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Current pre-release template: %s\n", cfg.PreRelease.Template)

			// Show what it would render to
			if cfg.PreRelease.Template != "" {
				vd, err := version.Load()
				if err == nil {
					templateData := emit.BuildTemplateDataFromVersion(vd)
					result, err := emit.RenderTemplateWithData(cfg.PreRelease.Template, templateData)
					if err == nil && result != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Rendered value: %s\n", result)
					}
				}
			}
			return nil
		}

		// Set new template in config
		cfg.PreRelease.Template = args[0]
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release template set to: %s\n", cfg.PreRelease.Template)

		// Render template and set in VERSION file
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error loading version: %w", err)
		}

		templateData := emit.BuildTemplateDataFromVersion(vd)
		result, err := emit.RenderTemplateWithData(cfg.PreRelease.Template, templateData)
		if err != nil {
			return fmt.Errorf("error rendering template: %w", err)
		}

		// Set the rendered value in VERSION file
		if err := version.SetPreRelease(result); err != nil {
			return fmt.Errorf("error setting pre-release: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release set to: %s\n", result)

		// Show updated version
		vd, err = version.Load()
		if err != nil {
			return fmt.Errorf("error loading version: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(prereleaseCmd)
	prereleaseCmd.AddCommand(prereleaseEnableCmd)
	prereleaseCmd.AddCommand(prereleaseDisableCmd)
	prereleaseCmd.AddCommand(prereleaseStatusCmd)
	prereleaseCmd.AddCommand(prereleaseSetCmd)
	prereleaseCmd.AddCommand(prereleaseClearCmd)
	prereleaseCmd.AddCommand(prereleaseTemplateCmd)
}
