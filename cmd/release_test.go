package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// ReleaseTestSuite defines the test suite for release command tests.
// The release command creates VCS tags and optionally branches for semantic versions.
type ReleaseTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	tempDir string
	origDir string
}

// SetupSuite runs once before all tests in the suite
func (suite *ReleaseTestSuite) SetupSuite() {
	// This runs once for the entire suite
}

// TearDownSuite runs once after all tests in the suite
func (suite *ReleaseTestSuite) TearDownSuite() {
	// This runs once after the entire suite
}

// SetupTest runs before each test
func (suite *ReleaseTestSuite) SetupTest() {
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
	suite.resetReleaseCommand()
}

// TearDownTest runs after each test
func (suite *ReleaseTestSuite) TearDownTest() {
	// Restore original directory
	if suite.origDir != "" {
		_ = os.Chdir(suite.origDir)
	}

	// Finish gomock controller
	if suite.ctrl != nil {
		suite.ctrl.Finish()
	}

	// Reset command state for other tests
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Clean up mock VCS and re-register real git VCS for other tests
	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
}

// resetReleaseCommand resets the release command state between tests
func (suite *ReleaseTestSuite) resetReleaseCommand() {
	// Reset command output and args
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Reset release command flags to their default values
	_ = releaseCmd.Flags().Set("message", "")
	_ = releaseCmd.Flags().Set("prefix", "v")
	_ = releaseCmd.Flags().Set("force", "false")
	_ = releaseCmd.Flags().Set("verbose", "false")
	_ = releaseCmd.Flags().Set("no-branch", "false")

	// Reset release push command flags
	_ = releasePushCmd.Flags().Set("message", "")
	_ = releasePushCmd.Flags().Set("prefix", "v")
	_ = releasePushCmd.Flags().Set("force", "false")
	_ = releasePushCmd.Flags().Set("verbose", "false")
	_ = releasePushCmd.Flags().Set("no-branch", "false")
}

// createTestFiles creates the standard test files needed for most tests
func (suite *ReleaseTestSuite) createTestFiles(version string) {
	suite.createTestFilesWithRelease(version, true)
}

// createTestFilesWithRelease creates test files with configurable release branching
func (suite *ReleaseTestSuite) createTestFilesWithRelease(version string, createBranch bool) {
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

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the release command: creating
// VCS tags and branches for semantic versions in a clean repository.
// =============================================================================

// TestReleaseCommand_Success validates that the release command creates both a tag
// and a release branch when the repository is clean and no conflicts exist.
//
// Why: This is the primary use case - users run `release` to mark a version in VCS.
// What: Given a clean repo with VERSION=1.2.3 and branch creation enabled,
// when release runs with --verbose, then it creates tag v1.2.3 and branch release/v1.2.3,
// outputting success messages including the git commit hash.
func (suite *ReleaseTestSuite) TestReleaseCommand_Success() {
	// Precondition: Repository has VERSION file with 1.2.3, config enables branch creation
	suite.createTestFiles("1.2.3")

	// Precondition: VCS reports clean working directory, no existing tag/branch
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

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "--verbose"})

	// Action: Execute the release command with verbose output
	err := rootCmd.Execute()

	// Expected: Command succeeds, output confirms tag, branch, message, and git ID
	suite.Require().NoError(err, "release command should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.2.3' for version 1.2.3 using git", "Should contain success message")
	suite.Contains(output, "Successfully created branch 'release/v1.2.3'", "Should contain branch success message")
	suite.Contains(output, "Message: Release 1.2.3", "Should contain verbose message output")
	suite.Contains(output, "git ID: abc1234", "Should contain verbose git ID output")
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows that modify the standard release behavior
// through flags or configuration options.
// =============================================================================

