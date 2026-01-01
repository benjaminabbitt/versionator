package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/filesystem"
	fstesting "github.com/benjaminabbitt/versionator/internal/filesystem/testing"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite defines the test suite for config command tests
type ConfigTestSuite struct {
	suite.Suite
	memFs   *fstesting.MemFs
	cleanup func()
}

// SetupTest runs before each test
func (suite *ConfigTestSuite) SetupTest() {
	suite.memFs, suite.cleanup = fstesting.SetupTestFs()
}

// TearDownTest runs after each test
func (suite *ConfigTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// createDefaultConfig creates a default config file
func (suite *ConfigTestSuite) createDefaultConfig() {
	configContent := `prefix: "v"
prerelease:
  elements: []
metadata:
  elements: []
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

// --- Prefix tests ---

func (suite *ConfigTestSuite) TestConfigPrefixGet_Default() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configPrefixGetCmd.SetOut(&stdout)
	err := configPrefixGetCmd.RunE(configPrefixGetCmd, []string{})
	suite.Require().NoError(err)
	suite.Equal("v\n", stdout.String())
}

func (suite *ConfigTestSuite) TestConfigPrefixGet_Empty() {
	configContent := `prefix: ""
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err)

	var stdout bytes.Buffer
	configPrefixGetCmd.SetOut(&stdout)
	err = configPrefixGetCmd.RunE(configPrefixGetCmd, []string{})
	suite.Require().NoError(err)
	suite.Equal("(empty)\n", stdout.String())
}

func (suite *ConfigTestSuite) TestConfigPrefixSet() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configPrefixSetCmd.SetOut(&stdout)
	err := configPrefixSetCmd.RunE(configPrefixSetCmd, []string{"release-"})
	suite.Require().NoError(err)
	suite.Contains(stdout.String(), "release-")

	// Verify the change persisted
	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.Equal("release-", cfg.Prefix)
}

// --- Prerelease tests ---

func (suite *ConfigTestSuite) TestConfigPrereleaseGet_Empty() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configPrereleaseGetCmd.SetOut(&stdout)
	err := configPrereleaseGetCmd.RunE(configPrereleaseGetCmd, []string{})
	suite.Require().NoError(err)
	suite.Equal("(none)\n", stdout.String())
}

func (suite *ConfigTestSuite) TestConfigPrereleaseSet() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configPrereleaseSetCmd.SetOut(&stdout)
	err := configPrereleaseSetCmd.RunE(configPrereleaseSetCmd, []string{"alpha", "CommitsSinceTag"})
	suite.Require().NoError(err)
	suite.Contains(stdout.String(), "alpha, CommitsSinceTag")

	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.Equal([]string{"alpha", "CommitsSinceTag"}, cfg.PreRelease.Elements)
}

func (suite *ConfigTestSuite) TestConfigPrereleaseClear() {
	// First set some elements
	configContent := `prefix: "v"
prerelease:
  elements: ["alpha", "beta"]
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err)

	var stdout bytes.Buffer
	configPrereleaseClearCmd.SetOut(&stdout)
	err = configPrereleaseClearCmd.RunE(configPrereleaseClearCmd, []string{})
	suite.Require().NoError(err)
	suite.Contains(stdout.String(), "cleared")

	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.Empty(cfg.PreRelease.Elements)
}

// --- Metadata tests ---

func (suite *ConfigTestSuite) TestConfigMetadataGet_Empty() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configMetadataGetCmd.SetOut(&stdout)
	err := configMetadataGetCmd.RunE(configMetadataGetCmd, []string{})
	suite.Require().NoError(err)
	suite.Equal("(none)\n", stdout.String())
}

func (suite *ConfigTestSuite) TestConfigMetadataSet() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configMetadataSetCmd.SetOut(&stdout)
	err := configMetadataSetCmd.RunE(configMetadataSetCmd, []string{"ShortHash", "BuildDateTimeCompact"})
	suite.Require().NoError(err)
	suite.Contains(stdout.String(), "ShortHash, BuildDateTimeCompact")

	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.Equal([]string{"ShortHash", "BuildDateTimeCompact"}, cfg.Metadata.Elements)
}

func (suite *ConfigTestSuite) TestConfigMetadataClear() {
	configContent := `prefix: "v"
metadata:
  elements: ["build", "hash"]
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err)

	var stdout bytes.Buffer
	configMetadataClearCmd.SetOut(&stdout)
	err = configMetadataClearCmd.RunE(configMetadataClearCmd, []string{})
	suite.Require().NoError(err)
	suite.Contains(stdout.String(), "cleared")

	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.Empty(cfg.Metadata.Elements)
}

// --- DotNet tests ---

func (suite *ConfigTestSuite) TestConfigDotNetStatus_Disabled() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configDotNetStatusCmd.SetOut(&stdout)
	err := configDotNetStatusCmd.RunE(configDotNetStatusCmd, []string{})
	suite.Require().NoError(err)
	output := stdout.String()
	suite.Contains(output, "DISABLED")
	suite.Contains(output, "3-component")
}

func (suite *ConfigTestSuite) TestConfigDotNetStatus_Enabled() {
	configContent := `prefix: "v"
dotnet: true
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err)

	var stdout bytes.Buffer
	configDotNetStatusCmd.SetOut(&stdout)
	err = configDotNetStatusCmd.RunE(configDotNetStatusCmd, []string{})
	suite.Require().NoError(err)
	output := stdout.String()
	suite.Contains(output, "ENABLED")
	suite.Contains(output, "4-component")
}

func (suite *ConfigTestSuite) TestConfigDotNetEnable() {
	suite.createDefaultConfig()

	var stdout bytes.Buffer
	configDotNetEnableCmd.SetOut(&stdout)
	err := configDotNetEnableCmd.RunE(configDotNetEnableCmd, []string{})
	suite.Require().NoError(err)
	suite.Contains(strings.ToLower(stdout.String()), "enabled")

	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.True(cfg.DotNet)
}

func (suite *ConfigTestSuite) TestConfigDotNetDisable() {
	configContent := `prefix: "v"
dotnet: true
logging:
  output: "console"
`
	err := filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err)

	var stdout bytes.Buffer
	configDotNetDisableCmd.SetOut(&stdout)
	err = configDotNetDisableCmd.RunE(configDotNetDisableCmd, []string{})
	suite.Require().NoError(err)
	suite.Contains(strings.ToLower(stdout.String()), "disabled")

	cfg, err := config.ReadConfig()
	suite.Require().NoError(err)
	suite.False(cfg.DotNet)
}

// TestConfigTestSuite runs the config test suite
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// Ensure unused import doesn't cause issues
var _ = os.Stdout
