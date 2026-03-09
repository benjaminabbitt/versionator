package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/commitparser"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// BUMP COMMAND TEST SUITE
// =============================================================================

// BumpTestSuite defines the test suite for bump command tests.
// It provides a controlled environment with temporary directories and mock VCS
// to test the bump command's behavior in isolation.
type BumpTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	tempDir string
	origDir string
}

func TestBumpSuite(t *testing.T) {
	suite.Run(t, new(BumpTestSuite))
}

func (suite *BumpTestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
	var err error
	suite.origDir, err = os.Getwd()
	suite.Require().NoError(err)
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)

	suite.ctrl = gomock.NewController(suite.T())
	suite.resetBumpCommand()
}

func (suite *BumpTestSuite) TearDownTest() {
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

func (suite *BumpTestSuite) resetBumpCommand() {
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
	_ = bumpCmd.Flags().Set("dry-run", "false")
	_ = bumpCmd.Flags().Set("no-amend", "false")
	_ = bumpCmd.Flags().Set("mode", "all")
}

func (suite *BumpTestSuite) createVersionFile(ver string) {
	err := os.WriteFile("VERSION", []byte(ver), 0644)
	suite.Require().NoError(err)
}

func (suite *BumpTestSuite) readVersionFile() string {
	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err)
	return strings.TrimSpace(string(content))
}

func (suite *BumpTestSuite) setupMockVCS() *mock.MockVersionControlSystem {
	mockVCS := mock.NewMockVersionControlSystem(suite.ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(suite.tempDir, nil).AnyTimes()
	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(mockVCS)
	return mockVCS
}

// =============================================================================
// CORE FUNCTIONALITY - Happy Path Tests
// =============================================================================

// TestRunBump_ActualBump_WithAmend validates the primary use case of the bump
// command: detecting version changes from commits and amending the last commit
// with the updated VERSION file.
//
// Why: This is the default behavior users expect when running 'bump' - automatic
// version detection and seamless integration with git workflow via amend.
//
// What: Given a VERSION file at 1.0.0 and a fix commit, when bump runs with
// default flags, then VERSION is updated to 1.0.1 and the commit is amended.
func (suite *BumpTestSuite) TestRunBump_ActualBump_WithAmend() {
	// Precondition: VERSION file exists with initial version, VCS configured
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"fix: bug fix",
	}, nil)
	mockVCS.EXPECT().AmendCommit([]string{"VERSION"}).Return(nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command with default settings
	err := rootCmd.Execute()

	// Expected: No error, version bumped, commit amended
	suite.NoError(err)
	output := buf.String()
	suite.Contains(output, "Version bumped from 1.0.0 to 1.0.1")
	suite.Contains(output, "Amended last commit")
}

// TestRunBump_ActualBump_WithNoAmend validates that the --no-amend flag
// correctly updates the VERSION file without modifying git history.
//
// Why: Users may want to update version without amending, especially in CI
// pipelines or when they want to create a separate version commit.
//
// What: Given a VERSION file and fix commit, when bump runs with --no-amend,
// then VERSION is updated but git history remains unchanged.
func (suite *BumpTestSuite) TestRunBump_ActualBump_WithNoAmend() {
	// Precondition: VERSION file exists, VCS configured
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"fix: bug fix",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump", "--no-amend"})

	// Action: Execute bump command with --no-amend flag
	err := rootCmd.Execute()

	// Expected: No error, version bumped, file updated
	suite.NoError(err)
	output := buf.String()
	suite.Contains(output, "Version bumped from 1.0.0 to 1.0.1")
	suite.Equal("1.0.1", suite.readVersionFile())
}

// TestMakeLevelCmd_MajorIncrement validates that 'bump major' correctly
// increments the major version and resets minor and patch to zero.
//
// Why: Major version bumps signal breaking changes; they must reset lower
// segments to maintain semver semantics.
//
// What: Given VERSION 1.2.3, when 'bump major' runs, then VERSION becomes 2.0.0.
func (suite *BumpTestSuite) TestMakeLevelCmd_MajorIncrement() {
	// Precondition: VERSION file exists with specific version
	suite.createVersionFile("1.2.3")
	rootCmd.SetArgs([]string{"bump", "major"})

	// Action: Execute bump major command
	err := rootCmd.Execute()

	// Expected: Major incremented, minor and patch reset
	suite.NoError(err)
	suite.Equal("2.0.0", suite.readVersionFile())
}

