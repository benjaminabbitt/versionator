package cmd

import (
	"bytes"
	"fmt"
	"testing"
	"versionator/internal/app"
	"versionator/internal/vcs/mock"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// SuffixTestSuite contains the test suite for suffix commands
type SuffixTestSuite struct {
	suite.Suite
	testApp      *app.App
	fs           afero.Fs
	restoreApp   func()
	outputBuffer *bytes.Buffer
	ctrl         *gomock.Controller
	mockVCS      *mock.MockVersionControlSystem
}

// SetupTest initializes the test environment before each test
func (suite *SuffixTestSuite) SetupTest() {
	var testApp *app.App
	suite.fs, testApp = createTestApp()
	suite.testApp = testApp
	suite.restoreApp = replaceAppInstance(testApp)
	suite.outputBuffer = &bytes.Buffer{}
	
	// Set up gomock controller and mock VCS
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockVCS = mock.NewMockVersionControlSystem(suite.ctrl)
	suite.testApp.VCS = suite.mockVCS
}

// TearDownTest cleans up after each test
func (suite *SuffixTestSuite) TearDownTest() {
	suite.restoreApp()
	suite.ctrl.Finish()
}

// TestSuffixEnableWithVCS tests enabling suffix when in a VCS repository
func (suite *SuffixTestSuite) TestSuffixEnableWithVCS() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.2.3")

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsRepository().Return(true)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the enable command
	suffixEnableCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix enabled", "Output should show suffix enabled")
	suite.Contains(output, "Current version:", "Output should show current version")

	// Verify config was updated
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.True(cfg.Suffix.Enabled, "Config should have suffix enabled")
	suite.Equal("git", cfg.Suffix.Type, "Config should have git suffix type")
}

// TestSuffixEnableWithoutVCS tests enabling suffix when not in a VCS repository
func (suite *SuffixTestSuite) TestSuffixEnableWithoutVCS() {
	// Setup: Create config and version files
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.0.0")

	// Mock VCS expectations - not in repository
	suite.mockVCS.EXPECT().IsRepository().Return(false)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the enable command
	suffixEnableCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix enabled", "Output should show suffix enabled")
	suite.Contains(output, "Git hash will be added when in a repository", "Output should indicate repository requirement")

	// Verify config was updated
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.True(cfg.Suffix.Enabled, "Config should have suffix enabled")
	suite.Equal("git", cfg.Suffix.Type, "Config should have git suffix type")
}

// TestSuffixEnableWithNilVCS tests enabling suffix when VCS is nil
func (suite *SuffixTestSuite) TestSuffixEnableWithNilVCS() {
	// Setup: Create config and version files with no VCS
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "3.1.0")
	
	// Set VCS to nil
	suite.testApp.VCS = nil

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the enable command
	suffixEnableCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix enabled", "Output should show suffix enabled")
	suite.Contains(output, "Git hash will be added when in a repository", "Output should indicate repository requirement")

	// Verify config was updated
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.True(cfg.Suffix.Enabled, "Config should have suffix enabled")
	suite.Equal("git", cfg.Suffix.Type, "Config should have git suffix type")
}

// TestSuffixDisable tests disabling suffix
func (suite *SuffixTestSuite) TestSuffixDisable() {
	// Setup: Create config and version files, first enable suffix
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.5.2")

	// First enable suffix
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Suffix.Enabled = true
	cfg.Suffix.Type = "git"
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Create and execute the disable command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the disable command
	suffixDisableCmd.Run(cmd, []string{})

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix disabled", "Output should show suffix disabled")
	suite.Contains(output, "Current version: 1.5.2", "Output should show version without suffix")

	// Verify config was updated
	cfg, err = suite.testApp.ReadConfig()
	suite.NoError(err, "Should be able to read config")
	suite.False(cfg.Suffix.Enabled, "Config should have suffix disabled")
}

// TestSuffixStatusEnabledWithVCS tests status when suffix is enabled and in VCS repo
func (suite *SuffixTestSuite) TestSuffixStatusEnabledWithVCS() {
	// Setup: Create config and version files with suffix enabled
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "2.1.0")

	// Enable suffix in config
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Suffix.Enabled = true
	cfg.Suffix.Type = "git"
	cfg.Suffix.Git.HashLength = 7
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Mock VCS expectations
	suite.mockVCS.EXPECT().IsRepository().Return(true)
	suite.mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the status command
	err = suffixStatusCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Status command should succeed")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix: ENABLED", "Output should show suffix enabled")
	suite.Contains(output, "Hash length: 7 characters", "Output should show hash length")
	suite.Contains(output, "Git hash: abc1234", "Output should show git hash")
	suite.Contains(output, "Current version:", "Output should show current version")
}

// TestSuffixStatusEnabledWithoutVCS tests status when suffix is enabled but not in VCS repo
func (suite *SuffixTestSuite) TestSuffixStatusEnabledWithoutVCS() {
	// Setup: Create config and version files with suffix enabled
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "3.0.0")

	// Enable suffix in config
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Suffix.Enabled = true
	cfg.Suffix.Type = "git"
	cfg.Suffix.Git.HashLength = 7
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Mock VCS expectations - not in repository
	suite.mockVCS.EXPECT().IsRepository().Return(false)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the status command
	err = suffixStatusCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Status command should succeed")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix: ENABLED", "Output should show suffix enabled")
	suite.Contains(output, "Hash length: 7 characters", "Output should show hash length")
	suite.Contains(output, "Not in a git repository", "Output should indicate no repository")
}

// TestSuffixStatusDisabled tests status when suffix is disabled
func (suite *SuffixTestSuite) TestSuffixStatusDisabled() {
	// Setup: Create config and version files with suffix disabled
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "1.0.0")

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the status command
	err := suffixStatusCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Status command should succeed")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix: DISABLED", "Output should show suffix disabled")
	suite.Contains(output, "Current version: 1.0.0", "Output should show current version")
}

