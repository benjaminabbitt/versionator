package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/update"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runConfiguredUpdates applies file updates from .versionator.yaml config.
// Called after every command that changes the VERSION file (set, bump, the
// config prerelease/metadata/prefix writers; release applies them itself) to
// keep manifest files such as Cargo.toml in sync.
func runConfiguredUpdates(cmd *cobra.Command) error {
	cfg, err := config.ReadConfig()
	if err != nil || cfg == nil || len(cfg.Updates) == 0 {
		return nil
	}

	v, err := version.Load()
	if err != nil {
		return fmt.Errorf("error loading version for updates: %w", err)
	}

	logger, _ := zap.NewProduction()
	updater := update.NewUpdater(cfg.Updates, update.NewDaselFileParser(), logger)

	templateData := updateTemplateData(v, cfg)

	if err := updater.UpdateFiles(templateData); err != nil {
		return fmt.Errorf("error updating files: %w", err)
	}

	if files := updater.GetFilesToCommit(); len(files) > 0 {
		cmd.Printf("Updated %d file(s)\n", len(files))
	}

	return nil
}

// updateTemplateData is the template data an update renders: the version with
// its pre-release and metadata as the version's output has them. A field stored
// in the VERSION file (stable, or left there) is that value; a dynamic field
// with nothing stored is its template, rendered.
func updateTemplateData(v *version.Version, cfg *config.Config) emit.TemplateData {
	preTemplate, metaTemplate := cfg.PreRelease.Template, cfg.Metadata.Template
	if cfg.PreRelease.Stable || v.PreRelease != "" {
		preTemplate = ""
	}
	if cfg.Metadata.Stable || v.BuildMetadata != "" {
		metaTemplate = ""
	}

	data := emit.BuildCompleteTemplateData(v, preTemplate, metaTemplate)
	if v.PreRelease != "" {
		data.PreRelease = v.PreRelease
		data.PreReleaseWithDash = "-" + v.PreRelease
	}
	if v.BuildMetadata != "" {
		data.Metadata = v.BuildMetadata
		data.MetadataWithPlus = "+" + v.BuildMetadata
	}
	return data
}
