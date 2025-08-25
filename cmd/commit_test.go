package cmd

import (
	"bytes"
	"testing"
	"versionator/internal/app"
	"versionator/internal/vcs/mock"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// CommitTestSuite contains the test suite for commit command
type CommitTestSuite struct {
	suite.Suite
	testApp      *app.App
	fs           afero.Fs
	restoreApp   func()
	outputBuffer *bytes.Buffer
	ctrl         *gomock.Controller
	mockVCS      *mock.MockVersionControlSystem
}

// SetupTest initializes the test environment before each test
func (suite *CommitTestSuite) SetupTest() {
	var testApp *app.App
	suite.fs, testApp = createTestApp()
	suite.testApp = testApp
	suite.restoreApp = replaceAppInstance(testApp)
	suite.outputBuffer = &bytes.Buffer{}
	
	// Set up gomock controller and mock VCS
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockVCS = mock.NewMockVersionControlSystem(suite.ctrl)
	suite.testApp.VCS = suite.mockVCS
}

// TearDownTest cleans up after each test
func (suite *CommitTestSuite) TearDownTest() {
	suite.restoreApp()
	suite.ctrl.Finish()
	
	// Reset flags to prevent test interference
	commitCmd.Flags().Set("message", "")
	commitCmd.Flags().Set("prefix", "v")
	commitCmd.Flags().Set("force", "false")
	commitCmd.Flags().Set("verbose", "false")
}

// TestCommitSuccessfulTagCreation tests successful tag creation with clean working directory
func (suite *CommitTestSuite) TestCommitSuccessfulTagCreation() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.2.3")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	suite.mockVCS.EXPECT().TagExists("v1.2.3").Return(false, nil)
	suite.mockVCS.EXPECT().CreateTag("v1.2.3", "Release 1.2.3").Return(nil)
	suite.mockVCS.EXPECT().Name().Return("git").AnyTimes()

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Commit should succeed with clean working directory")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Successfully created tag 'v1.2.3' for version 1.2.3 using git", "Output should show successful tag creation")
}

// TestCommitWithCustomPrefix tests tag creation with custom prefix
func (suite *CommitTestSuite) TestCommitWithCustomPrefix() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.0.0")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	suite.mockVCS.EXPECT().TagExists("release-2.0.0").Return(false, nil)
	suite.mockVCS.EXPECT().CreateTag("release-2.0.0", "Release 2.0.0").Return(nil)
	suite.mockVCS.EXPECT().Name().Return("git").AnyTimes()

	// Create command with custom prefix flag
	cmd := commitCmd
	cmd.SetOut(suite.outputBuffer)
	cmd.Flags().Set("prefix", "release-")

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Commit should succeed with custom prefix")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Successfully created tag 'release-2.0.0' for version 2.0.0", "Output should show tag with custom prefix")
}

// TestCommitWithCustomMessage tests tag creation with custom message
func (suite *CommitTestSuite) TestCommitWithCustomMessage() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.5.0")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	suite.mockVCS.EXPECT().TagExists("v1.5.0").Return(false, nil)
	suite.mockVCS.EXPECT().CreateTag("v1.5.0", "Custom release message").Return(nil)
	suite.mockVCS.EXPECT().Name().Return("git").AnyTimes()

	// Create command with custom message flag
	cmd := commitCmd
	cmd.SetOut(suite.outputBuffer)
	cmd.Flags().Set("message", "Custom release message")

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Commit should succeed with custom message")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Successfully created tag 'v1.5.0' for version 1.5.0", "Output should show successful tag creation")
}

// TestCommitWithVerboseOutput tests verbose output functionality
func (suite *CommitTestSuite) TestCommitWithVerboseOutput() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "3.1.0")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	suite.mockVCS.EXPECT().TagExists("v3.1.0").Return(false, nil)
	suite.mockVCS.EXPECT().CreateTag("v3.1.0", "Release 3.1.0").Return(nil)
	suite.mockVCS.EXPECT().Name().Return("git").AnyTimes()
	suite.mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

	// Create command with verbose flag
	cmd := commitCmd
	cmd.SetOut(suite.outputBuffer)
	cmd.Flags().Set("verbose", "true")

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Commit should succeed with verbose output")

	// Verify the output includes verbose information
	output := suite.outputBuffer.String()
	suite.Contains(output, "Successfully created tag 'v3.1.0'", "Output should show successful tag creation")
	suite.Contains(output, "Message: Release 3.1.0", "Output should show tag message")
	suite.Contains(output, "git ID: abc1234", "Output should show VCS identifier")
}

