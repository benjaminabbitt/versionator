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
// Called after any version change (set, bump, release) to keep manifest files in sync.
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

	templateData := emit.BuildCompleteTemplateData(v, cfg.PreRelease.Template, cfg.Metadata.Template)

	if err := updater.UpdateFiles(templateData); err != nil {
		return fmt.Errorf("error updating files: %w", err)
	}

	if files := updater.GetFilesToCommit(); len(files) > 0 {
		cmd.Printf("Updated %d file(s)\n", len(files))
	}

	return nil
}
