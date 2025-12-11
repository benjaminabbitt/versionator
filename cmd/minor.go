package cmd

import (
	"fmt"
	"os"
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
	Aliases: []string{"inc", "+"},
	Short:   "Increment minor version",
	Long:    "Increment the minor version and reset patch to 0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := version.Increment(version.MinorLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		version, err := version.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading updated version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Minor version incremented to: %s\n", version)
	},
}

var minorDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement minor version",
	Long:    "Decrement the minor version and reset patch to 0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := version.Decrement(version.MinorLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		version, err := version.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading updated version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Minor version decremented to: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(minorCmd)
	minorCmd.AddCommand(minorIncrementCmd)
	minorCmd.AddCommand(minorDecrementCmd)
}
