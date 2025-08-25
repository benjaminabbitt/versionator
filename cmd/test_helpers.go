package cmd

import (
	"strings"
	"testing"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Common test helper functions for DRY test setup across all test files

// createTestApp creates a fresh filesystem and test app instance
func createTestApp() (afero.Fs, *app.App) {
	fs := afero.NewMemMapFs()
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", nil),
		Versionator:    versionator.NewVersionator(fs, nil),
		VCS:            nil,
		FileSystem:     fs,
	}
	return fs, testApp
}

// getStandardConfigContent returns the standard config content used across tests
func getStandardConfigContent() string {
	return `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
}

// createConfigFile creates the standard config file in the filesystem
func createConfigFile(t *testing.T, fs afero.Fs) {
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(getStandardConfigContent()), 0644)
	require.NoError(t, err, "Failed to create config file")
}

// createVersionFile creates a VERSION file with the specified content if provided
func createVersionFile(t *testing.T, fs afero.Fs, version string) {
	if version != "" {
		err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
		require.NoError(t, err, "Failed to create VERSION file")
	}
}

// replaceAppInstance replaces the global app instance and returns a restore function
func replaceAppInstance(testApp *app.App) func() {
	originalApp := appInstance
	appInstance = testApp
	return func() {
		appInstance = originalApp
	}
}

// verifyVersionFile verifies the VERSION file contains the expected content
func verifyVersionFile(t *testing.T, fs afero.Fs, expectedVersion string) {
	content, err := afero.ReadFile(fs, "VERSION")
	require.NoError(t, err, "Should be able to read VERSION file")
	actualVersion := strings.TrimSpace(string(content))
	assert.Equal(t, expectedVersion, actualVersion, "VERSION file should contain '"+expectedVersion+"'")
}