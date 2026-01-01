package version

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/filesystem"
	fstesting "github.com/benjaminabbitt/versionator/internal/filesystem/testing"
	"github.com/stretchr/testify/suite"
)

// MinorTestSuite defines the test suite for minor command tests
type MinorTestSuite struct {
	suite.Suite
	memFs   *fstesting.MemFs
	cleanup func()
	cwd     string // current working directory for absolute paths
}

// SetupTest runs before each test
func (suite *MinorTestSuite) SetupTest() {
	// Get current working directory for absolute paths
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	suite.cwd = cwd

	// Set up in-memory filesystem
	suite.memFs, suite.cleanup = fstesting.SetupTestFs()
}

// TearDownTest runs after each test
func (suite *MinorTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// absPath returns absolute path for a filename relative to cwd
func (suite *MinorTestSuite) absPath(filename string) string {
	return filepath.Join(suite.cwd, filename)
}

// createTestFiles creates the standard test files needed for most tests
func (suite *MinorTestSuite) createTestFiles(version string) {
	// Write plain text VERSION file using absolute path
	err := filesystem.AppFs.WriteFile(suite.absPath("VERSION"), []byte(version+"\n"), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create a minimal config file with no prerelease/metadata elements
	configContent := `prefix: ""
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

func (suite *MinorTestSuite) TestMinorIncrementCommand() {
	suite.createTestFiles("1.2.3")

	var stdout bytes.Buffer
	minorIncrementCmd.SetOut(&stdout)
	err := minorIncrementCmd.RunE(minorIncrementCmd, []string{})
	suite.Require().NoError(err, "minor increment command should succeed")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.3.0", strings.TrimSpace(string(content)), "VERSION should contain 'v1.3.0'")
}

func (suite *MinorTestSuite) TestMinorDecrementCommand() {
	suite.createTestFiles("1.3.5")

	var stdout bytes.Buffer
	minorDecrementCmd.SetOut(&stdout)
	err := minorDecrementCmd.RunE(minorDecrementCmd, []string{})
	suite.Require().NoError(err, "minor decrement command should succeed")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.2.0", strings.TrimSpace(string(content)), "VERSION should contain 'v1.2.0'")
}

func (suite *MinorTestSuite) TestMinorDecrementCommand_AtZero() {
	suite.createTestFiles("1.0.3")

	err := minorDecrementCmd.RunE(minorDecrementCmd, []string{})
	suite.Error(err, "Expected minor decrement command to fail when minor version is at 0")
}

func (suite *MinorTestSuite) TestMinorIncrementCommand_NoVersionFile() {
	configContent := `prefix: ""
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

	err = minorIncrementCmd.RunE(minorIncrementCmd, []string{})
	suite.Require().NoError(err, "minor increment command should succeed with default version")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v0.1.0", strings.TrimSpace(string(content)), "VERSION should contain 'v0.1.0'")
}

func TestMinorTestSuite(t *testing.T) {
	suite.Run(t, new(MinorTestSuite))
}
