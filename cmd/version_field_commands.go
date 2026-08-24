package cmd

import (
	"fmt"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

// The `config prerelease ...` and `config metadata ...` subcommands are the same
// commands over two different SemVer fields. Their bodies live here once, driven
// by a versionFieldAccessor, so a fix to one field cannot be applied without the
// other receiving it.

// runStableCommand shows or sets whether the field is stored in the VERSION file
// (stable) or rendered from a template at output time (dynamic).
func runStableCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// If no argument, show current setting
	if len(args) == 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s stable: %t\n", acc.labelTitle, acc.getStable(cfg))
		if acc.getStable(cfg) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  Value is stored in VERSION file")
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  Value is generated from template at output time")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Template: %s\n", acc.getTemplate(cfg))
		}
		return nil
	}

	// Set new value
	switch args[0] {
	case "true", "1", "yes":
		acc.setStable(cfg, true)
	case "false", "0", "no":
		acc.setStable(cfg, false)
	default:
		return fmt.Errorf("invalid value '%s': use 'true' or 'false'", args[0])
	}

	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s stable set to: %t\n", acc.labelTitle, acc.getStable(cfg))

	// Switching to dynamic mode leaves a stale literal behind in the VERSION
	// file, which would then win over the template at output time.
	if !acc.getStable(cfg) {
		if err := acc.setVersion(""); err != nil {
			return fmt.Errorf("error clearing %s from VERSION: %w", acc.labelLower, err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s cleared from VERSION file (will be generated at output time)\n", acc.labelTitle)
	}

	return nil
}

// runEnableCommand renders the configured template into the VERSION file.
func runEnableCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	if !acc.getStable(cfg) {
		return fmt.Errorf("%s is configured as dynamic (stable: false)\n"+
			"In dynamic mode, %s is generated at output time.\n"+
			"To use this command, first run: versionator config %s stable true",
			acc.labelLower, acc.labelLower, acc.cmdName)
	}

	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	value := ""
	if template := acc.getTemplate(cfg); template != "" {
		templateData := emit.BuildTemplateDataFromVersion(vd)
		rendered, err := emit.RenderTemplateWithData(template, templateData)
		if err == nil && rendered != "" {
			value = rendered
		}
	}

	// No template, or it rendered empty: fall back to the field's own default.
	if value == "" {
		value = acc.defaultValue(cfg, vd)
	}

	if err := acc.setVersion(value); err != nil {
		return fmt.Errorf("error setting %s: %w", acc.labelLower, err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s enabled with value '%s'\n", acc.labelTitle, value)

	// Reload so the reported version reflects the write above.
	vd, err = version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
	return nil
}

// clearVersionValue backs both `disable` and `clear`: each requires stable mode,
// empties the field in the VERSION file and reports the resulting version. They
// differ only in the remedy offered when the field is dynamic, and in the word
// used to report success.
func clearVersionValue(cmd *cobra.Command, acc versionFieldAccessor, remedy, pastTense string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	if !acc.getStable(cfg) {
		return fmt.Errorf("%s is configured as dynamic (stable: false)\n"+
			"In dynamic mode, %s is not stored in VERSION file.\n"+
			"%s", acc.labelLower, acc.labelLower, remedy)
	}

	if err := acc.setVersion(""); err != nil {
		return fmt.Errorf("error clearing %s: %w", acc.labelLower, err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", acc.labelTitle, pastTense)

	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
	return nil
}

// runDisableCommand clears the field from the VERSION file.
func runDisableCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	return clearVersionValue(cmd, acc,
		fmt.Sprintf("To disable dynamic %s at output, use --%s=\"\" flag or clear the template",
			acc.labelLower, acc.cmdName),
		"disabled")
}

// runStatusCommand reports the field's configuration and its effective value.
func runStatusCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error reading version: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Stable: %t\n", acc.getStable(cfg))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Template: %s\n", acc.getTemplate(cfg))

	if acc.getStable(cfg) {
		// Show value from VERSION file
		if stored := acc.getVersion(vd); stored != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "VALUE (from VERSION file): %s\n", stored)
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "VALUE (from VERSION file): (none)")
		}
	} else {
		// Show what would be rendered
		if template := acc.getTemplate(cfg); template != "" {
			templateData := emit.BuildTemplateDataFromVersion(vd)
			result, err := emit.RenderTemplateWithData(template, templateData)
			if err == nil && result != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "VALUE (rendered from template): %s\n", result)
			}
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "VALUE: (no template configured)")
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "VERSION file: %s\n", vd.FullString())
	return nil
}

// runSetCommand sets a literal value for the field.
func runSetCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	value := args[0]

	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	if !acc.getStable(cfg) && !*acc.forceFlag {
		return fmt.Errorf("%s is configured as dynamic (stable: false)\n"+
			"Cannot set a static value when %s is generated at output time.\n"+
			"Options:\n"+
			"  1. Switch to stable mode: versionator config %s stable true\n"+
			"  2. Use --force to set the template to this literal value\n"+
			"  3. Use 'template' command to set a dynamic template",
			acc.labelLower, acc.labelLower, acc.cmdName)
	}

	// Update template in config
	acc.setTemplate(cfg, value)
	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	if acc.getStable(cfg) {
		// Write to VERSION file
		if err := acc.setVersion(value); err != nil {
			return fmt.Errorf("error setting %s: %w", acc.labelLower, err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s set to: %s\n", acc.labelTitle, value)

		vd, err := version.Load()
		if err != nil {
			return fmt.Errorf("error getting version: %w", err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", vd.FullString())
	} else {
		// --force was used: set template to literal, don't write to VERSION
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s template set to literal: %s\n", acc.labelTitle, value)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "(Value will be used at output time, not stored in VERSION file)")
	}

	return nil
}

// runClearCommand removes the field's value from the VERSION file.
func runClearCommand(cmd *cobra.Command, args []string, acc versionFieldAccessor) error {
	return clearVersionValue(cmd, acc,
		fmt.Sprintf("To clear the template, use: versionator config %s template \"\"", acc.cmdName),
		"cleared")
}