// TestReleaseCommand_CustomPrefix validates that the --prefix flag overrides
// the default "v" prefix for tag names.
//
// Why: Projects have different tag naming conventions (v1.0.0, release-1.0.0, etc.).
// What: Given VERSION=2.0.0, when release runs with --prefix=release-,
// then it creates tag "release-2.0.0" instead of "v2.0.0".
func (suite *ReleaseTestSuite) TestReleaseCommand_CustomPrefix() {
	// Precondition: Repository has VERSION file with 2.0.0
	suite.createTestFiles("2.0.0")

	// Precondition: VCS is clean, no existing tag/branch with custom prefix
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("release-2.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("release-2.0.0", "Release 2.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/release-2.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/release-2.0.0").Return(nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "--prefix", "release-"})

	// Action: Execute release with custom prefix
	err := rootCmd.Execute()

	// Expected: Tag uses custom prefix "release-2.0.0"
	suite.Require().NoError(err, "release command should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'release-2.0.0'", "Should contain success message with custom prefix")
}

// TestReleaseCommand_CustomMessage validates that the --message flag overrides
// the default "Release X.Y.Z" tag message.
//
// Why: Users may want descriptive release notes in the tag message.
// What: Given VERSION=1.5.0, when release runs with --message="Custom release message",
// then the tag is created with that custom message instead of the default.
func (suite *ReleaseTestSuite) TestReleaseCommand_CustomMessage() {
	// Precondition: Repository has VERSION file with 1.5.0
	suite.createTestFiles("1.5.0")

	// Precondition: VCS is clean, expects custom message in CreateTag call
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.5.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.5.0", "Custom release message").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.5.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.5.0").Return(nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "--message", "Custom release message"})

	// Action: Execute release with custom message
	err := rootCmd.Execute()

	// Expected: Command succeeds (mock verified the custom message was passed)
	suite.Require().NoError(err, "release command should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.5.0'", "Should contain success message")
}

// TestReleaseCommand_NoBranchFlag validates that the --no-branch flag prevents
// release branch creation even when enabled in config.
//
// Why: Users may want tag-only releases without branches for certain workflows.
// What: Given config with createBranch=true, when release runs with --no-branch,
// then only the tag is created, not the release branch.
func (suite *ReleaseTestSuite) TestReleaseCommand_NoBranchFlag() {
	// Precondition: Config has branch creation enabled
	suite.createTestFiles("1.0.0")

	// Precondition: VCS should NOT receive BranchExists or CreateBranch calls
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	// No branch calls expected due to --no-branch flag

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "--no-branch"})

	// Action: Execute release with --no-branch flag
	err := rootCmd.Execute()

	// Expected: Only tag created, no branch success message
	suite.Require().NoError(err, "release command should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain tag success message")
	suite.NotContains(output, "Successfully created branch", "Should NOT contain branch success message")
}

// TestReleaseCommand_BranchDisabledInConfig validates that branch creation can be
// disabled via the configuration file.
//
// Why: Some projects never use release branches and want to disable them globally.
// What: Given config with createBranch=false, when release runs,
// then only the tag is created, not the release branch.
func (suite *ReleaseTestSuite) TestReleaseCommand_BranchDisabledInConfig() {
	// Precondition: Config has branch creation disabled
	suite.createTestFilesWithRelease("1.0.0", false)

	// Precondition: VCS should NOT receive BranchExists or CreateBranch calls
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	// No branch calls expected due to config

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release (config disables branching)
	err := rootCmd.Execute()

	// Expected: Only tag created, no branch success message
	suite.Require().NoError(err, "release command should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain tag success message")
	suite.NotContains(output, "Successfully created branch", "Should NOT contain branch success message")
}

