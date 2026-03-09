package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// CITestSuite defines the test suite for ci command tests
type CITestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	tempDir string
	origDir string
}

func TestCISuite(t *testing.T) {
	suite.Run(t, new(CITestSuite))
}

func (suite *CITestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
	var err error
	suite.origDir, err = os.Getwd()
	suite.Require().NoError(err)
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)

	suite.ctrl = gomock.NewController(suite.T())
	suite.resetCICommand()
}

func (suite *CITestSuite) TearDownTest() {
	if suite.origDir != "" {
		_ = os.Chdir(suite.origDir)
	}
	if suite.ctrl != nil {
		suite.ctrl.Finish()
	}
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
}

func (suite *CITestSuite) resetCICommand() {
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
	_ = ciCmd.Flags().Set("format", "")
	_ = ciCmd.Flags().Set("output", "")
	_ = ciCmd.Flags().Set("prefix", "")
}

func (suite *CITestSuite) createVersionFile(ver string) {
	err := os.WriteFile("VERSION", []byte(ver), 0644)
	suite.Require().NoError(err)
}

func (suite *CITestSuite) setupMockVCS() *mock.MockVersionControlSystem {
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(mockVCS)
	return mockVCS
}

// setupMockVCSWithDefaults sets up mock VCS with default expectations for all methods
// that getVCSInfo calls (used by emit.BuildTemplateDataFromVersion)
func (suite *CITestSuite) setupMockVCSWithDefaults() *mock.MockVersionControlSystem {
	mockVCS := suite.setupMockVCS()
	// These are called by emit.getVCSInfo() via BuildTemplateDataFromVersion
	mockVCS.EXPECT().GetVCSIdentifier(40).Return("abc123def456789012345678901234567890dead", nil).AnyTimes()
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc123d", nil).AnyTimes()
	mockVCS.EXPECT().GetBranchName().Return("main", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitDate().Return(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), nil).AnyTimes()
	mockVCS.EXPECT().GetCommitsSinceTag().Return(5, nil).AnyTimes()
	mockVCS.EXPECT().GetLastTagCommit().Return("def456", nil).AnyTimes()
	mockVCS.EXPECT().GetUncommittedChanges().Return(0, nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthor().Return("Test Author", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthorEmail().Return("test@example.com", nil).AnyTimes()
	return mockVCS
}

// ============================================================================
// CORE FUNCTIONALITY
// Tests that verify the primary happy-path behavior of the CI command.
// ============================================================================

// TestRunCI_ShellFormat validates that CI outputs environment variables in shell format.
//
// Why: Shell format is the default and most common output format for CI/CD pipelines.
// CI scripts typically source the output file, so correct shell syntax is critical.
//
// What:
//   - Precondition: VERSION file exists with "1.2.3", mock VCS provides git metadata
//   - Action: Run "output ci --format=shell"
//   - Expected: Output contains properly quoted shell export statements for all version
//     components and git metadata
func (suite *CITestSuite) TestRunCI_ShellFormat() {
	// Precondition: VERSION file and mock VCS with default values
	suite.createVersionFile("1.2.3")
	suite.setupMockVCSWithDefaults()

	// Action: Execute CI command with shell format
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: Shell format uses quoted values with export statements
	output := buf.String()
	suite.Contains(output, `export VERSION="1.2.3"`)
	suite.Contains(output, `export VERSION_MAJOR="1"`)
	suite.Contains(output, `export VERSION_MINOR="2"`)
	suite.Contains(output, `export VERSION_PATCH="3"`)
	suite.Contains(output, `export GIT_SHA="abc123def456789012345678901234567890dead"`)
	suite.Contains(output, `export GIT_SHA_SHORT="abc123d"`)
	suite.Contains(output, `export GIT_BRANCH="main"`)
	suite.Contains(output, `export BUILD_NUMBER="5"`)
	suite.Contains(output, `export DIRTY="false"`)
}

// TestBuildCIVariables validates that buildCIVariables correctly extracts all version components.
//
// Why: The buildCIVariables function is the core data transformation that converts a
// Version struct into CI-friendly string variables. Incorrect mapping breaks all CI output.
//
// What:
//   - Precondition: Version with major=1, minor=2, patch=3, pre-release="beta",
//     metadata="build123", prefix="v"
//   - Action: Call buildCIVariables with this version
//   - Expected: All fields correctly extracted as strings with proper formatting
func TestBuildCIVariables(t *testing.T) {
	// Precondition: Create temp dir and VERSION file
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)
	_ = os.WriteFile("VERSION", []byte("1.2.3"), 0644)

	v := &version.Version{
		Major:         1,
		Minor:         2,
		Patch:         3,
		PreRelease:    "beta",
		BuildMetadata: "build123",
		Prefix:        "v",
	}

	// Unregister VCS so we test without git
	vcs.UnregisterVCS("git")
	defer vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	// Action: Build CI variables
	vars := buildCIVariables(v)

	// Expected: All fields correctly mapped
	if vars.Version != "v1.2.3-beta+build123" {
		t.Errorf("expected 'v1.2.3-beta+build123', got '%s'", vars.Version)
	}
	if vars.VersionSemver != "1.2.3-beta+build123" {
		t.Errorf("expected '1.2.3-beta+build123', got '%s'", vars.VersionSemver)
	}
	if vars.VersionCore != "1.2.3" {
		t.Errorf("expected '1.2.3', got '%s'", vars.VersionCore)
	}
	if vars.Major != "1" {
		t.Errorf("expected '1', got '%s'", vars.Major)
	}
	if vars.Minor != "2" {
		t.Errorf("expected '2', got '%s'", vars.Minor)
	}
	if vars.Patch != "3" {
		t.Errorf("expected '3', got '%s'", vars.Patch)
	}
	if vars.PreRelease != "beta" {
		t.Errorf("expected 'beta', got '%s'", vars.PreRelease)
	}
	if vars.Metadata != "build123" {
		t.Errorf("expected 'build123', got '%s'", vars.Metadata)
	}
}

// ============================================================================
// KEY VARIATIONS
// Tests that verify important alternate flows and configuration options.
// ============================================================================

// TestRunCI_GitHubFormat validates that CI outputs variables in GitHub Actions format.
//
// Why: GitHub Actions requires a specific format (name=value) for setting output
// variables. Using the wrong format would break GitHub CI workflows.
//
// What:
//   - Precondition: VERSION file exists with "2.0.0", mock VCS configured
//   - Action: Run "output ci --format=github"
//   - Expected: Output uses GitHub Actions format (name=value without quotes)
func (suite *CITestSuite) TestRunCI_GitHubFormat() {
	// Precondition: VERSION file and mock VCS
	suite.createVersionFile("2.0.0")
	suite.setupMockVCSWithDefaults()

	// Action: Execute CI command with GitHub format
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=github"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: GitHub format uses name=value
	output := buf.String()
	suite.Contains(output, "VERSION=2.0.0")
}

// TestRunCI_WithPrefix validates that custom variable prefixes are applied correctly.
//
// Why: In multi-project CI environments, variable name collisions can occur.
// Custom prefixes allow isolation of variables between projects.
//
// What:
//   - Precondition: VERSION file exists with "1.0.0", mock VCS configured
//   - Action: Run "output ci --format=shell --prefix=MYAPP_"
//   - Expected: All output variables use the custom prefix
func (suite *CITestSuite) TestRunCI_WithPrefix() {
	// Precondition: VERSION file and mock VCS
	suite.createVersionFile("1.0.0")
	suite.setupMockVCSWithDefaults()

	// Action: Execute CI command with custom prefix
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell", "--prefix=MYAPP_"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: Variables use custom prefix
	output := buf.String()
	suite.Contains(output, `export MYAPP_VERSION="1.0.0"`)
	suite.Contains(output, `export MYAPP_VERSION_MAJOR="1"`)
}

// TestRunCI_WriteToFile validates that CI output can be written to a file.
//
// Why: CI pipelines often need to persist variables to a file for sourcing in
// later steps or jobs. File output is essential for multi-stage pipelines.
//
// What:
//   - Precondition: VERSION file exists with "1.0.0", mock VCS configured
//   - Action: Run "output ci --format=shell --output=ci-vars.env"
//   - Expected: File is created with correct content
func (suite *CITestSuite) TestRunCI_WriteToFile() {
	// Precondition: VERSION file and mock VCS
	suite.createVersionFile("1.0.0")
	suite.setupMockVCSWithDefaults()

	outputFile := filepath.Join(suite.tempDir, "ci-vars.env")

	// Action: Execute CI command with file output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell", "--output=" + outputFile})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: File was created with correct content
	content, err := os.ReadFile(outputFile)
	suite.NoError(err)
	suite.Contains(string(content), `export VERSION="1.0.0"`)
}

// TestRunCI_DirtyRepo validates that dirty repository state is correctly reported.
//
// Why: Knowing if a build was created from a dirty working tree is critical for
// release validation. CI pipelines may need to fail or warn on dirty builds.
//
// What:
//   - Precondition: VERSION file exists, VCS reports 3 uncommitted changes
//   - Action: Run "output ci --format=shell"
//   - Expected: DIRTY variable is set to "true"
func (suite *CITestSuite) TestRunCI_DirtyRepo() {
	// Precondition: VERSION file and mock VCS with dirty state
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	// Set up defaults but override uncommitted changes to show dirty
	mockVCS.EXPECT().GetVCSIdentifier(40).Return("abc123def456789012345678901234567890dead", nil).AnyTimes()
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc123d", nil).AnyTimes()
	mockVCS.EXPECT().GetBranchName().Return("main", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitDate().Return(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), nil).AnyTimes()
	mockVCS.EXPECT().GetCommitsSinceTag().Return(0, nil).AnyTimes()
	mockVCS.EXPECT().GetLastTagCommit().Return("def456", nil).AnyTimes()
	mockVCS.EXPECT().GetUncommittedChanges().Return(3, nil).AnyTimes() // 3 dirty files
	mockVCS.EXPECT().GetCommitAuthor().Return("Test Author", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthorEmail().Return("test@example.com", nil).AnyTimes()

	// Action: Execute CI command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: Dirty flag is true
	output := buf.String()
	suite.Contains(output, `export DIRTY="true"`)
}

// TestBuildCIVariables_NoPreReleaseOrMetadata validates clean version handling.
//
// Why: Production releases typically have no pre-release or metadata suffixes.
// These fields should be empty strings, not nil or placeholder values.
//
// What:
//   - Precondition: Version with only major=1, minor=0, patch=0 (no extras)
//   - Action: Call buildCIVariables with this version
//   - Expected: PreRelease and Metadata fields are empty strings
func TestBuildCIVariables_NoPreReleaseOrMetadata(t *testing.T) {
	// Precondition: Create temp dir and VERSION file
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)
	_ = os.WriteFile("VERSION", []byte("1.0.0"), 0644)

	v := &version.Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}

	vcs.UnregisterVCS("git")
	defer vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	// Action: Build CI variables
	vars := buildCIVariables(v)

	// Expected: Clean version string without suffixes
	if vars.Version != "1.0.0" {
		t.Errorf("expected '1.0.0', got '%s'", vars.Version)
	}
	if vars.VersionSemver != "1.0.0" {
		t.Errorf("expected '1.0.0', got '%s'", vars.VersionSemver)
	}
	if vars.PreRelease != "" {
		t.Errorf("expected empty pre-release, got '%s'", vars.PreRelease)
	}
	if vars.Metadata != "" {
		t.Errorf("expected empty metadata, got '%s'", vars.Metadata)
	}
}

