package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a git tag for the current version",
	Long: `Create a git tag for the current version after ensuring the working directory is clean.

This command will:
1. Check that you're in a git repository
2. Verify there are no uncommitted changes
3. Get the current version
4. Create a git tag with the version (prefixed with 'v')

The command will fail if there are uncommitted changes or if the tag already exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get VCS from app instance
		vcsInstance := appInstance.VCS
		if vcsInstance == nil {
			return fmt.Errorf("not in a version control repository")
		}

		// Check if working directory is clean
		clean, err := vcsInstance.IsWorkingDirectoryClean()
		if err != nil {
			return fmt.Errorf("error checking %s status: %w", vcsInstance.Name(), err)
		}

		if !clean {
			return fmt.Errorf("working directory is not clean. Please commit or stash your changes first")
		}

		// Get current version
		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error getting current version: %w", err)
		}

		// Create tag name with 'v' prefix
		tagName := "v" + version

		// Check if custom tag prefix is specified
		prefix, _ := cmd.Flags().GetString("prefix")
		if prefix != "" {
			tagName = prefix + version
		}

		// Check if tag already exists
		exists, err := vcsInstance.TagExists(tagName)
		if err != nil {
			return fmt.Errorf("error checking if tag exists: %w", err)
		}

		if exists {
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				return fmt.Errorf("tag '%s' already exists. Use --force to overwrite", tagName)
			}
		}

		// Get custom message or use default
		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("Release %s", version)
		}

		// Create the tag
		if err := vcsInstance.CreateTag(tagName, message); err != nil {
			return fmt.Errorf("error creating tag: %w", err)
		}

		cmd.Printf("✓ Successfully created tag '%s' for version %s using %s\n", tagName, version, vcsInstance.Name())

		// Show additional information if requested
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			cmd.Printf("  Message: %s\n", message)

			// Get current VCS identifier
			if identifier, err := vcsInstance.GetVCSIdentifier(7); err == nil {
				cmd.Printf("  %s ID: %s\n", vcsInstance.Name(), identifier)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	// Add flags
	commitCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	commitCmd.Flags().StringP("prefix", "p", "v", "Tag prefix (default: 'v')")
	commitCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
	commitCmd.Flags().BoolP("verbose", "v", false, "Show additional information")
}
