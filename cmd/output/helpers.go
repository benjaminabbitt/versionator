package output

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"
	"github.com/spf13/cobra"
)

// BuildTemplateData creates template data with prerelease and metadata handling
func BuildTemplateData(cmd *cobra.Command, vd *version.Version, prefixOverride, prereleaseTemplate, metadataTemplate string) (emit.TemplateData, error) {
	// Handle prefix override
	var prefix string
	if cmd.Flags().Changed("prefix") {
		if prefixOverride == UseDefaultMarker {
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
		if prereleaseTemplate == UseDefaultMarker {
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
		if metadataTemplate == UseDefaultMarker {
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
