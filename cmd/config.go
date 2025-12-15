package cmd

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/spf13/cobra"
)

var configDumpOutput string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for managing versionator configuration files.`,
}

var configDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump default configuration to stdout or file",
	Long: `Dump a default .versionator.yaml configuration file.

By default, outputs to stdout. Use --output to write to a file.

Examples:
  versionator config dump                              # Print default config to stdout
  versionator config dump --output .versionator.yaml   # Write to file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultConfig := config.DefaultConfigYAML()

		if configDumpOutput == "" {
			fmt.Print(defaultConfig)
			return nil
		}

		if err := os.WriteFile(configDumpOutput, []byte(defaultConfig), FilePermission); err != nil {
			return fmt.Errorf("error writing config file to %s: %w", configDumpOutput, err)
		}

		fmt.Printf("Config file written to %s\n", configDumpOutput)
		return nil
	},
}

func init() {
	configDumpCmd.Flags().StringVarP(&configDumpOutput, "output", "o", "", "Output file path (default: stdout)")
	configCmd.AddCommand(configDumpCmd)
	rootCmd.AddCommand(configCmd)
}