// TestReleaseCommand_AutoCommitVersionFile validates that the release command
// automatically commits a dirty VERSION file before creating the tag.
//
// Why: Users often bump the version and immediately release; auto-commit simplifies this workflow.
// What: Given a dirty VERSION file (only), when release runs,
// then it commits the VERSION file with the release message before tagging.
func (suite *ReleaseTestSuite) TestReleaseCommand_AutoCommitVersionFile() {
	// Precondition: Repository has VERSION file with 1.2.3
	suite.createTestFiles("1.2.3")

	// Precondition: VCS reports dirty but only VERSION file is modified
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)
	mockVCS.EXPECT().GetDirtyFiles().Return([]string{"VERSION"}, nil)
	mockVCS.EXPECT().CommitFiles([]string{"VERSION"}, "Release 1.2.3").Return(nil)
	mockVCS.EXPECT().TagExists("v1.2.3").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.2.3", "Release 1.2.3").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.2.3").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.2.3").Return(nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release with dirty VERSION file
	err := rootCmd.Execute()

	// Expected: VERSION committed, then tag and branch created
	suite.Require().NoError(err, "release command should succeed with auto-commit")
	output := buf.String()
	suite.Contains(output, "Committed VERSION file: Release 1.2.3", "Should contain commit message")
	suite.Contains(output, "Successfully created tag 'v1.2.3'", "Should contain success message")
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes that should produce clear error messages
// and prevent invalid releases.
// =============================================================================

// TestReleaseCommand_NoVCS validates that the release command fails gracefully
// when no version control system is available.
//
// Why: Prevents cryptic errors when running outside a VCS repository.
// What: Given no VCS is registered, when release runs,
// then it fails with an appropriate error.
func (suite *ReleaseTestSuite) TestReleaseCommand_NoVCS() {
	// Precondition: VERSION file exists but no VCS
	suite.createTestFiles("1.0.0")

	// Precondition: Unregister git VCS to simulate no VCS available
	vcs.UnregisterVCS("git")

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release without VCS
	err := rootCmd.Execute()

	// Expected: Command fails with error
	suite.Error(err, "Expected release command to fail when no VCS is available")
}

// TestReleaseCommand_TagExists_NoForce validates that the release command refuses
// to overwrite an existing tag without the --force flag.
//
// Why: Prevents accidental tag overwrites that could cause confusion or break CI/CD.
// What: Given tag v1.0.0 already exists and --force is not used,
// when release runs, then it fails with an error about the existing tag.
func (suite *ReleaseTestSuite) TestReleaseCommand_TagExists_NoForce() {
	// Precondition: VERSION file with 1.0.0
	suite.createTestFiles("1.0.0")

	// Precondition: VCS reports tag already exists
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release without --force when tag exists
	err := rootCmd.Execute()

	// Expected: Command fails, refusing to overwrite tag
	suite.Error(err, "Expected release command to fail when tag exists and force is not used")
}

// TestReleaseCommand_DirtyWithOtherFiles validates that the release command refuses
// to proceed when files other than VERSION are modified.
//
// Why: Prevents releasing with uncommitted changes that could lead to version confusion.
// What: Given VERSION and other.txt are both dirty,
// when release runs, then it fails with an error about dirty files.
func (suite *ReleaseTestSuite) TestReleaseCommand_DirtyWithOtherFiles() {
	// Precondition: VERSION file exists
	suite.createTestFiles("1.0.0")

	// Precondition: VCS reports multiple dirty files
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)
	mockVCS.EXPECT().GetDirtyFiles().Return([]string{"VERSION", "other.txt"}, nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release with non-VERSION files dirty
	err := rootCmd.Execute()

	// Expected: Command fails, refusing to release with dirty files
	suite.Error(err, "Expected release command to fail when other files are dirty")
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and unusual but valid scenarios.
// =============================================================================

// TestReleaseCommand_TagExists_WithForce validates that the --force flag allows
// overwriting an existing tag.
//
// Why: Sometimes tags need to be recreated (e.g., build failed after initial tag).
// What: Given tag v1.0.0 already exists and --force is used,
// when release runs, then it successfully overwrites the tag.
func (suite *ReleaseTestSuite) TestReleaseCommand_TagExists_WithForce() {
	// Precondition: VERSION file with 1.0.0
	suite.createTestFiles("1.0.0")

	// Precondition: VCS reports tag exists but CreateTag will be called anyway
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.0.0").Return(nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "--force"})

	// Action: Execute release with --force to overwrite existing tag
	err := rootCmd.Execute()

	// Expected: Tag successfully overwritten
	suite.Require().NoError(err, "release command should succeed with force flag")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain success message")
}

// TestReleaseCommand_NoVersionFile validates that the release command uses
// a default version when no VERSION file exists.
//
// Why: Allows initial releases without requiring pre-existing version infrastructure.
// What: Given no VERSION file exists,
// when release runs, then it uses default version 0.0.1 for the tag.
func (suite *ReleaseTestSuite) TestReleaseCommand_NoVersionFile() {
	// Precondition: Only config file exists, no VERSION file
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

	// Precondition: VCS expects default version 0.0.1
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v0.0.1").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v0.0.1", "Release 0.0.1").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v0.0.1").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v0.0.1").Return(nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release without VERSION file
	err = rootCmd.Execute()

	// Expected: Uses default version 0.0.1
	suite.Require().NoError(err, "release command should succeed with default version")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v0.0.1' for version 0.0.1", "Should contain success message with default version")
}

