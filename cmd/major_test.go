package cmd

import (
	"bytes"
	"testing"
	"versionator/internal/app"
	"versionator/internal/version"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// MajorTestSuite contains the test suite for major version commands
type MajorTestSuite struct {
	suite.Suite
	testApp      *app.App
	fs           afero.Fs
	restoreApp   func()
	outputBuffer *bytes.Buffer
}

// SetupTest initializes the test environment before each test
func (suite *MajorTestSuite) SetupTest() {
	var testApp *app.App
	suite.fs, testApp = createTestApp()
	suite.testApp = testApp
	suite.restoreApp = replaceAppInstance(testApp)
	suite.outputBuffer = &bytes.Buffer{}
}

// TearDownTest cleans up after each test
func (suite *MajorTestSuite) TearDownTest() {
	suite.restoreApp()
}

// TestMajorIncrementFromDefaultVersion tests incrementing major version from default 0.0.0
func (suite *MajorTestSuite) TestMajorIncrementFromDefaultVersion() {
	// Setup: Create config file (VERSION file will be auto-created)
	createConfigFile(suite.T(), suite.fs)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the increment command
	err := majorIncrementCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Major increment should succeed")

	// Verify the version was incremented correctly
	verifyVersionFile(suite.T(), suite.fs, "1.0.0")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Major version incremented to: 1.0.0", "Output should show incremented version")
}

// TestMajorIncrementFromExistingVersion tests incrementing major version from an existing version
func (suite *MajorTestSuite) TestMajorIncrementFromExistingVersion() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the increment command
	err := majorIncrementCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Major increment should succeed")

	// Verify the version was incremented correctly (minor and patch should reset to 0)
	verifyVersionFile(suite.T(), suite.fs, "3.0.0")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Major version incremented to: 3.0.0", "Output should show incremented version")
}

// TestMajorDecrementFromExistingVersion tests decrementing major version from an existing version
func (suite *MajorTestSuite) TestMajorDecrementFromExistingVersion() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the decrement command
	err := majorDecrementCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Major decrement should succeed")

	// Verify the version was decremented correctly (minor and patch should reset to 0)
	verifyVersionFile(suite.T(), suite.fs, "1.0.0")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Major version decremented to: 1.0.0", "Output should show decremented version")
}

// TestMajorDecrementFromZeroMajorVersion tests error when trying to decrement from 0 major version
func (suite *MajorTestSuite) TestMajorDecrementFromZeroMajorVersion() {
	// Setup: Create config and version files with major version = 0
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "0.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the decrement command
	err := majorDecrementCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Major decrement from 0 should fail")
	suite.Contains(err.Error(), "cannot decrement major version below 0", "Error should indicate major version cannot go below 0")

	// Verify the version file wasn't changed
	verifyVersionFile(suite.T(), suite.fs, "0.5.3")
}

// TestMajorDecrementFromDefaultVersion tests error when trying to decrement from default 0.0.0
func (suite *MajorTestSuite) TestMajorDecrementFromDefaultVersion() {
	// Setup: Create config file only (VERSION file will be auto-created as 0.0.0)
	createConfigFile(suite.T(), suite.fs)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the decrement command
	err := majorDecrementCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Major decrement from default version should fail")
	suite.Contains(err.Error(), "cannot decrement major version below 0", "Error should indicate major version cannot go below 0")
}

// TestMajorIncrementWithAppInstanceError tests handling of appInstance errors during increment
func (suite *MajorTestSuite) TestMajorIncrementWithAppInstanceError() {
	// Setup: Create config but no version file and make filesystem read-only
	createConfigFile(suite.T(), suite.fs)

	// Make filesystem read-only by replacing with a filesystem that errors on write
	suite.testApp.FileSystem = afero.NewReadOnlyFs(suite.fs)
	suite.testApp.VersionManager = version.NewVersion(suite.testApp.FileSystem, ".", nil)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the increment command
	err := majorIncrementCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Increment should fail with read-only filesystem")
	suite.Contains(err.Error(), "error incrementing major version", "Error should indicate increment failure")
}

// TestMajorCommandStructure tests that the commands are properly structured
func (suite *MajorTestSuite) TestMajorCommandStructure() {
	// Test major command properties
	suite.Equal("major", majorCmd.Use, "Major command should have correct use")
	suite.Equal("Manage major version", majorCmd.Short, "Major command should have correct short description")
	suite.Contains(majorCmd.Long, "Commands to increment or decrement", "Major command should have correct long description")

	// Test increment command properties
	suite.Equal("increment", majorIncrementCmd.Use, "Increment command should have correct use")
	suite.Contains(majorIncrementCmd.Aliases, "inc", "Increment command should have 'inc' alias")
	suite.Contains(majorIncrementCmd.Aliases, "+", "Increment command should have '+' alias")
	suite.Equal("Increment major version", majorIncrementCmd.Short, "Increment command should have correct short description")
	suite.Contains(majorIncrementCmd.Long, "Increment the major version and reset minor and patch to 0", "Increment command should explain reset behavior")

	// Test decrement command properties
	suite.Equal("decrement", majorDecrementCmd.Use, "Decrement command should have correct use")
	suite.Contains(majorDecrementCmd.Aliases, "dec", "Decrement command should have 'dec' alias")
	suite.Equal("Decrement major version", majorDecrementCmd.Short, "Decrement command should have correct short description")
	suite.Contains(majorDecrementCmd.Long, "Decrement the major version and reset minor and patch to 0", "Decrement command should explain reset behavior")
}

// TestMajorCommandHierarchy tests that commands are properly registered in the command hierarchy
func (suite *MajorTestSuite) TestMajorCommandHierarchy() {
	// Find major command in root
	var foundMajorCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "major" {
			foundMajorCmd = cmd
			break
		}
	}
	suite.NotNil(foundMajorCmd, "Major command should be registered with root command")

	// Find increment and decrement subcommands
	var foundIncrementCmd, foundDecrementCmd *cobra.Command
	for _, cmd := range foundMajorCmd.Commands() {
		switch cmd.Use {
		case "increment":
			foundIncrementCmd = cmd
		case "decrement":
			foundDecrementCmd = cmd
		}
	}

	suite.NotNil(foundIncrementCmd, "Increment command should be registered with major command")
	suite.NotNil(foundDecrementCmd, "Decrement command should be registered with major command")
}

// TestMajorTestSuite runs the test suite
func TestMajorTestSuite(t *testing.T) {
	suite.Run(t, new(MajorTestSuite))
}