// TestSuffixStatusVCSError tests status when VCS returns error
func (suite *SuffixTestSuite) TestSuffixStatusVCSError() {
	// Setup: Create config and version files with suffix enabled
	createConfigFile(suite.T(), suite.fs)
	createVersionFile(suite.T(), suite.fs, "4.2.1")

	// Enable suffix in config
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Suffix.Enabled = true
	cfg.Suffix.Type = "git"
	cfg.Suffix.Git.HashLength = 7
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Mock VCS expectations - in repository but error getting hash
	suite.mockVCS.EXPECT().IsRepository().Return(true)
	suite.mockVCS.EXPECT().GetVCSIdentifier(7).Return("", fmt.Errorf("git command failed"))

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the status command
	err = suffixStatusCmd.RunE(cmd, []string{})

	// Verify no error occurred (errors are handled gracefully)
	suite.NoError(err, "Status command should succeed even with VCS errors")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Git hash suffix: ENABLED", "Output should show suffix enabled")
	suite.Contains(output, "cannot get hash: git command failed", "Output should show VCS error")
}

// TestSuffixConfigure tests configure command
func (suite *SuffixTestSuite) TestSuffixConfigure() {
	// Setup: Create config file with specific settings
	createConfigFile(suite.T(), suite.fs)

	// Set specific config values
	cfg, err := suite.testApp.ReadConfig()
	suite.NoError(err)
	cfg.Suffix.Enabled = true
	cfg.Suffix.Type = "git"
	cfg.Suffix.Git.HashLength = 8
	err = suite.testApp.WriteConfig(cfg)
	suite.NoError(err)

	// Create and execute the command
	cmd := &cobra.Command{}
	cmd.SetOut(suite.outputBuffer)

	// Execute the configure command
	err = suffixConfigureCmd.RunE(cmd, []string{})

	// Verify no error occurred
	suite.NoError(err, "Configure command should succeed")

	// Verify the output message
	output := suite.outputBuffer.String()
	suite.Contains(output, "Current configuration:", "Output should show configuration header")
	suite.Contains(output, "Git hash suffix enabled: true", "Output should show enabled status")
	suite.Contains(output, "Suffix type: git", "Output should show suffix type")
	suite.Contains(output, "Git hash length: 8", "Output should show hash length")
	suite.Contains(output, "Configuration is stored in .versionator.yaml", "Output should show config file location")
}

// TestSuffixCommandStructure tests that the commands are properly structured
func (suite *SuffixTestSuite) TestSuffixCommandStructure() {
	// Test suffix command properties
	suite.Equal("suffix", suffixCmd.Use, "Suffix command should have correct use")
	suite.Equal("Manage version suffix behavior", suffixCmd.Short, "Suffix command should have correct short description")
	suite.Contains(suffixCmd.Long, "Commands to enable or disable", "Suffix command should have correct long description")

	// Test enable command properties
	suite.Equal("enable", suffixEnableCmd.Use, "Enable command should have correct use")
	suite.Equal("Enable VCS identifier suffix", suffixEnableCmd.Short, "Enable command should have correct short description")
	suite.Contains(suffixEnableCmd.Long, "Enable appending VCS identifier", "Enable command should have correct long description")

	// Test disable command properties
	suite.Equal("disable", suffixDisableCmd.Use, "Disable command should have correct use")
	suite.Equal("Disable git hash suffix", suffixDisableCmd.Short, "Disable command should have correct short description")
	suite.Equal("Disable appending git hash to version numbers", suffixDisableCmd.Long, "Disable command should have correct long description")

	// Test status command properties
	suite.Equal("status", suffixStatusCmd.Use, "Status command should have correct use")
	suite.Equal("Show suffix configuration status", suffixStatusCmd.Short, "Status command should have correct short description")
	suite.Equal("Show whether git hash suffix is enabled or disabled", suffixStatusCmd.Long, "Status command should have correct long description")

	// Test configure command properties
	suite.Equal("configure", suffixConfigureCmd.Use, "Configure command should have correct use")
	suite.Equal("Configure git hash settings", suffixConfigureCmd.Short, "Configure command should have correct short description")
	suite.Contains(suffixConfigureCmd.Long, "Configure git hash settings", "Configure command should have correct long description")
}

// TestSuffixCommandHierarchy tests that commands are properly registered in the command hierarchy
func (suite *SuffixTestSuite) TestSuffixCommandHierarchy() {
	// Find suffix command in root
	var foundSuffixCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "suffix" {
			foundSuffixCmd = cmd
			break
		}
	}
	suite.NotNil(foundSuffixCmd, "Suffix command should be registered with root command")

	// Find all subcommands
	var foundEnableCmd, foundDisableCmd, foundStatusCmd, foundConfigureCmd *cobra.Command
	for _, cmd := range foundSuffixCmd.Commands() {
		switch cmd.Use {
		case "enable":
			foundEnableCmd = cmd
		case "disable":
			foundDisableCmd = cmd
		case "status":
			foundStatusCmd = cmd
		case "configure":
			foundConfigureCmd = cmd
		}
	}

	suite.NotNil(foundEnableCmd, "Enable command should be registered with suffix command")
	suite.NotNil(foundDisableCmd, "Disable command should be registered with suffix command")
	suite.NotNil(foundStatusCmd, "Status command should be registered with suffix command")
	suite.NotNil(foundConfigureCmd, "Configure command should be registered with suffix command")
}

// TestSuffixTestSuite runs the test suite
func TestSuffixTestSuite(t *testing.T) {
	suite.Run(t, new(SuffixTestSuite))
}