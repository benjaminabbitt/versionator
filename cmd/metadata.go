package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"

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
	Long: `Enable build metadata by rendering the config elements and setting it in VERSION file.

If elements are configured in .versionator.yaml, they will be rendered and joined with dots.
If no elements are configured, defaults to the git short hash.

The VERSION file is the source of truth - this command writes to it directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine metadata value: use config elements if set, else default to git hash
		metadata, err := versionator.RenderMetadata()
		if err != nil {
			return fmt.Errorf("error rendering metadata: %w", err)
		}

		// If no elements or render empty, use git hash as default
		if metadata == "" {
			cfg, _ := config.ReadConfig()
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
		vd, err := version.Load()
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

Also shows the configured elements from .versionator.yaml if set.`,
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

		// Show config elements if set
		if cfg, err := config.ReadConfig(); err == nil && len(cfg.Metadata.Elements) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Config elements: [%s]\n", strings.Join(cfg.Metadata.Elements, ", "))
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
		fmt.Printf("  Metadata elements: [%s]\n", strings.Join(cfg.Metadata.Elements, ", "))
		fmt.Printf("  Git hash length: %d\n", cfg.Metadata.Git.HashLength)
		fmt.Printf("\nConfiguration is stored in .versionator.yaml\n")
		fmt.Printf("Note: VERSION file is the source of truth for current metadata value.\n")
		return nil
	},
}

var metadataSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set metadata value",
	Long: `Set a static metadata value in VERSION file.

Use 'metadata elements' for dynamic values with variables like ShortHash.

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

var metadataElementsCmd = &cobra.Command{
	Use:   "elements [element1,element2,...]",
	Short: "Get or set the metadata elements",
	Long: `Get or set the metadata elements used for build metadata.

Elements are variable names that will be rendered and joined with dots (.)
per SemVer 2.0.0 specification. The leading plus (+) is added automatically.

When setting elements, provide a comma-separated list of variable names.
The elements are saved to .versionator.yaml config AND rendered
immediately to set the metadata value in VERSION file.

Available variables:
  ShortHash            - Short git commit hash, 7 chars (e.g., "abc1234")
  MediumHash           - Medium git commit hash, 12 chars (e.g., "abc1234def01")
  Hash                 - Full git commit hash (40 chars)
  BranchName           - Current branch name
  EscapedBranchName    - Branch name with / replaced by -
  CommitsSinceTag      - Commits since last tag
  BuildDateTimeCompact - Compact timestamp (20241211103045)
  BuildDateUTC         - Date only (2024-12-11)
  Dirty                - "dirty" if uncommitted changes

Literal values are also supported - they are used as-is.

Examples:
  versionator metadata elements                                        # Show current
  versionator metadata elements "BuildDateTimeCompact,ShortHash"       # Timestamp.hash
  versionator metadata elements "ShortHash"                            # Just git hash
  versionator metadata elements "BuildDateTimeCompact,ShortHash,Dirty" # Timestamp.hash.dirty`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		// If no argument, show current elements
		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Current metadata elements: [%s]\n", strings.Join(cfg.Metadata.Elements, ", "))

			// Show what it would render to
			if len(cfg.Metadata.Elements) > 0 {
				result, err := versionator.RenderMetadata()
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
		cfg.Metadata.Elements = elements
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata elements set to: [%s]\n", strings.Join(elements, ", "))

		// Render elements and set in VERSION file
		result, err := versionator.RenderMetadata()
		if err != nil {
			return fmt.Errorf("error rendering metadata: %w", err)
		}

		// Set the rendered value in VERSION file
		if err := version.SetMetadata(result); err != nil {
			return fmt.Errorf("error setting metadata: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Metadata set to: %s\n", result)

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
	rootCmd.AddCommand(metadataCmd)
	metadataCmd.AddCommand(metadataEnableCmd)
	metadataCmd.AddCommand(metadataDisableCmd)
	metadataCmd.AddCommand(metadataStatusCmd)
	metadataCmd.AddCommand(metadataConfigureCmd)
	metadataCmd.AddCommand(metadataSetCmd)
	metadataCmd.AddCommand(metadataClearCmd)
	metadataCmd.AddCommand(metadataElementsCmd)
}