// TestMakeLevelCmd_MinorIncrement validates that 'bump minor' correctly
// increments the minor version and resets patch to zero.
//
// Why: Minor version bumps signal new features; they must reset patch
// to maintain semver semantics.
//
// What: Given VERSION 1.2.3, when 'bump minor' runs, then VERSION becomes 1.3.0.
func (suite *BumpTestSuite) TestMakeLevelCmd_MinorIncrement() {
	// Precondition: VERSION file exists with specific version
	suite.createVersionFile("1.2.3")
	rootCmd.SetArgs([]string{"bump", "minor"})

	// Action: Execute bump minor command
	err := rootCmd.Execute()

	// Expected: Minor incremented, patch reset
	suite.NoError(err)
	suite.Equal("1.3.0", suite.readVersionFile())
}

// TestMakeLevelCmd_PatchIncrement validates that 'bump patch' correctly
// increments only the patch version.
//
// Why: Patch version bumps signal bug fixes; only patch should change.
//
// What: Given VERSION 1.2.3, when 'bump patch' runs, then VERSION becomes 1.2.4.
func (suite *BumpTestSuite) TestMakeLevelCmd_PatchIncrement() {
	// Precondition: VERSION file exists with specific version
	suite.createVersionFile("1.2.3")
	rootCmd.SetArgs([]string{"bump", "patch"})

	// Action: Execute bump patch command
	err := rootCmd.Execute()

	// Expected: Only patch incremented
	suite.NoError(err)
	suite.Equal("1.2.4", suite.readVersionFile())
}

// =============================================================================
// KEY VARIATIONS - Important Alternate Flows
// =============================================================================

// TestRunBump_DryRun_PatchBump validates that --dry-run correctly reports
// what would happen without modifying files.
//
// Why: Users need to preview version changes before committing, especially
// in CI pipelines to avoid unintended changes.
//
// What: Given a fix commit, when bump runs with --dry-run, then output shows
// the detected bump level and proposed version without file modification.
func (suite *BumpTestSuite) TestRunBump_DryRun_PatchBump() {
	// Precondition: VERSION file exists, VCS configured with fix commit
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"fix: bug fix",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump", "--dry-run"})

	// Action: Execute bump command with --dry-run
	err := rootCmd.Execute()

	// Expected: No error, output shows proposed changes
	suite.NoError(err)
	output := buf.String()
	suite.Contains(output, "Detected bump level: patch")
	suite.Contains(output, "Would bump from 1.0.0 to 1.0.1")
}

// TestRunBump_DryRun_MinorBump validates that minor bumps are correctly
// detected from feature commits.
//
// Why: Feature commits (feat:) should trigger minor version bumps per
// conventional commits specification.
//
// What: Given a feat commit, when bump runs with --dry-run, then output
// shows minor bump level.
func (suite *BumpTestSuite) TestRunBump_DryRun_MinorBump() {
	// Precondition: VERSION file exists, VCS configured with feat commit
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"feat: new feature",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump", "--dry-run"})

	// Action: Execute bump command with --dry-run
	err := rootCmd.Execute()

	// Expected: Minor bump detected
	suite.NoError(err)
	output := buf.String()
	suite.Contains(output, "Detected bump level: minor")
	suite.Contains(output, "Would bump from 1.0.0 to 1.1.0")
}

