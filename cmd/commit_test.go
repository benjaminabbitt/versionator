package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/benjaminabbitt/versionator/cmd/output"
	"github.com/benjaminabbitt/versionator/internal/filesystem"
	fstesting "github.com/benjaminabbitt/versionator/internal/filesystem/testing"
	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// CommitTestSuite defines the test suite for commit command tests
type CommitTestSuite struct {
	suite.Suite
	ctrl      *gomock.Controller
	memFs     *fstesting.MemFs
	fsCleanup func()
	cwd       string // current working directory for absolute paths
}

// SetupSuite runs once before all tests in the suite
func (suite *CommitTestSuite) SetupSuite() {
	// This runs once for the entire suite
}

// TearDownSuite runs once after all tests in the suite
func (suite *CommitTestSuite) TearDownSuite() {
	// This runs once after the entire suite
}

// SetupTest runs before each test
func (suite *CommitTestSuite) SetupTest() {
	// Get current working directory for absolute paths
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	suite.cwd = cwd

	// Set up in-memory filesystem
	suite.memFs, suite.fsCleanup = fstesting.SetupTestFs()

	// Initialize gomock controller
	suite.ctrl = gomock.NewController(suite.T())

	// Reset command state to prevent flag pollution
	suite.resetCommitCommand()
}

// TearDownTest runs after each test
func (suite *CommitTestSuite) TearDownTest() {
	// Finish gomock controller
	if suite.ctrl != nil {
		suite.ctrl.Finish()
	}
	// Clean up VCS
	vcs.UnregisterVCS("git")
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Restore real filesystem
	if suite.fsCleanup != nil {
		suite.fsCleanup()
	}
}

// absPath returns absolute path for a filename relative to cwd
func (suite *CommitTestSuite) absPath(filename string) string {
	return filepath.Join(suite.cwd, filename)
}

// resetCommitCommand resets the commit command state between tests
func (suite *CommitTestSuite) resetCommitCommand() {
	// Reset command output and args
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Reset commit command flags to their default values
	commitCmd.Flags().Set("message", "")
	commitCmd.Flags().Set("force", "false")

	// Also reset TagCmd flags since commit proxies to it
	output.TagCmd.Flags().Set("message", "")
	output.TagCmd.Flags().Set("force", "false")
	output.TagCmd.SetOut(nil)
	output.TagCmd.SetErr(nil)

	// Reset verbosity to prevent leakage between tests
	logging.ResetVerbosity()
	verboseCount = 0 // Reset the package-level verbose count
}

// createTestFiles creates the standard test files needed for most tests
// The version should include prefix if one is needed (e.g., "v1.2.3")
func (suite *CommitTestSuite) createTestFiles(version string) {
	// Create a VERSION file with absolute path
	err := filesystem.AppFs.WriteFile(suite.absPath("VERSION"), []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create a minimal config file
	configContent := `prefix: "v"
prerelease:
  elements: []
metadata:
  elements: []
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = filesystem.AppFs.WriteFile(suite.absPath(".versionator.yaml"), []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

func (suite *CommitTestSuite) TestCommitCommand_Success() {
	// Create test files - VERSION now contains full version with prefix
	suite.createTestFiles("v1.2.3")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.cwd, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.2.3").Return(false, nil)
	// Note: tag name is FullString() (with prefix), message uses String() (core version)
	mockVCS.EXPECT().CreateTag("v1.2.3", "Release 1.2.3").Return(nil)
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--verbose"})

	// Execute the commit command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "commit command should succeed")

	// Check output contains success message
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.2.3'", "Should contain success message")
}

func (suite *CommitTestSuite) TestCommitCommand_CustomMessage() {
	// Create test files - VERSION now contains full version with prefix
	suite.createTestFiles("v1.5.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.cwd, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.5.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.5.0", "Custom release message").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--message", "Custom release message"})

	// Execute the commit command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "commit command should succeed")

	// Check output contains success message
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.5.0'", "Should contain success message")
}

func (suite *CommitTestSuite) TestCommitCommand_NoVCS() {
	// Create test files
	suite.createTestFiles("v1.0.0")

	// Ensure no VCS is registered
	vcs.UnregisterVCS("git")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected commit command to fail when no VCS is available")
}

func (suite *CommitTestSuite) TestCommitCommand_DirtyWorkingDirectory() {
	// Create test files
	suite.createTestFiles("v1.0.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected commit command to fail when working directory is dirty")
}

func (suite *CommitTestSuite) TestCommitCommand_TagExists_NoForce() {
	// Create test files
	suite.createTestFiles("v1.0.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.cwd, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected commit command to fail when tag exists and force is not used")
}

func (suite *CommitTestSuite) TestCommitCommand_TagExists_WithForce() {
	// Create test files
	suite.createTestFiles("v1.0.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.cwd, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)
	// Note: message uses String() (core version without prefix)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--force"})

	// Execute the commit command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "commit command should succeed with force flag")

	// Check output contains success message
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain success message")
}

func (suite *CommitTestSuite) TestCommitCommand_NoVersionFile() {
	// Create only config file (no VERSION file)
	configContent := `prefix: "v"
prerelease:
  elements: []
metadata:
  elements: []
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(suite.absPath(".versionator.yaml"), []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Setup mock VCS - default version is 0.0.0 with "v" prefix from config
	// Load() creates VERSION with prefix from config when file doesn't exist
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.cwd, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	// Tag name is FullString() with prefix, message is String() (core version)
	mockVCS.EXPECT().TagExists("v0.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v0.0.0", "Release 0.0.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit"})

	// Execute the commit command - should succeed with default version
	err = rootCmd.Execute()
	suite.Require().NoError(err, "commit command should succeed with default version")

	// Check output contains success message with default version (includes prefix)
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v0.0.0'", "Should contain success message with default version")
}

// TestCommitTestSuite runs the commit test suite
func TestCommitTestSuite(t *testing.T) {
	suite.Run(t, new(CommitTestSuite))
}
