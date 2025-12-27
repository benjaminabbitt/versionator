package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/plugin"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var (
	buildVars               []string
	buildTemplate           string
	buildPrereleaseTemplate string
	buildMetadataTemplate   string
	buildPrefixOverride     string
)

var emitBuildCmd = &cobra.Command{
	Use: "build <language>",
	Short:   "Generate build flags for version injection",
	Long: `Generate build flags to inject version information at build time.

This allows injecting version information without generating source files.
The output can be used directly with your build tool.

Languages that support link-time injection are determined by registered plugins.
Use "versionator emit build --help" to see available languages.

Go Example:
  # Get ldflags for Go
  versionator emit build go --var main.Version
  # Output: -X main.Version=1.2.3

  # Use in build command
  go build -ldflags "$(versionator emit build go --var main.Version)" -o app .

  # Multiple variables
  versionator emit build go --var main.Version --var main.GitHash={{ShortHash}}

Rust Example:
  # Get cfg flags for Rust
  versionator emit build rust --var VERSION
  # Output: --cfg VERSION="1.2.3"

  # Use in build command
  RUSTFLAGS="$(versionator emit build rust)" cargo build

C/C++ Example:
  # Get defines for C/C++
  versionator emit build c --var APP_VERSION
  # Output: -DAPP_VERSION="1.2.3"

  # Use in build command
  gcc $(versionator emit build c --var APP_VERSION) -o app main.c`,
	Args: cobra.ExactArgs(1),
	RunE: runEmitBuild,
}

func runEmitBuild(cmd *cobra.Command, args []string) error {
	lang := strings.ToLower(args[0])

	// Get language plugin
	langPlugin, ok := plugin.GetLanguagePlugin(lang)
	if !ok {
		return fmt.Errorf("language '%s' is not supported\n%s", lang, getSupportedBuildLanguagesMessage())
	}

	// Check if language supports link-time injection
	linkConfig := langPlugin.GetBuildConfig()
	if linkConfig == nil {
		return fmt.Errorf("language '%s' does not support link-time version injection\n%s", lang, getSupportedBuildLanguagesMessage())
	}

	// Load version data
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Build template data
	templateData, err := buildTemplateData(cmd, vd, buildPrefixOverride, buildPrereleaseTemplate, buildMetadataTemplate)
	if err != nil {
		return err
	}

	// Build version string
	versionStr := templateData.MajorMinorPatch
	if templateData.PreReleaseWithDash != "" {
		versionStr += templateData.PreReleaseWithDash
	}
	if templateData.MetadataWithPlus != "" {
		versionStr += templateData.MetadataWithPlus
	}

	// If custom template provided, use it
	if buildTemplate != "" {
		result, err := emit.RenderTemplateWithData(buildTemplate, templateData)
		if err != nil {
			return fmt.Errorf("error rendering template: %w", err)
		}
		fmt.Print(strings.TrimSpace(result))
		return nil
	}

	// Generate output using plugin's link config
	output := generateFlags(linkConfig, templateData, versionStr)
	fmt.Println(strings.Join(output, " "))
	return nil
}

// getSupportedBuildLanguagesMessage returns a formatted message listing supported languages
func getSupportedBuildLanguagesMessage() string {
	var supported []string
	for name, lp := range plugin.GetLanguagePlugins() {
		if lp.GetBuildConfig() != nil {
			supported = append(supported, name)
		}
	}
	sort.Strings(supported)
	return fmt.Sprintf("Supported languages: %s", strings.Join(supported, ", "))
}

// generateFlags creates build flags using the plugin's link configuration
func generateFlags(linkConfig *plugin.LinkConfig, data emit.TemplateData, versionStr string) []string {
	var flags []string

	if len(buildVars) == 0 {
		// Default: use the plugin's default variable path
		flag := renderFlagTemplate(linkConfig.FlagTemplate, linkConfig.VariablePath, versionStr)
		flags = append(flags, flag)
	} else {
		for _, v := range buildVars {
			if strings.Contains(v, "=") {
				// Parse var=template
				parts := strings.SplitN(v, "=", 2)
				varName := parts[0]
				template := parts[1]

				// Render template
				value, err := emit.RenderTemplateWithData(template, data)
				if err != nil {
					value = template // Use as-is if template fails
				}
				flag := renderFlagTemplate(linkConfig.FlagTemplate, varName, strings.TrimSpace(value))
				flags = append(flags, flag)
			} else {
				// Just variable name, use version string
				flag := renderFlagTemplate(linkConfig.FlagTemplate, v, versionStr)
				flags = append(flags, flag)
			}
		}
	}

	return flags
}

// renderFlagTemplate replaces {{Variable}} and {{Value}} placeholders in the template
func renderFlagTemplate(template, variable, value string) string {
	result := strings.ReplaceAll(template, "{{Variable}}", variable)
	result = strings.ReplaceAll(result, "{{Value}}", value)
	return result
}

func init() {
	emitCmd.AddCommand(emitBuildCmd)

	emitBuildCmd.Flags().StringArrayVar(&buildVars, "var", nil, "Variable to set (e.g., 'main.Version' or 'main.Hash={{ShortHash}}')")
	emitBuildCmd.Flags().StringVarP(&buildTemplate, "template", "t", "", "Custom template for entire output")

	emitBuildCmd.Flags().StringVarP(&buildPrefixOverride, "prefix", "p", "", "Version prefix")
	emitBuildCmd.Flag("prefix").NoOptDefVal = useDefaultMarker

	emitBuildCmd.Flags().StringVar(&buildPrereleaseTemplate, "prerelease", "", "Pre-release template")
	emitBuildCmd.Flag("prerelease").NoOptDefVal = useDefaultMarker

	emitBuildCmd.Flags().StringVar(&buildMetadataTemplate, "metadata", "", "Metadata template")
	emitBuildCmd.Flag("metadata").NoOptDefVal = useDefaultMarker
}
