package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var prereleaseForceFlag bool

var prereleaseCmd = &cobra.Command{
	Use:   "prerelease",
	Short: "Manage pre-release identifier",
	Long: `Commands to manage pre-release identifiers.

Pre-release follows SemVer 2.0.0 specification:
- Appended with a dash (-) - this is added automatically
- Use DASHES (-) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Stability controls where the pre-release value lives:
  stable: true  - Value is written to VERSION file (traditional release workflow)
  stable: false - Value is generated from template at output time (default, CD workflow)

Example output: 1.2.3-build-5`,
}

var prereleaseStableCmd = &cobra.Command{
	Use:   "stable [true|false]",
	Short: "Get or set pre-release stability",
	Long: `Get or set whether pre-release is stable (written to VERSION file) or dynamic (generated at output).

When stable is true:
  - Pre-release value is stored in the VERSION file
  - Use 'set' and 'template' commands to modify it
  - Traditional release workflow (alpha, beta, rc.1, etc.)

When stable is false (default):
  - Pre-release is NOT stored in VERSION file
  - Value is generated from template at every output
  - Continuous delivery workflow (build-42, etc.)

Examples:
  versionator config prerelease stable         # Show current setting
  versionator config prerelease stable true    # Enable stable mode
  versionator config prerelease stable false   # Enable dynamic mode (default)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrereleaseStable,
}

func runPrereleaseStable(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// If no argument, show current setting
	if len(args) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release stable: %t\n", cfg.PreRelease.Stable)
		if cfg.PreRelease.Stable {
			fmt.Fprintln(cmd.OutOrStdout(), "  Value is stored in VERSION file")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  Value is generated from template at output time")
			fmt.Fprintf(cmd.OutOrStdout(), "  Template: %s\n", cfg.PreRelease.Template)
		}
		return nil
	}

	// Set new value
	switch args[0] {
	case "true", "1", "yes":
		cfg.PreRelease.Stable = true
	case "false", "0", "no":
		cfg.PreRelease.Stable = false
	default:
		return fmt.Errorf("invalid value '%s': use 'true' or 'false'", args[0])
	}

	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Pre-release stable set to: %t\n", cfg.PreRelease.Stable)

	// If switching to stable=false, clear prerelease from VERSION file
	if !cfg.PreRelease.Stable {
		if err := version.SetPreRelease(""); err != nil {
			return fmt.Errorf("error clearing pre-release from VERSION: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Pre-release cleared from VERSION file (will be generated at output time)")
	}

	return nil
}

var prereleaseEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable pre-release identifier (requires stable: true)",
	Long: `Enable pre-release identifier by rendering the config template and setting it in VERSION file.

This command requires stable: true. If pre-release is configured as dynamic (stable: false),
use 'versionator config prerelease stable true' first.

If a template is configured in .versionator.yaml, it will be rendered and set as a static value.
If no template is configured, defaults to "alpha".`,
	RunE: runPrereleaseEnable,
}

func runPrereleaseEnable(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// Check stability
	if !cfg.PreRelease.Stable {
		return fmt.Errorf("pre-release is configured as dynamic (stable: false)\n" +
			"In dynamic mode, pre-release is generated at output time.\n" +
			"To use this command, first run: versionator config prerelease stable true")
	}

	// Load current version
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Determine prerelease value: use config template if set, else default to "alpha"
	prerelease := "alpha"
	if cfg.PreRelease.Template != "" {
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
}

var prereleaseDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable pre-release identifier (requires stable: true)",
	Long: `Disable pre-release identifier by clearing it from the VERSION file.

This command requires stable: true. If pre-release is configured as dynamic (stable: false),
the pre-release is already not in the VERSION file.`,
	RunE: runPrereleaseDisable,
}

func runPrereleaseDisable(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// Check stability
	if !cfg.PreRelease.Stable {
		return fmt.Errorf("pre-release is configured as dynamic (stable: false)\n" +
			"In dynamic mode, pre-release is not stored in VERSION file.\n" +
			"To disable dynamic pre-release at output, use --prerelease=\"\" flag or clear the template.")
	}

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
}

var prereleaseStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pre-release status",
	Long:  `Show current pre-release configuration and value.`,
	RunE:  runPrereleaseStatus,
}

func runPrereleaseStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// Load version
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error reading version: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Stable: %t\n", cfg.PreRelease.Stable)
	fmt.Fprintf(cmd.OutOrStdout(), "Template: %s\n", cfg.PreRelease.Template)

	if cfg.PreRelease.Stable {
		// Show value from VERSION file
		if vd.PreRelease != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "VALUE (from VERSION file): %s\n", vd.PreRelease)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "VALUE (from VERSION file): (none)")
		}
	} else {
		// Show what would be rendered
		if cfg.PreRelease.Template != "" {
			templateData := emit.BuildTemplateDataFromVersion(vd)
			result, err := emit.RenderTemplateWithData(cfg.PreRelease.Template, templateData)
			if err == nil && result != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "VALUE (rendered from template): %s\n", result)
			}
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "VALUE: (no template configured)")
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "VERSION file: %s\n", vd.FullString())
	return nil
}

var prereleaseSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set pre-release value (requires stable: true)",
	Long: `Set a static pre-release value in the VERSION file.

This command requires stable: true. If pre-release is configured as dynamic (stable: false),
you will get an error. Use --force to override and set the template to a literal value.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dashes (e.g., "alpha-1")

Examples:
  versionator config prerelease set alpha
  versionator config prerelease set beta-1
  versionator config prerelease set rc-2
  versionator config prerelease set "build-42" --force  # Force on dynamic mode`,
	Args: cobra.ExactArgs(1),
	RunE: runPrereleaseSet,
}

