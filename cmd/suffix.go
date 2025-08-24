package cmd

import (
	"fmt"
	"os"
	"versionator/internal/config"
	"versionator/internal/logging"
	"versionator/internal/vcs"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/spf13/cobra"
)

var suffixCmd = &cobra.Command{
	Use:   "suffix",
	Short: "Manage version suffix behavior",
	Long:  "Commands to enable or disable appending git hash suffix to version numbers",
}

var suffixEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable VCS identifier suffix",
	Long:  "Enable appending VCS identifier to version numbers (format: version-identifier)",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()

		cfg, err := config.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Suffix.Enabled = true
		cfg.Suffix.Type = "git"

		if err := config.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		fmt.Println("Git hash suffix enabled")

		// Show current version with suffix
		version, err := versionator.GetVersionWithSuffix()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		gitVCS := vcs.GetVCS("git")
		if gitVCS != nil && gitVCS.IsRepository() {
			fmt.Printf("Current version: %s\n", version)
		} else {
			fmt.Printf("Current version: %s (Git hash will be added when in a repository)\n", version)
		}
	},
}

var suffixDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable git hash suffix",
	Long:  "Disable appending git hash to version numbers",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()

		cfg, err := config.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Suffix.Enabled = false

		if err := config.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		fmt.Println("Git hash suffix disabled")

		// Show current version without suffix
		version, err := version.GetCurrentVersion()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		fmt.Printf("Current version: %s\n", version)
	},
}

var suffixStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show suffix configuration status",
	Long:  "Show whether git hash suffix is enabled or disabled",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ReadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
			os.Exit(1)
		}

		if cfg.Suffix.Enabled && cfg.Suffix.Type == "git" {
			fmt.Println("Git hash suffix: ENABLED")
			fmt.Printf("Hash length: %d characters\n", cfg.Suffix.Git.HashLength)

			gitVCS := vcs.GetVCS("git")
			if gitVCS != nil && gitVCS.IsRepository() {
				gitHash, err := gitVCS.GetVCSIdentifier(cfg.Suffix.Git.HashLength)
				if err != nil {
					fmt.Printf("Git repository detected, but cannot get hash: %v\n", err)
				} else {
					fmt.Printf("Git hash: %s\n", gitHash)
				}
			} else {
				fmt.Println("Not in a git repository - hash will be applied when in git repo")
			}
		} else {
			fmt.Println("Git hash suffix: DISABLED")
		}

		// Show current version
		version, err := versionator.GetVersionWithSuffix()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current version: %s\n", version)
	},
}

var suffixConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure git hash settings",
	Long:  "Configure git hash settings (currently uses environment variable VERSIONATOR_HASH_LENGTH)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ReadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current configuration:\n")
		fmt.Printf("  Git hash suffix enabled: %t\n", cfg.Suffix.Enabled)
		fmt.Printf("  Suffix type: %s\n", cfg.Suffix.Type)
		fmt.Printf("  Git hash length: %d\n", cfg.Suffix.Git.HashLength)
		fmt.Printf("\nConfiguration is stored in .versionator.yaml\n")
	},
}

func init() {
	rootCmd.AddCommand(suffixCmd)
	suffixCmd.AddCommand(suffixEnableCmd)
	suffixCmd.AddCommand(suffixDisableCmd)
	suffixCmd.AddCommand(suffixStatusCmd)
	suffixCmd.AddCommand(suffixConfigureCmd)
}
