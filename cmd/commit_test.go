package cmd

import (
	"bytes"
	"os"
	"testing"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/vcs/mock"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// Helper functions for DRY test setup

// createCommitTestApp creates a fresh filesystem and test app instance
func createCommitTestApp(vcsInstance vcs.VersionControlSystem) (afero.Fs, *app.App) {
	fs := afero.NewMemMapFs()
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", vcsInstance),
		Versionator:    versionator.NewVersionator(fs, vcsInstance),
		VCS:            vcsInstance,
		FileSystem:     fs,
	}
	return fs, testApp
}

// getCommitStandardConfigContent returns the standard config content used across tests
func getCommitStandardConfigContent() string {
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

// createCommitConfigFile creates the standard config file in the filesystem
func createCommitConfigFile(t *testing.T, fs afero.Fs) {
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(getCommitStandardConfigContent()), 0644)
	require.NoError(t, err, "Failed to create config file")
}

// createCommitVersionFile creates a VERSION file with the specified content
func createCommitVersionFile(t *testing.T, fs afero.Fs, version string) {
	err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
	require.NoError(t, err, "Failed to create VERSION file")
}

// replaceCommitAppInstance replaces the global app instance and returns a restore function
func replaceCommitAppInstance(testApp *app.App) func() {
	originalApp := appInstance
	appInstance = testApp
	return func() {
		appInstance = originalApp
	}
}

// resetCommitCommand resets the commit command state between tests
func resetCommitCommand() {
	// Reset command output and args
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Reset commit command flags to their default values
	commitCmd.Flags().Set("message", "")
	commitCmd.Flags().Set("prefix", "v")
	commitCmd.Flags().Set("force", "false")
	commitCmd.Flags().Set("verbose", "false")
}

func TestCommitCommand_Success(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.2.3").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.2.3", "Release 1.2.3").Return(nil)
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "1.2.3")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--verbose"})

	// Execute the commit command
	err = rootCmd.Execute()
	require.NoError(t, err, "commit command should succeed")

	// Check output contains success message
	output := buf.String()
	require.Contains(t, output, "Successfully created tag 'v1.2.3' for version 1.2.3 using git", "Should contain success message")
	require.Contains(t, output, "Message: Release 1.2.3", "Should contain verbose message output")
	require.Contains(t, output, "git ID: abc1234", "Should contain verbose git ID output")
}

func TestCommitCommand_CustomPrefix(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("release-2.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("release-2.0.0", "Release 2.0.0").Return(nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "2.0.0")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--prefix", "release-"})

	// Execute the commit command
	err = rootCmd.Execute()
	require.NoError(t, err, "commit command should succeed")

	// Check output contains success message with custom prefix
	output := buf.String()
	require.Contains(t, output, "Successfully created tag 'release-2.0.0'", "Should contain success message with custom prefix")
}

func TestCommitCommand_CustomMessage(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.5.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.5.0", "Custom release message").Return(nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "1.5.0")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--message", "Custom release message"})

	// Execute the commit command
	err = rootCmd.Execute()
	require.NoError(t, err, "commit command should succeed")

	// Check output contains success message
	output := buf.String()
	require.Contains(t, output, "Successfully created tag 'v1.5.0'", "Should contain success message")
}

func TestCommitCommand_NoVCS(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	fs, testApp := createCommitTestApp(nil)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "1.0.0")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should fail
	err = rootCmd.Execute()
	require.Error(t, err, "Expected commit command to fail when no VCS is available")
}

func TestCommitCommand_DirtyWorkingDirectory(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "1.0.0")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should fail
	err = rootCmd.Execute()
	require.Error(t, err, "Expected commit command to fail when working directory is dirty")
}

func TestCommitCommand_TagExists_NoForce(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "1.0.0")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should fail
	err = rootCmd.Execute()
	require.Error(t, err, "Expected commit command to fail when tag exists and force is not used")
}

func TestCommitCommand_TagExists_WithForce(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create test files in memory filesystem
	createCommitConfigFile(t, fs)
	createCommitVersionFile(t, fs, "1.0.0")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--force"})

	// Execute the commit command
	err = rootCmd.Execute()
	require.NoError(t, err, "commit command should succeed with force flag")

	// Check output contains success message
	output := buf.String()
	require.Contains(t, output, "Successfully created tag 'v1.0.0'", "Should contain success message")
}

func TestCommitCommand_NoVersionFile(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		if origDir != "" {
			os.Chdir(origDir)
		}
		vcs.UnregisterVCS("git")
		resetCommitCommand()
	}()

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v0.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v0.0.0", "Release 0.0.0").Return(nil)

	fs, testApp := createCommitTestApp(mockVCS)
	defer replaceCommitAppInstance(testApp)()

	// Create only config file (no VERSION file)
	createCommitConfigFile(t, fs)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should succeed with default version
	err = rootCmd.Execute()
	require.NoError(t, err, "commit command should succeed with default version")

	// Check output contains success message with default version
	output := buf.String()
	require.Contains(t, output, "Successfully created tag 'v0.0.0' for version 0.0.0", "Should contain success message with default version")
}