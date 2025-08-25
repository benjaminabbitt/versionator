package cmd

import (
	"bytes"
	"strings"
	"testing"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// Helper functions for DRY test setup

// createMinorTestApp creates a fresh filesystem and test app instance
func createMinorTestApp() (afero.Fs, *app.App) {
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

// getMinorStandardConfigContent returns the standard config content used across tests
func getMinorStandardConfigContent() string {
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

// createMinorConfigFile creates the standard config file in the filesystem
func createMinorConfigFile(t *testing.T, fs afero.Fs) {
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(getMinorStandardConfigContent()), 0644)
	require.NoError(t, err, "Failed to create config file")
}

// createMinorVersionFile creates a VERSION file with the specified content
func createMinorVersionFile(t *testing.T, fs afero.Fs, version string) {
	err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
	require.NoError(t, err, "Failed to create VERSION file")
}

// replaceMinorAppInstance replaces the global app instance and returns a restore function
func replaceMinorAppInstance(testApp *app.App) func() {
	originalApp := appInstance
	appInstance = testApp
	return func() {
		appInstance = originalApp
	}
}

// verifyMinorVersionFile verifies the VERSION file contains the expected content
func verifyMinorVersionFile(t *testing.T, fs afero.Fs, expectedVersion string) {
	content, err := afero.ReadFile(fs, "VERSION")
	require.NoError(t, err, "Should be able to read VERSION file")
	require.Equal(t, expectedVersion, strings.TrimSpace(string(content)), "VERSION file should contain '"+expectedVersion+"'")
}

func TestMinorIncrementCommand(t *testing.T) {
	// Unregister Git VCS to prevent interference with tests
	vcs.UnregisterVCS("git")
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMinorTestApp()
	defer replaceMinorAppInstance(testApp)()

	createMinorConfigFile(t, fs)
	createMinorVersionFile(t, fs, "1.2.3")

	// Execute the minor increment command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"minor", "increment"})
	err := rootCmd.Execute()
	require.NoError(t, err, "minor increment command should succeed")
	
	// Reset command state
	rootCmd.SetArgs([]string{})

	verifyMinorVersionFile(t, fs, "1.3.0")
}

func TestMinorIncrementCommand_Aliases(t *testing.T) {
	testCases := []string{"inc", "+"}

	for _, alias := range testCases {
		t.Run("alias_"+alias, func(t *testing.T) {
			// Unregister Git VCS to prevent interference with tests
			vcs.UnregisterVCS("git")
			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			fs, testApp := createMinorTestApp()
			defer replaceMinorAppInstance(testApp)()

			createMinorConfigFile(t, fs)
			createMinorVersionFile(t, fs, "0.5.7")

			// Execute the minor increment command with alias
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"minor", alias})
			err := rootCmd.Execute()
			require.NoError(t, err, "minor %s command should succeed", alias)
			
			// Reset command state
			rootCmd.SetArgs([]string{})

			verifyMinorVersionFile(t, fs, "0.6.0")
		})
	}
}

func TestMinorDecrementCommand(t *testing.T) {
	// Unregister Git VCS to prevent interference with tests
	vcs.UnregisterVCS("git")
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMinorTestApp()
	defer replaceMinorAppInstance(testApp)()

	createMinorConfigFile(t, fs)
	createMinorVersionFile(t, fs, "1.3.5")

	// Execute the minor decrement command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"minor", "decrement"})
	err := rootCmd.Execute()
	require.NoError(t, err, "minor decrement command should succeed")
	
	// Reset command state
	rootCmd.SetArgs([]string{})

	verifyMinorVersionFile(t, fs, "1.2.0")
}

func TestMinorDecrementCommand_Aliases(t *testing.T) {
	testCases := []string{"dec"}

	for _, alias := range testCases {
		t.Run("alias_"+alias, func(t *testing.T) {
			// Unregister Git VCS to prevent interference with tests
			vcs.UnregisterVCS("git")
			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			fs, testApp := createMinorTestApp()
			defer replaceMinorAppInstance(testApp)()

			createMinorConfigFile(t, fs)
			createMinorVersionFile(t, fs, "2.5.1")

			// Execute the minor decrement command with alias
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"minor", alias})
			err := rootCmd.Execute()
			require.NoError(t, err, "minor %s command should succeed", alias)
			
			// Reset command state
			rootCmd.SetArgs([]string{})

			verifyMinorVersionFile(t, fs, "2.4.0")
		})
	}
}

func TestMinorIncrementCommand_NoVersionFile(t *testing.T) {
	// Unregister Git VCS to prevent interference with tests
	vcs.UnregisterVCS("git")
	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	fs, testApp := createMinorTestApp()
	defer replaceMinorAppInstance(testApp)()

	// Create only config file (no VERSION file)
	createMinorConfigFile(t, fs)

	// Execute the minor increment command - should succeed with default version
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"minor", "increment"})
	err := rootCmd.Execute()
	require.NoError(t, err, "minor increment command should succeed with default version")
	
	// Reset command state
	rootCmd.SetArgs([]string{})

	verifyMinorVersionFile(t, fs, "0.1.0")
}

func TestMinorCommandHelp(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "minor help",
			args: []string{"minor", "--help"},
		},
		{
			name: "minor increment help",
			args: []string{"minor", "increment", "--help"},
		},
		{
			name: "minor decrement help",
			args: []string{"minor", "decrement", "--help"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tc.args)

			err := rootCmd.Execute()
			require.NoError(t, err, "Help command should succeed")

			output := buf.String()
			require.Contains(t, output, "Usage:", "Help output should contain usage information")
		})
	}
}