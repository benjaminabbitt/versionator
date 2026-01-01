package cmd

import (
	"fmt"
	"os"
	"strings"

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

// --- Prefix subcommand ---

var configPrefixCmd = &cobra.Command{
	Use:   "prefix",
	Short: "Manage version prefix configuration",
	Long:  `Get or set the version prefix in configuration.`,
}

var configPrefixGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current prefix value",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		if cfg.Prefix == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "(empty)")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), cfg.Prefix)
		}
		return nil
	},
}

var configPrefixSetCmd = &cobra.Command{
	Use:   "set <value>",
	Short: "Set prefix value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.Prefix = args[0]
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Prefix set to: %s\n", args[0])
		return nil
	},
}

// --- Prerelease subcommand ---

var configPrereleaseCmd = &cobra.Command{
	Use:   "prerelease",
	Short: "Manage prerelease configuration",
	Long:  `Get, set, or clear the prerelease elements in configuration.`,
}

var configPrereleaseGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current prerelease elements",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		if len(cfg.PreRelease.Elements) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "(none)")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(cfg.PreRelease.Elements, ", "))
		}
		return nil
	},
}

var configPrereleaseSetCmd = &cobra.Command{
	Use:   "set <element1> [element2] ...",
	Short: "Set prerelease elements",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.PreRelease.Elements = args
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Prerelease elements set to: %s\n", strings.Join(args, ", "))
		return nil
	},
}

var configPrereleaseClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear prerelease elements",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.PreRelease.Elements = nil
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Prerelease elements cleared")
		return nil
	},
}

// --- Metadata subcommand ---

var configMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Manage metadata configuration",
	Long:  `Get, set, or clear the metadata elements in configuration.`,
}

var configMetadataGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current metadata elements",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		if len(cfg.Metadata.Elements) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "(none)")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(cfg.Metadata.Elements, ", "))
		}
		return nil
	},
}

var configMetadataSetCmd = &cobra.Command{
	Use:   "set <element1> [element2] ...",
	Short: "Set metadata elements",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.Metadata.Elements = args
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Metadata elements set to: %s\n", strings.Join(args, ", "))
		return nil
	},
}

var configMetadataClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear metadata elements",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.Metadata.Elements = nil
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Metadata elements cleared")
		return nil
	},
}

// --- DotNet subcommand ---

var configDotNetCmd = &cobra.Command{
	Use:   "dotnet",
	Short: "Manage .NET 4-component version mode",
	Long:  `Enable, disable, or check status of .NET 4-component version mode.`,
}

var configDotNetStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current .NET mode status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		if cfg.DotNet {
			fmt.Fprintln(cmd.OutOrStdout(), "DotNet mode: ENABLED")
			fmt.Fprintln(cmd.OutOrStdout(), "Versions will use 4-component format: Major.Minor.Patch.Revision")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "DotNet mode: DISABLED")
			fmt.Fprintln(cmd.OutOrStdout(), "Versions use standard 3-component format: Major.Minor.Patch")
		}
		return nil
	},
}

var configDotNetEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable .NET 4-component version mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.DotNet = true
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "DotNet mode enabled")
		return nil
	},
}

var configDotNetDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable .NET 4-component version mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.DotNet = false
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "DotNet mode disabled")
		return nil
	},
}

func init() {
	// Dump subcommand
	configDumpCmd.Flags().StringVarP(&configDumpOutput, "output", "o", "", "Output file path (default: stdout)")
	configCmd.AddCommand(configDumpCmd)

	// Prefix subcommands
	configPrefixCmd.AddCommand(configPrefixGetCmd)
	configPrefixCmd.AddCommand(configPrefixSetCmd)
	configCmd.AddCommand(configPrefixCmd)

	// Prerelease subcommands
	configPrereleaseCmd.AddCommand(configPrereleaseGetCmd)
	configPrereleaseCmd.AddCommand(configPrereleaseSetCmd)
	configPrereleaseCmd.AddCommand(configPrereleaseClearCmd)
	configCmd.AddCommand(configPrereleaseCmd)

	// Metadata subcommands
	configMetadataCmd.AddCommand(configMetadataGetCmd)
	configMetadataCmd.AddCommand(configMetadataSetCmd)
	configMetadataCmd.AddCommand(configMetadataClearCmd)
	configCmd.AddCommand(configMetadataCmd)

	// DotNet subcommands
	configDotNetCmd.AddCommand(configDotNetStatusCmd)
	configDotNetCmd.AddCommand(configDotNetEnableCmd)
	configDotNetCmd.AddCommand(configDotNetDisableCmd)
	configCmd.AddCommand(configDotNetCmd)

	rootCmd.AddCommand(configCmd)
}
