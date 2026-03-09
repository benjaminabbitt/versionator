package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage versionator configuration",
	Long: `Manage versionator configuration including version prefix, pre-release,
metadata, and custom variables.

Use subcommands to configure specific aspects:
  config prefix      - Manage version prefix (v, V)
  config prerelease  - Manage pre-release identifiers (includes stability setting)
  config metadata    - Manage build metadata (includes stability setting)
  config custom      - Manage custom key-value pairs
  config vars        - Show all available template variables`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
