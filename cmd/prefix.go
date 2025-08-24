package cmd

import (
	"versionator/internal/logging"

	"github.com/spf13/cobra"
)

var prefixCmd = &cobra.Command{
	Use:   "prefix",
	Short: "Manage version prefix behavior",
	Long:  "Commands to enable, disable, or set version prefix",
}

var prefixEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable version prefix",
	Long:  "Enable version prefix with default value 'v'",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()

		cfg, err := appInstance.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Prefix = "v"

		if err := appInstance.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		cmd.Println("Version prefix enabled with default value 'v'")

		// Show current version with prefix
		version, err := appInstance.GetVersionWithSuffix()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		cmd.Printf("Current version: %s\n", version)
	},
}

var prefixDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable version prefix",
	Long:  "Disable version prefix by setting it to empty string",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()

		cfg, err := appInstance.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Prefix = ""

		if err := appInstance.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		cmd.Println("Version prefix disabled")

		// Show current version without prefix
		version, err := appInstance.GetCurrentVersion()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		cmd.Printf("Current version: %s\n", version)
	},
}

var prefixSetCmd = &cobra.Command{
	Use:   "set <prefix>",
	Short: "Set version prefix",
	Long:  "Set a custom version prefix",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()
		prefix := args[0]

		cfg, err := appInstance.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		cfg.Prefix = prefix

		if err := appInstance.WriteConfig(cfg); err != nil {
			logger.Fatalw("Error writing config", "error", err)
		}

		if prefix == "" {
			cmd.Println("Version prefix disabled (set to empty)")
		} else {
			cmd.Printf("Version prefix set to: %s\n", prefix)
		}

		// Show current version with new prefix
		version, err := appInstance.GetVersionWithSuffix()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		cmd.Printf("Current version: %s\n", version)
	},
}

var prefixStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show prefix configuration status",
	Long:  "Show current version prefix configuration",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetSugaredLogger()

		cfg, err := appInstance.ReadConfig()
		if err != nil {
			logger.Fatalw("Error reading config", "error", err)
		}

		if cfg.Prefix == "" {
			cmd.Println("Version prefix: DISABLED")
		} else {
			cmd.Printf("Version prefix: ENABLED\n")
			cmd.Printf("Prefix value: %s\n", cfg.Prefix)
		}

		// Show current version
		version, err := appInstance.GetVersionWithSuffix()
		if err != nil {
			logger.Fatalw("Error getting version", "error", err)
		}

		cmd.Printf("Current version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(prefixCmd)
	prefixCmd.AddCommand(prefixEnableCmd)
	prefixCmd.AddCommand(prefixDisableCmd)
	prefixCmd.AddCommand(prefixSetCmd)
	prefixCmd.AddCommand(prefixStatusCmd)
}
