package cmd

import (
	"fmt"
	"versionator/internal/logging"

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

		cfg, err := appInstance.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Suffix.Enabled = true
		cfg.Suffix.Type = "git"

		if err := appInstance.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		cmd.Println("Git hash suffix enabled")

		// Show current version with suffix
		version, err := appInstance.GetVersionWithSuffix()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		// Use the VCS from app instance instead of global registry
		if appInstance.VCS != nil && appInstance.VCS.IsRepository() {
			cmd.Printf("Current version: %s\n", version)
		} else {
			cmd.Printf("Current version: %s (Git hash will be added when in a repository)\n", version)
		}
	},
}

var suffixDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable git hash suffix",
	Long:  "Disable appending git hash to version numbers",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()

		cfg, err := appInstance.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Suffix.Enabled = false

		if err := appInstance.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		cmd.Println("Git hash suffix disabled")

		// Show current version without suffix
		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		cmd.Printf("Current version: %s\n", version)
	},
}

var suffixStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show suffix configuration status",
	Long:  "Show whether git hash suffix is enabled or disabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := appInstance.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		if cfg.Suffix.Enabled && cfg.Suffix.Type == "git" {
			cmd.Println("Git hash suffix: ENABLED")
			cmd.Printf("Hash length: %d characters\n", cfg.Suffix.Git.HashLength)

			// Use the VCS from app instance instead of global registry
			if appInstance.VCS != nil && appInstance.VCS.IsRepository() {
				gitHash, err := appInstance.VCS.GetVCSIdentifier(cfg.Suffix.Git.HashLength)
				if err != nil {
					cmd.Printf("Git repository detected, but cannot get hash: %v\n", err)
				} else {
					cmd.Printf("Git hash: %s\n", gitHash)
				}
			} else {
				cmd.Println("Not in a git repository - hash will be applied when in git repo")
			}
		} else {
			cmd.Println("Git hash suffix: DISABLED")
		}

		// Show current version
		version, err := appInstance.GetVersionWithSuffix()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		cmd.Printf("Current version: %s\n", version)
		return nil
	},
}

var suffixConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure git hash settings",
	Long:  "Configure git hash settings (currently uses environment variable VERSIONATOR_HASH_LENGTH)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := appInstance.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		cmd.Printf("Current configuration:\n")
		cmd.Printf("  Git hash suffix enabled: %t\n", cfg.Suffix.Enabled)
		cmd.Printf("  Suffix type: %s\n", cfg.Suffix.Type)
		cmd.Printf("  Git hash length: %d\n", cfg.Suffix.Git.HashLength)
		cmd.Printf("\nConfiguration is stored in .application.yaml\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(suffixCmd)
	suffixCmd.AddCommand(suffixEnableCmd)
	suffixCmd.AddCommand(suffixDisableCmd)
	suffixCmd.AddCommand(suffixStatusCmd)
	suffixCmd.AddCommand(suffixConfigureCmd)
}
