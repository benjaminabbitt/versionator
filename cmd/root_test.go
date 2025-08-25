package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// Helper functions for DRY test setup

// createRootTestFiles creates test files in the temp directory
func createRootTestFiles(t *testing.T, version string, prefix string) {
	// Create VERSION file
	err := afero.WriteFile(afero.NewOsFs(), "VERSION", []byte(version), 0644)
	require.NoError(t, err, "Failed to create VERSION file")

	// Create config file
	configContent := `prefix: "` + prefix + `"
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = afero.WriteFile(afero.NewOsFs(), ".versionator.yaml", []byte(configContent), 0644)
	require.NoError(t, err, "Failed to create config file")
}

func TestExecute_Success(t *testing.T) {
	// Get original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Create temporary directory and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)

		// Change back to original directory
		if originalDir != "" {
			err := os.Chdir(originalDir)
			require.NoError(t, err, "Failed to restore original directory")
		}
	}()

	// Create test files
	createRootTestFiles(t, "1.0.0", "")

	// Test Execute function doesn't panic
	err = Execute()
	// Since Execute() runs the root command without args, it should show help and return nil
	require.NoError(t, err, "Execute() should not return error")
}

func TestVersionCommand(t *testing.T) {
	// Get original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Create temporary directory and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)

		// Change back to original directory
		if originalDir != "" {
			err := os.Chdir(originalDir)
			require.NoError(t, err, "Failed to restore original directory")
		}
	}()

	// Create test files
	createRootTestFiles(t, "2.1.0", "")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err = rootCmd.Execute()
	require.NoError(t, err, "version command should succeed")

	// Check output contains the version
	output := buf.String()
	require.Equal(t, "2.1.0\n", output, "Version command should output correct version")
}

func TestVersionCommand_WithPrefix(t *testing.T) {
	// Get original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Create temporary directory and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)

		// Change back to original directory
		if originalDir != "" {
			err := os.Chdir(originalDir)
			require.NoError(t, err, "Failed to restore original directory")
		}
	}()

	// Create test files with prefix
	createRootTestFiles(t, "3.0.0", "v")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err = rootCmd.Execute()
	require.NoError(t, err, "version command should succeed")

	// Check output contains the prefixed version
	output := buf.String()
	require.Equal(t, "v3.0.0\n", output, "Version command should output prefixed version")
}

func TestVersionCommand_NoVersionFile(t *testing.T) {
	// Get original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Create temporary directory and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	defer func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)

		// Change back to original directory
		if originalDir != "" {
			err := os.Chdir(originalDir)
			require.NoError(t, err, "Failed to restore original directory")
		}
	}()

	// Create only config file (no VERSION file)
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = afero.WriteFile(afero.NewOsFs(), ".versionator.yaml", []byte(configContent), 0644)
	require.NoError(t, err, "Failed to create config file")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err = rootCmd.Execute()
	require.NoError(t, err, "version command should succeed with default version")

	// Check output contains the default version
	output := buf.String()
	require.Equal(t, "0.0.0\n", output, "Version command should output default version when no VERSION file exists")
}

func TestLogFormatFlag(t *testing.T) {
	testCases := []struct {
		name   string
		flag   string
		format string
	}{
		{
			name:   "console format",
			flag:   "--log-format=console",
			format: "console",
		},
		{
			name:   "json format",
			flag:   "--log-format=json",
			format: "json",
		},
		{
			name:   "development format",
			flag:   "--log-format=development",
			format: "development",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get original directory
			originalDir, err := os.Getwd()
			require.NoError(t, err, "Failed to get current working directory")

			// Create temporary directory and change to it
			tempDir := t.TempDir()
			err = os.Chdir(tempDir)
			require.NoError(t, err, "Failed to change to temp directory")

			defer func() {
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)

				// Change back to original directory
				if originalDir != "" {
					err := os.Chdir(originalDir)
					require.NoError(t, err, "Failed to restore original directory")
				}
			}()

			// Create test files
			createRootTestFiles(t, "1.0.0", "")

			// Capture stdout
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs([]string{tc.flag, "version"})

			// Execute the command with log format flag
			err = rootCmd.Execute()
			require.NoError(t, err, "Command with log format flag should succeed")

			// Check that version is still output correctly
			output := buf.String()
			require.Equal(t, "1.0.0\n", output, "Version should be output correctly regardless of log format")
		})
	}
}