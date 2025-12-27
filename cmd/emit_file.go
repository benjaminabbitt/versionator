package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"
	"github.com/spf13/cobra"
)

var (
	fileOutput             string
	filePrereleaseTemplate string
	fileMetadataTemplate   string
	filePrefixOverride     string
)

var emitFileCmd = &cobra.Command{
	Use:   "file <format>",
	Short: "Generate version source file for a language",
	Long: `Generate a version source file for the specified programming language.

Supported formats: ` + strings.Join(emit.SupportedFormats(), ", ") + `

This generates a source file that can be compiled into your application,
providing version information as constants or variables.

Examples:
  # Generate Python version file
  versionator emit file python --output _version.py

  # Generate Go version file
  versionator emit file go --output version/version.go

  # Generate with prerelease
  versionator emit file python --prerelease="alpha" --output _version.py`,
	Args: cobra.ExactArgs(1),
	RunE: runEmitFile,
}

func runEmitFile(cmd *cobra.Command, args []string) error {
	format := emit.Format(args[0])
	if !emit.IsValidFormat(string(format)) {
		return fmt.Errorf("unsupported format '%s'\nSupported formats: %s", format, strings.Join(emit.SupportedFormats(), ", "))
	}

	// Load version data
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Build template data
	templateData, err := buildTemplateData(cmd, vd, filePrefixOverride, filePrereleaseTemplate, fileMetadataTemplate)
	if err != nil {
		return err
	}

	// Get and render template
	tmplStr, err := emit.GetEmbeddedTemplate(format)
	if err != nil {
		return fmt.Errorf("error getting template: %w", err)
	}

	content, err := emit.RenderTemplateWithData(tmplStr, templateData)
	if err != nil {
		return fmt.Errorf("error rendering format: %w", err)
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

// buildTemplateData creates template data with prerelease and metadata handling
func buildTemplateData(cmd *cobra.Command, vd *version.Version, prefixOverride, prereleaseTemplate, metadataTemplate string) (emit.TemplateData, error) {
	// Handle prefix override
	var prefix string
	if cmd.Flags().Changed("prefix") {
		if prefixOverride == useDefaultMarker {
			prefix = "v"
		} else {
			prefix = prefixOverride
		}
	} else {
		prefix = vd.Prefix
	}

	// Handle prerelease
	var prereleaseResult string
	var err error
	if cmd.Flags().Changed("prerelease") {
		if prereleaseTemplate == useDefaultMarker {
			prereleaseResult, err = versionator.RenderPreRelease()
			if err != nil {
				return emit.TemplateData{}, fmt.Errorf("error rendering prerelease: %w", err)
			}
		} else {
			baseData := emit.BuildTemplateDataFromVersion(vd)
			prereleaseResult, err = emit.RenderTemplateWithData(prereleaseTemplate, baseData)
			if err != nil {
				return emit.TemplateData{}, fmt.Errorf("error rendering prerelease template: %w", err)
			}
			prereleaseResult = strings.TrimSpace(prereleaseResult)
		}
	}

	// Handle metadata
	var metadataResult string
	if cmd.Flags().Changed("metadata") {
		if metadataTemplate == useDefaultMarker {
			metadataResult, err = versionator.RenderMetadata()
			if err != nil {
				return emit.TemplateData{}, fmt.Errorf("error rendering metadata: %w", err)
			}
		} else {
			baseData := emit.BuildTemplateDataFromVersion(vd)
			metadataResult, err = emit.RenderTemplateWithData(metadataTemplate, baseData)
			if err != nil {
				return emit.TemplateData{}, fmt.Errorf("error rendering metadata template: %w", err)
			}
			metadataResult = strings.TrimSpace(metadataResult)
		}
	}

	// Build template data
	templateData := emit.BuildTemplateDataFromVersion(vd)
	templateData.Prefix = prefix
	templateData.PreRelease = prereleaseResult
	if prereleaseResult != "" {
		templateData.PreReleaseWithDash = "-" + prereleaseResult
	}
	templateData.Metadata = metadataResult
	if metadataResult != "" {
		templateData.MetadataWithPlus = "+" + metadataResult
	}

	return templateData, nil
}

func init() {
	emitCmd.AddCommand(emitFileCmd)

	emitFileCmd.Flags().StringVarP(&fileOutput, "output", "o", "", "Output file path (default: stdout)")

	emitFileCmd.Flags().StringVarP(&filePrefixOverride, "prefix", "p", "", "Version prefix (default 'v' if flag provided without value)")
	emitFileCmd.Flag("prefix").NoOptDefVal = useDefaultMarker

	emitFileCmd.Flags().StringVar(&filePrereleaseTemplate, "prerelease", "", "Pre-release template (uses config default if flag provided without value)")
	emitFileCmd.Flag("prerelease").NoOptDefVal = useDefaultMarker

	emitFileCmd.Flags().StringVar(&fileMetadataTemplate, "metadata", "", "Metadata template (uses config default if flag provided without value)")
	emitFileCmd.Flag("metadata").NoOptDefVal = useDefaultMarker
}
