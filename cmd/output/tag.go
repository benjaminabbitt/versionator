package output

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var TagCmd = &cobra.Command{
	Use:     "tag",
	Aliases: []string{"commit"},
	Short:   "Create a git tag for the current version",
	Long: `Create a git tag for the current version from VERSION file.

This command reads the VERSION file and creates a git tag with that exact value.
VERSION is the source of truth - use 'version render' first to update it with
fresh prerelease/metadata from config elements.

Workflow:
  versionator version patch bump    # Bump version, renders from config
  versionator output tag            # Create git tag from VERSION

The command will fail if:
- Not in a git repository
- Working directory has uncommitted changes
- Tag already exists (unless --force is used)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get active VCS
		activeVCS := vcs.GetActiveVCS()
		if activeVCS == nil {
			return fmt.Errorf("not in a version control repository")
		}

		// Check if working directory is clean
		clean, err := activeVCS.IsWorkingDirectoryClean()
		if err != nil {
			return fmt.Errorf("error checking %s status: %w", activeVCS.Name(), err)
		}

		if !clean {
			return fmt.Errorf("working directory is not clean. Please commit or stash your changes first")
		}

		// Get current version data from VERSION file
		// VERSION is the source of truth - no re-rendering here
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting current version: %w", err)
		}

		// Tag uses the full version string from VERSION file
		tagName := vd.FullString()

		// Check if tag already exists
		exists, err := activeVCS.TagExists(tagName)
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
			message = fmt.Sprintf("Release %s", vd.String())
		}

		// Create the tag
		if err := activeVCS.CreateTag(tagName, message); err != nil {
			return fmt.Errorf("error creating tag: %w", err)
		}

		cmd.Printf("Successfully created tag '%s' using %s\n", tagName, activeVCS.Name())

		// Show additional information if verbosity is enabled
		if logging.GetVerbosity() > 0 {
			cmd.Printf("  Message: %s\n", message)

			// Get current VCS identifier
			if identifier, err := activeVCS.GetVCSIdentifier(7); err == nil {
				cmd.Printf("  %s ID: %s\n", activeVCS.Name(), identifier)
			}
		}

		return nil
	},
}

func init() {
	TagCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	TagCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
}
