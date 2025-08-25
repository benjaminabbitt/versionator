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

// MinorTestSuite contains the test suite for minor version commands
type MinorTestSuite struct {
	suite.Suite
	testApp      *app.App
	fs           afero.Fs
	restoreApp   func()
	outputBuffer *bytes.Buffer
}

// SetupTest initializes the test environment before each test
func (suite *MinorTestSuite) SetupTest() {
	var testApp *app.App
	suite.fs, testApp = createTestApp()
	suite.testApp = testApp
	suite.restoreApp = replaceAppInstance(testApp)
	suite.outputBuffer = &bytes.Buffer{}
}

// TearDownTest cleans up after each test
func (suite *MinorTestSuite) TearDownTest() {
	suite.restoreApp()
}

// TestMinorIncrementFromDefaultVersion tests incrementing minor version from default 0.0.0
func (suite *MinorTestSuite) TestMinorIncrementFromDefaultVersion() {
	// Setup: Create config file (VERSION file will be auto-created)
	createConfigFile(suite.T(), suite.fs)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the increment command
	err := minorIncrementCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Minor increment should succeed")

	// Verify the version was incremented correctly
	verifyVersionFile(suite.T(), suite.fs, "0.1.0")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Minor version incremented to: 0.1.0", "Output should show incremented version")
}

// TestMinorIncrementFromExistingVersion tests incrementing minor version from an existing version
func (suite *MinorTestSuite) TestMinorIncrementFromExistingVersion() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the increment command
	err := minorIncrementCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Minor increment should succeed")

	// Verify the version was incremented correctly (patch should reset to 0)
	verifyVersionFile(suite.T(), suite.fs, "2.6.0")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Minor version incremented to: 2.6.0", "Output should show incremented version")
}

// TestMinorDecrementFromExistingVersion tests decrementing minor version from an existing version
func (suite *MinorTestSuite) TestMinorDecrementFromExistingVersion() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the decrement command
	err := minorDecrementCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Minor decrement should succeed")

	// Verify the version was decremented correctly (patch should reset to 0)
	verifyVersionFile(suite.T(), suite.fs, "2.4.0")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Minor version decremented to: 2.4.0", "Output should show decremented version")
}

// TestMinorDecrementFromZeroMinorVersion tests error when trying to decrement from 0 minor version
func (suite *MinorTestSuite) TestMinorDecrementFromZeroMinorVersion() {
	// Setup: Create config and version files with minor version = 0
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.0.5")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the decrement command
	err := minorDecrementCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Minor decrement from 0 should fail")
	suite.Contains(err.Error(), "cannot decrement minor version below 0", "Error should indicate minor version cannot go below 0")

	// Verify the version file wasn't changed
	verifyVersionFile(suite.T(), suite.fs, "1.0.5")
}

// TestMinorDecrementFromDefaultVersion tests error when trying to decrement from default 0.0.0
func (suite *MinorTestSuite) TestMinorDecrementFromDefaultVersion() {
	// Setup: Create config file only (VERSION file will be auto-created as 0.0.0)
	createConfigFile(suite.T(), suite.fs)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the decrement command
	err := minorDecrementCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Minor decrement from default version should fail")
	suite.Contains(err.Error(), "cannot decrement minor version below 0", "Error should indicate minor version cannot go below 0")
}

// TestMinorIncrementWithAppInstanceError tests handling of appInstance errors during increment
func (suite *MinorTestSuite) TestMinorIncrementWithAppInstanceError() {
	// Setup: Create config but no version file and make filesystem read-only
	createConfigFile(suite.T(), suite.fs)

	// Make filesystem read-only by replacing with a filesystem that errors on write
	suite.testApp.FileSystem = afero.NewReadOnlyFs(suite.fs)
	suite.testApp.VersionManager = version.NewVersion(suite.testApp.FileSystem, ".", nil)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the increment command
	err := minorIncrementCmd.RunE(cmd, []string{})

	// Verify error occurred
	suite.Error(err, "Increment should fail with read-only filesystem")
	suite.Contains(err.Error(), "error incrementing minor version", "Error should indicate increment failure")
}

// TestMinorCommandHierarchy tests that commands are properly registered in the command hierarchy
func (suite *MinorTestSuite) TestMinorCommandHierarchy() {
	// Find minor command in root
	var foundMinorCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "minor" {
			foundMinorCmd = cmd
			break
		}
	}
	suite.NotNil(foundMinorCmd, "Minor command should be registered with root command")

	// Find increment and decrement subcommands
	var foundIncrementCmd, foundDecrementCmd *cobra.Command
	for _, cmd := range foundMinorCmd.Commands() {
		switch cmd.Use {
		case "increment":
			foundIncrementCmd = cmd
		case "decrement":
			foundDecrementCmd = cmd
		}
	}

	suite.NotNil(foundIncrementCmd, "Increment command should be registered with minor command")
	suite.NotNil(foundDecrementCmd, "Decrement command should be registered with minor command")
}

// TestMinorTestSuite runs the test suite
func TestMinorTestSuite(t *testing.T) {
	suite.Run(t, new(MinorTestSuite))
}
