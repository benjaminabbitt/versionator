package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var majorCmd = &cobra.Command{
	Use:   "major",
	Short: "Manage major version",
	Long:  "Commands to increment or decrement the major version component",
}

var majorIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "bump", "+"},
	Short:   "Increment major version",
	Long:    "Increment the major version and reset minor and patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Increment(version.MajorLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Major version incremented to: %s\n", ver)
		return nil
	},
}

var majorDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement major version",
	Long:    "Decrement the major version and reset minor and patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Decrement(version.MajorLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Major version decremented to: %s\n", ver)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(majorCmd)
	majorCmd.AddCommand(majorIncrementCmd)
	majorCmd.AddCommand(majorDecrementCmd)
}
