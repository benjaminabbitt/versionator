package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

// validPrefix checks if a prefix is valid (empty, "v", or "V")
func validPrefix(p string) bool {
	return p == "" || p == "v" || p == "V"
}

var prefixCmd = &cobra.Command{
	Use:   "prefix",
	Short: "Manage version prefix",
	Long: `Commands to enable, disable, or set version prefix in VERSION file.

Only 'v' or 'V' prefixes are allowed per SemVer convention.`,
}

var prefixEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable version prefix",
	Long:  "Enable version prefix using config value if set (must be 'v' or 'V'), otherwise 'v'",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use config prefix if set and valid, otherwise default to "v"
		prefix := "v"
		if cfg, err := config.ReadConfig(); err == nil && cfg.Prefix != "" {
			if !validPrefix(cfg.Prefix) {
				return fmt.Errorf("invalid config prefix %q: only 'v' or 'V' allowed per SemVer convention", cfg.Prefix)
			}
			prefix = cfg.Prefix
		}

		if err := version.SetPrefix(prefix); err != nil {
			return fmt.Errorf("error setting prefix: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Version prefix enabled with value '%s'\n", prefix)

		// Show current version with prefix
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prefixDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable version prefix",
	Long:  "Disable version prefix by setting it to empty string",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := version.SetPrefix(""); err != nil {
			return fmt.Errorf("error setting prefix: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Version prefix disabled")

		// Show current version without prefix
		v, err := version.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", v)
		return nil
	},
}

var prefixSetCmd = &cobra.Command{
	Use:   "set <prefix>",
	Short: "Set version prefix (v or V only)",
	Long: `Set version prefix in both config and VERSION file.

Only 'v' or 'V' prefixes are allowed per SemVer convention.
Use 'prefix disable' to remove the prefix.

This updates:
1. The config file (.versionator.yaml) - so 'prefix enable' can restore it
2. The VERSION file - the source of truth for the current version`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := args[0]

		// Validate prefix
		if !validPrefix(prefix) {
			return fmt.Errorf("invalid prefix %q: only 'v' or 'V' allowed per SemVer convention", prefix)
		}

		// Update config with new prefix
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		cfg.Prefix = prefix
		if err := config.WriteConfig(cfg); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		// Update VERSION file
		if err := version.SetPrefix(prefix); err != nil {
			return fmt.Errorf("error setting prefix: %w", err)
		}

		if prefix == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Version prefix disabled (set to empty)")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Version prefix set to: %s\n", prefix)
		}

		// Show current version with new prefix
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

var prefixStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show prefix status",
	Long: `Show current prefix status from VERSION file (source of truth).

Also shows the configured prefix from .versionator.yaml that will be used on 'prefix enable'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load version - VERSION file is source of truth
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error reading version: %w", err)
		}

		if vd.Prefix == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Prefix: DISABLED")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Prefix: ENABLED")
			fmt.Fprintf(cmd.OutOrStdout(), "Value: %s\n", vd.Prefix)
		}

		// Show config prefix (what will be restored on enable)
		configPrefix := "v" // default
		if cfg, err := config.ReadConfig(); err == nil && cfg.Prefix != "" {
			configPrefix = cfg.Prefix
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config prefix (for enable): %s\n", configPrefix)

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
		return nil
	},
}

func init() {
	configCmd.AddCommand(prefixCmd)
	prefixCmd.AddCommand(prefixEnableCmd)
	prefixCmd.AddCommand(prefixDisableCmd)
	prefixCmd.AddCommand(prefixSetCmd)
	prefixCmd.AddCommand(prefixStatusCmd)
}
