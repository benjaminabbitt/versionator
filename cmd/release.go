package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/update"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// releaseResult holds the results of a release operation
type releaseResult struct {
	tagName    string
	branchName string
	vcsImpl    vcs.VersionControlSystem
}

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
	RunE: runReleaseCmd,
}

func runReleaseCmd(cmd *cobra.Command, args []string) error {
	_, err := runRelease(cmd)
	return err
}

var releasePushCmd = &cobra.Command{
	Use:   "push",
	Short: "Create release and push tag and branch to remote",
	Long: `Create a release (tag and branch) and push both to the remote repository.

This combines the release command with git push operations:
1. Perform the standard release (create tag and branch)
2. Push the tag to origin
3. Push the release branch to origin (if created)

Example:
  versionator release push           # Release and push to remote
  versionator release push --no-branch  # Release and push tag only`,
	RunE: runReleasePush,
}

func runReleasePush(cmd *cobra.Command, args []string) error {
	result, err := runRelease(cmd)
	if err != nil {
		return err
	}

	// Push the tag
	cmd.Printf("Pushing tag '%s' to remote...\n", result.tagName)
	if err := result.vcsImpl.PushTag(result.tagName); err != nil {
		return fmt.Errorf("failed to push tag: %w", err)
	}
	cmd.Printf("Successfully pushed tag '%s'\n", result.tagName)

	// Push the branch if it was created
	if result.branchName != "" {
		cmd.Printf("Pushing branch '%s' to remote...\n", result.branchName)
		if err := result.vcsImpl.PushBranch(result.branchName); err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}
		cmd.Printf("Successfully pushed branch '%s'\n", result.branchName)
	}

	return nil
}

