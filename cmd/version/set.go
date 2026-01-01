package version

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

// SetCmd sets an absolute version
var SetCmd = &cobra.Command{
	Use:   "set <version>",
	Short: "Set absolute version",
	Long: `Set an absolute version (3 or 4 components).

Examples:
  versionator version set 1.2.3        # Set 3-component version
  versionator version set 1.2.3.4      # Set 4-component version (enables .NET mode)
  versionator version set v1.2.3       # Set version with prefix (updates config)
  versionator version set release-2.0.0 # Set version with custom prefix

If a 4-component version is provided, .NET mode is automatically enabled in config.
If a prefix is present, the config prefix is updated to match.
If no prefix is provided, the config prefix is cleared.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		versionStr := args[0]

		// Parse the version
		v := version.Parse(versionStr)

		// Validate parsed version - must have at least Major component parsed
		if v.Raw != versionStr || (v.Major == 0 && v.Minor == 0 && v.Patch == 0 && v.Revision == 0 && v.Prefix == "" && v.PreRelease == "" && v.BuildMetadata == "") {
			// Check if parsing produced meaningful results
			reparsed := version.Parse(versionStr)
			if reparsed.String() == "0.0.0" && !hasDigit(versionStr) {
				return fmt.Errorf("invalid version format: %s", versionStr)
			}
		}

		// Read current config
		cfg, err := config.ReadConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		configChanged := false

		// Update prefix in config based on parsed version
		if v.Prefix != cfg.Prefix {
			cfg.Prefix = v.Prefix
			configChanged = true
		}

		// Enable/disable DotNet mode based on 4-component version
		if v.FourComponent {
			if !cfg.DotNet {
				cfg.DotNet = true
				configChanged = true
			}
		} else {
			// Disable DotNet mode for 3-component versions
			if cfg.DotNet {
				cfg.DotNet = false
				configChanged = true
			}
		}

		// Save config if changed
		if configChanged {
			if err := config.WriteConfig(cfg); err != nil {
				return fmt.Errorf("error writing config: %w", err)
			}
		}

		// Save version
		if err := version.Save(&v); err != nil {
			return fmt.Errorf("error saving version: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Version set to: %s\n", v.FullString())
		return nil
	},
}

// hasDigit checks if a string contains at least one digit
func hasDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

func init() {
	// SetCmd is added to versionCmd in cmd/version.go
}
