package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"

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
	Long: `Enable pre-release identifier by rendering the config elements and setting it in VERSION file.

If elements are configured in .versionator.yaml, they will be rendered and joined with dashes.
If no elements are configured, defaults to "alpha".

The VERSION file is the source of truth - this command writes to it directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine prerelease value: use config elements if set, else default to "alpha"
		prerelease, err := versionator.RenderPreRelease()
		if err != nil {
			return fmt.Errorf("error rendering prerelease: %w", err)
		}

		// If no elements or render empty, default to "alpha"
		if prerelease == "" {
			prerelease = "alpha"
		}

		// Set prerelease in VERSION file
		if err := version.SetPreRelease(prerelease); err != nil {
			return fmt.Errorf("error setting pre-release: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release enabled with value '%s'\n", prerelease)

		// Show current version
		vd, err := version.Load()
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

Also shows the configured elements from .versionator.yaml if set.`,
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

		// Show config elements if set
		if cfg, err := config.ReadConfig(); err == nil && len(cfg.PreRelease.Elements) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Config elements: [%s]\n", strings.Join(cfg.PreRelease.Elements, ", "))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prereleaseSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set pre-release value",
	Long: `Set a static pre-release value in VERSION file.

Use 'prerelease elements' for dynamic values with variables like CommitsSinceTag.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dashes (e.g., "alpha-1")

Examples:
  versionator prerelease set alpha
  versionator prerelease set beta-1
  versionator prerelease set rc-2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := args[0]

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

var prereleaseElementsCmd = &cobra.Command{
	Use:   "elements [element1,element2,...]",
	Short: "Get or set the pre-release elements",
	Long: `Get or set the pre-release elements.

Elements are variable names that will be rendered and joined with dashes (-)
per SemVer 2.0.0 specification. The leading dash (-) is added automatically.

When setting elements, provide a comma-separated list of variable names.
The elements are saved to .versionator.yaml config AND rendered
immediately to set the pre-release value in VERSION file.

Available variables:
  ShortHash            - Short git commit hash, 7 chars (e.g., "abc1234")
  MediumHash           - Medium git commit hash, 12 chars (e.g., "abc1234def01")
  Hash                 - Full git commit hash (40 chars)
  BranchName           - Current branch name
  EscapedBranchName    - Branch name with / replaced by -
  CommitsSinceTag      - Commits since last tag
  BuildDateTimeCompact - Compact timestamp (20241211103045)
  Dirty                - "dirty" if uncommitted changes

Literal values are also supported - they are used as-is.

Examples:
  versionator prerelease elements                                              # Show current
  versionator prerelease elements "alpha,CommitsSinceTag"                      # alpha-5
  versionator prerelease elements "CommitsSinceTag,BuildDateTimeCompact,ShortHash" # Go-style
  versionator prerelease elements "beta,EscapedBranchName"                     # beta-feature-foo`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		// If no argument, show current elements
		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Current pre-release elements: [%s]\n", strings.Join(cfg.PreRelease.Elements, ", "))

			// Show what it would render to
			if len(cfg.PreRelease.Elements) > 0 {
				result, err := versionator.RenderPreRelease()
				if err == nil && result != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Rendered value: %s\n", result)
				}
			}
			return nil
		}

		// Parse comma-separated elements
		elements := strings.Split(args[0], ",")
		for i, e := range elements {
			elements[i] = strings.TrimSpace(e)
		}

		// Set new elements in config
		cfg.PreRelease.Elements = elements
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release elements set to: [%s]\n", strings.Join(elements, ", "))

		// Render elements and set in VERSION file
		result, err := versionator.RenderPreRelease()
		if err != nil {
			return fmt.Errorf("error rendering prerelease: %w", err)
		}

		// Set the rendered value in VERSION file
		if err := version.SetPreRelease(result); err != nil {
			return fmt.Errorf("error setting pre-release: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release set to: %s\n", result)

		// Show updated version
		vd, err := version.Load()
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
	prereleaseCmd.AddCommand(prereleaseElementsCmd)
}
