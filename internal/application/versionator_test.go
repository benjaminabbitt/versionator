package application

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"versionator/internal/vcs/mock"
)

// VersionatorTestSuite is the test suite for application functionality
type VersionatorTestSuite struct {
	suite.Suite
	fs          afero.Fs
	versionator *Versionator
	mockCtrl    *gomock.Controller
	mockVCS     *mock.MockVersionControlSystem
	tempDir     string
	testVersion string
}

// SetupTest runs before each test
func (suite *VersionatorTestSuite) SetupTest() {
	suite.fs = afero.NewMemMapFs()
	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockVCS = mock.NewMockVersionControlSystem(suite.mockCtrl)
	suite.versionator = NewVersionator(suite.fs, nil) // Pass nil VCS for most tests
	suite.tempDir = "/test"
	suite.testVersion = "1.2.3"
}

// TearDownTest runs after each test
func (suite *VersionatorTestSuite) TearDownTest() {
	suite.mockCtrl.Finish()
}

// createVersionFile creates a VERSION file with the specified content
func (suite *VersionatorTestSuite) createVersionFile(version string) {
	// Create VERSION file in root directory where getVersionFilePath will find it
	// Since we're using memory filesystem and no VCS, it will fallback to current directory
	err := afero.WriteFile(suite.fs, "VERSION", []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")
}

// createConfigFile creates a config file with the specified content
func (suite *VersionatorTestSuite) createConfigFile(content string) {
	err := afero.WriteFile(suite.fs, ".application.yaml", []byte(content), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

// getConfigContent returns config content for different scenarios
func (suite *VersionatorTestSuite) getConfigContent(prefix string, suffixEnabled bool, suffixType string, hashLength int) string {
	return `prefix: "` + prefix + `"
suffix:
  enabled: ` + boolToString(suffixEnabled) + `
  type: "` + suffixType + `"
  git:
    hashLength: ` + intToString(hashLength) + `
logging:
  output: "console"
`
}

// Helper functions
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func intToString(i int) string {
	switch i {
	case 7:
		return "7"
	case 8:
		return "8"
	case 10:
		return "10"
	default:
		return "7"
	}
}

// TestGetVersionWithSuffix_NoSuffix_NoPrefix tests version with no suffix and no prefix
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_NoSuffix_NoPrefix() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("", false, "git", 7))

	// Test
	version, err := suite.versionator.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("1.2.3", version, "Expected version '1.2.3'")
}

// TestGetVersionWithSuffix_WithPrefix_NoSuffix tests version with prefix but no suffix
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_WithPrefix_NoSuffix() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("v", false, "git", 7))

	// Test
	version, err := suite.versionator.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("v1.2.3", version, "Expected version 'v1.2.3'")
}

// TestGetVersionWithSuffix_WithSuffix_NoVCS tests version with suffix enabled but no VCS
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_WithSuffix_NoVCS() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("", true, "git", 7))

	// Test
	version, err := suite.versionator.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("1.2.3", version, "Expected version '1.2.3' (no suffix due to no VCS)")
}

// TestGetVersionWithSuffix_WithSuffix_WithVCS tests version with suffix and VCS
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_WithSuffix_WithVCS() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("", true, "git", 7))

	// Mock VCS setup
	suite.mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	suite.mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	suite.mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

	// Create application with mock VCS
	versionatorWithVCS := NewVersionator(suite.fs, suite.mockVCS)

	// Test
	version, err := versionatorWithVCS.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("1.2.3-abc1234", version, "Expected version '1.2.3-abc1234'")
}

// TestGetVersionWithSuffix_WithPrefixAndSuffix tests version with both prefix and suffix
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_WithPrefixAndSuffix() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("v", true, "git", 8))

	// Mock VCS setup
	suite.mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	suite.mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	suite.mockVCS.EXPECT().GetVCSIdentifier(8).Return("abc12345", nil)

	// Create application with mock VCS
	versionatorWithVCS := NewVersionator(suite.fs, suite.mockVCS)

	// Test
	version, err := versionatorWithVCS.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("v1.2.3-abc12345", version, "Expected version 'v1.2.3-abc12345'")
}

// TestGetVersionWithSuffix_VCSNotInRepository tests when VCS is not in a repository
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_VCSNotInRepository() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("", true, "git", 7))

	// Mock VCS setup - not in repository
	suite.mockVCS.EXPECT().IsRepository().Return(false).AnyTimes()

	// Create application with mock VCS
	versionatorWithVCS := NewVersionator(suite.fs, suite.mockVCS)

	// Test
	version, err := versionatorWithVCS.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("1.2.3", version, "Expected version '1.2.3' (no suffix when not in repository)")
}

// TestGetVersionWithSuffix_VCSIdentifierError tests when VCS identifier fails
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_VCSIdentifierError() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("v", true, "git", 7))

	// Mock VCS setup - identifier error
	suite.mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	suite.mockVCS.EXPECT().GetRepositoryRoot().Return(".", nil).AnyTimes()
	suite.mockVCS.EXPECT().GetVCSIdentifier(7).Return("", suite.error("VCS identifier failed"))

	// Create application with mock VCS
	versionatorWithVCS := NewVersionator(suite.fs, suite.mockVCS)

	// Test
	version, err := versionatorWithVCS.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error (should continue without suffix)")
	suite.Equal("v1.2.3", version, "Expected version 'v1.2.3' (no suffix due to VCS error)")
}

// TestGetVersionWithSuffix_NonGitSuffixType tests with non-git suffix type
func (suite *VersionatorTestSuite) TestGetVersionWithSuffix_NonGitSuffixType() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("", true, "svn", 7))

	// Test
	version, err := suite.versionator.GetVersionWithSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("1.2.3", version, "Expected version '1.2.3' (no suffix for non-git type)")
}

// TestGetVersionWithPrefixAndSuffix_BackwardCompatibility tests backward compatibility
func (suite *VersionatorTestSuite) TestGetVersionWithPrefixAndSuffix_BackwardCompatibility() {
	// Setup
	suite.createVersionFile("1.2.3")
	suite.createConfigFile(suite.getConfigContent("v", false, "git", 7))

	// Test
	version, err := suite.versionator.GetVersionWithPrefixAndSuffix()
	suite.NoError(err, "Expected no error")
	suite.Equal("v1.2.3", version, "Expected version 'v1.2.3'")
}

// error is a helper to create errors in tests
func (suite *VersionatorTestSuite) error(message string) error {
	return &testError{message: message}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// TestVersionatorTestSuite runs the test suite
func TestVersionatorTestSuite(t *testing.T) {
	suite.Run(t, new(VersionatorTestSuite))
}
