package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var minorCmd = &cobra.Command{
	Use:   "minor",
	Short: "Manage minor version",
	Long:  "Commands to increment or decrement the minor version component",
}

var minorIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "bump", "+"},
	Short:   "Increment minor version",
	Long:    "Increment the minor version and reset patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Increment(version.MinorLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Minor version incremented to: %s\n", ver)
		return nil
	},
}

var minorDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement minor version",
	Long:    "Decrement the minor version and reset patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Decrement(version.MinorLevel); err != nil {
			return err
		}

		ver, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		fmt.Printf("Minor version decremented to: %s\n", ver)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(minorCmd)
	minorCmd.AddCommand(minorIncrementCmd)
	minorCmd.AddCommand(minorDecrementCmd)
}
