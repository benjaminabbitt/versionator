package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// CommitTestSuite defines the test suite for commit command tests
type CommitTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	tempDir string
	origDir string
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
	// Create a temporary directory for testing
	suite.tempDir = suite.T().TempDir()
	var err error
	suite.origDir, err = os.Getwd()
	suite.Require().NoError(err)
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)

	// Initialize gomock controller
	suite.ctrl = gomock.NewController(suite.T())

	// Reset command state to prevent flag pollution
	suite.resetCommitCommand()
}

// TearDownTest runs after each test
func (suite *CommitTestSuite) TearDownTest() {
	// Restore original directory
	if suite.origDir != "" {
		os.Chdir(suite.origDir)
	}

	// Finish gomock controller
	if suite.ctrl != nil {
		suite.ctrl.Finish()
	}

	// Clean up any registered VCS
	vcs.UnregisterVCS("git")
}

// resetCommitCommand resets the commit command state between tests
func (suite *CommitTestSuite) resetCommitCommand() {
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

// createTestFiles creates the standard test files needed for most tests
func (suite *CommitTestSuite) createTestFiles(version string) {
	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create a minimal config file
	configContent := `prefix: ""
metadata:
  template: ""
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

func (suite *CommitTestSuite) TestCommitCommand_Success() {
	// Create test files
	suite.createTestFiles("1.2.3")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.2.3").Return(false, nil)
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
	suite.Contains(output, "Successfully created tag 'v1.2.3' for version 1.2.3 using git", "Should contain success message")
	suite.Contains(output, "Message: Release 1.2.3", "Should contain verbose message output")
	suite.Contains(output, "git ID: abc1234", "Should contain verbose git ID output")
}

func (suite *CommitTestSuite) TestCommitCommand_CustomPrefix() {
	// Create test files
	suite.createTestFiles("2.0.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("release-2.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("release-2.0.0", "Release 2.0.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"commit", "--prefix", "release-"})

	// Execute the commit command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "commit command should succeed")

	// Check output contains success message with custom prefix
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'release-2.0.0'", "Should contain success message with custom prefix")
}

func (suite *CommitTestSuite) TestCommitCommand_CustomMessage() {
	// Create test files
	suite.createTestFiles("1.5.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
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
	suite.createTestFiles("1.0.0")

	// Ensure no VCS is registered (already handled in TearDownTest)
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
	suite.createTestFiles("1.0.0")

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
	suite.createTestFiles("1.0.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
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
	suite.createTestFiles("1.0.0")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)
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
	configContent := `prefix: ""
metadata:
  template: ""
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Setup mock VCS
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
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

	// Check output contains success message with default version
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v0.0.0' for version 0.0.0", "Should contain success message with default version")
}

// TestCommitTestSuite runs the commit test suite
func TestCommitTestSuite(t *testing.T) {
	suite.Run(t, new(CommitTestSuite))
}