// TestCommitDirtyWorkingDirectory tests error when working directory is dirty
func (suite *CommitTestSuite) TestCommitDirtyWorkingDirectory() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.0.0")

	// Mock VCS to return dirty working directory
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Commit should fail with dirty working directory")
	suite.Contains(err.Error(), "working directory is not clean", "Error should indicate dirty working directory")
}

// TestCommitTagAlreadyExists tests error when tag already exists without force flag
func (suite *CommitTestSuite) TestCommitTagAlreadyExists() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.0.0")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	suite.mockVCS.EXPECT().TagExists("v2.0.0").Return(true, nil)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Commit should fail when tag already exists")
	suite.Contains(err.Error(), "tag 'v2.0.0' already exists", "Error should indicate tag already exists")
	suite.Contains(err.Error(), "Use --force to overwrite", "Error should suggest force flag")
}

// TestCommitForceOverwriteExistingTag tests successful tag overwrite with force flag
func (suite *CommitTestSuite) TestCommitForceOverwriteExistingTag() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.1.0")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	suite.mockVCS.EXPECT().TagExists("v2.1.0").Return(true, nil)
	suite.mockVCS.EXPECT().CreateTag("v2.1.0", "Release 2.1.0").Return(nil)
	suite.mockVCS.EXPECT().Name().Return("git").AnyTimes()

	// Create command with force flag
	cmd := commitCmd
	cmd.SetOut(suite.outputBuffer)
	cmd.Flags().Set("force", "true")

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Commit should succeed with force flag")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Successfully created tag 'v2.1.0' for version 2.1.0", "Output should show successful tag creation")
}

// TestCommitNoVCS tests error when no VCS is available
func (suite *CommitTestSuite) TestCommitNoVCS() {
	// Setup: Create config and version files but no VCS
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.0.0")
	
	// Set VCS to nil
	suite.testApp.VCS = nil

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the commit command
	err := commitCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Commit should fail when no VCS is available")
	suite.Contains(err.Error(), "not in a version control repository", "Error should indicate no VCS")
}

// TestCommitCommandStructure tests that the command is properly structured
func (suite *CommitTestSuite) TestCommitCommandStructure() {
	// Test command properties
	suite.Equal("commit", commitCmd.Use, "Commit command should have correct use")
	suite.Equal("Create a git tag for the current version", commitCmd.Short, "Commit command should have correct short description")
	suite.Contains(commitCmd.Long, "Create a git tag for the current version", "Commit command should have correct long description")
	
	// Test flags exist
	messageFlag := commitCmd.Flags().Lookup("message")
	suite.NotNil(messageFlag, "Should have message flag")
	suite.Equal("m", messageFlag.Shorthand, "Message flag should have correct shorthand")
	
	prefixFlag := commitCmd.Flags().Lookup("prefix")
	suite.NotNil(prefixFlag, "Should have prefix flag")
	suite.Equal("p", prefixFlag.Shorthand, "Prefix flag should have correct shorthand")
	
	forceFlag := commitCmd.Flags().Lookup("force")
	suite.NotNil(forceFlag, "Should have force flag")
	suite.Equal("f", forceFlag.Shorthand, "Force flag should have correct shorthand")
	
	verboseFlag := commitCmd.Flags().Lookup("verbose")
	suite.NotNil(verboseFlag, "Should have verbose flag")
	suite.Equal("v", verboseFlag.Shorthand, "Verbose flag should have correct shorthand")
}

// TestCommitCommandHierarchy tests that command is properly registered in the command hierarchy
func (suite *CommitTestSuite) TestCommitCommandHierarchy() {
	// Find commit command in root
	var foundCommitCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "commit" {
			foundCommitCmd = cmd
			break
		}
	}
	suite.NotNil(foundCommitCmd, "Commit command should be registered with root command")
}

// TestCommitTestSuite runs the test suite
func TestCommitTestSuite(t *testing.T) {
	suite.Run(t, new(CommitTestSuite))
}