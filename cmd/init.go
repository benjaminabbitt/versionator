package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var (
	initVersion     string
	initPrefix      string
	initWithConfig  bool
	initForce       bool
	hookUninstall   bool
)

const postCommitHookScript = `#!/bin/sh
# versionator post-commit hook
# Automatically bump VERSION based on +semver: tags in commit messages

# Get the commit message
MSG=$(git log -1 --pretty=%B)

# Check if message contains +semver: tag (major, minor, or patch)
if echo "$MSG" | grep -qE '\+semver:(major|minor|patch)'; then
    versionator bump
fi
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize versionator in this directory",
	Long: `Initialize versionator by creating a VERSION file.

Creates a VERSION file with the specified initial version and prefix.
Optionally creates a .versionator.yaml configuration file.

Only 'v' or 'V' prefixes are allowed per SemVer convention.

Examples:
  versionator init                        # Create VERSION with 0.0.1
  versionator init --version 1.0.0        # Create VERSION with 1.0.0
  versionator init --prefix v             # Create VERSION with v0.0.1
  versionator init --config               # Also create .versionator.yaml
  versionator init --force                # Overwrite existing files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		versionPath := "VERSION"
		configPath := ".versionator.yaml"

		// Validate prefix early with clear error message
		if initPrefix != "" && initPrefix != "v" && initPrefix != "V" {
			return fmt.Errorf("invalid prefix %q: only 'v' or 'V' allowed per SemVer convention", initPrefix)
		}

		// Check if VERSION exists
		if _, err := os.Stat(versionPath); err == nil && !initForce {
			return fmt.Errorf("VERSION file already exists (use --force to overwrite)")
		}

		// Check if config exists when --config is specified
		if initWithConfig {
			if _, err := os.Stat(configPath); err == nil && !initForce {
				return fmt.Errorf(".versionator.yaml already exists (use --force to overwrite)")
			}
		}

		// Parse the initial version
		v := version.Parse(initVersion)
		if initPrefix != "" {
			v.Prefix = initPrefix
		}

		// Validate version
		if err := v.Validate(); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}

		// Write VERSION file
		content := v.FullString() + "\n"
		if err := os.WriteFile(versionPath, []byte(content), FilePermission); err != nil {
			return fmt.Errorf("error writing VERSION file: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created VERSION: %s\n", v.FullString())

		// Write config if requested
		if initWithConfig {
			defaultConfig := config.DefaultConfigYAML()
			if err := os.WriteFile(configPath, []byte(defaultConfig), FilePermission); err != nil {
				return fmt.Errorf("error writing .versionator.yaml: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created .versionator.yaml\n")
		}

		return nil
	},
}

var initHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Install post-commit hook for automatic version bumping",
	Long: `Install a git post-commit hook that runs 'versionator bump'.

This automatically bumps the VERSION file based on +semver: tags in commit
messages and amends the commit to include the VERSION change.

The hook only triggers when the commit message contains:
  +semver:major - Bump major version
  +semver:minor - Bump minor version
  +semver:patch - Bump patch version

Examples:
  versionator init hook              # Install the post-commit hook
  versionator init hook --uninstall  # Remove the post-commit hook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get active VCS
		activeVCS := vcs.GetActiveVCS()
		if activeVCS == nil {
			return fmt.Errorf("not in a git repository")
		}

		hooksPath, err := activeVCS.GetHooksPath()
		if err != nil {
			return fmt.Errorf("failed to get hooks path: %w", err)
		}

		hookPath := filepath.Join(hooksPath, "post-commit")

		if hookUninstall {
			// Check if hook exists
			if _, err := os.Stat(hookPath); os.IsNotExist(err) {
				cmd.Println("No post-commit hook installed")
				return nil
			}

			// Read existing hook to check if it's ours
			content, err := os.ReadFile(hookPath)
			if err != nil {
				return fmt.Errorf("failed to read hook: %w", err)
			}

			if string(content) != postCommitHookScript {
				return fmt.Errorf("post-commit hook exists but was not installed by versionator (use --force to remove anyway)")
			}

			if err := os.Remove(hookPath); err != nil {
				return fmt.Errorf("failed to remove hook: %w", err)
			}
			cmd.Println("Removed post-commit hook")
			return nil
		}

		// Install hook
		// Check if hook already exists
		if _, err := os.Stat(hookPath); err == nil && !initForce {
			return fmt.Errorf("post-commit hook already exists (use --force to overwrite)")
		}

		// Ensure hooks directory exists
		if err := os.MkdirAll(hooksPath, 0755); err != nil {
			return fmt.Errorf("failed to create hooks directory: %w", err)
		}

		// Write hook script
		if err := os.WriteFile(hookPath, []byte(postCommitHookScript), 0755); err != nil {
			return fmt.Errorf("failed to write hook: %w", err)
		}

		cmd.Println("Installed post-commit hook")
		cmd.Println("Commits with +semver:major/minor/patch will automatically bump VERSION")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initVersion, "version", "v", "0.0.1", "Initial version")
	initCmd.Flags().StringVarP(&initPrefix, "prefix", "p", "", "Version prefix ('v' or 'V' only)")
	initCmd.Flags().BoolVar(&initWithConfig, "config", false, "Also create .versionator.yaml")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing files")

	initHookCmd.Flags().BoolVar(&hookUninstall, "uninstall", false, "Remove the post-commit hook")
	initHookCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing hook")

	initCmd.AddCommand(initHookCmd)
	rootCmd.AddCommand(initCmd)
}
