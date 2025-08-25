package cmd

import (
	"bytes"
	"strings"
	"testing"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// Helper functions for DRY test setup

// createMajorTestApp creates a fresh filesystem and test app instance
func createMajorTestApp() (afero.Fs, *app.App) {
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

// getMajorStandardConfigContent returns the standard config content used across tests
func getMajorStandardConfigContent() string {
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

// createMajorConfigFile creates the standard config file in the filesystem
func createMajorConfigFile(t *testing.T, fs afero.Fs) {
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(getMajorStandardConfigContent()), 0644)
	require.NoError(t, err, "Failed to create config file")
}

// createMajorVersionFile creates a VERSION file with the specified content
func createMajorVersionFile(t *testing.T, fs afero.Fs, version string) {
	err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
	require.NoError(t, err, "Failed to create VERSION file")
}

// replaceMajorAppInstance replaces the global app instance and returns a restore function
func replaceMajorAppInstance(testApp *app.App) func() {
	originalApp := appInstance
	appInstance = testApp
	return func() {
		appInstance = originalApp
	}
}

// verifyMajorVersionFile verifies the VERSION file contains the expected content
func verifyMajorVersionFile(t *testing.T, fs afero.Fs, expectedVersion string) {
	content, err := afero.ReadFile(fs, "VERSION")
	require.NoError(t, err, "Should be able to read VERSION file")
	require.Equal(t, expectedVersion, strings.TrimSpace(string(content)), "VERSION file should contain '"+expectedVersion+"'")
}

func TestMajorIncrementCommand(t *testing.T) {
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMajorTestApp()
	defer replaceMajorAppInstance(testApp)()

	createMajorConfigFile(t, fs)
	createMajorVersionFile(t, fs, "1.2.3")

	// Execute the major increment command
	rootCmd.SetArgs([]string{"major", "increment"})
	err := rootCmd.Execute()
	require.NoError(t, err, "major increment command should succeed")

	verifyMajorVersionFile(t, fs, "2.0.0")
}

func TestMajorIncrementCommand_Aliases(t *testing.T) {
	testCases := []string{"inc", "+"}

	for _, alias := range testCases {
		t.Run("alias_"+alias, func(t *testing.T) {
			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			fs, testApp := createMajorTestApp()
			defer replaceMajorAppInstance(testApp)()

			createMajorConfigFile(t, fs)
			createMajorVersionFile(t, fs, "0.1.0")

			// Execute the major increment command with alias
			rootCmd.SetArgs([]string{"major", alias})
			err := rootCmd.Execute()
			require.NoError(t, err, "major %s command should succeed", alias)

			verifyMajorVersionFile(t, fs, "1.0.0")
		})
	}
}

func TestMajorDecrementCommand(t *testing.T) {
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMajorTestApp()
	defer replaceMajorAppInstance(testApp)()

	createMajorConfigFile(t, fs)
	createMajorVersionFile(t, fs, "3.5.7")

	// Execute the major decrement command
	rootCmd.SetArgs([]string{"major", "decrement"})
	err := rootCmd.Execute()
	require.NoError(t, err, "major decrement command should succeed")

	verifyMajorVersionFile(t, fs, "2.0.0")
}

func TestMajorDecrementCommand_Aliases(t *testing.T) {
	testCases := []string{"dec"}

	for _, alias := range testCases {
		t.Run("alias_"+alias, func(t *testing.T) {
			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			fs, testApp := createMajorTestApp()
			defer replaceMajorAppInstance(testApp)()

			createMajorConfigFile(t, fs)
			createMajorVersionFile(t, fs, "2.1.0")

			// Execute the major decrement command with alias
			rootCmd.SetArgs([]string{"major", alias})
			err := rootCmd.Execute()
			require.NoError(t, err, "major %s command should succeed", alias)

			verifyMajorVersionFile(t, fs, "1.0.0")
		})
	}
}

func TestMajorIncrementCommand_NoVersionFile(t *testing.T) {
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMajorTestApp()
	defer replaceMajorAppInstance(testApp)()

	// Create only config file (no VERSION file)
	createMajorConfigFile(t, fs)

	// Execute the major increment command - should succeed with default version
	rootCmd.SetArgs([]string{"major", "increment"})
	err := rootCmd.Execute()
	require.NoError(t, err, "major increment command should succeed with default version")

	verifyMajorVersionFile(t, fs, "1.0.0")
}

func TestMajorDecrementCommand_AtZero(t *testing.T) {
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMajorTestApp()
	defer replaceMajorAppInstance(testApp)()

	createMajorConfigFile(t, fs)
	createMajorVersionFile(t, fs, "0.5.3")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"major", "decrement"})

	// Execute the major decrement command - should fail
	err := rootCmd.Execute()
	require.Error(t, err, "Expected major decrement command to fail when major version is at 0")
}

func TestMajorCommand_InvalidVersionFile(t *testing.T) {
	// Test both increment and decrement with invalid version
	testCases := []string{"increment", "decrement"}

	for _, operation := range testCases {
		t.Run(operation, func(t *testing.T) {
			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			fs, testApp := createMajorTestApp()
			defer replaceMajorAppInstance(testApp)()

			createMajorConfigFile(t, fs)
			createMajorVersionFile(t, fs, "invalid.version")

			// Capture stderr
			var buf bytes.Buffer
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"major", operation})

			// Execute the command - should fail
			err := rootCmd.Execute()
			require.Error(t, err, "Expected major %s command to fail with invalid version file", operation)
		})
	}
}