// TestReleaseCommand_BranchAlreadyExists validates that the release command
// warns but succeeds when the release branch already exists.
//
// Why: Allows re-running release after partial failures without branch conflicts.
// What: Given tag v1.0.0 doesn't exist but branch release/v1.0.0 does,
// when release runs, then it creates the tag and warns about the existing branch.
func (suite *ReleaseTestSuite) TestReleaseCommand_BranchAlreadyExists() {
	// Precondition: Config has branch creation enabled
	suite.createTestFiles("1.0.0")

	// Precondition: VCS reports branch exists, so CreateBranch won't be called
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(true, nil)
	// CreateBranch should NOT be called since branch exists

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release"})

	// Action: Execute release when branch already exists
	err := rootCmd.Execute()

	// Expected: Tag created, warning about existing branch
	suite.Require().NoError(err, "release command should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'", "Should contain tag success message")
	suite.Contains(output, "Warning: branch 'release/v1.0.0' already exists", "Should contain warning about existing branch")
}

// =============================================================================
// ERROR HANDLING
// Tests that verify expected failure modes and error responses.
// =============================================================================

// TestReleaseCommand_IsWorkingDirectoryCleanError validates error handling when status check fails.
//
// Why: Repository status checks can fail due to corruption or access issues.
// The release command must report these errors clearly rather than proceeding.
//
// What:
//   - Precondition: VERSION file exists, mock VCS returns error from IsWorkingDirectoryClean
//   - Action: Run "release" command
//   - Expected: Command returns error about status check
func (suite *ReleaseTestSuite) TestReleaseCommand_IsWorkingDirectoryCleanError() {
	// Precondition: VERSION file and mock VCS that errors on status check
	suite.createTestFiles("1.0.0")
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, fmt.Errorf("repository corrupted"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about status check
	suite.Error(err)
	suite.Contains(err.Error(), "error checking")
}

// TestReleaseCommand_GetDirtyFilesError validates error handling when dirty files query fails.
//
// Why: Dirty file enumeration can fail due to filesystem issues. The release
// command must report these errors rather than making assumptions.
//
// What:
//   - Precondition: VERSION file exists, VCS reports dirty but GetDirtyFiles fails
//   - Action: Run "release" command
//   - Expected: Command returns error about getting dirty files
func (suite *ReleaseTestSuite) TestReleaseCommand_GetDirtyFilesError() {
	// Precondition: VERSION file and mock VCS that reports dirty but fails on GetDirtyFiles
	suite.createTestFiles("1.0.0")
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)
	mockVCS.EXPECT().GetDirtyFiles().Return(nil, fmt.Errorf("cannot enumerate files"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about getting dirty files
	suite.Error(err)
	suite.Contains(err.Error(), "error getting dirty files")
}

// TestReleaseCommand_CommitFilesError validates error handling when auto-commit fails.
//
// Why: Auto-committing the VERSION file can fail due to hooks or permissions.
// The release must fail clearly rather than leaving the repo in an inconsistent state.
//
// What:
//   - Precondition: Only VERSION file is dirty, CommitFiles returns error
//   - Action: Run "release" command
//   - Expected: Command returns error about committing VERSION file
func (suite *ReleaseTestSuite) TestReleaseCommand_CommitFilesError() {
	// Precondition: VERSION file and mock VCS that fails on commit
	suite.createTestFiles("1.0.0")
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)
	mockVCS.EXPECT().GetDirtyFiles().Return([]string{"VERSION"}, nil)
	mockVCS.EXPECT().CommitFiles([]string{"VERSION"}, gomock.Any()).Return(fmt.Errorf("pre-commit hook failed"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about committing
	suite.Error(err)
	suite.Contains(err.Error(), "error committing VERSION file")
}

// TestReleaseCommand_TagExistsError validates error handling when tag existence check fails.
//
// Why: Tag existence checks can fail due to repository issues. The release
// must report these errors rather than potentially overwriting existing tags.
//
// What:
//   - Precondition: VERSION file exists, VCS is clean, TagExists returns error
//   - Action: Run "release" command
//   - Expected: Command returns error about checking tag existence
func (suite *ReleaseTestSuite) TestReleaseCommand_TagExistsError() {
	// Precondition: VERSION file and mock VCS that errors on TagExists
	suite.createTestFiles("1.0.0")
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, fmt.Errorf("tags unavailable"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about checking tag existence
	suite.Error(err)
	suite.Contains(err.Error(), "error checking if tag exists")
}

// TestReleaseCommand_CreateTagError validates error handling when tag creation fails.
//
// Why: Tag creation can fail due to permissions, signing issues, or other problems.
// These failures must be reported to prevent false success assumptions.
//
// What:
//   - Precondition: VERSION file exists, VCS is clean, tag doesn't exist, CreateTag fails
//   - Action: Run "release" command
//   - Expected: Command returns error about creating tag
func (suite *ReleaseTestSuite) TestReleaseCommand_CreateTagError() {
	// Precondition: VERSION file and mock VCS that errors on CreateTag
	suite.createTestFiles("1.0.0")
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", gomock.Any()).Return(fmt.Errorf("GPG signing failed"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about creating tag
	suite.Error(err)
	suite.Contains(err.Error(), "error creating tag")
}

// TestReleaseCommand_BranchExistsError validates error handling when branch check fails.
//
// Why: Branch existence checks can fail due to repository issues. The release
// must report these errors rather than skipping branch creation silently.
//
// What:
//   - Precondition: VERSION file, config with createBranch=true, BranchExists returns error
//   - Action: Run "release" command
//   - Expected: Command returns error about checking branch existence
func (suite *ReleaseTestSuite) TestReleaseCommand_BranchExistsError() {
	// Precondition: VERSION file with branch creation enabled
	suite.createTestFilesWithRelease("1.0.0", true)
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", gomock.Any()).Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, fmt.Errorf("branches unavailable"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about checking branch existence
	suite.Error(err)
	suite.Contains(err.Error(), "error checking if branch exists")
}

// TestReleaseCommand_CreateBranchError validates error handling when branch creation fails.
//
// Why: Branch creation can fail due to permissions or ref conflicts. These
// failures must be reported after tag creation to indicate partial success.
//
// What:
//   - Precondition: VERSION file, config with createBranch=true, CreateBranch returns error
//   - Action: Run "release" command
//   - Expected: Command returns error about creating branch
func (suite *ReleaseTestSuite) TestReleaseCommand_CreateBranchError() {
	// Precondition: VERSION file with branch creation enabled
	suite.createTestFilesWithRelease("1.0.0", true)
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", gomock.Any()).Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.0.0").Return(fmt.Errorf("ref locked"))
	vcs.RegisterVCS(mockVCS)

	// Action: Execute release command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release"})

	err := rootCmd.Execute()

	// Expected: Returns error about creating branch
	suite.Error(err)
	suite.Contains(err.Error(), "error creating release branch")
}

// =============================================================================
// RELEASE PUSH TESTS
// =============================================================================

// TestReleasePushCommand_Success validates that the release push command creates
// tag and branch, then pushes both to the remote.
//
// Why: Users need a single command to create a release and push it to remote.
//
// What: Run "release push" with clean repo, verify tag and branch are created and pushed.
func (suite *ReleaseTestSuite) TestReleasePushCommand_Success() {
	// Precondition: VERSION file exists
	suite.createTestFiles("1.0.0")

	// Precondition: VCS is clean and all operations succeed
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.0.0").Return(nil)
	mockVCS.EXPECT().PushTag("v1.0.0").Return(nil)
	mockVCS.EXPECT().PushBranch("release/v1.0.0").Return(nil)

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "push"})

	// Action: Execute release push
	err := rootCmd.Execute()

	// Expected: Tag and branch created and pushed
	suite.Require().NoError(err, "release push should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'")
	suite.Contains(output, "Successfully pushed tag 'v1.0.0'")
	suite.Contains(output, "Successfully pushed branch 'release/v1.0.0'")
}

// TestReleasePushCommand_PushTagError validates error handling when tag push fails.
//
// Why: Push failures must be reported clearly so users know their release is incomplete.
//
// What: Run "release push" where PushTag fails, verify error is returned.
func (suite *ReleaseTestSuite) TestReleasePushCommand_PushTagError() {
	// Precondition: VERSION file exists
	suite.createTestFiles("1.0.0")

	// Precondition: VCS operations succeed until PushTag fails
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.0.0").Return(nil)
	mockVCS.EXPECT().PushTag("v1.0.0").Return(fmt.Errorf("remote rejected"))

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release", "push"})

	// Action: Execute release push
	err := rootCmd.Execute()

	// Expected: Error about pushing tag
	suite.Error(err)
	suite.Contains(err.Error(), "failed to push tag")
}

// TestReleasePushCommand_PushBranchError validates error handling when branch push fails.
//
// Why: Branch push failures must be reported even if tag push succeeded.
//
// What: Run "release push" where PushBranch fails, verify error is returned.
func (suite *ReleaseTestSuite) TestReleasePushCommand_PushBranchError() {
	// Precondition: VERSION file exists
	suite.createTestFiles("1.0.0")

	// Precondition: VCS operations succeed until PushBranch fails
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().BranchExists("release/v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateBranch("release/v1.0.0").Return(nil)
	mockVCS.EXPECT().PushTag("v1.0.0").Return(nil)
	mockVCS.EXPECT().PushBranch("release/v1.0.0").Return(fmt.Errorf("branch exists on remote"))

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"release", "push"})

	// Action: Execute release push
	err := rootCmd.Execute()

	// Expected: Error about pushing branch
	suite.Error(err)
	suite.Contains(err.Error(), "failed to push branch")
}

// TestReleasePushCommand_NoBranch validates that push works with --no-branch flag.
//
// Why: Users may want to push only the tag, not a release branch.
//
// What: Run "release push --no-branch", verify only tag is pushed.
func (suite *ReleaseTestSuite) TestReleasePushCommand_NoBranch() {
	// Precondition: VERSION file exists
	suite.createTestFiles("1.0.0")

	// Precondition: VCS is clean and all operations succeed (no branch operations)
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, nil)
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release 1.0.0").Return(nil)
	mockVCS.EXPECT().PushTag("v1.0.0").Return(nil)
	// No branch operations expected

	vcs.RegisterVCS(mockVCS)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"release", "push", "--no-branch"})

	// Action: Execute release push with no-branch
	err := rootCmd.Execute()

	// Expected: Only tag created and pushed
	suite.Require().NoError(err, "release push should succeed")
	output := buf.String()
	suite.Contains(output, "Successfully created tag 'v1.0.0'")
	suite.Contains(output, "Successfully pushed tag 'v1.0.0'")
	suite.NotContains(output, "branch")
}

// =============================================================================
// TEST SUITE RUNNER
// =============================================================================

// TestReleaseTestSuite runs the release test suite
func TestReleaseTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseTestSuite))
}