// TestRunBump_DryRun_MajorBump validates that major bumps are correctly
// detected from breaking change commits.
//
// Why: Breaking changes (feat!: or BREAKING CHANGE:) should trigger major
// version bumps per conventional commits specification.
//
// What: Given a breaking change commit, when bump runs with --dry-run,
// then output shows major bump level.
func (suite *BumpTestSuite) TestRunBump_DryRun_MajorBump() {
	// Precondition: VERSION file exists, VCS configured with breaking change
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"feat!: breaking change",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump", "--dry-run"})

	// Action: Execute bump command with --dry-run
	err := rootCmd.Execute()

	// Expected: Major bump detected
	suite.NoError(err)
	output := buf.String()
	suite.Contains(output, "Detected bump level: major")
	suite.Contains(output, "Would bump from 1.0.0 to 2.0.0")
}

// TestRunBump_SemverMarker validates that +semver: markers in commit messages
// are correctly parsed and used for version bumps.
//
// Why: Semver markers provide explicit control over version bumps when
// conventional commit prefixes don't apply or need to be overridden.
//
// What: Given a commit with +semver:minor marker, when bump runs, then
// minor bump is detected regardless of commit prefix.
func (suite *BumpTestSuite) TestRunBump_SemverMarker() {
	// Precondition: VERSION file exists, VCS configured with semver marker
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"update something +semver:minor",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump", "--dry-run"})

	// Action: Execute bump command
	err := rootCmd.Execute()

	// Expected: Minor bump detected from semver marker
	suite.NoError(err)
	output := buf.String()
	suite.Contains(output, "Detected bump level: minor")
}

// TestRunBump_SemverMode validates that --mode=semver restricts parsing
// to only semver markers, ignoring conventional commit prefixes.
//
// Why: Some projects use non-standard commit messages but still want
// semver markers for version control.
//
// What: Given a feat commit (normally minor) with --mode=semver, when bump
// runs, then no bump is detected because semver markers are absent.
func (suite *BumpTestSuite) TestRunBump_SemverMode() {
	// Precondition: VERSION file exists, VCS configured with feat commit
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"feat: new feature", // Would be minor in all mode, but ignored in semver mode
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump", "--mode=semver"})

	// Action: Execute bump command with semver-only mode
	err := rootCmd.Execute()

	// Expected: No bump detected (conventional commits ignored)
	suite.NoError(err)
	suite.Contains(buf.String(), "No version bump detected")
}

// TestMakeLevelCmd_MajorDecrement validates that 'bump major decrement'
// correctly decrements the major version.
//
// Why: Version decrement is needed for rollback scenarios or correcting
// accidental version bumps.
//
// What: Given VERSION 2.0.0, when 'bump major decrement' runs, then VERSION
// becomes 1.0.0.
func (suite *BumpTestSuite) TestMakeLevelCmd_MajorDecrement() {
	// Precondition: VERSION file exists with major > 0
	suite.createVersionFile("2.0.0")
	rootCmd.SetArgs([]string{"bump", "major", "decrement"})

	// Action: Execute bump major decrement command
	err := rootCmd.Execute()

	// Expected: Major decremented
	suite.NoError(err)
	suite.Equal("1.0.0", suite.readVersionFile())
}

// TestMakeLevelCmd_MinorDecrement validates that 'bump minor decrement'
// correctly decrements the minor version.
//
// Why: Minor version decrement is needed for rollback scenarios.
//
// What: Given VERSION 1.2.0, when 'bump minor decrement' runs, then VERSION
// becomes 1.1.0.
func (suite *BumpTestSuite) TestMakeLevelCmd_MinorDecrement() {
	// Precondition: VERSION file exists with minor > 0
	suite.createVersionFile("1.2.0")
	rootCmd.SetArgs([]string{"bump", "minor", "decrement"})

	// Action: Execute bump minor decrement command
	err := rootCmd.Execute()

	// Expected: Minor decremented
	suite.NoError(err)
	suite.Equal("1.1.0", suite.readVersionFile())
}

