package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Manage version build metadata behavior",
	Long: `Commands to enable or disable appending build metadata to version numbers.

Build metadata follows SemVer 2.0.0 specification:
- Appended with a plus sign (+) - this is added automatically
- Use DOTS (.) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3+20241211103045.abc1234def5`,
}

var metadataEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable build metadata",
	Long: `Enable build metadata by rendering the config template and setting it in VERSION file.

If a template is configured in .versionator.yaml, it will be rendered and set as a static value.
If no template is configured, defaults to the git short hash.

The VERSION file is the source of truth - this command writes to it directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load current version
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		// Determine metadata value: use config template if set, else default to git hash
		metadata := ""
		cfg, _ := config.ReadConfig()
		if cfg != nil && cfg.Metadata.Template != "" {
			templateData := emit.BuildTemplateDataFromVersion(vd)
			rendered, err := emit.RenderTemplateWithData(cfg.Metadata.Template, templateData)
			if err == nil && rendered != "" {
				metadata = rendered
			}
		}

		// If no template or render failed, use git hash as default
		if metadata == "" {
			gitVCS := vcs.GetVCS("git")
			hashLength := 7
			if cfg != nil && cfg.Metadata.Git.HashLength > 0 {
				hashLength = cfg.Metadata.Git.HashLength
			}
			if gitVCS != nil && gitVCS.IsRepository() {
				if hash, err := gitVCS.GetVCSIdentifier(hashLength); err == nil {
					metadata = hash
				}
			}
		}

		// If still no metadata, use a default
		if metadata == "" {
			metadata = "build"
		}

		// Set metadata in VERSION file
		if err := version.SetMetadata(metadata); err != nil {
			return fmt.Errorf("error setting metadata: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata enabled with value '%s'\n", metadata)

		// Show current version
		vd, err = version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var metadataDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable build metadata",
	Long: `Disable build metadata by clearing it from the VERSION file.

The VERSION file is the source of truth - this command removes the metadata from it directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Clear metadata from VERSION file
		if err := version.SetMetadata(""); err != nil {
			return fmt.Errorf("error clearing metadata: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Metadata disabled")

		// Show current version without metadata
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var metadataStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show metadata status",
	Long: `Show current metadata status from VERSION file (source of truth).

Also shows the configured template from .versionator.yaml if set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load version - VERSION file is source of truth
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading version: %w", err)
		}

		if vd.BuildMetadata != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Metadata: ENABLED")
			fmt.Fprintf(cmd.OutOrStdout(), "Value: %s\n", vd.BuildMetadata)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Metadata: DISABLED")
		}

		// Show config template if set
		if cfg, err := config.ReadConfig(); err == nil && cfg.Metadata.Template != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Config template: %s\n", cfg.Metadata.Template)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var metadataConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Show metadata configuration",
	Long:  "Show metadata configuration from .versionator.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		fmt.Printf("Current configuration:\n")
		fmt.Printf("  Metadata template: %s\n", cfg.Metadata.Template)
		fmt.Printf("  Git hash length: %d\n", cfg.Metadata.Git.HashLength)
		fmt.Printf("\nConfiguration is stored in .versionator.yaml\n")
		fmt.Printf("Note: VERSION file is the source of truth for current metadata value.\n")
		return nil
	},
}

var metadataSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set metadata value",
	Long: `Set a static metadata value in both config and VERSION file.

This updates:
1. The config file (.versionator.yaml) - so 'metadata enable' can restore it
2. The VERSION file - the source of truth for the current version

Use 'metadata template' for dynamic values with variables like {{ShortHash}}.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dots (e.g., "build.123")

Examples:
  versionator metadata set build.123
  versionator metadata set 20241211103045
  versionator metadata set ci.456.linux`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := args[0]

		// Update config with static value as template
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.Metadata.Template = value
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		// Update VERSION file
		if err := version.SetMetadata(value); err != nil {
			return fmt.Errorf("error setting metadata: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata set to: %s\n", value)

		// Show current version with metadata
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var metadataClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear metadata value from VERSION file",
	Long:  "Remove the build metadata from VERSION file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.SetMetadata(""); err != nil {
			return fmt.Errorf("error clearing metadata: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Metadata cleared")

		// Show current version without metadata
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var metadataTemplateCmd = &cobra.Command{
	Use:   "template [template-string]",
	Short: "Get or set the metadata template",
	Long: `Get or set the metadata template used for build metadata.

When setting a template, it is saved to .versionator.yaml config AND rendered
immediately to set the metadata value in VERSION file.

IMPORTANT: Use DOTS (.) to separate metadata identifiers per SemVer 2.0.0.
The leading plus (+) is added automatically - do NOT include it in your template.

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
  versionator metadata template                                              # Show current
  versionator metadata template "{{BuildDateTimeCompact}}.{{MediumHash}}"    # Timestamp.hash
  versionator metadata template "{{ShortHash}}"                              # Just git hash
  versionator metadata template "{{CommitsSinceTag}}.{{ShortHash}}"          # Build number.hash`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		// If no argument, show current template
		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Current metadata template: %s\n", cfg.Metadata.Template)

			// Show what it would render to
			if cfg.Metadata.Template != "" {
				vd, err := version.Load()
				if err == nil {
					templateData := emit.BuildTemplateDataFromVersion(vd)
					result, err := emit.RenderTemplateWithData(cfg.Metadata.Template, templateData)
					if err == nil && result != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Rendered value: %s\n", result)
					}
				}
			}
			return nil
		}

		// Set new template in config
		cfg.Metadata.Template = args[0]
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata template set to: %s\n", cfg.Metadata.Template)

		// Render template and set in VERSION file
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error loading version: %w", err)
		}

		templateData := emit.BuildTemplateDataFromVersion(vd)
		result, err := emit.RenderTemplateWithData(cfg.Metadata.Template, templateData)
		if err != nil {
			return fmt.Errorf("error rendering template: %w", err)
		}

		// Set the rendered value in VERSION file
		if err := version.SetMetadata(result); err != nil {
			return fmt.Errorf("error setting metadata: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata set to: %s\n", result)

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
	rootCmd.AddCommand(metadataCmd)
	metadataCmd.AddCommand(metadataEnableCmd)
	metadataCmd.AddCommand(metadataDisableCmd)
	metadataCmd.AddCommand(metadataStatusCmd)
	metadataCmd.AddCommand(metadataConfigureCmd)
	metadataCmd.AddCommand(metadataSetCmd)
	metadataCmd.AddCommand(metadataClearCmd)
	metadataCmd.AddCommand(metadataTemplateCmd)
}
