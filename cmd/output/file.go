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
	fileOutput             string
	filePrereleaseTemplate string
	fileMetadataTemplate   string
	filePrefixOverride     string
)

var FileCmd = &cobra.Command{
	Use:   "file <plugin>",
	Short: "Generate version source file",
	Long: `Generate a version source file using the specified emit plugin.

Use 'versionator plugin list emit' to see available emit plugins.

This generates a source file that can be compiled into your application,
providing version information as constants or variables.

Examples:
  # Generate Python version file
  versionator out file emit-python --output _version.py

  # Generate Go version file
  versionator out file emit-go --output version/version.go

  # Generate with prerelease
  versionator out file emit-python --prerelease="alpha" --output _version.py`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 || PluginLoader == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, 0, len(PluginLoader.EmitPlugins))
		for _, p := range PluginLoader.EmitPlugins {
			names = append(names, p.Name())
		}
		sort.Strings(names)
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runFile,
}

func runFile(cmd *cobra.Command, args []string) error {
	if PluginLoader == nil {
		return fmt.Errorf("no plugins loaded")
	}

	pluginArg := args[0]

	// Find emit plugin by name or format
	var emitPlugin = findEmitPlugin(pluginArg)
	if emitPlugin == nil {
		names := getEmitPluginNames()
		return fmt.Errorf("emit plugin '%s' not found\nAvailable plugins: %s", pluginArg, strings.Join(names, ", "))
	}

	// Load version data
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Build template data
	templateData, err := BuildTemplateData(cmd, vd, filePrefixOverride, filePrereleaseTemplate, fileMetadataTemplate)
	if err != nil {
		return err
	}

	// Convert template data to map for plugin
	vars := emit.TemplateDataToMap(templateData)

	// Generate content using plugin
	content, err := emitPlugin.Emit(vars)
	if err != nil {
		return fmt.Errorf("error generating content: %w", err)
	}

	// Output to file or stdout
	if fileOutput != "" {
		if err := emit.WriteToFile(content, fileOutput); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
		fmt.Printf("Version %s written to %s\n", vd.CoreVersion(), fileOutput)
	} else {
		fmt.Print(content)
	}
	return nil
}

// findEmitPlugin finds an emit plugin by name or format
func findEmitPlugin(nameOrFormat string) interface{ Emit(map[string]string) (string, error) } {
	// First try by format (key in the map)
	if p, ok := PluginLoader.EmitPlugins[nameOrFormat]; ok {
		return p
	}

	// Then try by plugin name
	for _, p := range PluginLoader.EmitPlugins {
		if p.Name() == nameOrFormat {
			return p
		}
	}

	return nil
}

// getEmitPluginNames returns sorted list of emit plugin names
func getEmitPluginNames() []string {
	names := make([]string, 0, len(PluginLoader.EmitPlugins))
	for _, p := range PluginLoader.EmitPlugins {
		names = append(names, p.Name())
	}
	sort.Strings(names)
	return names
}

func init() {
	FileCmd.Flags().StringVarP(&fileOutput, "output", "o", "", "Output file path (default: stdout)")

	FileCmd.Flags().StringVarP(&filePrefixOverride, "prefix", "p", "", "Version prefix (default 'v' if flag provided without value)")
	FileCmd.Flag("prefix").NoOptDefVal = UseDefaultMarker

	FileCmd.Flags().StringVar(&filePrereleaseTemplate, "prerelease", "", "Pre-release template (uses config default if flag provided without value)")
	FileCmd.Flag("prerelease").NoOptDefVal = UseDefaultMarker

	FileCmd.Flags().StringVar(&fileMetadataTemplate, "metadata", "", "Metadata template (uses config default if flag provided without value)")
	FileCmd.Flag("metadata").NoOptDefVal = UseDefaultMarker
}
