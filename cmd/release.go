package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create git tag and release branch for current version",
	Long: `Create a git tag and release branch for the current version.

This command will:
1. Check that you're in a git repository
2. If only the VERSION file is dirty, commit it automatically
3. Verify there are no other uncommitted changes
4. Get the current version
5. Create a git tag with the version (prefixed with 'v')
6. Create a release branch (e.g., 'release/v1.2.3') if enabled

This is the recommended workflow after bumping a version:
  versionator patch increment
  versionator release

Release branch creation is enabled by default. Configure in .versionator.yaml:
  release:
    createBranch: true    # set to false to disable
    branchPrefix: "release/"

Use --no-branch to skip branch creation for a single invocation.

The command will fail if there are uncommitted changes (other than VERSION)
or if the tag already exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get active VCS
		vcsImpl := vcs.GetActiveVCS()
		if vcsImpl == nil {
			return fmt.Errorf("not in a version control repository")
		}

		// Check if working directory is clean
		clean, err := vcsImpl.IsWorkingDirectoryClean()
		if err != nil {
			return fmt.Errorf("error checking %s status: %w", vcsImpl.Name(), err)
		}

		if !clean {
			// Check if only VERSION file is dirty
			dirtyFiles, err := vcsImpl.GetDirtyFiles()
			if err != nil {
				return fmt.Errorf("error getting dirty files: %w", err)
			}

			// Only auto-commit if exactly VERSION file is dirty
			if len(dirtyFiles) == 1 && dirtyFiles[0] == "VERSION" {
				// Load version to get the version string for commit message
				vd, err := version.Load()
				if err != nil {
					return fmt.Errorf("error loading version: %w", err)
				}

				commitMsg := fmt.Sprintf("Release %s", vd.String())
				if err := vcsImpl.CommitFiles([]string{"VERSION"}, commitMsg); err != nil {
					return fmt.Errorf("error committing VERSION file: %w", err)
				}
				cmd.Printf("Committed VERSION file: %s\n", commitMsg)
			} else {
				return fmt.Errorf("working directory is not clean. Please commit or stash your changes first (dirty files: %v)", dirtyFiles)
			}
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
		exists, err := vcsImpl.TagExists(tagName)
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
		if err := vcsImpl.CreateTag(tagName, message); err != nil {
			return fmt.Errorf("error creating tag: %w", err)
		}

		cmd.Printf("Successfully created tag '%s' for version %s using %s\n", tagName, vd.String(), vcsImpl.Name())

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
			branchExists, err := vcsImpl.BranchExists(branchName)
			if err != nil {
				return fmt.Errorf("error checking if branch exists: %w", err)
			}

			if branchExists {
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					cmd.Printf("Warning: branch '%s' already exists, skipping branch creation\n", branchName)
				}
			} else {
				if err := vcsImpl.CreateBranch(branchName); err != nil {
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
			if identifier, err := vcsImpl.GetVCSIdentifier(7); err == nil {
				cmd.Printf("  %s ID: %s\n", vcsImpl.Name(), identifier)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Add flags
	releaseCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	releaseCmd.Flags().StringP("prefix", "p", "v", "Tag prefix (default: 'v')")
	releaseCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
	releaseCmd.Flags().BoolP("verbose", "v", false, "Show additional information")
	releaseCmd.Flags().Bool("no-branch", false, "Skip creating release branch")
}