// TestBuildCIVariables_WithVCS validates that VCS metadata is correctly extracted.
//
// Why: CI pipelines need git metadata (SHA, branch, build number) for traceability,
// artifact tagging, and deployment decisions.
//
// What:
//   - Precondition: Mock VCS returning SHA, branch, and commit count
//   - Action: Call buildCIVariables with a version
//   - Expected: Git fields populated with VCS data, dirty flag reflects uncommitted changes
func TestBuildCIVariables_WithVCS(t *testing.T) {
	// Precondition: Create temp dir and VERSION file
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)
	_ = os.WriteFile("VERSION", []byte("1.0.0"), 0644)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(tempDir, nil).AnyTimes()
	// All methods called by emit.getVCSInfo() via BuildTemplateDataFromVersion
	mockVCS.EXPECT().GetVCSIdentifier(40).Return("fullsha123456789012345678901234567890dead", nil).AnyTimes()
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("fullsha", nil).AnyTimes()
	mockVCS.EXPECT().GetBranchName().Return("feature/test", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitDate().Return(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), nil).AnyTimes()
	mockVCS.EXPECT().GetCommitsSinceTag().Return(10, nil).AnyTimes()
	mockVCS.EXPECT().GetLastTagCommit().Return("def456", nil).AnyTimes()
	mockVCS.EXPECT().GetUncommittedChanges().Return(2, nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthor().Return("Test Author", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthorEmail().Return("test@example.com", nil).AnyTimes()

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	v := &version.Version{Major: 1, Minor: 0, Patch: 0}

	// Action: Build CI variables
	vars := buildCIVariables(v)

	// Expected: VCS fields populated correctly
	if vars.GitSHA != "fullsha123456789012345678901234567890dead" {
		t.Errorf("expected full SHA, got '%s'", vars.GitSHA)
	}
	if vars.GitSHAShort != "fullsha" {
		t.Errorf("expected short SHA, got '%s'", vars.GitSHAShort)
	}
	if vars.GitBranch != "feature/test" {
		t.Errorf("expected 'feature/test', got '%s'", vars.GitBranch)
	}
	if vars.BuildNumber != "10" {
		t.Errorf("expected '10', got '%s'", vars.BuildNumber)
	}
	if vars.Dirty != "true" {
		t.Errorf("expected 'true', got '%s'", vars.Dirty)
	}
}

// ============================================================================
// ERROR HANDLING
// Tests that verify expected failure modes and error responses.
// ============================================================================

// TestRunCI_InvalidFormat validates that invalid format names are rejected.
//
// Why: Users may typo the format name or use an unsupported format.
// Clear error messages help users fix configuration issues quickly.
//
// What:
//   - Precondition: VERSION file exists
//   - Action: Run "output ci --format=invalid"
//   - Expected: Command returns error containing "invalid format"
func (suite *CITestSuite) TestRunCI_InvalidFormat() {
	// Precondition: VERSION file exists
	suite.createVersionFile("1.0.0")

	// Action: Execute CI command with invalid format
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=invalid"})

	err := rootCmd.Execute()

	// Expected: Error returned with descriptive message
	suite.Error(err)
	suite.Contains(err.Error(), "invalid format")
}

// ============================================================================
// EDGE CASES
// Tests that verify behavior at boundary conditions and unusual states.
// ============================================================================

// TestRunCI_NoVersionFile_CreatesDefault validates auto-creation of VERSION file.
//
// Why: New projects or CI environments may not have a VERSION file yet.
// Auto-creation with a sensible default prevents CI failures on first run.
//
// What:
//   - Precondition: No VERSION file exists, mock VCS configured
//   - Action: Run "output ci --format=shell"
//   - Expected: Command succeeds, uses default version "v0.0.1"
func (suite *CITestSuite) TestRunCI_NoVersionFile_CreatesDefault() {
	// Precondition: Don't create VERSION file - version.Load() auto-creates one with v0.0.1
	suite.setupMockVCSWithDefaults()

	// Action: Execute CI command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: Should have created a default version file with prefix
	output := buf.String()
	suite.Contains(output, `export VERSION="v0.0.1"`)
	suite.Contains(output, `export VERSION_CORE="0.0.1"`)
}

// TestRunCI_NoVCS validates graceful degradation when no VCS is available.
//
// Why: Some CI environments may not have git installed or may build from
// tarballs without git history. The command should still provide version info.
//
// What:
//   - Precondition: VERSION file exists, no VCS registered
//   - Action: Run "output ci --format=shell"
//   - Expected: Version fields populated, git fields empty
func (suite *CITestSuite) TestRunCI_NoVCS() {
	// Precondition: VERSION file but no VCS
	suite.createVersionFile("1.0.0")
	vcs.UnregisterVCS("git")

	// Action: Execute CI command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: Version info present, git info empty
	output := buf.String()
	suite.Contains(output, `export VERSION="1.0.0"`)
	suite.Contains(output, `export GIT_SHA=""`)
}

// TestRunCI_AutoDetect_NoCI validates behavior when no CI environment is detected.
//
// Why: When run locally without explicit format, the command should detect
// there's no CI environment and provide helpful guidance to the user.
//
// What:
//   - Precondition: VERSION file exists, no CI environment variables set
//   - Action: Run "output ci" (no format flag)
//   - Expected: Output indicates no CI environment detected
func (suite *CITestSuite) TestRunCI_AutoDetect_NoCI() {
	// Precondition: VERSION file, mock VCS, no CI env vars
	suite.createVersionFile("1.0.0")
	suite.setupMockVCSWithDefaults()

	// Clear CI env vars to ensure no CI is detected
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITLAB_CI")
	os.Unsetenv("CIRCLECI")

	// Action: Execute CI command without format flag
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "ci"})

	err := rootCmd.Execute()
	suite.NoError(err)

	// Expected: Informative message about no CI detected
	output := buf.String()
	suite.Contains(output, "No CI environment detected")
}

