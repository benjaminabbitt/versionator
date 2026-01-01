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

// MajorTestSuite defines the test suite for major command tests
type MajorTestSuite struct {
	suite.Suite
	memFs   *fstesting.MemFs
	cleanup func()
	cwd     string // current working directory for absolute paths
}

// SetupTest runs before each test
func (suite *MajorTestSuite) SetupTest() {
	// Get current working directory for absolute paths
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	suite.cwd = cwd

	// Set up in-memory filesystem
	suite.memFs, suite.cleanup = fstesting.SetupTestFs()
}

// TearDownTest runs after each test
func (suite *MajorTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// absPath returns absolute path for a filename relative to cwd
func (suite *MajorTestSuite) absPath(filename string) string {
	return filepath.Join(suite.cwd, filename)
}

// createTestFiles creates the standard test files needed for most tests
func (suite *MajorTestSuite) createTestFiles(version string) {
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

func (suite *MajorTestSuite) TestMajorIncrementCommand() {
	// Create test files
	suite.createTestFiles("1.2.3")

	// Execute the major increment command directly
	var stdout bytes.Buffer
	majorIncrementCmd.SetOut(&stdout)
	err := majorIncrementCmd.RunE(majorIncrementCmd, []string{})
	suite.Require().NoError(err, "major increment command should succeed")

	// Verify VERSION file was updated correctly
	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v2.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'v2.0.0'")
}

func (suite *MajorTestSuite) TestMajorDecrementCommand() {
	// Create test files
	suite.createTestFiles("3.5.7")

	// Execute the major decrement command directly
	var stdout bytes.Buffer
	majorDecrementCmd.SetOut(&stdout)
	err := majorDecrementCmd.RunE(majorDecrementCmd, []string{})
	suite.Require().NoError(err, "major decrement command should succeed")

	// Verify VERSION file was updated correctly
	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v2.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'v2.0.0'")
}

func (suite *MajorTestSuite) TestMajorDecrementCommand_AtZero() {
	// Create test files with major version at 0
	suite.createTestFiles("0.5.3")

	// Execute the major decrement command - should fail
	err := majorDecrementCmd.RunE(majorDecrementCmd, []string{})
	suite.Error(err, "Expected major decrement command to fail when major version is at 0")
}

func (suite *MajorTestSuite) TestMajorIncrementCommand_NoVersionFile() {
	// Create only config file (no VERSION file)
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

	// Execute the major increment command - should succeed with default version
	err = majorIncrementCmd.RunE(majorIncrementCmd, []string{})
	suite.Require().NoError(err, "major increment command should succeed with default version")

	// Verify VERSION file was created and updated correctly
	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'v1.0.0'")
}

// TestMajorTestSuite runs the major test suite
func TestMajorTestSuite(t *testing.T) {
	suite.Run(t, new(MajorTestSuite))
}
