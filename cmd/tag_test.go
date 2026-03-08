package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// TagTestSuite defines the test suite for tag command tests
type TagTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	tempDir string
	origDir string
}

// SetupSuite runs once before all tests in the suite
func (suite *TagTestSuite) SetupSuite() {
	// This runs once for the entire suite
}

// TearDownSuite runs once after all tests in the suite
func (suite *TagTestSuite) TearDownSuite() {
	// This runs once after the entire suite
}

// SetupTest runs before each test
func (suite *TagTestSuite) SetupTest() {
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
	suite.resetTagCommand()
}

// TearDownTest runs after each test
func (suite *TagTestSuite) TearDownTest() {
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

// resetTagCommand resets the tag command state between tests
func (suite *TagTestSuite) resetTagCommand() {
	// Reset command output and args
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Reset tag command flags to their default values
	tagCmd.Flags().Set("message", "")
	tagCmd.Flags().Set("prefix", "v")
	tagCmd.Flags().Set("force", "false")
	tagCmd.Flags().Set("verbose", "false")
	tagCmd.Flags().Set("no-branch", "false")
}

// createTestFiles creates the standard test files needed for most tests
func (suite *TagTestSuite) createTestFiles(version string) {
	suite.createTestFilesWithRelease(version, true)
}

// createTestFilesWithRelease creates test files with configurable release branching
func (suite *TagTestSuite) createTestFilesWithRelease(version string, createBranch bool) {
	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create a config file with release settings
	configContent := `prefix: ""
metadata:
  template: ""
  git:
    hashLength: 7
release:
  createBranch: %t
  branchPrefix: "release/"
logging:
  output: "console"
`
	err = os.WriteFile(".versionator.yaml", []byte([]byte(fmt.Sprintf(configContent, createBranch))), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

func (suite *TagTestSuite) TestTagCommand_Success() {
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
	mockVCS.EXPECT().BranchExists("release/v1.2.3").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.2.3").Return(nil)
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag", "--verbose"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed")

	// Check output contains success message
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.2.3' for version 1.2.3 using git", "Should contain success message")
	suite.Contains(output, "Successfully created branch 'release/v1.2.3'", "Should contain branch success message")
	suite.Contains(output, "Message: Release 1.2.3", "Should contain verbose message output")
	suite.Contains(output, "git ID: abc1234", "Should contain verbose git ID output")
}

func (suite *TagTestSuite) TestTagCommand_CustomPrefix() {
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
	mockVCS.EXPECT().BranchExists("release/release-2.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/release-2.0.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag", "--prefix", "release-"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed")

	// Check output contains success message with custom prefix
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'release-2.0.0'", "Should contain success message with custom prefix")
}

func (suite *TagTestSuite) TestTagCommand_CustomMessage() {
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
	mockVCS.EXPECT().BranchExists("release/v1.5.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.5.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag", "--message", "Custom release message"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed")

	// Check output contains success message
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.5.0'", "Should contain success message")
}

func (suite *TagTestSuite) TestTagCommand_NoVCS() {
	// Create test files
	suite.createTestFiles("1.0.0")

	// Ensure no VCS is registered (already handled in TearDownTest)
	vcs.UnregisterVCS("git")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"tag"})

	// Execute the tag command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected tag command to fail when no VCS is available")
}

func (suite *TagTestSuite) TestTagCommand_DirtyWorkingDirectory() {
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
	rootCmd.SetArgs([]string{"tag"})

	// Execute the tag command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected tag command to fail when working directory is dirty")
}

func (suite *TagTestSuite) TestTagCommand_TagExists_NoForce() {
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
	rootCmd.SetArgs([]string{"tag"})

	// Execute the tag command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected tag command to fail when tag exists and force is not used")
}

func (suite *TagTestSuite) TestTagCommand_TagExists_WithForce() {
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
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.0.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag", "--force"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed with force flag")

	// Check output contains success message
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain success message")
}

func (suite *TagTestSuite) TestTagCommand_NoVersionFile() {
	// Create only config file (no VERSION file)
	configContent := `prefix: ""
metadata:
  template: ""
  git:
    hashLength: 7
release:
  createBranch: true
  branchPrefix: "release/"
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
	mockVCS.EXPECT().BranchExists("release/v0.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v0.0.0").Return(nil)

	// Register mock VCS and set as active
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag"})

	// Execute the tag command - should succeed with default version
	err = rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed with default version")

	// Check output contains success message with default version
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v0.0.0' for version 0.0.0", "Should contain success message with default version")
}

func (suite *TagTestSuite) TestTagCommand_NoBranchFlag() {
	// Create test files with branch creation enabled
	suite.createTestFiles("1.0.0")

	// Setup mock VCS - should NOT expect CreateBranch or BranchExists calls
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	// No branch calls expected due to --no-branch flag

	// Register mock VCS
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag", "--no-branch"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed")

	// Check output
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain tag success message")
	suite.NotContains(output, "Successfully created branch", "Should NOT contain branch success message")
}

func (suite *TagTestSuite) TestTagCommand_BranchDisabledInConfig() {
	// Create test files with branch creation disabled
	suite.createTestFilesWithRelease("1.0.0", false)

	// Setup mock VCS - should NOT expect CreateBranch or BranchExists calls
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	// No branch calls expected due to config

	// Register mock VCS
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed")

	// Check output
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain tag success message")
	suite.NotContains(output, "Successfully created branch", "Should NOT contain branch success message")
}

func (suite *TagTestSuite) TestTagCommand_BranchAlreadyExists() {
	// Create test files with branch creation enabled
	suite.createTestFiles("1.0.0")

	// Setup mock VCS - branch already exists
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(true, nil)
	// CreateBranch should NOT be called since branch exists

	// Register mock VCS
	vcs.RegisterVCS(mockVCS)

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"tag"})

	// Execute the tag command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "tag command should succeed")

	// Check output contains warning about existing branch
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain tag success message")
	suite.Contains(output, "Warning: branch 'release/v1.0.0' already exists", "Should contain warning about existing branch")
}

// TestTagTestSuite runs the tag test suite
func TestTagTestSuite(t *testing.T) {
	suite.Run(t, new(TagTestSuite))
}
