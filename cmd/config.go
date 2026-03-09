package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage versionator configuration",
	Long: `Manage versionator configuration including version prefix, pre-release,
metadata, custom variables, and versioning mode.

Use subcommands to configure specific aspects:
  config prefix      - Manage version prefix (v, V)
  config prerelease  - Manage pre-release identifiers
  config metadata    - Manage build metadata
  config custom      - Manage custom key-value pairs
  config mode        - Switch between release and continuous-delivery modes
  config vars        - Show all available template variables`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
