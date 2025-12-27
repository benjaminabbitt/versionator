package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/plugin"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

const configFileName = ".versionator.yaml"

var initGoMode bool

var initCmd = &cobra.Command{
	Use:   "init [language]",
	Short: "Initialize versionator in the current directory",
	Long: `Initialize versionator by creating VERSION and .versionator.yaml files.

This command creates:
  - VERSION file with initial version 0.0.0 (or v0.0.0 if prefix configured)
  - .versionator.yaml configuration file with language-specific defaults

Arguments:
  language  Target language for version file emission (optional)
            Use 'versionator init --help' to see supported languages

Flags:
  --go    Use Go pseudo-version prerelease pattern
          ({{CommitsSinceTag}}.{{BuildDateTimeCompact}}.{{ShortHash}})
          Can be combined with any language for Go-compatible versioning

Examples:
  versionator init              # Default configuration
  versionator init go           # Go project with Go-specific settings
  versionator init python       # Python project settings
  versionator init rust --go    # Rust project with Go pseudo-version pattern

Existing files are not overwritten.`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		langs := plugin.GetSupportedLanguages()
		sort.Strings(langs)
		return langs, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var createdVersion, createdConfig bool
		var language string

		if len(args) > 0 {
			language = strings.ToLower(args[0])
			if !plugin.IsLanguageSupported(language) {
				langs := plugin.GetSupportedLanguages()
				sort.Strings(langs)
				return fmt.Errorf("unsupported language: %s\nSupported languages: %s",
					language, strings.Join(langs, ", "))
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
			configContent := buildConfigYAML(language, initGoMode)
			msg := describeConfig(language, initGoMode)
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s %s\n", configFileName, msg)

			if err := os.WriteFile(configFileName, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			createdConfig = true
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

// buildConfigYAML generates the configuration YAML based on language and versioning pattern
func buildConfigYAML(language string, goVersioning bool) string {
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

	// Get language plugin if specified
	var langPlugin plugin.LanguagePlugin
	if language != "" {
		if lp, ok := plugin.GetLanguagePlugin(language); ok {
			langPlugin = lp
		}
	}

	return generateConfigYAML(langPlugin, versioningCfg)
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

// generateConfigYAML creates the YAML configuration from language plugin and versioning configs
func generateConfigYAML(langPlugin plugin.LanguagePlugin, verCfg *plugin.VersioningConfig) string {
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

	if langPlugin != nil {
		// Add emit configuration
		emitCfg := langPlugin.GetEmitConfig()
		if emitCfg != nil {
			sb.WriteString("# Emit configuration - generate version source file\n")
			sb.WriteString("emit:\n")
			sb.WriteString(fmt.Sprintf("  language: \"%s\"\n", langPlugin.LanguageName()))
			sb.WriteString(fmt.Sprintf("  outputPath: \"%s\"  # Override default output path\n", emitCfg.DefaultOutputPath))
			if emitCfg.DefaultPackageName != "" {
				sb.WriteString(fmt.Sprintf("  packageName: \"%s\"\n", emitCfg.DefaultPackageName))
			}
			sb.WriteString("\n")
		}

		// Add link configuration if supported
		linkCfg := langPlugin.GetBuildConfig()
		if linkCfg != nil {
			sb.WriteString("# Link configuration - linker flag injection\n")
			sb.WriteString("link:\n")
			sb.WriteString(fmt.Sprintf("  variablePath: \"%s\"  # Override variable to inject\n", linkCfg.VariablePath))
			sb.WriteString(fmt.Sprintf("  flagTemplate: \"%s\"\n", escapeYAMLString(linkCfg.FlagTemplate)))
			sb.WriteString("\n")
		}

		// Add patch configuration if supported
		patchConfigs := langPlugin.GetPatchConfigs()
		if len(patchConfigs) > 0 {
			sb.WriteString("# Patch configuration - update manifest/config files\n")
			sb.WriteString("patch:\n")
			sb.WriteString(fmt.Sprintf("  filePath: \"%s\"  # Override file to patch\n", patchConfigs[0].FilePath))
			sb.WriteString(fmt.Sprintf("  versionPath: \"%s\"  # Override path to version field\n", patchConfigs[0].VersionPath))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("# Logging configuration\n")
	sb.WriteString("logging:\n")
	sb.WriteString("  output: \"console\"\n")

	return sb.String()
}

// describeConfig returns a description of the configuration being created
func describeConfig(language string, goVersioning bool) string {
	if language != "" && goVersioning {
		return fmt.Sprintf("for %s with Go versioning pattern", language)
	}
	if language != "" {
		return fmt.Sprintf("for %s", language)
	}
	if goVersioning {
		return "with Go versioning pattern"
	}
	return "with default configuration"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	initCmd.Flags().BoolVar(&initGoMode, "go", false, "Configure for Go projects with prerelease versioning enabled")
	rootCmd.AddCommand(initCmd)
}
