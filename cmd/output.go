package cmd

import (
	"github.com/benjaminabbitt/versionator/cmd/output"
	"github.com/spf13/cobra"
)

var outputCmd = &cobra.Command{
	Use:     "output",
	Aliases: []string{"out"},
	Short:   "Output version information in various formats",
	Long: `Output version information to files, build systems, or manifests.

Subcommands:
  patch  - Patch version in manifest files (pyproject.toml, package.json, etc.)
  build  - Generate build flags for version injection (Go ldflags, Rust cfg, etc.)
  file   - Generate version source files for compilation

Examples:
  # Patch all manifest files
  versionator output patch

  # Generate Go ldflags
  versionator output build build-go --var main.Version

  # Generate Python version file
  versionator output file emit-python --output _version.py`,
}

func init() {
	rootCmd.AddCommand(outputCmd)

	// Add output subcommands from subpackage
	outputCmd.AddCommand(output.BuildCmd)
	outputCmd.AddCommand(output.FileCmd)
	outputCmd.AddCommand(output.PatchCmd)
	outputCmd.AddCommand(output.TagCmd)
}
