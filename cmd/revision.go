package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var revisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "Manage revision version (4th component for .NET)",
	Long: `Commands to increment or decrement the revision version component.

The revision is the 4th version component, primarily used in .NET ecosystems
where the version format is Major.Minor.Build.Revision (e.g., 1.2.3.4).

When revision is non-zero, the full version format is:
  Major.Minor.Patch.Revision[-PreRelease][+BuildMetadata]

Examples:
  1.2.3.4
  1.2.3.4-alpha.1
  1.2.3.4+build.123
  1.2.3.4-beta.2+build.456

When revision is 0, it is omitted from the version string to maintain
standard SemVer compatibility:
  1.2.3 (not 1.2.3.0)`,
}

var revisionIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "bump", "+"},
	Short:   "Increment revision version",
	Long:    "Increment the revision version (4th component, primarily for .NET)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Increment(version.RevisionLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Revision version incremented to: %s\n", ver)
		return nil
	},
}

var revisionDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement revision version",
	Long:    "Decrement the revision version (4th component, primarily for .NET)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Decrement(version.RevisionLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Revision version decremented to: %s\n", ver)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(revisionCmd)
	revisionCmd.AddCommand(revisionIncrementCmd)
	revisionCmd.AddCommand(revisionDecrementCmd)
}
