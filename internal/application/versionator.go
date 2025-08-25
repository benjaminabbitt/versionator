package application

import (
	"fmt"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/version"

	"github.com/spf13/afero"
)

// Versionator handles version operations with configurable dependencies
type Versionator struct {
	configManager  *config.ConfigManager
	versionManager *version.Version
	vcs            vcs.VersionControlSystem
}

// NewVersionator creates a new Versionator with the provided filesystem and VCS
func NewVersionator(fs afero.Fs, vcsInstance vcs.VersionControlSystem) *Versionator {
	return &Versionator{
		configManager:  config.NewConfigManager(fs),
		versionManager: version.NewVersion(fs, ".", vcsInstance),
		vcs:            vcsInstance,
	}
}

// GetVersionWithSuffix returns the version with optional prefix and git hash suffix
func (v *Versionator) GetVersionWithSuffix() (string, error) {
	currentVersion, err := v.versionManager.GetCurrentVersion()
	if err != nil {
		return "", err
	}

	config, err := v.configManager.ReadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	// Start with the base version
	finalVersion := currentVersion

	// Add VCS suffix if enabled
	if config.Suffix.Enabled && config.Suffix.Type == "git" {
		// Use the injected VCS implementation
		if v.vcs != nil && v.vcs.IsRepository() {
			hashLength := config.Suffix.Git.HashLength
			vcsHash, err := v.vcs.GetVCSIdentifier(hashLength)
			if err == nil {
				finalVersion = fmt.Sprintf("%s-%s", finalVersion, vcsHash)
			}
			// If VCS hash fails, continue without suffix
		}
	}

	// Apply prefix if configured
	if config.Prefix != "" {
		finalVersion = config.Prefix + finalVersion
	}

	return finalVersion, nil
}

// GetVersionWithPrefixAndSuffix is an alias for GetVersionWithSuffix for backward compatibility
func (v *Versionator) GetVersionWithPrefixAndSuffix() (string, error) {
	return v.GetVersionWithSuffix()
}
