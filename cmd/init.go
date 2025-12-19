package cmd

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

const configFileName = ".versionator.yaml"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize versionator in the current directory",
	Long: `Initialize versionator by creating VERSION and .versionator.yaml files.

This command creates:
  - VERSION file with initial version 0.0.0 (or v0.0.0 if prefix configured)
  - .versionator.yaml configuration file with documented defaults

Existing files are not overwritten.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var createdVersion, createdConfig bool

		// Check if VERSION file exists
		versionExists := fileExists("VERSION")
		if !versionExists {
			// Load will create the VERSION file if it doesn't exist
			v, err := version.Load()
			if err != nil {
				return fmt.Errorf("failed to create VERSION file: %w", err)
			}
			createdVersion = true
			fmt.Fprintf(cmd.OutOrStdout(), "Created VERSION file: %s\n", v.FullString())
		} else {
			v, err := version.Load()
			if err != nil {
				return fmt.Errorf("failed to read VERSION file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "VERSION file exists: %s\n", v.FullString())
		}

		// Check if config file exists
		configExists := fileExists(configFileName)
		if !configExists {
			// Write default config
			if err := os.WriteFile(configFileName, []byte(config.DefaultConfigYAML()), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			createdConfig = true
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s with default configuration\n", configFileName)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s exists\n", configFileName)
		}

		if createdVersion || createdConfig {
			fmt.Fprintln(cmd.OutOrStdout(), "\nInitialization complete!")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "\nAlready initialized.")
		}

		return nil
	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