// TestMakeLevelCmd_IncrementAlias validates that 'inc' is accepted as an
// alias for 'increment'.
//
// Why: Short aliases improve CLI usability for frequent operations.
//
// What: Given VERSION 1.2.3, when 'bump patch inc' runs, then VERSION
// becomes 1.2.4 (same as 'increment').
func (suite *BumpTestSuite) TestMakeLevelCmd_IncrementAlias() {
	// Precondition: VERSION file exists
	suite.createVersionFile("1.2.3")
	rootCmd.SetArgs([]string{"bump", "patch", "inc"})

	// Action: Execute bump patch with inc alias
	err := rootCmd.Execute()

	// Expected: Same as increment
	suite.NoError(err)
	suite.Equal("1.2.4", suite.readVersionFile())
}

// TestMakeLevelCmd_DecrementAlias validates that 'dec' is accepted as an
// alias for 'decrement'.
//
// Why: Short aliases improve CLI usability for frequent operations.
//
// What: Given VERSION 1.2.3, when 'bump patch dec' runs, then VERSION
// becomes 1.2.2 (same as 'decrement').
func (suite *BumpTestSuite) TestMakeLevelCmd_DecrementAlias() {
	// Precondition: VERSION file exists
	suite.createVersionFile("1.2.3")
	rootCmd.SetArgs([]string{"bump", "patch", "dec"})

	// Action: Execute bump patch with dec alias
	err := rootCmd.Execute()

	// Expected: Same as decrement
	suite.NoError(err)
	suite.Equal("1.2.2", suite.readVersionFile())
}

// =============================================================================
// ERROR HANDLING - Expected Failure Modes
// =============================================================================

// TestRunBump_NoVCSDetected validates that bump fails gracefully when
// run outside a version control repository.
//
// Why: Clear error messaging helps users understand they must run the
// command from within a git repository.
//
// What: When bump runs without a VCS, then an error is returned with
// a descriptive message.
func (suite *BumpTestSuite) TestRunBump_NoVCSDetected() {
	// Precondition: No VCS registered
	vcs.UnregisterVCS("git")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command without VCS
	err := rootCmd.Execute()

	// Expected: Error indicating no VCS detected
	suite.Error(err)
	suite.Contains(err.Error(), "not in a version control repository")
}

// TestRunBump_GetCommitsError validates that errors from VCS when fetching
// commits are properly propagated with context.
//
// Why: VCS errors should be surfaced clearly so users can diagnose
// repository issues.
//
// What: When VCS returns an error for GetCommitMessagesSinceTag, then
// bump returns an error with appropriate context.
func (suite *BumpTestSuite) TestRunBump_GetCommitsError() {
	// Precondition: VCS configured but GetCommitMessagesSinceTag fails
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return(nil, assert.AnError)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command
	err := rootCmd.Execute()

	// Expected: Error with context about commit fetching
	suite.Error(err)
	suite.Contains(err.Error(), "failed to get commits")
}

// TestRunBump_AmendError validates that errors during commit amendment
// are properly propagated with context.
//
// Why: Amend failures can leave the repository in an unexpected state;
// users need clear error messages to recover.
//
// What: When VCS returns an error for AmendCommit, then bump returns
// an error with appropriate context.
func (suite *BumpTestSuite) TestRunBump_AmendError() {
	// Precondition: VERSION file exists, VCS configured but AmendCommit fails
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"fix: bug fix",
	}, nil)
	mockVCS.EXPECT().AmendCommit([]string{"VERSION"}).Return(assert.AnError)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command
	err := rootCmd.Execute()

	// Expected: Error with context about amend failure
	suite.Error(err)
	suite.Contains(err.Error(), "failed to amend commit")
}

// =============================================================================
// EDGE CASES - Boundary Conditions
// =============================================================================

// TestRunBump_NoCommitsSinceTag validates behavior when no commits exist
// since the last version tag.
//
// Why: This is a common scenario in freshly tagged releases; users should
// see a clear message rather than an error.
//
// What: When no commits exist since the last tag, then bump outputs an
// informative message and succeeds without changes.
func (suite *BumpTestSuite) TestRunBump_NoCommitsSinceTag() {
	// Precondition: VERSION file exists, VCS returns empty commit list
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command
	err := rootCmd.Execute()

	// Expected: No error, informative message
	suite.NoError(err)
	suite.Contains(buf.String(), "No commits since last tag")
}

