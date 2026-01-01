package version

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync config from VERSION file",
	Long: `Sync the .versionator.yaml config file to match the VERSION file.

This command updates config values to match what's in VERSION:
- prefix: Updated to match VERSION prefix

VERSION is the source of truth. Use this command to ensure config
stays in sync after manually editing the VERSION file.

Examples:
  # Sync config from VERSION
  versionator version sync`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load current version
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading version: %w", err)
		}

		// Sync config from VERSION
		if err := version.SyncConfigFromVersion(vd); err != nil {
			return fmt.Errorf("error syncing config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Config synced from VERSION: %s\n", vd.FullString())
		return nil
	},
}

func init() {
	// SyncCmd is added to the version command in cmd/version.go
}
