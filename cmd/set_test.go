package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// SetTestSuite defines the test suite for set command tests.
type SetTestSuite struct {
	suite.Suite
	tempDir string
	origDir string
}

// SetupTest runs before each test
func (s *SetTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
	var err error
	s.origDir, err = os.Getwd()
	s.Require().NoError(err)
	err = os.Chdir(s.tempDir)
	s.Require().NoError(err)

	// Create a default VERSION file
	err = os.WriteFile("VERSION", []byte("0.0.1\n"), 0644)
	s.Require().NoError(err)
}

// TearDownTest runs after each test
func (s *SetTestSuite) TearDownTest() {
	if s.origDir != "" {
		_ = os.Chdir(s.origDir)
	}
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
}

func TestSetTestSuite(t *testing.T) {
	suite.Run(t, new(SetTestSuite))
}

// TestSetCommand_ValidSemver_WritesVersionFile validates setting a basic semver version.
func (s *SetTestSuite) TestSetCommand_ValidSemver_WritesVersionFile() {
	rootCmd.SetArgs([]string{"set", "1.2.3"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("1.2.3", strings.TrimSpace(string(content)))
}

// TestSetCommand_WithPrefix_PreservesPrefix validates that prefix is preserved.
func (s *SetTestSuite) TestSetCommand_WithPrefix_PreservesPrefix() {
	rootCmd.SetArgs([]string{"set", "v2.0.0"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("v2.0.0", strings.TrimSpace(string(content)))
}

// TestSetCommand_WithPreRelease_PreservesPreRelease validates pre-release handling.
func (s *SetTestSuite) TestSetCommand_WithPreRelease_PreservesPreRelease() {
	rootCmd.SetArgs([]string{"set", "1.0.0-alpha.1"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("1.0.0-alpha.1", strings.TrimSpace(string(content)))
}

// TestSetCommand_WithMetadata_PreservesMetadata validates build metadata handling.
func (s *SetTestSuite) TestSetCommand_WithMetadata_PreservesMetadata() {
	rootCmd.SetArgs([]string{"set", "1.0.0+build.42"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("1.0.0+build.42", strings.TrimSpace(string(content)))
}

// TestSetCommand_WithRevision_PreservesRevision validates 4-component version handling.
func (s *SetTestSuite) TestSetCommand_WithRevision_PreservesRevision() {
	rootCmd.SetArgs([]string{"set", "1.2.3.4"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("1.2.3.4", strings.TrimSpace(string(content)))
}

// TestSetCommand_FullVersion_PreservesAll validates a version with all components.
func (s *SetTestSuite) TestSetCommand_FullVersion_PreservesAll() {
	rootCmd.SetArgs([]string{"set", "v1.2.3-rc.1+build.42"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("v1.2.3-rc.1+build.42", strings.TrimSpace(string(content)))
}

// TestSetCommand_RevisionWithPreRelease_PreservesAll validates 4-component with pre-release.
func (s *SetTestSuite) TestSetCommand_RevisionWithPreRelease_PreservesAll() {
	rootCmd.SetArgs([]string{"set", "1.2.3.4-beta.1"})
	err := rootCmd.Execute()

	s.Require().NoError(err)
	content, err := os.ReadFile("VERSION")
	s.Require().NoError(err)
	s.Equal("1.2.3.4-beta.1", strings.TrimSpace(string(content)))
}

// TestSetCommand_InvalidVersion_ReturnsError validates that invalid input is rejected.
func (s *SetTestSuite) TestSetCommand_InvalidVersion_ReturnsError() {
	rootCmd.SetArgs([]string{"set", "not-a-version"})
	err := rootCmd.Execute()

	s.Error(err)
}

// TestSetCommand_NoArgs_ReturnsError validates that missing argument is caught.
func (s *SetTestSuite) TestSetCommand_NoArgs_ReturnsError() {
	rootCmd.SetArgs([]string{"set"})
	err := rootCmd.Execute()

	s.Error(err)
}