// TestRunBump_NoBumpDetected validates behavior when commits exist but
// none trigger a version bump (e.g., chore, docs).
//
// Why: Non-bumping commits are common; users should understand why no
// version change occurred.
//
// What: When commits don't trigger bumps, then bump outputs an informative
// message and succeeds without changes.
func (suite *BumpTestSuite) TestRunBump_NoBumpDetected() {
	// Precondition: VERSION file exists, commits don't trigger bumps
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"chore: update deps",
		"docs: update readme",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command
	err := rootCmd.Execute()

	// Expected: No error, informative message
	suite.NoError(err)
	suite.Contains(buf.String(), "No version bump detected")
}

// TestRunBump_SkipDetected validates behavior when a commit contains
// the +semver:skip marker.
//
// Why: The skip marker allows users to explicitly prevent version bumps
// for specific commits, even if they would normally trigger one.
//
// What: When a commit contains +semver:skip, then bump outputs an
// informative message and succeeds without changes.
func (suite *BumpTestSuite) TestRunBump_SkipDetected() {
	// Precondition: VERSION file exists, commit has skip marker
	suite.createVersionFile("1.0.0")
	mockVCS := suite.setupMockVCS()
	mockVCS.EXPECT().GetCommitMessagesSinceTag().Return([]string{
		"feat: new feature +semver:skip",
	}, nil)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"bump"})

	// Action: Execute bump command
	err := rootCmd.Execute()

	// Expected: No error, skip message
	suite.NoError(err)
	suite.Contains(buf.String(), "Version bump skipped")
}

// =============================================================================
// HELPER FUNCTION TESTS
// =============================================================================

// -----------------------------------------------------------------------------
// getParseMode - CORE FUNCTIONALITY
// -----------------------------------------------------------------------------

// TestGetParseMode validates that mode flag strings are correctly converted
// to ParseMode enum values for the commit parser.
//
// Why: Mode conversion is critical for controlling which commit message
// formats are recognized; incorrect mapping could cause missed version bumps.
//
// What: Given various mode strings (case-insensitive), the function should
// return the corresponding ParseMode constant.
func TestGetParseMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected commitparser.ParseMode
	}{
		// Core functionality - recognized modes
		{name: "semver lowercase", input: "semver", expected: commitparser.ModeSemverMarkers},
		{name: "conventional lowercase", input: "conventional", expected: commitparser.ModeConventionalCommits},
		{name: "all lowercase", input: "all", expected: commitparser.ModeAll},

		// Key variations - case insensitivity
		{name: "semver uppercase", input: "SEMVER", expected: commitparser.ModeSemverMarkers},
		{name: "semver mixed case", input: "Semver", expected: commitparser.ModeSemverMarkers},
		{name: "conventional uppercase", input: "CONVENTIONAL", expected: commitparser.ModeConventionalCommits},
		{name: "all uppercase", input: "ALL", expected: commitparser.ModeAll},

		// Edge cases - defaults
		{name: "default for empty", input: "", expected: commitparser.ModeAll},
		{name: "default for unknown", input: "unknown", expected: commitparser.ModeAll},
		{name: "default for invalid", input: "xyz", expected: commitparser.ModeAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Convert mode string
			result := getParseMode(tt.input)

			// Expected: Correct ParseMode returned
			assert.Equal(t, tt.expected, result)
		})
	}
}

// -----------------------------------------------------------------------------
// calculateNewVersion - CORE FUNCTIONALITY
// -----------------------------------------------------------------------------