func runRelease(cmd *cobra.Command) (*releaseResult, error) {
	// Get active VCS
	vcsImpl := vcs.GetActiveVCS()
	if vcsImpl == nil {
		return nil, fmt.Errorf("not in a version control repository")
	}

	// Read config early for file updates
	cfg, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	// Build set of allowed dirty files: VERSION + .versionator.yaml + files from updates config
	allowedDirty := map[string]bool{"VERSION": true, ".versionator.yaml": true}
	for _, u := range cfg.Updates {
		allowedDirty[u.File] = true
	}

	// Check if working directory is clean
	clean, err := vcsImpl.IsWorkingDirectoryClean()
	if err != nil {
		return nil, fmt.Errorf("error checking %s status: %w", vcsImpl.Name(), err)
	}

	var versionDirty bool
	if !clean {
		// Check what files are dirty
		dirtyFiles, err := vcsImpl.GetDirtyFiles()
		if err != nil {
			return nil, fmt.Errorf("error getting dirty files: %w", err)
		}

		// If no updates configured, use original behavior: only allow VERSION dirty
		if len(cfg.Updates) == 0 {
			if len(dirtyFiles) == 1 && dirtyFiles[0] == "VERSION" {
				// Load version to get the version string for commit message
				vd, err := version.Load()
				if err != nil {
					return nil, fmt.Errorf("error loading version: %w", err)
				}

				commitMsg := fmt.Sprintf("Release %s", vd.String())
				if err := vcsImpl.CommitFiles([]string{"VERSION"}, commitMsg); err != nil {
					return nil, fmt.Errorf("error committing VERSION file: %w", err)
				}
				cmd.Printf("Committed VERSION file: %s\n", commitMsg)
			} else {
				return nil, fmt.Errorf("working directory is not clean. Please commit or stash your changes first (dirty files: %v)", dirtyFiles)
			}
		} else {
			// With updates configured, allow VERSION + update target files to be dirty
			for _, f := range dirtyFiles {
				if !allowedDirty[f] {
					return nil, fmt.Errorf("working directory is not clean. Please commit or stash your changes first (dirty files: %v)", dirtyFiles)
				}
				if f == "VERSION" {
					versionDirty = true
				}
			}
		}
	}

	// Get current version data (includes prefix)
	vd, err := version.Load()
	if err != nil {
		return nil, fmt.Errorf("error getting current version: %w", err)
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
		return nil, fmt.Errorf("error checking if tag exists: %w", err)
	}

	if exists {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return nil, fmt.Errorf("tag '%s' already exists. Use --force to overwrite", tagName)
		}
	}

	// Get custom message or use default
	message, _ := cmd.Flags().GetString("message")
	if message == "" {
		message = fmt.Sprintf("Release %s", vd.String())
	}

	// Apply file updates if configured
	var updatedFiles []string
	if len(cfg.Updates) > 0 {
		logger, _ := zap.NewProduction()
		updater := update.NewUpdater(cfg.Updates, update.NewDaselFileParser(), logger)

		// Build template data for rendering update templates
		templateData := emit.BuildCompleteTemplateData(vd, cfg.PreRelease.Template, cfg.Metadata.Template)

		if err := updater.UpdateFiles(templateData); err != nil {
			return nil, fmt.Errorf("error updating files: %w", err)
		}
		updatedFiles = updater.GetFilesToCommit()
		cmd.Printf("Updated %d file(s)\n", len(updatedFiles))
	}

	// Commit VERSION + updated files if there are changes to commit
	filesToCommit := make([]string, 0)
	if versionDirty {
		filesToCommit = append(filesToCommit, "VERSION")
	}
	filesToCommit = append(filesToCommit, updatedFiles...)

	if len(filesToCommit) > 0 {
		commitMsg := fmt.Sprintf("Release %s", vd.String())
		if err := vcsImpl.CommitFiles(filesToCommit, commitMsg); err != nil {
			return nil, fmt.Errorf("error committing release files: %w", err)
		}
		cmd.Printf("Committed: %v\n", filesToCommit)
	}

	// Create the tag
	if err := vcsImpl.CreateTag(tagName, message); err != nil {
		return nil, fmt.Errorf("error creating tag: %w", err)
	}

	cmd.Printf("Successfully created tag '%s' for version %s using %s\n", tagName, vd.String(), vcsImpl.Name())

	result := &releaseResult{
		tagName: tagName,
		vcsImpl: vcsImpl,
	}

	// Check command-line flag for branch creation (overrides config)
	noBranch, _ := cmd.Flags().GetBool("no-branch")
	createBranch := cfg.Release.CreateBranch && !noBranch

	if createBranch {
		branchName := cfg.Release.BranchPrefix + tagName

		// Check if branch already exists
		branchExists, err := vcsImpl.BranchExists(branchName)
		if err != nil {
			return nil, fmt.Errorf("error checking if branch exists: %w", err)
		}

		if branchExists {
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				cmd.Printf("Warning: branch '%s' already exists, skipping branch creation\n", branchName)
			}
		} else {
			if err := vcsImpl.CreateBranch(branchName); err != nil {
				return nil, fmt.Errorf("error creating release branch: %w", err)
			}
			cmd.Printf("Successfully created branch '%s'\n", branchName)
			result.branchName = branchName
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

	return result, nil
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Add flags to release command
	releaseCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	releaseCmd.Flags().StringP("prefix", "p", "v", "Tag prefix (default: 'v')")
	releaseCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
	releaseCmd.Flags().BoolP("verbose", "v", false, "Show additional information")
	releaseCmd.Flags().Bool("no-branch", false, "Skip creating release branch")

	// Add push subcommand
	releaseCmd.AddCommand(releasePushCmd)

	// Add same flags to push subcommand
	releasePushCmd.Flags().StringP("message", "m", "", "Tag message (default: 'Release <version>')")
	releasePushCmd.Flags().StringP("prefix", "p", "v", "Tag prefix (default: 'v')")
	releasePushCmd.Flags().BoolP("force", "f", false, "Force creation even if tag exists")
	releasePushCmd.Flags().BoolP("verbose", "v", false, "Show additional information")
	releasePushCmd.Flags().Bool("no-branch", false, "Skip creating release branch")
}