func runPrereleaseSet(cmd *cobra.Command, args []string) error {
	value := args[0]

	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// Check stability
	if !cfg.PreRelease.Stable && !prereleaseForceFlag {
		return fmt.Errorf("pre-release is configured as dynamic (stable: false)\n" +
			"Cannot set a static value when pre-release is generated at output time.\n" +
			"Options:\n" +
			"  1. Switch to stable mode: versionator config prerelease stable true\n" +
			"  2. Use --force to set the template to this literal value\n" +
			"  3. Use 'template' command to set a dynamic template")
	}

	// Update template in config
	cfg.PreRelease.Template = value
	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	if cfg.PreRelease.Stable {
		// Write to VERSION file
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
	} else {
		// --force was used: set template to literal, don't write to VERSION
		fmt.Fprintf(cmd.OutOrStdout(), "Pre-release template set to literal: %s\n", value)
		fmt.Fprintln(cmd.OutOrStdout(), "(Value will be used at output time, not stored in VERSION file)")
	}

	return nil
}

var prereleaseClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear pre-release value from VERSION file (requires stable: true)",
	Long: `Remove the pre-release identifier from VERSION file.

This command requires stable: true. If pre-release is configured as dynamic (stable: false),
the pre-release is already not in the VERSION file.`,
	RunE: runPrereleaseClear,
}

func runPrereleaseClear(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// Check stability
	if !cfg.PreRelease.Stable {
		return fmt.Errorf("pre-release is configured as dynamic (stable: false)\n" +
			"In dynamic mode, pre-release is not stored in VERSION file.\n" +
			"To clear the template, use: versionator config prerelease template \"\"")
	}

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
}

var prereleaseTemplateCmd = &cobra.Command{
	Use:   "template [template-string]",
	Short: "Get or set the pre-release template",
	Long: `Get or set the pre-release template.

Behavior depends on stability setting:
  stable: true  - Template is rendered and written to VERSION file
  stable: false - Template is saved to config only (rendered at output time)

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
  versionator config prerelease template                              # Show current template
  versionator config prerelease template "alpha"                      # Static "alpha"
  versionator config prerelease template "alpha-{{CommitsSinceTag}}"  # "alpha-5"
  versionator config prerelease template "rc-{{CommitsSinceTag}}"     # "rc-5"
  versionator config prerelease template "beta-{{EscapedBranchName}}" # "beta-feature-foo"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrereleaseTemplate,
}

func runPrereleaseTemplate(cmd *cobra.Command, args []string) error {
	return runTemplateCommand(cmd, args, prereleaseAccessor)
}

func init() {
	configCmd.AddCommand(prereleaseCmd)
	prereleaseCmd.AddCommand(prereleaseStableCmd)
	prereleaseCmd.AddCommand(prereleaseEnableCmd)
	prereleaseCmd.AddCommand(prereleaseDisableCmd)
	prereleaseCmd.AddCommand(prereleaseStatusCmd)
	prereleaseCmd.AddCommand(prereleaseSetCmd)
	prereleaseCmd.AddCommand(prereleaseClearCmd)
	prereleaseCmd.AddCommand(prereleaseTemplateCmd)

	// Add --force flag to set command
	prereleaseSetCmd.Flags().BoolVarP(&prereleaseForceFlag, "force", "f", false, "Force set on dynamic mode (sets template to literal value)")
}
