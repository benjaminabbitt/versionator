package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// InitTestSuite defines the test suite for init command tests.
// The init command creates VERSION files and optional configuration files
// for projects adopting versionator's version management approach.
type InitTestSuite struct {
	suite.Suite
	tempDir string
	origDir string
}

// SetupTest runs before each test
func (suite *InitTestSuite) SetupTest() {
	// Create a temporary directory for testing
	suite.tempDir = suite.T().TempDir()
	var err error
	suite.origDir, err = os.Getwd()
	suite.Require().NoError(err)
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)
}

// TearDownTest runs after each test
func (suite *InitTestSuite) TearDownTest() {
	// Restore original directory
	if suite.origDir != "" {
		_ = os.Chdir(suite.origDir)
	}

	// Reset command state
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Reset flags to defaults
	initVersion = "0.0.1"
	initPrefix = ""
	initWithConfig = false
	initForce = false
}

// =============================================================================
// CORE FUNCTIONALITY - Happy path tests for primary init behaviors
// =============================================================================

// TestInitCommand_Default_CreatesVersionFile validates that the init command
// creates a VERSION file with the default version when run without arguments.
//
// Why: This is the primary use case for init - bootstrapping a new project with
// version tracking. Users expect a sensible default (0.0.1) when starting fresh.
//
// What: Run init with no arguments in an empty directory; expect VERSION file
// with "0.0.1" content and no config file created.
func (suite *InitTestSuite) TestInitCommand_Default_CreatesVersionFile() {
	// Precondition: Empty directory (handled by SetupTest)

	// Action: Run init command with no arguments
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()

	// Expected: Command succeeds, VERSION file created with default version
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("0.0.1", strings.TrimSpace(string(content)), "VERSION should contain '0.0.1'")

	// Expected: Config file should NOT be created by default
	_, err = os.Stat(".versionator.yaml")
	suite.True(os.IsNotExist(err), "Config file should not exist without --config flag")
}

// TestInitCommand_WithConfig_CreatesBothFiles validates that the --config flag
// creates both VERSION and .versionator.yaml files.
//
// Why: Some users want a config file to customize versionator behavior. The
// --config flag provides an opt-in way to bootstrap configuration.
//
// What: Run init with --config flag; expect both VERSION and config files created.
func (suite *InitTestSuite) TestInitCommand_WithConfig_CreatesBothFiles() {
	// Precondition: Empty directory (handled by SetupTest)

	// Action: Run init with --config flag
	rootCmd.SetArgs([]string{"init", "--config"})
	err := rootCmd.Execute()

	// Expected: Command succeeds
	suite.Require().NoError(err, "init command should succeed")

	// Expected: VERSION file exists
	_, err = os.Stat("VERSION")
	suite.Require().NoError(err, "VERSION file should exist")

	// Expected: Config file exists
	_, err = os.Stat(".versionator.yaml")
	suite.Require().NoError(err, "Config file should exist with --config flag")
}

// =============================================================================
// KEY VARIATIONS - Important alternate flows for init customization
// =============================================================================

// TestInitCommand_WithVersion_UsesSpecifiedVersion validates that the --version
// flag allows users to specify a custom starting version.
//
// Why: Projects may adopt versionator after already having releases. They need
// to initialize with their current version rather than starting at 0.0.1.
//
// What: Run init with --version 1.2.3; expect VERSION file with "1.2.3".
func (suite *InitTestSuite) TestInitCommand_WithVersion_UsesSpecifiedVersion() {
	// Precondition: Empty directory (handled by SetupTest)

	// Action: Run init with custom version
	rootCmd.SetArgs([]string{"init", "--version", "1.2.3"})
	err := rootCmd.Execute()

	// Expected: Command succeeds with specified version in file
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.2.3", strings.TrimSpace(string(content)), "VERSION should contain '1.2.3'")
}

// TestInitCommand_WithPrefix_PrependsToVersion validates that the --prefix flag
// adds a prefix (typically "v") to the version string.
//
// Why: Many projects use "v" prefixed tags (e.g., v1.0.0). The prefix flag
// allows version strings to match the project's tagging convention.
//
// What: Run init with --prefix v; expect VERSION file with "v0.0.1".
func (suite *InitTestSuite) TestInitCommand_WithPrefix_PrependsToVersion() {
	// Precondition: Empty directory (handled by SetupTest)

	// Action: Run init with prefix
	rootCmd.SetArgs([]string{"init", "--prefix", "v"})
	err := rootCmd.Execute()

	// Expected: Command succeeds with prefixed version
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v0.0.1", strings.TrimSpace(string(content)), "VERSION should contain 'v0.0.1'")
}

