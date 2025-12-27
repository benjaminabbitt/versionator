// Package goversion provides the Go versioning pattern plugin for versionator.
//
// CANONICAL PRE-RELEASE USE CASE
//
// Pre-release is the canonical versioning feature for Go projects. While
// pre-release can be used in any ecosystem, it was designed primarily to
// address Go's unique versioning requirements. Other ecosystems may use
// pre-release for Go-compatible versioning via 'versionator init <lang> --go'.
//
// Go modules use a specific versioning pattern that includes:
//   - Version prefix: "v" (required by Go modules)
//   - Pre-release format compatible with Go pseudo-versions
//   - No build metadata (Go ignores +suffix entirely)
//
// Go pseudo-versions follow the pattern:
//
//	v0.0.0-20191109021931-daa7c04131f5
//
// This plugin configures the pre-release template to produce compatible output:
//
//	{{CommitsSinceTag}}{{BuildDateTimeCompactWithDash}}{{ShortHashWithDash}}{{DirtyWithDash}}
//
// Which produces versions like:
//
//	v1.2.3-5-20241211103045-abc1234
//	v1.2.3-5-20241211103045-abc1234-dirty  (with uncommitted changes)
//
// This format is compatible with Go's module system and allows proper
// version ordering when the commit count increases.
//
// Use the --go flag with versionator init to apply this versioning pattern
// to any language target:
//
//	versionator init --go         # Go project
//	versionator init rust --go    # Rust project with Go-style versioning
//	versionator init python --go  # Python project with Go-style versioning
package goversion

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
)

// Plugin implements the VersioningPlugin interface for Go versioning pattern
type Plugin struct{}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "goversion"
}

// Types returns the plugin types this plugin implements
func (p *Plugin) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeVersioning)
}

// PatternName returns the versioning pattern identifier
func (p *Plugin) PatternName() string {
	return "go"
}

// GetVersioningConfig returns Go versioning pattern configuration
func (p *Plugin) GetVersioningConfig() *plugin.VersioningConfig {
	return &plugin.VersioningConfig{
		Name:   "go",
		Prefix: "v",
		// Pre-release: Go pseudo-version compatible format (joined with dashes)
		PreReleaseElements: []string{"CommitsSinceTag", "BuildDateTimeCompact", "ShortHash", "Dirty"},
		// Metadata: Same data (joined with dots, Go ignores but useful for --go with other ecosystems)
		MetadataElements: []string{"BuildDateTimeCompact", "ShortHash", "Dirty"},
	}
}

func init() {
	plugin.Register(&Plugin{})
}
