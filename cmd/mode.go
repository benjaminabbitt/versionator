package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/mode"

	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Manage versioning mode (release or continuous-delivery)",
	Long: `Manage versioning mode configuration.

Versioning modes control how pre-release and metadata are generated:

  release (default):
    - Pre-release and metadata come from VERSION file
    - Used for standard release workflows
    - Developer controls version components

  continuous-delivery:
    - Pre-release and metadata are auto-generated from templates
    - Every build gets a unique version (e.g., 1.2.3-build-42+abc1234)
    - Templates use Mustache syntax with VCS variables

Examples:
  versionator mode                           # Show current mode
  versionator mode release                   # Set to release mode
  versionator mode cd                        # Set to continuous-delivery mode
  versionator mode cd --prerelease "build-{{CommitsSinceTag}}"
  versionator mode cd --metadata "{{ShortHash}}"`,
	RunE: runMode,
}

func init() {
	rootCmd.AddCommand(modeCmd)

	modeCmd.Flags().String("prerelease", "", "Pre-release template for CD mode (Mustache)")
	modeCmd.Flags().String("metadata", "", "Metadata template for CD mode (Mustache)")
}

func runMode(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// No args - show current mode
	if len(args) == 0 {
		currentMode := mode.GetMode(cfg)
		cmd.Printf("Current mode: %s\n", currentMode.Name())

		if !currentMode.IsReleaseMode() {
			// Show CD mode templates
			cdMode, ok := currentMode.(*mode.ContinuousDeliveryMode)
			if ok {
				prereleaseTemplate := cdMode.PrereleaseTemplate
				if prereleaseTemplate == "" {
					prereleaseTemplate = mode.DefaultCDPrereleaseTemplate
				}
				metadataTemplate := cdMode.MetadataTemplate
				if metadataTemplate == "" {
					metadataTemplate = mode.DefaultCDMetadataTemplate
				}
				cmd.Printf("  Pre-release template: %s\n", prereleaseTemplate)
				cmd.Printf("  Metadata template: %s\n", metadataTemplate)
			}
		}
		return nil
	}

	// Parse mode argument
	modeArg := args[0]
	var newModeType string

	switch modeArg {
	case "release", "rel":
		newModeType = "release"
	case "continuous-delivery", "cd":
		newModeType = "continuous-delivery"
	default:
		return fmt.Errorf("unknown mode '%s' (use 'release' or 'cd'/'continuous-delivery')", modeArg)
	}

	// Update config
	cfg.Mode.Type = newModeType

	// Set CD mode templates if provided
	if newModeType == "continuous-delivery" {
		prereleaseFlag, _ := cmd.Flags().GetString("prerelease")
		metadataFlag, _ := cmd.Flags().GetString("metadata")

		if prereleaseFlag != "" {
			// Validate template
			if err := config.ValidateTemplate(prereleaseFlag); err != nil {
				return fmt.Errorf("invalid prerelease template: %w", err)
			}
			cfg.Mode.ContinuousDelivery.PrereleaseTemplate = prereleaseFlag
		}
		if metadataFlag != "" {
			// Validate template
			if err := config.ValidateTemplate(metadataFlag); err != nil {
				return fmt.Errorf("invalid metadata template: %w", err)
			}
			cfg.Mode.ContinuousDelivery.MetadataTemplate = metadataFlag
		}
	}

	// Save config
	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("Mode set to: %s\n", newModeType)
	return nil
}
