package mode

import (
	"github.com/cbroglie/mustache"

	"github.com/benjaminabbitt/versionator/internal/version"
)

const (
	// DefaultCDPrereleaseTemplate is the default pre-release template for CD mode
	DefaultCDPrereleaseTemplate = "build-{{CommitsSinceTag}}"
	// DefaultCDMetadataTemplate is the default metadata template for CD mode
	DefaultCDMetadataTemplate = "{{ShortHash}}"
)

// ContinuousDeliveryMode generates unique versions for every build
// Pre-release and metadata are auto-generated from templates
type ContinuousDeliveryMode struct {
	PrereleaseTemplate string
	MetadataTemplate   string
}

// Name returns the mode name
func (m *ContinuousDeliveryMode) Name() string {
	return "continuous-delivery"
}

// GetPreRelease generates pre-release from template
func (m *ContinuousDeliveryMode) GetPreRelease(_ *version.Version, templateData map[string]string) (string, error) {
	template := m.PrereleaseTemplate
	if template == "" {
		template = DefaultCDPrereleaseTemplate
	}

	return renderTemplate(template, templateData)
}

// GetMetadata generates metadata from template
func (m *ContinuousDeliveryMode) GetMetadata(_ *version.Version, templateData map[string]string) (string, error) {
	template := m.MetadataTemplate
	if template == "" {
		template = DefaultCDMetadataTemplate
	}

	return renderTemplate(template, templateData)
}

// IsReleaseMode returns false for CD mode
func (m *ContinuousDeliveryMode) IsReleaseMode() bool {
	return false
}

// renderTemplate renders a Mustache template with the given data
func renderTemplate(template string, data map[string]string) (string, error) {
	if template == "" {
		return "", nil
	}

	result, err := mustache.Render(template, data)
	if err != nil {
		return "", err
	}

	return result, nil
}
