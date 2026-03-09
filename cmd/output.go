package cmd

import (
	"github.com/spf13/cobra"
)

var outputCmd = &cobra.Command{
	Use:   "output",
	Short: "Output version in various formats",
	Long: `Output the current version in various formats for different use cases.

Use subcommands to output version information:
  output version  - Show current version (with optional template)
  output emit     - Generate version files for programming languages
  output ci       - Output version variables for CI/CD systems`,
}

func init() {
	rootCmd.AddCommand(outputCmd)
}
