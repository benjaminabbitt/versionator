package version

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"
	"github.com/spf13/cobra"
)

var RenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render and save version with config elements",
	Long: `Render the VERSION file with fresh prerelease and metadata from config elements.

This command:
1. Loads the current VERSION file
2. Renders prerelease from config prerelease.elements (if configured)
3. Renders metadata from config metadata.elements (if configured)
4. Applies prefix from VERSION or config
5. Saves the updated VERSION file

After running this command, the VERSION file will contain the fully rendered
version string with current dynamic values (e.g., CommitsSinceTag, ShortHash).

Use 'versionator commit' to create a git tag from the VERSION file.

Examples:
  # Render VERSION with config elements
  versionator version render

  # Check what will be rendered without saving
  versionator prerelease status
  versionator metadata status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load current version
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading version: %w", err)
		}

		// Load config
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		// Render prerelease from config elements if configured
		if len(cfg.PreRelease.Elements) > 0 {
			prerelease, err := versionator.RenderPreRelease()
			if err != nil {
				return fmt.Errorf("error rendering prerelease: %w", err)
			}
			vd.PreRelease = prerelease
		}

		// Render metadata from config elements if configured
		if len(cfg.Metadata.Elements) > 0 {
			metadata, err := versionator.RenderMetadata()
			if err != nil {
				return fmt.Errorf("error rendering metadata: %w", err)
			}
			vd.BuildMetadata = metadata
		}

		// Apply prefix from VERSION or config
		if vd.Prefix == "" && cfg.Prefix != "" {
			vd.Prefix = cfg.Prefix
		}

		// Save the updated VERSION file
		if err := version.Save(vd); err != nil {
			return fmt.Errorf("error saving version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Version rendered and saved: %s\n", vd.FullString())
		return nil
	},
}

func init() {
	// RenderCmd is added to the version command in cmd/version.go
}
