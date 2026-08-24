package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var metadataForceFlag bool

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Manage build metadata",
	Long: `Commands to manage build metadata.

Build metadata follows SemVer 2.0.0 specification:
- Appended with a plus sign (+) - added automatically
- Multiple identifiers separated by DOTS (.)
- Each identifier: alphanumerics and hyphens only [0-9A-Za-z-]

Stability controls where the metadata value lives:
  stable: true  - Value is written to VERSION file
  stable: false - Value is generated from template at output time (default)

Example: 1.2.3+abc1234
         └─────┘
          metadata`,
}

var metadataStableCmd = &cobra.Command{
	Use:   "stable [true|false]",
	Short: "Get or set metadata stability",
	Long: `Get or set whether metadata is stable (written to VERSION file) or dynamic (generated at output).

When stable is true:
  - Metadata value is stored in the VERSION file
  - Use 'set' and 'template' commands to modify it
  - Note: This is rarely used since metadata is usually build-time info

When stable is false (default):
  - Metadata is NOT stored in VERSION file
  - Value is generated from template at every output
  - Ideal for commit hashes, build timestamps, etc.

Examples:
  versionator config metadata stable         # Show current setting
  versionator config metadata stable true    # Enable stable mode
  versionator config metadata stable false   # Enable dynamic mode (default)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMetadataStable,
}

func runMetadataStable(cmd *cobra.Command, args []string) error {
	return runStableCommand(cmd, args, metadataAccessor)
}

var metadataEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable build metadata (requires stable: true)",
	Long: `Enable build metadata by rendering the config template and setting it in VERSION file.

This command requires stable: true. If metadata is configured as dynamic (stable: false),
use 'versionator config metadata stable true' first.

If a template is configured in .versionator.yaml, it will be rendered and set as a static value.
If no template is configured, defaults to the git short hash.`,
	RunE: runMetadataEnable,
}

func runMetadataEnable(cmd *cobra.Command, args []string) error {
	return runEnableCommand(cmd, args, metadataAccessor)
}

var metadataDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable build metadata (requires stable: true)",
	Long: `Disable build metadata by clearing it from the VERSION file.

This command requires stable: true. If metadata is configured as dynamic (stable: false),
the metadata is already not in the VERSION file.`,
	RunE: runMetadataDisable,
}

func runMetadataDisable(cmd *cobra.Command, args []string) error {
	return runDisableCommand(cmd, args, metadataAccessor)
}

var metadataStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show metadata status",
	Long:  `Show current metadata configuration and value.`,
	RunE:  runMetadataStatus,
}

func runMetadataStatus(cmd *cobra.Command, args []string) error {
	return runStatusCommand(cmd, args, metadataAccessor)
}

var metadataConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Show metadata configuration",
	Long:  "Show metadata configuration from .versionator.yaml",
	RunE:  runMetadataConfigure,
}

func runMetadataConfigure(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	fmt.Printf("Current configuration:\n")
	fmt.Printf("  Stable: %t\n", cfg.Metadata.Stable)
	fmt.Printf("  Metadata template: %s\n", cfg.Metadata.Template)
	fmt.Printf("  Git hash length: %d\n", cfg.Metadata.Git.HashLength)
	fmt.Printf("\nConfiguration is stored in .versionator.yaml\n")
	return nil
}

var metadataSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set metadata value (requires stable: true)",
	Long: `Set a static metadata value in the VERSION file.

This command requires stable: true. If metadata is configured as dynamic (stable: false),
you will get an error. Use --force to override and set the template to a literal value.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dots (e.g., "build.123")

Examples:
  versionator config metadata set build.123
  versionator config metadata set 20241211103045
  versionator config metadata set ci.456.linux
  versionator config metadata set "abc1234" --force  # Force on dynamic mode`,
	Args: cobra.ExactArgs(1),
	RunE: runMetadataSet,
}

func runMetadataSet(cmd *cobra.Command, args []string) error {
	return runSetCommand(cmd, args, metadataAccessor)
}

var metadataClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear metadata value from VERSION file (requires stable: true)",
	Long: `Remove the build metadata from VERSION file.

This command requires stable: true. If metadata is configured as dynamic (stable: false),
the metadata is already not in the VERSION file.`,
	RunE: runMetadataClear,
}

func runMetadataClear(cmd *cobra.Command, args []string) error {
	return runClearCommand(cmd, args, metadataAccessor)
}

var metadataTemplateCmd = &cobra.Command{
	Use:   "template [template-string]",
	Short: "Get or set the metadata template",
	Long: `Get or set the metadata template used for build metadata.

Behavior depends on stability setting:
  stable: true  - Template is rendered and written to VERSION file
  stable: false - Template is saved to config only (rendered at output time)

IMPORTANT: Separate identifiers with DOTS (.) per SemVer 2.0.0.
Each identifier can only contain alphanumerics and hyphens [0-9A-Za-z-].
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
  versionator config metadata template                                              # Show current
  versionator config metadata template "{{BuildDateTimeCompact}}.{{MediumHash}}"    # Timestamp.hash
  versionator config metadata template "{{ShortHash}}"                              # Just git hash
  versionator config metadata template "{{CommitsSinceTag}}.{{ShortHash}}"          # Build number.hash`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMetadataTemplate,
}

func runMetadataTemplate(cmd *cobra.Command, args []string) error {
	return runTemplateCommand(cmd, args, metadataAccessor)
}

func init() {
	configCmd.AddCommand(metadataCmd)
	metadataCmd.AddCommand(metadataStableCmd)
	metadataCmd.AddCommand(metadataEnableCmd)
	metadataCmd.AddCommand(metadataDisableCmd)
	metadataCmd.AddCommand(metadataStatusCmd)
	metadataCmd.AddCommand(metadataConfigureCmd)
	metadataCmd.AddCommand(metadataSetCmd)
	metadataCmd.AddCommand(metadataClearCmd)
	metadataCmd.AddCommand(metadataTemplateCmd)

	// Add --force flag to set command
	metadataSetCmd.Flags().BoolVarP(&metadataForceFlag, "force", "f", false, "Force set on dynamic mode (sets template to literal value)")
}

// defaultMetadataFallback is used when metadata has no template and the
// repository cannot supply a commit hash.
const defaultMetadataFallback = "build"

// defaultMetadataHashLength is the short-hash width used when the config does
// not specify one.
const defaultMetadataHashLength = 7

// metadataDefaultValue supplies `enable`'s value when the template renders
// empty: the short commit hash, falling back to a fixed label outside a repo.
func metadataDefaultValue(cfg *config.Config, _ *version.Version) string {
	hashLength := defaultMetadataHashLength
	if cfg.Metadata.Git.HashLength > 0 {
		hashLength = cfg.Metadata.Git.HashLength
	}

	gitVCS := vcs.GetVCS("git")
	if gitVCS != nil && gitVCS.IsRepository() {
		if hash, err := gitVCS.GetVCSIdentifier(hashLength); err == nil && hash != "" {
			return hash
		}
	}

	return defaultMetadataFallback
}
