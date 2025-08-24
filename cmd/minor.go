package cmd

import (
	"fmt"
	"versionator/internal/version"

	"github.com/spf13/cobra"
)

var minorCmd = &cobra.Command{
	Use:   "minor",
	Short: "Manage minor version",
	Long:  "Commands to increment or decrement the minor version component",
}

var minorIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "+"},
	Short:   "Increment minor version",
	Long:    "Increment the minor version and reset patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := appInstance.Increment(version.MinorLevel); err != nil {
			return fmt.Errorf("error incrementing minor version: %w", err)
		}

		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		cmd.Printf("Minor version incremented to: %s\n", version)
		return nil
	},
}

var minorDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec"},
	Short:   "Decrement minor version",
	Long:    "Decrement the minor version and reset patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := appInstance.Decrement(version.MinorLevel); err != nil {
			return fmt.Errorf("error decrementing minor version: %w", err)
		}

		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error reading updated version: %w", err)
		}

		cmd.Printf("Minor version decremented to: %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(minorCmd)
	minorCmd.AddCommand(minorIncrementCmd)
	minorCmd.AddCommand(minorDecrementCmd)
}
