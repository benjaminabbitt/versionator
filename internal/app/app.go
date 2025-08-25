package app

import (
	"versionator/internal/application"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/vcs/git"
	"versionator/internal/version"

	"github.com/spf13/afero"
)

// App holds all application dependencies and provides simplified method calls
type App struct {
	ConfigManager  *config.ConfigManager
	VersionManager *version.Version
	Versionator    *application.Versionator
	VCS            vcs.VersionControlSystem
	FileSystem     afero.Fs
}

// NewApp creates a new App instance with all dependencies initialized
func NewApp() *App {
	fs := afero.NewOsFs()
	gitVCS := git.NewGitVCS(fs)
	return &App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", gitVCS),
		Versionator:    application.NewVersionator(fs, gitVCS),
		VCS:            gitVCS,
		FileSystem:     fs,
	}
}

// Version delegation methods

// GetCurrentVersion gets the current version from VERSION file
func (a *App) GetCurrentVersion() (string, error) {
	return a.VersionManager.GetCurrentVersion()
}

// WriteVersion writes version to VERSION file
func (a *App) WriteVersion(version string) error {
	return a.VersionManager.WriteVersion(version)
}

// Increment increments the specified version level
func (a *App) Increment(level version.VersionLevel) error {
	return a.VersionManager.Increment(level)
}

// Decrement decrements the specified version level
func (a *App) Decrement(level version.VersionLevel) error {
	return a.VersionManager.Decrement(level)
}

// Versionator delegation methods

// GetVersionWithSuffix returns the version with optional prefix and git hash suffix
func (a *App) GetVersionWithSuffix() (string, error) {
	return a.Versionator.GetVersionWithSuffix()
}

// GetVersionWithPrefixAndSuffix is an alias for GetVersionWithSuffix for backward compatibility
func (a *App) GetVersionWithPrefixAndSuffix() (string, error) {
	return a.Versionator.GetVersionWithPrefixAndSuffix()
}

// Config delegation methods

// ReadConfig reads configuration from file
func (a *App) ReadConfig() (*config.Config, error) {
	return a.ConfigManager.ReadConfig()
}

// WriteConfig writes configuration to file
func (a *App) WriteConfig(cfg *config.Config) error {
	return a.ConfigManager.WriteConfig(cfg)
}
