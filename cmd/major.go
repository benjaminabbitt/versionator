package cmd

import (
	"fmt"
	"os"
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
	Aliases: []string{"inc", "+"},
	Short:   "Increment major version",
	Long:    "Increment the major version and reset minor and patch to 0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := version.Increment(version.MajorLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		version, err := version.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading updated version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Major version incremented to: %s\n", version)
	},
}

var majorDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement major version",
	Long:    "Decrement the major version and reset minor and patch to 0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := version.Decrement(version.MajorLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		version, err := version.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading updated version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Major version decremented to: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(majorCmd)
	majorCmd.AddCommand(majorIncrementCmd)
	majorCmd.AddCommand(majorDecrementCmd)
}
