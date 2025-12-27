// Package standard provides the standard SemVer versioning pattern plugin.
//
// Standard semantic versioning (SemVer 2.0.0) uses the format:
//
//	MAJOR.MINOR.PATCH[-PRERELEASE][+METADATA]
//
// This plugin configures default SemVer patterns:
//   - No version prefix (bare version numbers)
//   - Pre-release format: alpha-{{CommitsSinceTag}}
//   - Metadata format: {{BuildDateTimeCompact}}.{{ShortHash}}{{DirtyWithDot}}
//
// Example versions:
//
//	1.2.3
//	1.2.3-alpha-5
//	1.2.3-alpha-5+20241211103045.abc1234
//	1.2.3-alpha-5+20241211103045.abc1234.dirty
//
// This is the default versioning pattern when no --go flag is specified.
package standard

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
)

// Plugin implements the VersioningPlugin interface for standard SemVer
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "standard"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeVersioning)
}

// PatternName returns the versioning pattern identifier
func (p *Plugin) PatternName() string {
	return "standard"
}

// GetVersioningConfig returns standard SemVer configuration
func (p *Plugin) GetVersioningConfig() *plugin.VersioningConfig {
	return &plugin.VersioningConfig{
		Name:   "standard",
		Prefix: "",
		// Pre-release: simple alpha tag with commit count (joined with dashes)
		PreReleaseElements: []string{"alpha", "CommitsSinceTag"},
		// Metadata: timestamp, hash, dirty status (joined with dots)
		MetadataElements: []string{"BuildDateTimeCompact", "ShortHash", "Dirty"},
	}
}

func init() {
	plugin.Register(&Plugin{})
}
