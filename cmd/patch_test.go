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

// PatchTestSuite contains the test suite for patch version commands
type PatchTestSuite struct {
	suite.Suite
	testApp      *app.App
	fs           afero.Fs
	restoreApp   func()
	outputBuffer *bytes.Buffer
}

// SetupTest initializes the test environment before each test
func (suite *PatchTestSuite) SetupTest() {
	var testApp *app.App
	suite.fs, testApp = createTestApp()
	suite.testApp = testApp
	suite.restoreApp = replaceAppInstance(testApp)
	suite.outputBuffer = &bytes.Buffer{}
}

// TearDownTest cleans up after each test
func (suite *PatchTestSuite) TearDownTest() {
	suite.restoreApp()
}

// TestPatchIncrementFromDefaultVersion tests incrementing patch version from default 0.0.0
func (suite *PatchTestSuite) TestPatchIncrementFromDefaultVersion() {
	// Setup: Create config file (VERSION file will be auto-created)
	createConfigFile(suite.T(), suite.fs)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)
	
	// Execute the increment command
	err := patchIncrementCmd.RunE(cmd, []string{})
	
	// Verify no error occurred
	suite.NoError(err, "Patch increment should succeed")
	
	// Verify the version was incremented correctly
	verifyVersionFile(suite.T(), suite.fs, "0.0.1")
	
	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Patch version incremented to: 0.0.1", "Output should show incremented version")
}

// TestPatchIncrementFromExistingVersion tests incrementing patch version from an existing version
func (suite *PatchTestSuite) TestPatchIncrementFromExistingVersion() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)
	
	// Execute the increment command
	err := patchIncrementCmd.RunE(cmd, []string{})
	
	// Verify no error occurred
	suite.NoError(err, "Patch increment should succeed")
	
	// Verify the version was incremented correctly
	verifyVersionFile(suite.T(), suite.fs, "2.5.4")
	
	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Patch version incremented to: 2.5.4", "Output should show incremented version")
}

// TestPatchDecrementFromExistingVersion tests decrementing patch version from an existing version
func (suite *PatchTestSuite) TestPatchDecrementFromExistingVersion() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.5.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)
	
	// Execute the decrement command
	err := patchDecrementCmd.RunE(cmd, []string{})
	
	// Verify no error occurred
	suite.NoError(err, "Patch decrement should succeed")
	
	// Verify the version was decremented correctly
	verifyVersionFile(suite.T(), suite.fs, "2.5.2")
	
	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Patch version decremented to: 2.5.2", "Output should show decremented version")
}

// TestPatchDecrementFromZeroPatchVersion tests error when trying to decrement from 0 patch version
func (suite *PatchTestSuite) TestPatchDecrementFromZeroPatchVersion() {
	// Setup: Create config and version files with patch version = 0
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.2.0")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)
	
	// Execute the decrement command
	err := patchDecrementCmd.RunE(cmd, []string{})
	
	// Verify error occurred
	suite.Error(err, "Patch decrement from 0 should fail")
	suite.Contains(err.Error(), "cannot decrement patch version below 0", "Error should indicate patch version cannot go below 0")
	
	// Verify the version file wasn't changed
	verifyVersionFile(suite.T(), suite.fs, "1.2.0")
}

// TestPatchDecrementFromDefaultVersion tests error when trying to decrement from default 0.0.0
func (suite *PatchTestSuite) TestPatchDecrementFromDefaultVersion() {
	// Setup: Create config file only (VERSION file will be auto-created as 0.0.0)
	createConfigFile(suite.T(), suite.fs)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)
	
	// Execute the decrement command
	err := patchDecrementCmd.RunE(cmd, []string{})
	
	// Verify error occurred
	suite.Error(err, "Patch decrement from default version should fail")
	suite.Contains(err.Error(), "cannot decrement patch version below 0", "Error should indicate patch version cannot go below 0")
}

// TestPatchIncrementWithAppInstanceError tests handling of appInstance errors during increment
func (suite *PatchTestSuite) TestPatchIncrementWithAppInstanceError() {
	// Setup: Create config but no version file and make filesystem read-only
	createConfigFile(suite.T(), suite.fs)
	
	// Make filesystem read-only by replacing with a filesystem that errors on write
	suite.testApp.FileSystem = afero.NewReadOnlyFs(suite.fs)
	suite.testApp.VersionManager = version.NewVersion(suite.testApp.FileSystem, ".", nil)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)
	
	// Execute the increment command
	err := patchIncrementCmd.RunE(cmd, []string{})
	
	// Verify error occurred
	suite.Error(err, "Increment should fail with read-only filesystem")
	suite.Contains(err.Error(), "error incrementing patch version", "Error should indicate increment failure")
}

// TestPatchCommandStructure tests that the commands are properly structured
func (suite *PatchTestSuite) TestPatchCommandStructure() {
	// Test patch command properties
	suite.Equal("patch", patchCmd.Use, "Patch command should have correct use")
	suite.Equal("Manage patch version", patchCmd.Short, "Patch command should have correct short description")
	suite.Contains(patchCmd.Long, "Commands to increment or decrement", "Patch command should have correct long description")

	// Test increment command properties
	suite.Equal("increment", patchIncrementCmd.Use, "Increment command should have correct use")
	suite.Contains(patchIncrementCmd.Aliases, "inc", "Increment command should have 'inc' alias")
	suite.Contains(patchIncrementCmd.Aliases, "+", "Increment command should have '+' alias")
	suite.Equal("Increment patch version", patchIncrementCmd.Short, "Increment command should have correct short description")
	suite.Equal("Increment the patch version", patchIncrementCmd.Long, "Increment command should have correct long description")

	// Test decrement command properties
	suite.Equal("decrement", patchDecrementCmd.Use, "Decrement command should have correct use")
	suite.Contains(patchDecrementCmd.Aliases, "dec", "Decrement command should have 'dec' alias")
	suite.Equal("Decrement patch version", patchDecrementCmd.Short, "Decrement command should have correct short description")
	suite.Equal("Decrement the patch version", patchDecrementCmd.Long, "Decrement command should have correct long description")
}

// TestPatchCommandHierarchy tests that commands are properly registered in the command hierarchy
func (suite *PatchTestSuite) TestPatchCommandHierarchy() {
	// Find patch command in root
	var foundPatchCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "patch" {
			foundPatchCmd = cmd
			break
		}
	}
	suite.NotNil(foundPatchCmd, "Patch command should be registered with root command")
	
	// Find increment and decrement subcommands
	var foundIncrementCmd, foundDecrementCmd *cobra.Command
	for _, cmd := range foundPatchCmd.Commands() {
		switch cmd.Use {
		case "increment":
			foundIncrementCmd = cmd
		case "decrement":
			foundDecrementCmd = cmd
		}
	}
	
	suite.NotNil(foundIncrementCmd, "Increment command should be registered with patch command")
	suite.NotNil(foundDecrementCmd, "Decrement command should be registered with patch command")
}

// TestPatchTestSuite runs the test suite
func TestPatchTestSuite(t *testing.T) {
	suite.Run(t, new(PatchTestSuite))
}