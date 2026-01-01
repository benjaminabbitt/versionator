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

// PatchTestSuite defines the test suite for patch command tests
type PatchTestSuite struct {
	suite.Suite
	memFs   *fstesting.MemFs
	cleanup func()
	cwd     string // current working directory for absolute paths
}

// SetupTest runs before each test
func (suite *PatchTestSuite) SetupTest() {
	// Get current working directory for absolute paths
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	suite.cwd = cwd

	// Set up in-memory filesystem
	suite.memFs, suite.cleanup = fstesting.SetupTestFs()
}

// TearDownTest runs after each test
func (suite *PatchTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// absPath returns absolute path for a filename relative to cwd
func (suite *PatchTestSuite) absPath(filename string) string {
	return filepath.Join(suite.cwd, filename)
}

// createTestFiles creates the standard test files needed for most tests
func (suite *PatchTestSuite) createTestFiles(version string) {
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

func (suite *PatchTestSuite) TestPatchIncrementCommand() {
	suite.createTestFiles("1.2.3")

	var stdout bytes.Buffer
	patchIncrementCmd.SetOut(&stdout)
	err := patchIncrementCmd.RunE(patchIncrementCmd, []string{})
	suite.Require().NoError(err, "patch increment command should succeed")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.2.4", strings.TrimSpace(string(content)), "VERSION should contain 'v1.2.4'")
}

func (suite *PatchTestSuite) TestPatchDecrementCommand() {
	suite.createTestFiles("1.3.5")

	var stdout bytes.Buffer
	patchDecrementCmd.SetOut(&stdout)
	err := patchDecrementCmd.RunE(patchDecrementCmd, []string{})
	suite.Require().NoError(err, "patch decrement command should succeed")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.3.4", strings.TrimSpace(string(content)), "VERSION should contain 'v1.3.4'")
}

func (suite *PatchTestSuite) TestPatchDecrementCommand_AtZero() {
	suite.createTestFiles("1.3.0")

	err := patchDecrementCmd.RunE(patchDecrementCmd, []string{})
	suite.Error(err, "Expected patch decrement command to fail when patch version is at 0")
}

func (suite *PatchTestSuite) TestPatchIncrementCommand_NoVersionFile() {
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

	err = patchIncrementCmd.RunE(patchIncrementCmd, []string{})
	suite.Require().NoError(err, "patch increment command should succeed with default version")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v0.0.1", strings.TrimSpace(string(content)), "VERSION should contain 'v0.0.1'")
}

func (suite *PatchTestSuite) TestPatchIncrementCommand_LargeNumber() {
	suite.createTestFiles("1.0.999")

	err := patchIncrementCmd.RunE(patchIncrementCmd, []string{})
	suite.Require().NoError(err, "patch increment command should succeed")

	// Note: default config prefix is "v", so version will have that prefix
	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.0.1000", strings.TrimSpace(string(content)), "VERSION should contain 'v1.0.1000'")
}

func TestPatchTestSuite(t *testing.T) {
	suite.Run(t, new(PatchTestSuite))
}
