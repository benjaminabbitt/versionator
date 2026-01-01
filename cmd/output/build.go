package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
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

var BuildCmd = &cobra.Command{
	Use:   "build <plugin>",
	Short: "Generate build flags for version injection",
	Long: `Generate build flags to inject version information at build time.

This allows injecting version information without generating source files.
The output can be used directly with your build tool.

Use "versionator plugin list build" to see available build plugins.

Go Example:
  # Get ldflags for Go
  versionator out build build-go --var main.Version
  # Output: -X main.Version=1.2.3

  # Use in build command
  go build -ldflags "$(versionator out build build-go --var main.Version)" -o app .

Rust Example:
  # Get cfg flags for Rust
  versionator out build build-rust --var VERSION
  # Output: VERSION=1.2.3

C/C++ Example:
  # Get defines for C/C++
  versionator out build build-c --var APP_VERSION
  # Output: -DAPP_VERSION="1.2.3"`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 || PluginLoader == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, 0, len(PluginLoader.BuildPlugins))
		for _, p := range PluginLoader.BuildPlugins {
			names = append(names, p.Name())
		}
		sort.Strings(names)
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	pluginArg := args[0]

	if PluginLoader == nil {
		return fmt.Errorf("no plugins loaded")
	}

	// Find build plugin by name or format
	buildPlugin := findBuildPlugin(pluginArg)
	if buildPlugin == nil {
		names := getBuildPluginNames()
		return fmt.Errorf("build plugin '%s' not found\nAvailable plugins: %s", pluginArg, strings.Join(names, ", "))
	}

	// Load version data
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Build template data
	templateData, err := BuildTemplateData(cmd, vd, buildPrefixOverride, buildPrereleaseTemplate, buildMetadataTemplate)
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

	// If custom --var flags provided, generate flags ourselves
	if len(buildVars) > 0 {
		output := generateCustomVarFlags(buildPlugin.Format(), buildVars, versionStr, templateData)
		fmt.Println(strings.TrimSpace(output))
		return nil
	}

	// Build vars map for plugin (default behavior - no custom vars)
	vars := emit.TemplateDataToMap(templateData)
	vars["Version"] = versionStr

	// Generate flags using plugin
	output, err := buildPlugin.GenerateFlags(vars)
	if err != nil {
		return fmt.Errorf("error generating flags: %w", err)
	}

	fmt.Println(strings.TrimSpace(output))
	return nil
}

// generateCustomVarFlags generates build flags for custom --var entries
func generateCustomVarFlags(format string, vars []string, versionStr string, templateData emit.TemplateData) string {
	var flags []string

	for _, v := range vars {
		var varName, value string

		if strings.Contains(v, "=") {
			parts := strings.SplitN(v, "=", 2)
			varName = parts[0]
			template := parts[1]

			// Render template
			rendered, err := emit.RenderTemplateWithData(template, templateData)
			if err != nil {
				value = template
			} else {
				value = strings.TrimSpace(rendered)
			}
		} else {
			varName = v
			value = versionStr
		}

		// Format based on plugin format
		switch format {
		case "go":
			flags = append(flags, fmt.Sprintf("-X %s=%s", varName, value))
		case "rust":
			flags = append(flags, fmt.Sprintf("%s=%s", varName, value))
		case "c", "cpp", "csharp":
			flags = append(flags, fmt.Sprintf("-D%s=\"%s\"", varName, value))
		default:
			// Generic format
			flags = append(flags, fmt.Sprintf("%s=%s", varName, value))
		}
	}

	return strings.Join(flags, " ")
}

// buildPluginInterface defines the interface for build plugins
type buildPluginInterface interface {
	GenerateFlags(map[string]string) (string, error)
	Format() string
}

// findBuildPlugin finds a build plugin by name or format
func findBuildPlugin(nameOrFormat string) buildPluginInterface {
	// First try by format (key in the map)
	if p, ok := PluginLoader.BuildPlugins[nameOrFormat]; ok {
		return p
	}

	// Then try by plugin name
	for _, p := range PluginLoader.BuildPlugins {
		if p.Name() == nameOrFormat {
			return p
		}
	}

	return nil
}

// getBuildPluginNames returns sorted list of build plugin names
func getBuildPluginNames() []string {
	names := make([]string, 0, len(PluginLoader.BuildPlugins))
	for _, p := range PluginLoader.BuildPlugins {
		names = append(names, p.Name())
	}
	sort.Strings(names)
	return names
}

func init() {
	BuildCmd.Flags().StringArrayVar(&buildVars, "var", nil, "Variable to set (e.g., 'main.Version' or 'main.Hash={{ShortHash}}')")
	BuildCmd.Flags().StringVarP(&buildTemplate, "template", "t", "", "Custom template for entire output")

	BuildCmd.Flags().StringVarP(&buildPrefixOverride, "prefix", "p", "", "Version prefix")
	BuildCmd.Flag("prefix").NoOptDefVal = UseDefaultMarker

	BuildCmd.Flags().StringVar(&buildPrereleaseTemplate, "prerelease", "", "Pre-release template")
	BuildCmd.Flag("prerelease").NoOptDefVal = UseDefaultMarker

	BuildCmd.Flags().StringVar(&buildMetadataTemplate, "metadata", "", "Metadata template")
	BuildCmd.Flag("metadata").NoOptDefVal = UseDefaultMarker
}
