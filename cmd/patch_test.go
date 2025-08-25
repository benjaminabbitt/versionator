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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper methods for DRY test setup

// createPatchTestApp creates a fresh filesystem and test app instance
func createPatchTestApp() (afero.Fs, *app.App) {
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

// getPatchStandardConfigContent returns the standard config content used across tests
func getPatchStandardConfigContent() string {
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

// createPatchConfigFile creates the standard config file in the filesystem
func createPatchConfigFile(t *testing.T, fs afero.Fs) {
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(getPatchStandardConfigContent()), 0644)
	require.NoError(t, err)
}

// createPatchVersionFile creates a VERSION file with the specified content if provided
func createPatchVersionFile(t *testing.T, fs afero.Fs, version string) {
	if version != "" {
		err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
		require.NoError(t, err)
	}
}

// replacePatchAppInstance replaces the global app instance and returns a restore function
func replacePatchAppInstance(testApp *app.App) func() {
	originalApp := appInstance
	appInstance = testApp
	return func() {
		appInstance = originalApp
	}
}

// verifyPatchVersionFile verifies the VERSION file contains the expected content
func verifyPatchVersionFile(t *testing.T, fs afero.Fs, expectedVersion string) {
	content, err := afero.ReadFile(fs, "VERSION")
	require.NoError(t, err)
	actualVersion := strings.TrimSpace(string(content))
	assert.Equal(t, expectedVersion, actualVersion)
}

func TestPatchCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		initialVersion string
		expectedVersion string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "increment from 1.2.3",
			args:           []string{"patch", "increment"},
			initialVersion: "1.2.3",
			expectedVersion: "1.2.4",
			expectError:    false,
		},
		{
			name:           "increment with inc alias",
			args:           []string{"patch", "inc"},
			initialVersion: "0.5.7",
			expectedVersion: "0.5.8",
			expectError:    false,
		},
		{
			name:           "increment with + alias",
			args:           []string{"patch", "+"},
			initialVersion: "2.1.9",
			expectedVersion: "2.1.10",
			expectError:    false,
		},
		{
			name:           "decrement from 1.3.5",
			args:           []string{"patch", "decrement"},
			initialVersion: "1.3.5",
			expectedVersion: "1.3.4",
			expectError:    false,
		},
		{
			name:           "decrement with dec alias",
			args:           []string{"patch", "dec"},
			initialVersion: "2.5.1",
			expectedVersion: "2.5.0",
			expectError:    false,
		},
		{
			name:           "increment from default version",
			args:           []string{"patch", "increment"},
			initialVersion: "", // No VERSION file
			expectedVersion: "0.0.1",
			expectError:    false,
		},
		{
			name:           "increment large patch number",
			args:           []string{"patch", "increment"},
			initialVersion: "1.0.999",
			expectedVersion: "1.0.1000",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, testApp := createPatchTestApp()
			defer replacePatchAppInstance(testApp)()

			createPatchConfigFile(t, fs)
			createPatchVersionFile(t, fs, tt.initialVersion)

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				verifyPatchVersionFile(t, fs, tt.expectedVersion)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPatchCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "patch help",
			args: []string{"patch", "--help"},
		},
		{
			name: "patch increment help",
			args: []string{"patch", "increment", "--help"},
		},
		{
			name: "patch decrement help",
			args: []string{"patch", "decrement", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "Usage:")

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPatchCommandEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		initialVersion string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "increment with whitespace in version file",
			args:           []string{"patch", "increment"},
			initialVersion: "  1.2.3  \n",
			expectError:    false,
		},
		{
			name:           "decrement with trailing newline",
			args:           []string{"patch", "decrement"},
			initialVersion: "1.2.3\n",
			expectError:    false,
		},
		{
			name:           "invalid command",
			args:           []string{"patch", "invalid"},
			initialVersion: "1.2.3",
			expectError:    false, // Cobra shows help instead of returning error
			errorContains:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, testApp := createPatchTestApp()
			defer replacePatchAppInstance(testApp)()

			createPatchConfigFile(t, fs)
			createPatchVersionFile(t, fs, tt.initialVersion)

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			rootCmd.SetArgs(nil)
		})
	}
}
