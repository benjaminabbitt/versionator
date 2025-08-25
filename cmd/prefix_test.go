package cmd

import (
	"bytes"
	"testing"
	"versionator/internal/app"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// PrefixTestSuite contains the test suite for prefix commands
type PrefixTestSuite struct {
	suite.Suite
	testApp      *app.App
	fs           afero.Fs
	restoreApp   func()
	outputBuffer *bytes.Buffer
}

// SetupTest initializes the test environment before each test
func (suite *PrefixTestSuite) SetupTest() {
	var testApp *app.App
	suite.fs, testApp = createTestApp()
	suite.testApp = testApp
	suite.restoreApp = replaceAppInstance(testApp)
	suite.outputBuffer = &bytes.Buffer{}
}

// TearDownTest cleans up after each test
func (suite *PrefixTestSuite) TearDownTest() {
	suite.restoreApp()
}

// TestPrefixEnable tests enabling version prefix with default 'v'
func (suite *PrefixTestSuite) TestPrefixEnable() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.2.3")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the enable command
	prefixEnableCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Version prefix enabled with default value 'v'", "Output should show prefix enabled")
	suite.Contains(output, "Current version: v1.2.3", "Output should show version with prefix")

	// Verify config was updated
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.Equal("v", cfg.Prefix, "Config should have 'v' prefix")
}

// TestPrefixDisable tests disabling version prefix
func (suite *PrefixTestSuite) TestPrefixDisable() {
	// Setup: Create config and version files, first enable prefix
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.0.0")

	// First enable prefix
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Prefix = "v"
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Create and execute the disable command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the disable command
	prefixDisableCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Version prefix disabled", "Output should show prefix disabled")
	suite.Contains(output, "Current version: 2.0.0", "Output should show version without prefix")

	// Verify config was updated
	cfg, err = suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.Equal("", cfg.Prefix, "Config should have empty prefix")
}

// TestPrefixSetCustomValue tests setting a custom prefix value
func (suite *PrefixTestSuite) TestPrefixSetCustomValue() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "3.1.0")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the set command with custom prefix
	prefixSetCmd.Run(cmd, []string{"release-"})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Version prefix set to: release-", "Output should show custom prefix set")
	suite.Contains(output, "Current version: release-3.1.0", "Output should show version with custom prefix")

	// Verify config was updated
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.Equal("release-", cfg.Prefix, "Config should have custom prefix")
}

// TestPrefixSetEmptyValue tests setting prefix to empty string via set command
func (suite *PrefixTestSuite) TestPrefixSetEmptyValue() {
	// Setup: Create config and version files with existing prefix
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.5.2")

	// First set a prefix
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Prefix = "v"
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the set command with empty prefix
	prefixSetCmd.Run(cmd, []string{""})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Version prefix disabled (set to empty)", "Output should show prefix disabled")
	suite.Contains(output, "Current version: 1.5.2", "Output should show version without prefix")

	// Verify config was updated
	cfg, err = suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.Equal("", cfg.Prefix, "Config should have empty prefix")
}

// TestPrefixStatusWithPrefix tests status command when prefix is enabled
func (suite *PrefixTestSuite) TestPrefixStatusWithPrefix() {
	// Setup: Create config and version files with prefix
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "4.0.1")

	// Set a prefix in config
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Prefix = "app-"
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the status command
	prefixStatusCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Version prefix: ENABLED", "Output should show prefix enabled")
	suite.Contains(output, "Prefix value: app-", "Output should show prefix value")
	suite.Contains(output, "Current version: app-4.0.1", "Output should show version with prefix")
}

// TestPrefixStatusWithoutPrefix tests status command when prefix is disabled
func (suite *PrefixTestSuite) TestPrefixStatusWithoutPrefix() {
	// Setup: Create config and version files without prefix
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "0.9.0")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the status command
	prefixStatusCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Version prefix: DISABLED", "Output should show prefix disabled")
	suite.Contains(output, "Current version: 0.9.0", "Output should show version without prefix")
}

// TestPrefixCommandStructure tests that the commands are properly structured
func (suite *PrefixTestSuite) TestPrefixCommandStructure() {
	// Test prefix command properties
	suite.Equal("prefix", prefixCmd.Use, "Prefix command should have correct use")
	suite.Equal("Manage version prefix behavior", prefixCmd.Short, "Prefix command should have correct short description")
	suite.Contains(prefixCmd.Long, "Commands to enable, disable, or set", "Prefix command should have correct long description")

	// Test enable command properties
	suite.Equal("enable", prefixEnableCmd.Use, "Enable command should have correct use")
	suite.Equal("Enable version prefix", prefixEnableCmd.Short, "Enable command should have correct short description")
	suite.Contains(prefixEnableCmd.Long, "Enable version prefix with default value 'v'", "Enable command should have correct long description")

	// Test disable command properties
	suite.Equal("disable", prefixDisableCmd.Use, "Disable command should have correct use")
	suite.Equal("Disable version prefix", prefixDisableCmd.Short, "Disable command should have correct short description")
	suite.Contains(prefixDisableCmd.Long, "Disable version prefix by setting it to empty string", "Disable command should have correct long description")

	// Test set command properties
	suite.Equal("set <prefix>", prefixSetCmd.Use, "Set command should have correct use")
	suite.Equal("Set version prefix", prefixSetCmd.Short, "Set command should have correct short description")
	suite.Equal("Set a custom version prefix", prefixSetCmd.Long, "Set command should have correct long description")

	// Test status command properties
	suite.Equal("status", prefixStatusCmd.Use, "Status command should have correct use")
	suite.Equal("Show prefix configuration status", prefixStatusCmd.Short, "Status command should have correct short description")
	suite.Equal("Show current version prefix configuration", prefixStatusCmd.Long, "Status command should have correct long description")
}

// TestPrefixCommandHierarchy tests that commands are properly registered in the command hierarchy
func (suite *PrefixTestSuite) TestPrefixCommandHierarchy() {
	// Find prefix command in root
	var foundPrefixCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "prefix" {
			foundPrefixCmd = cmd
			break
		}
	}
	suite.NotNil(foundPrefixCmd, "Prefix command should be registered with root command")

	// Find all subcommands
	var foundEnableCmd, foundDisableCmd, foundSetCmd, foundStatusCmd *cobra.Command
	for _, cmd := range foundPrefixCmd.Commands() {
		switch cmd.Use {
		case "enable":
			foundEnableCmd = cmd
		case "disable":
			foundDisableCmd = cmd
		case "set <prefix>":
			foundSetCmd = cmd
		case "status":
			foundStatusCmd = cmd
		}
	}

	suite.NotNil(foundEnableCmd, "Enable command should be registered with prefix command")
	suite.NotNil(foundDisableCmd, "Disable command should be registered with prefix command")
	suite.NotNil(foundSetCmd, "Set command should be registered with prefix command")
	suite.NotNil(foundStatusCmd, "Status command should be registered with prefix command")
}

// TestPrefixSetCommandArgs tests that set command requires exactly one argument
func (suite *PrefixTestSuite) TestPrefixSetCommandArgs() {
	// Verify Args validation - test that it's set (function pointers can't be compared)
	suite.NotNil(prefixSetCmd.Args, "Set command should have Args validation set")
}

// TestPrefixTestSuite runs the test suite
func TestPrefixTestSuite(t *testing.T) {
	suite.Run(t, new(PrefixTestSuite))
}