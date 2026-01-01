package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/cmd/output"
	"github.com/spf13/cobra"
)

// commitCmd is a deprecated alias for 'output tag'
var commitCmd = &cobra.Command{
	Use:        "commit",
	Short:      "Create a git tag (deprecated: use 'output tag')",
	Long:       "DEPRECATED: Use 'versionator output tag' instead.\n\nThis command is an alias for 'output tag' for backwards compatibility.",
	Deprecated: "use 'output tag' instead",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.ErrOrStderr(), "Note: 'commit' is deprecated. Use 'output tag' instead.")

		// Pass flags from commit command to tag command
		message, _ := cmd.Flags().GetString("message")
		force, _ := cmd.Flags().GetBool("force")
		output.TagCmd.Flags().Set("message", message)
		output.TagCmd.Flags().Set("force", fmt.Sprintf("%t", force))

		// Copy output settings
		output.TagCmd.SetOut(cmd.OutOrStdout())
		output.TagCmd.SetErr(cmd.ErrOrStderr())

		return output.TagCmd.RunE(output.TagCmd, args)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	// Mirror flags from output.TagCmd for compatibility
	commitCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	commitCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
}
