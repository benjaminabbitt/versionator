package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin management",
	Long:  `Commands for listing and inspecting loaded plugins.`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list [type]",
	Short: "List loaded plugins",
	Long: `List loaded plugins, optionally filtered by type.

Types:
  emit   - Plugins that generate version source files
  build  - Plugins that generate linker/build flags
  patch  - Plugins that update manifest/config files
  all    - All plugins (default)`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return []string{"emit", "build", "patch", "all"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		filterType := "all"
		if len(args) > 0 {
			filterType = args[0]
		}

		if pluginLoader == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "No plugins loaded")
			return
		}

		if filterType == "all" || filterType == "emit" {
			if len(pluginLoader.EmitPlugins) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Emit plugins:")
				names := make([]string, 0, len(pluginLoader.EmitPlugins))
				nameToFormat := make(map[string]string)
				for format, p := range pluginLoader.EmitPlugins {
					name := p.Name()
					names = append(names, name)
					nameToFormat[name] = format
				}
				sort.Strings(names)
				for _, name := range names {
					format := nameToFormat[name]
					p := pluginLoader.EmitPlugins[format]
					fmt.Fprintf(cmd.OutOrStdout(), "  %-20s  format: %-10s  output: %s\n", name, format, p.DefaultOutput())
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
		}

		if filterType == "all" || filterType == "build" {
			if len(pluginLoader.BuildPlugins) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Build plugins:")
				names := make([]string, 0, len(pluginLoader.BuildPlugins))
				nameToFormat := make(map[string]string)
				for format, p := range pluginLoader.BuildPlugins {
					name := p.Name()
					names = append(names, name)
					nameToFormat[name] = format
				}
				sort.Strings(names)
				for _, name := range names {
					format := nameToFormat[name]
					fmt.Fprintf(cmd.OutOrStdout(), "  %-20s  format: %s\n", name, format)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
		}

		if filterType == "all" || filterType == "patch" {
			if len(pluginLoader.PatchPlugins) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Patch plugins:")
				names := make([]string, 0, len(pluginLoader.PatchPlugins))
				nameToPattern := make(map[string]string)
				for pattern, p := range pluginLoader.PatchPlugins {
					name := p.Name()
					names = append(names, name)
					nameToPattern[name] = pattern
				}
				sort.Strings(names)
				for _, name := range names {
					pattern := nameToPattern[name]
					p := pluginLoader.PatchPlugins[pattern]
					fmt.Fprintf(cmd.OutOrStdout(), "  %-20s  pattern: %-20s  %s\n", name, pattern, p.Description())
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
		}
	},
}

var pluginShowCmd = &cobra.Command{
	Use:   "show <format>",
	Short: "Show plugin details for a format",
	Long: `Show details for all plugins that support a given format.

This aggregates information from emit, build, and patch plugins
that match the specified format.`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 || pluginLoader == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// Return unique formats from emit plugins
		formats := make([]string, 0, len(pluginLoader.EmitPlugins))
		for format := range pluginLoader.EmitPlugins {
			formats = append(formats, format)
		}
		sort.Strings(formats)
		return formats, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		format := args[0]

		if pluginLoader == nil {
			return fmt.Errorf("no plugins loaded")
		}

		found := false

		// Check emit plugin
		if ep, ok := pluginLoader.EmitPlugins[format]; ok {
			found = true
			fmt.Fprintf(cmd.OutOrStdout(), "Format: %s\n\n", format)
			fmt.Fprintln(cmd.OutOrStdout(), "Emit Plugin:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:           %s\n", ep.Name())
			fmt.Fprintf(cmd.OutOrStdout(), "  Default Output: %s\n", ep.DefaultOutput())
			fmt.Fprintf(cmd.OutOrStdout(), "  File Extension: %s\n", ep.FileExtension())
			fmt.Fprintln(cmd.OutOrStdout())
		}

		// Check build plugin
		if bp, ok := pluginLoader.BuildPlugins[format]; ok {
			if !found {
				fmt.Fprintf(cmd.OutOrStdout(), "Format: %s\n\n", format)
				found = true
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Build Plugin:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Name: %s\n", bp.Name())
			fmt.Fprintln(cmd.OutOrStdout())
		} else if found {
			fmt.Fprintln(cmd.OutOrStdout(), "Build Plugin: none")
			fmt.Fprintln(cmd.OutOrStdout())
		}

		// Check patch plugins - need to find by format match
		// Patch plugins are keyed by file pattern, not format
		// For now, skip patch plugin matching by format

		if !found {
			return fmt.Errorf("no plugins found for format: %s", format)
		}

		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginShowCmd)
	rootCmd.AddCommand(pluginCmd)
}