// TestInitCommand_WithVersionAndPrefix_CombinesBoth validates that version and
// prefix flags work together correctly.
//
// Why: Users adopting versionator with existing versioned releases that use
// prefixed tags need both options to properly initialize.
//
// What: Run init with --version 2.0.0 --prefix V; expect VERSION with "V2.0.0".
func (suite *InitTestSuite) TestInitCommand_WithVersionAndPrefix_CombinesBoth() {
	// Precondition: Empty directory (handled by SetupTest)

	// Action: Run init with both version and prefix
	rootCmd.SetArgs([]string{"init", "--version", "2.0.0", "--prefix", "V"})
	err := rootCmd.Execute()

	// Expected: Command succeeds with both applied
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("V2.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'V2.0.0'")
}

// TestInitCommand_ForceOverwrite_ReplacesExistingVersion validates that --force
// allows overwriting an existing VERSION file.
//
// Why: Users may need to reinitialize a project (e.g., after corruption or to
// reset to a specific version). The force flag provides this escape hatch.
//
// What: Create existing VERSION with "1.0.0", run init --force --version 2.0.0;
// expect VERSION overwritten to "2.0.0".
func (suite *InitTestSuite) TestInitCommand_ForceOverwrite_ReplacesExistingVersion() {
	// Precondition: VERSION file already exists
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)

	// Action: Run init with --force and new version
	rootCmd.SetArgs([]string{"init", "--force", "--version", "2.0.0"})
	err = rootCmd.Execute()

	// Expected: Command succeeds and file is overwritten
	suite.Require().NoError(err, "init --force should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err)
	suite.Equal("2.0.0", strings.TrimSpace(string(content)), "VERSION should be overwritten to '2.0.0'")
}

// TestInitCommand_ForceOverwriteConfig_ReplacesBothFiles validates that --force
// with --config overwrites both VERSION and config files.
//
// Why: When reinitializing, users may want to reset both files to defaults.
// This ensures --force applies to all created files.
//
// What: Create existing VERSION and config files, run init --force --config;
// expect both files overwritten successfully.
func (suite *InitTestSuite) TestInitCommand_ForceOverwriteConfig_ReplacesBothFiles() {
	// Precondition: Both files already exist
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)
	err = os.WriteFile(".versionator.yaml", []byte("prefix: old\n"), 0644)
	suite.Require().NoError(err)

	// Action: Run init with --force and --config
	rootCmd.SetArgs([]string{"init", "--force", "--config"})
	err = rootCmd.Execute()

	// Expected: Command succeeds with both files overwritten
	suite.Require().NoError(err, "init --force --config should succeed")

	_, err = os.Stat("VERSION")
	suite.Require().NoError(err)
	_, err = os.Stat(".versionator.yaml")
	suite.Require().NoError(err)
}

// =============================================================================
// ERROR HANDLING - Expected failure modes for invalid inputs or states
// =============================================================================

// TestInitCommand_InvalidPrefixRejected validates that only valid prefixes
// (v or V) are accepted.
//
// Why: Arbitrary prefixes could lead to inconsistent version formats across
// the ecosystem. Restricting to v/V maintains semver convention compatibility.
//
// What: Run init with --prefix "release-"; expect error mentioning valid prefixes.
func (suite *InitTestSuite) TestInitCommand_InvalidPrefixRejected() {
	// Precondition: Empty directory (handled by SetupTest)

	// Action: Run init with invalid prefix
	rootCmd.SetArgs([]string{"init", "--prefix", "release-"})
	err := rootCmd.Execute()

	// Expected: Command fails with descriptive error
	suite.Error(err, "init command should reject invalid prefix")
	suite.Contains(err.Error(), "only 'v' or 'V' allowed", "error should mention valid prefixes")
}

// TestInitCommand_FailsIfVersionExists validates that init refuses to overwrite
// an existing VERSION file without --force.
//
// Why: Accidentally overwriting a VERSION file could cause version tracking
// issues. Requiring --force makes overwrites intentional.
//
// What: Create existing VERSION file, run init without --force; expect error
// indicating file already exists.
func (suite *InitTestSuite) TestInitCommand_FailsIfVersionExists() {
	// Precondition: VERSION file already exists
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)

	// Action: Run init without --force
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()

	// Expected: Command fails with "already exists" error
	suite.Error(err, "init should fail if VERSION exists")
	suite.Contains(err.Error(), "already exists")
}

// TestInitCommand_ConfigFailsIfExists validates that init --config refuses to
// overwrite an existing config file without --force.
//
// Why: Config files may contain customizations. Accidentally overwriting them
// could lose user settings. Requiring --force makes overwrites intentional.
//
// What: Create existing config file, run init --config without --force; expect
// error indicating file already exists.
func (suite *InitTestSuite) TestInitCommand_ConfigFailsIfExists() {
	// Precondition: Config file already exists
	err := os.WriteFile(".versionator.yaml", []byte("prefix: v\n"), 0644)
	suite.Require().NoError(err)

	// Action: Run init --config without --force
	rootCmd.SetArgs([]string{"init", "--config"})
	err = rootCmd.Execute()

	// Expected: Command fails with "already exists" error
	suite.Error(err, "init --config should fail if config exists")
	suite.Contains(err.Error(), "already exists")
}

