package mode

import (
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
)

// Mode defines how version output is calculated
type Mode interface {
	// Name returns the mode name
	Name() string

	// GetPreRelease returns the pre-release string for this mode
	// For release mode, returns the VERSION file pre-release
	// For CD mode, generates from template
	GetPreRelease(v *version.Version, templateData map[string]string) (string, error)

	// GetMetadata returns the metadata string for this mode
	// For release mode, returns the VERSION file metadata
	// For CD mode, generates from template
	GetMetadata(v *version.Version, templateData map[string]string) (string, error)

	// IsReleaseMode returns true if this is the default release mode
	IsReleaseMode() bool
}

// ModeType represents the type of versioning mode
type ModeType string

const (
	ModeTypeRelease            ModeType = "release"
	ModeTypeContinuousDelivery ModeType = "continuous-delivery"
)

// GetMode returns the active mode based on configuration
func GetMode(cfg *config.Config) Mode {
	if cfg == nil {
		return &ReleaseMode{}
	}

	switch ModeType(cfg.Mode.Type) {
	case ModeTypeContinuousDelivery:
		return &ContinuousDeliveryMode{
			PrereleaseTemplate: cfg.Mode.ContinuousDelivery.PrereleaseTemplate,
			MetadataTemplate:   cfg.Mode.ContinuousDelivery.MetadataTemplate,
		}
	default:
		return &ReleaseMode{}
	}
}
