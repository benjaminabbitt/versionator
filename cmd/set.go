package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <version>",
	Short: "Set version to an arbitrary value",
	Long: `Set the VERSION file to an arbitrary version string.

Accepts any version string the parser grammar supports:
  [v|V]Major.Minor[.Patch[.Revision]][-PreRelease][+BuildMetadata]

Examples:
  versionator set 1.2.3
  versionator set v2.0.0-rc.1
  versionator set 1.2.3.4
  versionator set 1.0.0-alpha.1+build.42`,
	Args: cobra.ExactArgs(1),
	RunE: runSet,
}

func runSet(cmd *cobra.Command, args []string) error {
	if err := version.SetVersion(args[0]); err != nil {
		return err
	}

	v, err := version.Load()
	if err != nil {
		return fmt.Errorf("version set but error reading back: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Version set to: %s\n", v.FullString())

	if err := runConfiguredUpdates(cmd); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(setCmd)
}
