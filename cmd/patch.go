package cmd

import (
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		if err := version.Increment(version.PatchLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		version, err := version.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading updated version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Patch version incremented to: %s\n", version)
	},
}

var patchDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "-"},
	Short:   "Decrement patch version",
	Long:    "Decrement the patch version",
	Run: func(cmd *cobra.Command, args []string) {
		if err := version.Decrement(version.PatchLevel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		version, err := version.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading updated version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Patch version decremented to: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
	patchCmd.AddCommand(patchIncrementCmd)
	patchCmd.AddCommand(patchDecrementCmd)
}
