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

	data, err := buildTemplateData(v, cfg, effectivePrefix(v.Prefix))
	if err != nil {
		return err
	}

	if err := updater.UpdateFiles(data); err != nil {
		return fmt.Errorf("error updating files: %w", err)
	}

	if files := updater.GetFilesToCommit(); len(files) > 0 {
		cmd.Printf("Updated %d file(s)\n", len(files))
	}

	return nil
}

// withConfigRollback runs change, which writes the config and then the VERSION
// file; when change fails (a version change the update check refused), the
// config is put back as it was, byte for byte.
func withConfigRollback(change func() error) error {
	restore, err := config.Snapshot()
	if err != nil {
		return err
	}
	if err := change(); err != nil {
		if restoreErr := restore(); restoreErr != nil {
			return fmt.Errorf("%w (restoring .versionator.yaml also failed: %v)", err, restoreErr)
		}
		return err
	}
	return nil
}

// checkConfiguredUpdates is the version change check (version.SetChangeCheck)
// the root command sets: before VERSION is written with next, every configured
// update is rendered for next and applied in memory. A refusal refuses the
// version change, so neither VERSION nor any configured file changes.
func checkConfiguredUpdates(next *version.Version) error {
	cfg, err := config.ReadConfig()
	if err != nil || cfg == nil || len(cfg.Updates) == 0 {
		return nil
	}

	logger, _ := zap.NewProduction()
	updater := update.NewUpdater(cfg.Updates, update.NewDaselFileParser(), logger)
	data, err := buildTemplateData(next, cfg, effectivePrefix(next.Prefix))
	if err != nil {
		return err
	}
	if err := updater.Check(data); err != nil {
		return fmt.Errorf("error updating files: %w", err)
	}
	return nil
}

// effectivePrefix is the {{Prefix}} release and the file updates render: the
// VERSION file's prefix, else "v" (release tags "v1.2.3" when no prefix is
// configured anywhere). release's --prefix flag replaces it.
func effectivePrefix(versionPrefix string) string {
	if versionPrefix == "" {
		return "v"
	}
	return versionPrefix
}

// buildTemplateData is the ONE builder of the template data a version renders
// with: every version change's file updates, the pre-save check of those
// updates, and release's tag, branch, message and file updates. So the check
// validates exactly what release will write.
//
// Pre-release and metadata are the version's output values: a field stored in
// the VERSION file (stable, or left there) is that value; a dynamic field with
// nothing stored is its template, rendered. On top are the release variables:
// {{Prefix}} (the given prefix), {{ModulePath}}, {{ModuleName}} and
// {{BranchPrefix}}.
func buildTemplateData(v *version.Version, cfg *config.Config, prefix string) (emit.TemplateData, error) {
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

	modulePath, err := version.ModulePath()
	if err != nil {
		return data, fmt.Errorf("error resolving module path: %w", err)
	}
	moduleName, err := version.ModuleName()
	if err != nil {
		return data, fmt.Errorf("error resolving module name: %w", err)
	}
	emit.MergeCustomVars(&data, map[string]string{
		"Prefix":       prefix,
		"ModulePath":   modulePath,
		"ModuleName":   moduleName,
		"BranchPrefix": cfg.Release.BranchPrefix,
	})
	return data, nil
}
