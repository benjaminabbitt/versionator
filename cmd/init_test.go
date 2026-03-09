package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// InitTestSuite defines the test suite for init command tests
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

func (suite *InitTestSuite) TestInitCommand_Default() {
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify VERSION file was created
	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("0.0.1", strings.TrimSpace(string(content)), "VERSION should contain '0.0.1'")

	// Verify config was NOT created
	_, err = os.Stat(".versionator.yaml")
	suite.True(os.IsNotExist(err), "Config file should not exist without --config flag")
}

func (suite *InitTestSuite) TestInitCommand_WithVersion() {
	rootCmd.SetArgs([]string{"init", "--version", "1.2.3"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.2.3", strings.TrimSpace(string(content)), "VERSION should contain '1.2.3'")
}

func (suite *InitTestSuite) TestInitCommand_WithPrefix() {
	rootCmd.SetArgs([]string{"init", "--prefix", "v"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v0.0.1", strings.TrimSpace(string(content)), "VERSION should contain 'v0.0.1'")
}

func (suite *InitTestSuite) TestInitCommand_WithVersionAndPrefix() {
	rootCmd.SetArgs([]string{"init", "--version", "2.0.0", "--prefix", "V"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("V2.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'V2.0.0'")
}

func (suite *InitTestSuite) TestInitCommand_InvalidPrefixRejected() {
	rootCmd.SetArgs([]string{"init", "--prefix", "release-"})
	err := rootCmd.Execute()
	suite.Error(err, "init command should reject invalid prefix")
	suite.Contains(err.Error(), "only 'v' or 'V' allowed", "error should mention valid prefixes")
}

func (suite *InitTestSuite) TestInitCommand_WithConfig() {
	rootCmd.SetArgs([]string{"init", "--config"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify VERSION file was created
	_, err = os.Stat("VERSION")
	suite.Require().NoError(err, "VERSION file should exist")

	// Verify config was created
	_, err = os.Stat(".versionator.yaml")
	suite.Require().NoError(err, "Config file should exist with --config flag")
}

func (suite *InitTestSuite) TestInitCommand_FailsIfVersionExists() {
	// Create existing VERSION file
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)

	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Error(err, "init should fail if VERSION exists")
	suite.Contains(err.Error(), "already exists")
}

func (suite *InitTestSuite) TestInitCommand_ForceOverwrite() {
	// Create existing VERSION file
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)

	rootCmd.SetArgs([]string{"init", "--force", "--version", "2.0.0"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init --force should succeed")

	content, err := os.ReadFile("VERSION")
	suite.Require().NoError(err)
	suite.Equal("2.0.0", strings.TrimSpace(string(content)), "VERSION should be overwritten to '2.0.0'")
}

func (suite *InitTestSuite) TestInitCommand_ConfigFailsIfExists() {
	// Create existing config file
	err := os.WriteFile(".versionator.yaml", []byte("prefix: v\n"), 0644)
	suite.Require().NoError(err)

	rootCmd.SetArgs([]string{"init", "--config"})
	err = rootCmd.Execute()
	suite.Error(err, "init --config should fail if config exists")
	suite.Contains(err.Error(), "already exists")
}

func (suite *InitTestSuite) TestInitCommand_ForceOverwriteConfig() {
	// Create existing files
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)
	err = os.WriteFile(".versionator.yaml", []byte("prefix: old\n"), 0644)
	suite.Require().NoError(err)

	rootCmd.SetArgs([]string{"init", "--force", "--config"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init --force --config should succeed")

	// Verify both files were overwritten
	_, err = os.Stat("VERSION")
	suite.Require().NoError(err)
	_, err = os.Stat(".versionator.yaml")
	suite.Require().NoError(err)
}

// TestInitTestSuite runs the init test suite
func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}
