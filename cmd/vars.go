package cmd

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/pkg/plugin"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"

	"github.com/spf13/cobra"
)

var varsCmd = &cobra.Command{
	Use:   "vars",
	Short: "Show all template variables and their current values",
	Long: `Display all available template variables and their current values.

This is useful for understanding what variables are available when
creating custom templates for version, prerelease, or metadata output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error loading version data: %w", err)
		}

		// Build template data
		templateData := emit.BuildTemplateDataFromVersion(vd)

		// Populate PreRelease and Metadata from config
		prerelease, _ := versionator.RenderPreRelease()
		templateData.PreRelease = prerelease
		if prerelease != "" {
			templateData.PreReleaseWithDash = "-" + prerelease
		}

		metadata, _ := versionator.RenderMetadata()
		templateData.Metadata = metadata
		if metadata != "" {
			templateData.MetadataWithPlus = "+" + metadata
		}

		// Use reflection to iterate over all fields
		v := reflect.ValueOf(templateData)
		t := v.Type()

		fmt.Println("Template Variables")
		fmt.Println(strings.Repeat("=", 60))

		// Group variables by category
		categories := map[string][]string{
			"Version Components": {
				"Major", "Minor", "Patch", "MajorMinorPatch", "MajorMinor", "Prefix",
			},
			"Pre-release (template-based)": {
				"PreRelease", "PreReleaseWithDash", "PreReleaseLabel", "PreReleaseNumber",
			},
			"Metadata (template-based)": {
				"Metadata", "MetadataWithPlus",
			},
			"VCS/Git": {
				"Hash", "ShortHash", "MediumHash",
				"ShortHashWithDot", "MediumHashWithDot",
				"ShortHashWithDash", "MediumHashWithDash",
				"BranchName", "EscapedBranchName",
				"CommitsSinceTag", "BuildNumber", "BuildNumberPadded",
				"UncommittedChanges", "Dirty", "DirtyWithDot", "DirtyWithDash",
				"VersionSourceHash",
			},
			"Commit Author": {
				"CommitAuthor", "CommitAuthorEmail",
			},
			"Commit Timestamps": {
				"CommitDate", "CommitDateCompact", "CommitDateShort",
				"CommitYear", "CommitMonth", "CommitDay",
			},
			"Build Timestamps": {
				"BuildDateTimeUTC", "BuildDateTimeCompact",
				"BuildDateTimeCompactWithDot", "BuildDateTimeCompactWithDash",
				"BuildDateUTC", "BuildYear", "BuildMonth", "BuildDay",
			},
		}

		categoryOrder := []string{
			"Version Components",
			"Pre-release (template-based)",
			"Metadata (template-based)",
			"VCS/Git",
			"Commit Author",
			"Commit Timestamps",
			"Build Timestamps",
		}

		for _, category := range categoryOrder {
			fields := categories[category]
			fmt.Printf("\n%s\n", category)
			fmt.Println(strings.Repeat("-", len(category)))

			for _, fieldName := range fields {
				field, found := t.FieldByName(fieldName)
				if !found {
					continue
				}
				value := v.FieldByName(fieldName)

				var valueStr string
				switch value.Kind() {
				case reflect.String:
					valueStr = value.String()
					if valueStr == "" {
						valueStr = "(empty)"
					}
				case reflect.Int, reflect.Int64:
					valueStr = fmt.Sprintf("%d", value.Int())
				default:
					valueStr = fmt.Sprintf("%v", value.Interface())
				}

				// Truncate long values
				if len(valueStr) > 40 {
					valueStr = valueStr[:37] + "..."
				}

				fmt.Printf("  {{%s}}%s = %s\n",
					field.Name,
					strings.Repeat(" ", 30-len(field.Name)),
					valueStr)
			}
		}

		// Display custom variables from config file
		if len(templateData.Custom) > 0 {
			fmt.Printf("\nCustom Variables (from .versionator.yaml)\n")
			fmt.Println(strings.Repeat("-", 36))

			// Sort keys for consistent output
			keys := make([]string, 0, len(templateData.Custom))
			for k := range templateData.Custom {
				keys = append(keys, k)
			}
			sortStrings(keys)

			for _, k := range keys {
				valueStr := templateData.Custom[k]
				if len(valueStr) > 40 {
					valueStr = valueStr[:37] + "..."
				}
				padding := 30 - len(k)
				if padding < 1 {
					padding = 1
				}
				fmt.Printf("  {{%s}}%s = %s\n", k, strings.Repeat(" ", padding), valueStr)
			}
		}

		// Display plugin-provided variables
		pluginVars := plugin.GetAllTemplateVariables(map[string]string{
			"ShortHash":  templateData.ShortHash,
			"MediumHash": templateData.MediumHash,
			"Hash":       templateData.Hash,
		})
		if len(pluginVars) > 0 {
			fmt.Printf("\nPlugin Variables\n")
			fmt.Println(strings.Repeat("-", 16))

			// Sort keys for consistent output
			keys := make([]string, 0, len(pluginVars))
			for k := range pluginVars {
				keys = append(keys, k)
			}
			sortStrings(keys)

			for _, k := range keys {
				valueStr := pluginVars[k]
				if len(valueStr) > 40 {
					valueStr = valueStr[:37] + "..."
				}
				padding := 30 - len(k)
				if padding < 1 {
					padding = 1
				}
				fmt.Printf("  {{%s}}%s = %s\n", k, strings.Repeat(" ", padding), valueStr)
			}
		}

		fmt.Println()
		return nil
	},
}

// sortStrings sorts a slice of strings in place using standard library
func sortStrings(s []string) {
	sort.Strings(s)
}

func init() {
	rootCmd.AddCommand(varsCmd)
}
