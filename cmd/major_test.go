package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// MajorTestSuite defines the test suite for major command tests
type MajorTestSuite struct {
	suite.Suite
	tempDir string
	origDir string
}

// SetupTest runs before each test
func (suite *MajorTestSuite) SetupTest() {
	// Create a temporary directory for testing
	suite.tempDir = suite.T().TempDir()
	var err error
	suite.origDir, err = os.Getwd()
	suite.Require().NoError(err)
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)
}

// TearDownTest runs after each test
func (suite *MajorTestSuite) TearDownTest() {
	// Restore original directory
	if suite.origDir != "" {
		os.Chdir(suite.origDir)
	}

	// Reset command state
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
}

// createTestFiles creates the standard test files needed for most tests
func (suite *MajorTestSuite) createTestFiles(version string) {
	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create a minimal config file
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

func (suite *MajorTestSuite) TestMajorIncrementCommand() {
	// Create test files
	suite.createTestFiles("1.2.3")

	// Execute the major increment command
	rootCmd.SetArgs([]string{"major", "increment"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "major increment command should succeed")

	// Verify VERSION file was updated correctly
	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("2.0.0", strings.TrimSpace(string(content)), "VERSION file should contain '2.0.0'")
}

func (suite *MajorTestSuite) TestMajorIncrementCommand_Aliases() {
	testCases := []string{"inc", "+"}

	for _, alias := range testCases {
		suite.Run("alias_"+alias, func() {
			// Create test files
			suite.createTestFiles("0.1.0")

			// Execute the major increment command with alias
			rootCmd.SetArgs([]string{"major", alias})
			err := rootCmd.Execute()
			suite.Require().NoError(err, "major %s command should succeed", alias)

			// Verify VERSION file was updated correctly
			content, err := os.ReadFile("VERSION")
			suite.Require().NoError(err, "Should be able to read VERSION file")
			suite.Equal("1.0.0", strings.TrimSpace(string(content)), "VERSION file should contain '1.0.0'")
		})
	}
}

func (suite *MajorTestSuite) TestMajorDecrementCommand() {
	// Create test files
	suite.createTestFiles("3.5.7")

	// Execute the major decrement command
	rootCmd.SetArgs([]string{"major", "decrement"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "major decrement command should succeed")

	// Verify VERSION file was updated correctly
	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("2.0.0", strings.TrimSpace(string(content)), "VERSION file should contain '2.0.0'")
}

func (suite *MajorTestSuite) TestMajorDecrementCommand_Aliases() {
	testCases := []string{"dec", "-"}

	for _, alias := range testCases {
		suite.Run("alias_"+alias, func() {
			// Create test files
			suite.createTestFiles("2.1.0")

			// Execute the major decrement command with alias
			rootCmd.SetArgs([]string{"major", alias})
			err := rootCmd.Execute()
			suite.Require().NoError(err, "major %s command should succeed", alias)

			// Verify VERSION file was updated correctly
			content, err := os.ReadFile("VERSION")
			suite.Require().NoError(err, "Should be able to read VERSION file")
			suite.Equal("1.0.0", strings.TrimSpace(string(content)), "VERSION file should contain '1.0.0'")
		})
	}
}

func (suite *MajorTestSuite) TestMajorIncrementCommand_NoVersionFile() {
	// Create only config file (no VERSION file)
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Execute the major increment command - should succeed with default version
	rootCmd.SetArgs([]string{"major", "increment"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "major increment command should succeed with default version")

	// Verify VERSION file was created and updated correctly
	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.0.0", strings.TrimSpace(string(content)), "VERSION file should contain '1.0.0'")
}

func (suite *MajorTestSuite) TestMajorDecrementCommand_AtZero() {
	// Create test files with major version at 0
	suite.createTestFiles("0.5.3")

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"major", "decrement"})

	// Execute the major decrement command - should fail
	err := rootCmd.Execute()
	suite.Error(err, "Expected major decrement command to fail when major version is at 0")
}

func (suite *MajorTestSuite) TestMajorCommand_InvalidVersionFile() {
	// Create an invalid VERSION file
	err := os.WriteFile("VERSION", []byte("invalid.version"), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create a minimal config file
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Test both increment and decrement with invalid version
	testCases := []string{"increment", "decrement"}

	for _, operation := range testCases {
		suite.Run(operation, func() {
			// Capture stderr
			var buf bytes.Buffer
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"major", operation})

			// Execute the command - should fail
			err := rootCmd.Execute()
			suite.Error(err, "Expected major %s command to fail with invalid version file", operation)
		})
	}
}

// TestMajorTestSuite runs the major test suite
func TestMajorTestSuite(t *testing.T) {
	suite.Run(t, new(MajorTestSuite))
}
