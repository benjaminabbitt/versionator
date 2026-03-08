package mode

import (
	"github.com/benjaminabbitt/versionator/internal/version"
)

// ReleaseMode is the standard versioning mode
// Uses pre-release and metadata from the VERSION file as-is
type ReleaseMode struct{}

// Name returns the mode name
func (m *ReleaseMode) Name() string {
	return "release"
}

// GetPreRelease returns the pre-release from the VERSION file
func (m *ReleaseMode) GetPreRelease(v *version.Version, _ map[string]string) (string, error) {
	return v.PreRelease, nil
}

// GetMetadata returns the metadata from the VERSION file
func (m *ReleaseMode) GetMetadata(v *version.Version, _ map[string]string) (string, error) {
	return v.BuildMetadata, nil
}

// IsReleaseMode returns true for release mode
func (m *ReleaseMode) IsReleaseMode() bool {
	return true
}
