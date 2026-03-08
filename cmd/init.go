package cmd

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var (
	initVersion    string
	initPrefix     string
	initWithConfig bool
	initForce      bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize versionator in this directory",
	Long: `Initialize versionator by creating a VERSION file.

Creates a VERSION file with the specified initial version and prefix.
Optionally creates a .versionator.yaml configuration file.

Examples:
  versionator init                        # Create VERSION with 0.0.0
  versionator init --version 1.0.0        # Create VERSION with 1.0.0
  versionator init --prefix v             # Create VERSION with v0.0.0
  versionator init --config               # Also create .versionator.yaml
  versionator init --force                # Overwrite existing files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		versionPath := "VERSION"
		configPath := ".versionator.yaml"

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

func init() {
	initCmd.Flags().StringVarP(&initVersion, "version", "v", "0.0.0", "Initial version")
	initCmd.Flags().StringVarP(&initPrefix, "prefix", "p", "", "Version prefix (e.g., 'v')")
	initCmd.Flags().BoolVar(&initWithConfig, "config", false, "Also create .versionator.yaml")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing files")
	rootCmd.AddCommand(initCmd)
}
