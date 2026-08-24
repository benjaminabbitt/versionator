package cmd

import (
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
	return runStableCommand(cmd, args, prereleaseAccessor)
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
	return runEnableCommand(cmd, args, prereleaseAccessor)
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
	return runDisableCommand(cmd, args, prereleaseAccessor)
}

var prereleaseStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pre-release status",
	Long:  `Show current pre-release configuration and value.`,
	RunE:  runPrereleaseStatus,
}

func runPrereleaseStatus(cmd *cobra.Command, args []string) error {
	return runStatusCommand(cmd, args, prereleaseAccessor)
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
	return runSetCommand(cmd, args, prereleaseAccessor)
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
	return runClearCommand(cmd, args, prereleaseAccessor)
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

// defaultPreReleaseValue is used when pre-release has no template configured.
const defaultPreReleaseValue = "alpha"
