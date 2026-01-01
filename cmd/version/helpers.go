package version

import (
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/internal/versionator"
)

// RenderFromConfig renders prerelease and metadata from config elements
// and saves the updated VERSION file. Called after version modifications.
func RenderFromConfig() error {
	// Load current version
	vd, err := version.Load()
	if err != nil {
		return err
	}

	// Load config
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	updated := false

	// Render prerelease from config elements if configured
	if len(cfg.PreRelease.Elements) > 0 {
		prerelease, err := versionator.RenderPreRelease()
		if err == nil {
			vd.PreRelease = prerelease
			updated = true
		}
	}

	// Render metadata from config elements if configured
	if len(cfg.Metadata.Elements) > 0 {
		metadata, err := versionator.RenderMetadata()
		if err == nil {
			vd.BuildMetadata = metadata
			updated = true
		}
	}

	// Apply prefix from config if VERSION has none
	if vd.Prefix == "" && cfg.Prefix != "" {
		vd.Prefix = cfg.Prefix
		updated = true
	}

	// Save if anything changed
	if updated {
		return version.Save(vd)
	}

	return nil
}
