package version

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Manage patch version",
	Long:  "Commands to increment or decrement the patch version component",
}

var patchIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "bump", "up", "+"},
	Short:   "Increment patch version",
	Long:    "Increment the patch version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Increment(version.PatchLevel); err != nil {
			return err
		}

		// Render prerelease/metadata from config elements
		if err := RenderFromConfig(); err != nil {
			return fmt.Errorf("error rendering from config: %w", err)
		}

		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Patch version incremented to: %s\n", vd.FullString())
		return nil
	},
}

var patchDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "down", "-"},
	Short:   "Decrement patch version",
	Long:    "Decrement the patch version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Decrement(version.PatchLevel); err != nil {
			return err
		}

		// Render prerelease/metadata from config elements
		if err := RenderFromConfig(); err != nil {
			return fmt.Errorf("error rendering from config: %w", err)
		}

		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Patch version decremented to: %s\n", vd.FullString())
		return nil
	},
}

func init() {
	PatchCmd.AddCommand(patchIncrementCmd)
	PatchCmd.AddCommand(patchDecrementCmd)
}
