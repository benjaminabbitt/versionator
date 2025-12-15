package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Manage patch version",
	Long:  "Commands to increment or decrement the patch version component",
}

var patchIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "+"},
	Short:   "Increment patch version",
	Long:    "Increment the patch version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Increment(version.PatchLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Patch version incremented to: %s\n", ver)
		return nil
	},
}

var patchDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement patch version",
	Long:    "Decrement the patch version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Decrement(version.PatchLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Patch version decremented to: %s\n", ver)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
	patchCmd.AddCommand(patchIncrementCmd)
	patchCmd.AddCommand(patchDecrementCmd)
}
