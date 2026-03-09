package cmd

import (
	"github.com/spf13/cobra"
)

var supportCmd = &cobra.Command{
	Use:   "support",
	Short: "Shell completion and tooling support",
	Long: `Support commands for shell completion and tooling integration.

Use subcommands for different support features:
  support completion  - Generate shell completion scripts
  support schema      - Generate machine-readable CLI schema for tooling`,
}

func init() {
	rootCmd.AddCommand(supportCmd)
}