// TestCalculateNewVersion validates that version bumps are calculated correctly
// for each bump level without modifying the original version.
//
// Why: Version calculation is the core logic of semantic versioning; incorrect
// bumps could lead to version collisions or broken dependency resolution.
//
// What: Given a version and bump level, the function should return the
// correctly bumped version string while leaving the original unchanged.
func TestCalculateNewVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  *version.Version
		level    commitparser.BumpLevel
		expected string
	}{
		// Core functionality - standard bumps
		{
			name:     "major bump resets minor and patch",
			version:  &version.Version{Major: 1, Minor: 2, Patch: 3},
			level:    commitparser.BumpMajor,
			expected: "2.0.0",
		},
		{
			name:     "minor bump resets patch",
			version:  &version.Version{Major: 1, Minor: 2, Patch: 3},
			level:    commitparser.BumpMinor,
			expected: "1.3.0",
		},
		{
			name:     "patch bump increments patch only",
			version:  &version.Version{Major: 1, Minor: 2, Patch: 3},
			level:    commitparser.BumpPatch,
			expected: "1.2.4",
		},

		// Key variations - prefix preservation
		{
			name:     "major bump with prefix",
			version:  &version.Version{Prefix: "v", Major: 1, Minor: 0, Patch: 0},
			level:    commitparser.BumpMajor,
			expected: "v2.0.0",
		},

		// Edge cases - zero versions
		{
			name:     "minor bump from zero version",
			version:  &version.Version{Major: 0, Minor: 0, Patch: 0},
			level:    commitparser.BumpMinor,
			expected: "0.1.0",
		},
		{
			name:     "patch bump from zero version",
			version:  &version.Version{Major: 0, Minor: 0, Patch: 0},
			level:    commitparser.BumpPatch,
			expected: "0.0.1",
		},

		// Edge cases - high numbers
		{
			name:     "major bump from high numbers",
			version:  &version.Version{Major: 99, Minor: 99, Patch: 99},
			level:    commitparser.BumpMajor,
			expected: "100.0.0",
		},

		// Edge cases - no bump
		{
			name:     "no bump (BumpNone)",
			version:  &version.Version{Major: 1, Minor: 2, Patch: 3},
			level:    commitparser.BumpNone,
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Calculate new version
			result := calculateNewVersion(tt.version, tt.level)

			// Expected: Correct version string returned
			assert.Equal(t, tt.expected, result)
		})
	}
}

// -----------------------------------------------------------------------------
// truncateCommit - CORE FUNCTIONALITY
// -----------------------------------------------------------------------------

// TestTruncateCommit validates that commit messages are properly truncated
// for display purposes while preserving important information.
//
// Why: Truncation ensures consistent output formatting in logs and CLI
// without losing the essential commit summary.
//
// What: Given commit messages of various lengths and formats, the function
// should return truncated strings of at most 60 characters.
func TestTruncateCommit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Core functionality - normal messages
		{
			name:     "short message unchanged",
			input:    "fix: simple bug fix",
			expected: "fix: simple bug fix",
		},

		// Edge cases - boundary conditions
		{
			name:     "exactly 60 chars unchanged",
			input:    "123456789012345678901234567890123456789012345678901234567890",
			expected: "123456789012345678901234567890123456789012345678901234567890",
		},
		{
			name:     "61 chars truncated with ellipsis",
			input:    "1234567890123456789012345678901234567890123456789012345678901",
			expected: "123456789012345678901234567890123456789012345678901234567...",
		},

		// Key variations - long messages
		{
			name:     "long message truncated",
			input:    "feat: this is a very long commit message that exceeds the maximum display length and should be truncated",
			expected: "feat: this is a very long commit message that exceeds the...",
		},

		// Key variations - multiline handling
		{
			name:     "multiline takes first line only",
			input:    "fix: first line\n\nThis is the body of the commit message with more details.",
			expected: "fix: first line",
		},
		{
			name:     "multiline with long first line truncated",
			input:    "feat: this is a very long first line that exceeds the maximum display length and should be truncated\n\nBody text here.",
			expected: "feat: this is a very long first line that exceeds the max...",
		},

		// Edge cases - whitespace handling
		{
			name:     "whitespace trimmed",
			input:    "   fix: message with leading space   ",
			expected: "fix: message with leading space",
		},

		// Edge cases - empty input
		{
			name:     "empty message",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n   \n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Truncate commit message
			result := truncateCommit(tt.input)

			// Expected: Correctly truncated string
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 60, "truncated message should be at most 60 chars")
		})
	}
}
