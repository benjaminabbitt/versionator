package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	versioncmd "github.com/benjaminabbitt/versionator/cmd/version"
	"github.com/benjaminabbitt/versionator/internal/filesystem"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/pkg/plugin"
	"github.com/spf13/cobra"
)

const configFileName = ".versionator.yaml"

var initGoMode bool
var initFormat string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize versionator in the current directory",
	Long: `Initialize versionator by creating VERSION and .versionator.yaml files.

This command creates:
  - VERSION file with initial version 0.0.0 (or v0.0.0 if prefix configured)
  - .versionator.yaml configuration file

Flags:
  --go           Use Go pseudo-version prerelease pattern
  --format, -f   Set emit format (e.g., go, python, rust)

Examples:
  versionator init                # Default configuration
  versionator init --go           # Go versioning pattern (v prefix, pseudo-version prerelease)
  versionator init -f python      # Set emit format to python
  versionator init -f rust --go   # Rust format with Go versioning pattern

Existing files are not overwritten.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var createdVersion, createdConfig bool

		// Validate format if specified
		if initFormat != "" && pluginLoader != nil {
			if _, ok := pluginLoader.EmitPlugins[initFormat]; !ok {
				formats := make([]string, 0, len(pluginLoader.EmitPlugins))
				for f := range pluginLoader.EmitPlugins {
					formats = append(formats, f)
				}
				sort.Strings(formats)
				return fmt.Errorf("unsupported format: %s\nAvailable formats: %s",
					initFormat, strings.Join(formats, ", "))
			}
		}

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
			configContent := buildConfigYAML(initFormat, initGoMode)
			msg := describeConfig(initFormat, initGoMode)
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s %s\n", configFileName, msg)

			if err := filesystem.AppFs.WriteFile(getAbsPath(configFileName), []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			createdConfig = true
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s exists\n", configFileName)
		}

		// Render VERSION with config elements if we created config or version
		if createdVersion || createdConfig {
			if err := versioncmd.RenderFromConfig(); err != nil {
				return fmt.Errorf("failed to render version from config: %w", err)
			}

			// Show the rendered version
			v, err := version.Load()
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "VERSION rendered: %s\n", v.FullString())
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\nInitialization complete!")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "\nAlready initialized.")
		}

		return nil
	},
}

// buildConfigYAML generates the configuration YAML based on format and versioning pattern
func buildConfigYAML(format string, goVersioning bool) string {
	// Get versioning config (go or standard)
	var versioningCfg *plugin.VersioningConfig
	if goVersioning {
		if vp, ok := plugin.GetVersioningPlugin("go"); ok {
			versioningCfg = vp.GetVersioningConfig()
		}
	} else {
		if vp, ok := plugin.GetVersioningPlugin("standard"); ok {
			versioningCfg = vp.GetVersioningConfig()
		}
	}

	// Fall back to defaults if plugins not found
	if versioningCfg == nil {
		versioningCfg = &plugin.VersioningConfig{
			Name:               "standard",
			Prefix:             "",
			PreReleaseElements: []string{"alpha", "CommitsSinceTag"},
			MetadataElements:   []string{"BuildDateTimeCompact", "ShortHash", "Dirty"},
		}
	}

	return generateConfigYAML(format, versioningCfg)
}

// escapeYAMLString escapes quotes in a string for YAML double-quoted output
func escapeYAMLString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// formatElementsList formats a slice of strings as a YAML list
func formatElementsList(elements []string) string {
	if len(elements) == 0 {
		return "[]"
	}
	var parts []string
	for _, e := range elements {
		parts = append(parts, fmt.Sprintf("\"%s\"", e))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// generateConfigYAML creates the YAML configuration
func generateConfigYAML(format string, verCfg *plugin.VersioningConfig) string {
	prefix := verCfg.Prefix
	if prefix == "" {
		prefix = `""`
	} else {
		prefix = `"` + prefix + `"`
	}

	var sb strings.Builder
	sb.WriteString("# Versionator Configuration\n")
	sb.WriteString("# See https://github.com/benjaminabbitt/versionator for documentation\n\n")

	sb.WriteString("# Version prefix\n")
	sb.WriteString(fmt.Sprintf("prefix: %s\n\n", prefix))

	sb.WriteString("# Pre-release configuration (elements joined with dashes)\n")
	sb.WriteString("prerelease:\n")
	sb.WriteString(fmt.Sprintf("  elements: %s\n\n", formatElementsList(verCfg.PreReleaseElements)))

	sb.WriteString("# Build metadata configuration (elements joined with dots)\n")
	sb.WriteString("metadata:\n")
	sb.WriteString(fmt.Sprintf("  elements: %s\n", formatElementsList(verCfg.MetadataElements)))
	sb.WriteString("  git:\n")
	sb.WriteString("    hashLength: 12\n\n")

	// Add emit configuration if format specified
	if format != "" {
		sb.WriteString("# Emit configuration - generate version source file\n")
		sb.WriteString("emit:\n")
		sb.WriteString(fmt.Sprintf("  format: \"%s\"\n", format))

		// Get default output from emit plugin if available
		if pluginLoader != nil {
			if ep, ok := pluginLoader.EmitPlugins[format]; ok {
				sb.WriteString(fmt.Sprintf("  outputPath: \"%s\"\n", ep.DefaultOutput()))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("# Logging configuration\n")
	sb.WriteString("logging:\n")
	sb.WriteString("  output: \"console\"\n")

	return sb.String()
}

// describeConfig returns a description of the configuration being created
func describeConfig(format string, goVersioning bool) string {
	if format != "" && goVersioning {
		return fmt.Sprintf("for %s with Go versioning pattern", format)
	}
	if format != "" {
		return fmt.Sprintf("for %s", format)
	}
	if goVersioning {
		return "with Go versioning pattern"
	}
	return "with default configuration"
}

// getAbsPath returns the absolute path for a file relative to the current directory
func getAbsPath(filename string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return filename
	}
	return filepath.Join(cwd, filename)
}

func fileExists(path string) bool {
	absPath := getAbsPath(path)
	_, err := filesystem.AppFs.Stat(absPath)
	return err == nil
}

func init() {
	initCmd.Flags().BoolVar(&initGoMode, "go", false, "Use Go versioning pattern (v prefix, pseudo-version prerelease)")
	initCmd.Flags().StringVarP(&initFormat, "format", "f", "", "Emit format (e.g., go, python, rust)")
	rootCmd.AddCommand(initCmd)
}
