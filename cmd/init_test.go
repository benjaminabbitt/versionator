package cmd

import (
	"bytes"
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
	suite.tempDir = suite.T().TempDir()
	var err error
	suite.origDir, err = os.Getwd()
	suite.Require().NoError(err)
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)
}

// TearDownTest runs after each test
func (suite *InitTestSuite) TearDownTest() {
	if suite.origDir != "" {
		os.Chdir(suite.origDir)
	}
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
}

func (suite *InitTestSuite) TestInitCommand_CreatesVersionFile() {
	// Verify VERSION doesn't exist
	_, err := os.Stat("VERSION")
	suite.True(os.IsNotExist(err), "VERSION file should not exist before init")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify VERSION was created
	data, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "VERSION file should exist after init")

	content := strings.TrimSpace(string(data))
	suite.True(content == "0.0.0" || content == "v0.0.0",
		"VERSION should be 0.0.0 or v0.0.0, got %s", content)
}

func (suite *InitTestSuite) TestInitCommand_CreatesConfigFile() {
	// Verify config doesn't exist
	_, err := os.Stat(".versionator.yaml")
	suite.True(os.IsNotExist(err), ".versionator.yaml should not exist before init")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify config was created
	data, err := os.ReadFile(".versionator.yaml")
	suite.Require().NoError(err, ".versionator.yaml should exist after init")

	content := string(data)
	suite.Contains(content, "prefix:", "config should contain prefix setting")
	suite.Contains(content, "prerelease:", "config should contain prerelease setting")
}

func (suite *InitTestSuite) TestInitCommand_DoesNotOverwriteExistingVersion() {
	// Create existing VERSION file with custom content
	existingVersion := "1.2.3\n"
	err := os.WriteFile("VERSION", []byte(existingVersion), 0644)
	suite.Require().NoError(err, "failed to create existing VERSION")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify VERSION was NOT overwritten
	data, err := os.ReadFile("VERSION")
	suite.Require().NoError(err, "failed to read VERSION")
	suite.Equal(existingVersion, string(data), "VERSION should not be overwritten")
}

func (suite *InitTestSuite) TestInitCommand_DoesNotOverwriteExistingConfig() {
	// Create existing config file
	existingConfig := "# custom config\nprefix: \"release-\"\n"
	err := os.WriteFile(".versionator.yaml", []byte(existingConfig), 0644)
	suite.Require().NoError(err, "failed to create existing config")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify config was NOT overwritten
	data, err := os.ReadFile(".versionator.yaml")
	suite.Require().NoError(err, "failed to read config")
	suite.Equal(existingConfig, string(data), "config should not be overwritten")
}

func (suite *InitTestSuite) TestInitCommand_OutputsCreatedStatus() {
	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	output := stdout.String()
	suite.Contains(output, "Created VERSION", "output should mention VERSION file created")
	suite.Contains(output, "Created .versionator.yaml", "output should mention config file created")
	suite.Contains(output, "Initialization complete", "output should indicate completion")
}

func (suite *InitTestSuite) TestInitCommand_OutputsExistsStatus() {
	// Create existing files
	err := os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)
	err = os.WriteFile(".versionator.yaml", []byte("prefix: \"v\"\n"), 0644)
	suite.Require().NoError(err)

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	output := stdout.String()
	suite.Contains(output, "VERSION file exists", "output should mention VERSION exists")
	suite.Contains(output, ".versionator.yaml exists", "output should mention config exists")
	suite.Contains(output, "Already initialized", "output should indicate already initialized")
}

// TestInitTestSuite runs the init test suite
func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}
