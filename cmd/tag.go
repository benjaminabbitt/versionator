package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create git tag and release branch for current version",
	Long: `Create a git tag and release branch for the current version.

This command will:
1. Check that you're in a git repository
2. Verify there are no uncommitted changes
3. Get the current version
4. Create a git tag with the version (prefixed with 'v')
5. Create a release branch (e.g., 'release/v1.2.3') if enabled

Release branch creation is enabled by default. Configure in .versionator.yaml:
  release:
    createBranch: true    # set to false to disable
    branchPrefix: "release/"

Use --no-branch to skip branch creation for a single invocation.

The command will fail if there are uncommitted changes or if the tag already exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get active VCS
		vcs := vcs.GetActiveVCS()
		if vcs == nil {
			return fmt.Errorf("not in a version control repository")
		}

		// Check if working directory is clean
		clean, err := vcs.IsWorkingDirectoryClean()
		if err != nil {
			return fmt.Errorf("error checking %s status: %w", vcs.Name(), err)
		}

		if !clean {
			return fmt.Errorf("working directory is not clean. Please commit or stash your changes first")
		}

		// Get current version data (includes prefix)
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting current version: %w", err)
		}

		// Use VERSION file prefix by default, or command-line override
		prefix := vd.Prefix
		if cmdPrefix, _ := cmd.Flags().GetString("prefix"); cmdPrefix != "" {
			prefix = cmdPrefix
		}
		// If no prefix configured anywhere, default to "v"
		if prefix == "" {
			prefix = "v"
		}

		tagName := prefix + vd.String()

		// Check if tag already exists
		exists, err := vcs.TagExists(tagName)
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
		if err := vcs.CreateTag(tagName, message); err != nil {
			return fmt.Errorf("error creating tag: %w", err)
		}

		cmd.Printf("Successfully created tag '%s' for version %s using %s\n", tagName, vd.String(), vcs.Name())

		// Read config for release branch settings
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		// Check command-line flag for branch creation (overrides config)
		noBranch, _ := cmd.Flags().GetBool("no-branch")
		createBranch := cfg.Release.CreateBranch && !noBranch

		if createBranch {
			branchName := cfg.Release.BranchPrefix + tagName

			// Check if branch already exists
			branchExists, err := vcs.BranchExists(branchName)
			if err != nil {
				return fmt.Errorf("error checking if branch exists: %w", err)
			}

			if branchExists {
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					cmd.Printf("Warning: branch '%s' already exists, skipping branch creation\n", branchName)
				}
			} else {
				if err := vcs.CreateBranch(branchName); err != nil {
					return fmt.Errorf("error creating release branch: %w", err)
				}
				cmd.Printf("Successfully created branch '%s'\n", branchName)
			}
		}

		// Show additional information if requested
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			cmd.Printf("  Message: %s\n", message)

			// Get current VCS identifier
			if identifier, err := vcs.GetVCSIdentifier(7); err == nil {
				cmd.Printf("  %s ID: %s\n", vcs.Name(), identifier)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)

	// Add flags
	tagCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	tagCmd.Flags().StringP("prefix", "p", "v", "Tag prefix (default: 'v')")
	tagCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
	tagCmd.Flags().BoolP("verbose", "v", false, "Show additional information")
	tagCmd.Flags().Bool("no-branch", false, "Skip creating release branch")
}
