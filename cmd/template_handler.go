package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

// templateKind identifies which template type we're working with
type templateKind string

const (
	templateKindPreRelease templateKind = "pre-release"
	templateKindMetadata   templateKind = "metadata"
)

// versionFieldAccessor names everything that differs between the two SemVer
// fields versionator manages as commands (pre-release and build metadata).
// Every `config <field>` subcommand body is shared and parameterized by one of
// these, so a change to one field's behaviour cannot silently skip the other.
type versionFieldAccessor struct {
	kind        templateKind
	getStable   func(*config.Config) bool
	setStable   func(*config.Config, bool)
	getTemplate func(*config.Config) string
	setTemplate func(*config.Config, string)
	setVersion  func(string) error
	getVersion  func(*version.Version) string
	// defaultValue supplies the value `enable` writes when the template renders
	// empty. This is the one genuinely field-specific rule: metadata falls back
	// to a git hash, pre-release to a fixed label.
	defaultValue func(*config.Config, *version.Version) string
	forceFlag    *bool
	cmdName      string // subcommand and flag spelling, e.g. "prerelease"
	labelTitle   string // sentence-initial label, e.g., "Pre-release" or "Metadata"
	labelLower   string // mid-sentence label, e.g., "pre-release" or "metadata"
}

var prereleaseAccessor = versionFieldAccessor{
	kind:         templateKindPreRelease,
	getStable:    func(c *config.Config) bool { return c.PreRelease.Stable },
	setStable:    func(c *config.Config, v bool) { c.PreRelease.Stable = v },
	getTemplate:  func(c *config.Config) string { return c.PreRelease.Template },
	setTemplate:  func(c *config.Config, t string) { c.PreRelease.Template = t },
	setVersion:   version.SetPreRelease,
	getVersion:   func(v *version.Version) string { return v.PreRelease },
	defaultValue: func(*config.Config, *version.Version) string { return defaultPreReleaseValue },
	forceFlag:    &prereleaseForceFlag,
	cmdName:      "prerelease",
	labelTitle:   "Pre-release",
	labelLower:   "pre-release",
}

var metadataAccessor = versionFieldAccessor{
	kind:         templateKindMetadata,
	getStable:    func(c *config.Config) bool { return c.Metadata.Stable },
	setStable:    func(c *config.Config, v bool) { c.Metadata.Stable = v },
	getTemplate:  func(c *config.Config) string { return c.Metadata.Template },
	setTemplate:  func(c *config.Config, t string) { c.Metadata.Template = t },
	setVersion:   version.SetMetadata,
	getVersion:   func(v *version.Version) string { return v.BuildMetadata },
	defaultValue: metadataDefaultValue,
	forceFlag:    &metadataForceFlag,
	cmdName:      "metadata",
	labelTitle:   "Metadata",
	labelLower:   "metadata",
}

// runTemplateCommand handles the template subcommand for both prerelease and metadata
func runTemplateCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// If no argument, show current template
	if len(args) == 0 {
		return showTemplate(cmd, cfg, acc)
	}

	// Set new template
	return setTemplate(cmd, cfg, args[0], acc)
}

// showTemplate displays the current template configuration
func showTemplate(cmd *cobra.Command, cfg *config.Config, acc versionFieldAccessor) error {
	cmd.Printf("Stable: %t\n", acc.getStable(cfg))
	cmd.Printf("Template: %s\n", acc.getTemplate(cfg))

	// Show what it would render to
	template := acc.getTemplate(cfg)
	if template != "" {
		vd, err := version.Load()
		if err == nil {
			templateData := emit.BuildTemplateDataFromVersion(vd)
			result, err := emit.RenderTemplateWithData(template, templateData)
			if err == nil && result != "" {
				cmd.Printf("Rendered value: %s\n", result)
			}
		}
	}
	return nil
}

// setTemplate sets a new template value
func setTemplate(cmd *cobra.Command, cfg *config.Config, template string, acc versionFieldAccessor) error {
	acc.setTemplate(cfg, template)
	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	cmd.Printf("%s template set to: %s\n", acc.labelTitle, template)

	// If stable, also render and write to VERSION file
	if acc.getStable(cfg) {
		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error loading version: %w", err)
		}

		templateData := emit.BuildTemplateDataFromVersion(vd)
		result, err := emit.RenderTemplateWithData(template, templateData)
		if err != nil {
			return fmt.Errorf("error rendering template: %w", err)
		}

		if err := acc.setVersion(result); err != nil {
			return fmt.Errorf("error setting %s: %w", acc.labelLower, err)
		}

		cmd.Printf("%s set to: %s\n", acc.labelTitle, result)

		// Show updated version
		vd, err = version.Load()
		if err != nil {
			return fmt.Errorf("error loading version: %w", err)
		}
		cmd.Printf("Current version: %s\n", vd.FullString())
	} else {
		cmd.Println("(Template will be rendered at output time, not stored in VERSION file)")
	}

	return nil
}