// ============================================================================
// MINUTIAE
// Tests that verify obscure scenarios and minor behavioral details.
// ============================================================================

// TestBuildCIVariables_OnlyPreRelease validates versions with pre-release but no metadata.
//
// Why: Pre-release versions (alpha, beta, rc) are common during development.
// The version string should include the pre-release suffix but not have a
// trailing "+" when there's no metadata.
//
// What:
//   - Precondition: Version with pre-release="alpha" but no metadata
//   - Action: Call buildCIVariables with this version
//   - Expected: Version ends with "-alpha" and has no "+" character
func TestBuildCIVariables_OnlyPreRelease(t *testing.T) {
	// Precondition: Create temp dir and VERSION file
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)
	_ = os.WriteFile("VERSION", []byte("1.0.0-alpha"), 0644)

	v := &version.Version{
		Major:      1,
		Minor:      0,
		Patch:      0,
		PreRelease: "alpha",
	}

	vcs.UnregisterVCS("git")
	defer vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	// Action: Build CI variables
	vars := buildCIVariables(v)

	// Expected: Version has pre-release suffix, no metadata separator
	if !strings.HasSuffix(vars.Version, "-alpha") {
		t.Errorf("expected version to end with '-alpha', got '%s'", vars.Version)
	}
	if strings.Contains(vars.Version, "+") {
		t.Errorf("expected version to not contain '+', got '%s'", vars.Version)
	}
}