// =============================================================================
// INIT HOOK TESTS
// =============================================================================

// TestInitHookCommand_InstallsHook validates that init hook installs the
// post-commit hook script.
//
// Why: Users want automatic version bumping on commit.
//
// What: Run "init hook" in a git repo, verify hook file is created.
func (suite *InitTestSuite) TestInitHookCommand_InstallsHook() {
	// Precondition: Create mock VCS that returns hooks path
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	hooksDir := filepath.Join(suite.tempDir, ".git", "hooks")
	err := os.MkdirAll(hooksDir, 0755)
	suite.Require().NoError(err)

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().GetHooksPath().Return(hooksDir, nil)

	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	// Action: Execute "init hook"
	rootCmd.SetArgs([]string{"init", "hook"})
	err = rootCmd.Execute()

	// Expected: Hook file is created
	suite.Require().NoError(err, "init hook should succeed")

	hookPath := filepath.Join(hooksDir, "post-commit")
	_, err = os.Stat(hookPath)
	suite.Require().NoError(err, "post-commit hook should exist")

	// Verify hook content
	content, err := os.ReadFile(hookPath)
	suite.Require().NoError(err)
	suite.Contains(string(content), "versionator bump")
}

// TestInitHookCommand_UninstallsHook validates that init hook --uninstall
// removes the versionator hook.
//
// Why: Users need to remove the hook if they no longer want automatic bumping.
//
// What: Install hook, then run "init hook --uninstall", verify hook is removed.
func (suite *InitTestSuite) TestInitHookCommand_UninstallsHook() {
	// Precondition: Create mock VCS and hook file
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	hooksDir := filepath.Join(suite.tempDir, ".git", "hooks")
	err := os.MkdirAll(hooksDir, 0755)
	suite.Require().NoError(err)

	// Create hook file with versionator content
	hookPath := filepath.Join(hooksDir, "post-commit")
	err = os.WriteFile(hookPath, []byte(postCommitHookScript), 0755)
	suite.Require().NoError(err)

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().GetHooksPath().Return(hooksDir, nil)

	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	// Reset uninstall flag
	hookUninstall = true
	defer func() { hookUninstall = false }()

	// Action: Execute "init hook --uninstall"
	rootCmd.SetArgs([]string{"init", "hook", "--uninstall"})
	err = rootCmd.Execute()

	// Expected: Hook file is removed
	suite.Require().NoError(err, "init hook --uninstall should succeed")

	_, err = os.Stat(hookPath)
	suite.True(os.IsNotExist(err), "post-commit hook should be removed")
}

// TestInitHookCommand_NoVCS_ReturnsError validates that init hook fails
// gracefully when not in a git repository.
//
// Why: Users need clear feedback when running outside a VCS repository.
//
// What: Run "init hook" without VCS, verify error is returned.
func (suite *InitTestSuite) TestInitHookCommand_NoVCS_ReturnsError() {
	// Precondition: Unregister VCS
	vcs.UnregisterVCS("git")
	defer vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	// Action: Execute "init hook"
	rootCmd.SetArgs([]string{"init", "hook"})
	err := rootCmd.Execute()

	// Expected: Error about not being in a git repository
	suite.Error(err)
	suite.Contains(err.Error(), "not in a git repository")
}

// TestInitHookCommand_HookExistsNoForce_ReturnsError validates that init hook
// fails when hook already exists and --force is not used.
//
// Why: Prevents accidental overwriting of existing hooks.
//
// What: Create existing hook, run "init hook" without --force, verify error.
func (suite *InitTestSuite) TestInitHookCommand_HookExistsNoForce_ReturnsError() {
	// Precondition: Create mock VCS and existing hook file
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	hooksDir := filepath.Join(suite.tempDir, ".git", "hooks")
	err := os.MkdirAll(hooksDir, 0755)
	suite.Require().NoError(err)

	// Create existing hook file
	hookPath := filepath.Join(hooksDir, "post-commit")
	err = os.WriteFile(hookPath, []byte("#!/bin/sh\necho existing"), 0755)
	suite.Require().NoError(err)

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	mockVCS.EXPECT().GetHooksPath().Return(hooksDir, nil)

	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	// Action: Execute "init hook" without --force
	rootCmd.SetArgs([]string{"init", "hook"})
	err = rootCmd.Execute()

	// Expected: Error about existing hook
	suite.Error(err)
	suite.Contains(err.Error(), "already exists")
}

// =============================================================================
// EDGE CASES - Boundary conditions (none currently identified)
// =============================================================================

// =============================================================================
// MINUTIAE - Obscure scenarios (none currently identified)
// =============================================================================

// TestInitTestSuite runs the init test suite
func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}
