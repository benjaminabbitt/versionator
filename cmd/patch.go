package cmd

import (
	"fmt"
	"versionator/internal/version"

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
		if err := appInstance.Increment(version.PatchLevel); err != nil {
			return fmt.Errorf("error incrementing patch version: %w", err)
		}

		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		cmd.Printf("Patch version incremented to: %s\n", version)
		return nil
	},
}

var patchDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec"},
	Short:   "Decrement patch version",
	Long:    "Decrement the patch version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := appInstance.Decrement(version.PatchLevel); err != nil {
			return fmt.Errorf("error decrementing patch version: %w", err)
		}

		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		cmd.Printf("Patch version decremented to: %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
	patchCmd.AddCommand(patchIncrementCmd)
	patchCmd.AddCommand(patchDecrementCmd)
}
