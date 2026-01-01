package version

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var MajorCmd = &cobra.Command{
	Use:   "major",
	Short: "Manage major version",
	Long:  "Commands to increment or decrement the major version component",
}

var majorIncrementCmd = &cobra.Command{
	Use:     "increment",
	Aliases: []string{"inc", "bump", "up", "+"},
	Short:   "Increment major version",
	Long:    "Increment the major version and reset minor and patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Increment(version.MajorLevel); err != nil {
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

		fmt.Printf("Major version incremented to: %s\n", vd.FullString())
		return nil
	},
}

var majorDecrementCmd = &cobra.Command{
	Use:     "decrement",
	Aliases: []string{"dec", "down", "-"},
	Short:   "Decrement major version",
	Long:    "Decrement the major version and reset minor and patch to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.Decrement(version.MajorLevel); err != nil {
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

		fmt.Printf("Major version decremented to: %s\n", vd.FullString())
		return nil
	},
}

func init() {
	MajorCmd.AddCommand(majorIncrementCmd)
	MajorCmd.AddCommand(majorDecrementCmd)
}
