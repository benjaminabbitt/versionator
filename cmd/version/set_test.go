package version

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/filesystem"
	fstesting "github.com/benjaminabbitt/versionator/internal/filesystem/testing"
	"github.com/stretchr/testify/suite"
)

// SetTestSuite defines the test suite for set command tests
type SetTestSuite struct {
	suite.Suite
	memFs   *fstesting.MemFs
	cleanup func()
	cwd     string
}

// SetupTest runs before each test
func (suite *SetTestSuite) SetupTest() {
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	suite.cwd = cwd

	suite.memFs, suite.cleanup = fstesting.SetupTestFs()
}

// TearDownTest runs after each test
func (suite *SetTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// absPath returns absolute path for a filename relative to cwd
func (suite *SetTestSuite) absPath(filename string) string {
	return filepath.Join(suite.cwd, filename)
}

// createTestFiles creates the standard test files needed for most tests
func (suite *SetTestSuite) createTestFiles(version string) {
	// VERSION uses absolute path via getVersionPath()
	err := filesystem.AppFs.WriteFile(suite.absPath("VERSION"), []byte(version+"\n"), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Config uses relative path in config.ReadConfig()
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
	err = filesystem.AppFs.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

// readConfig reads the config using the config package (ensures path consistency)
func (suite *SetTestSuite) readConfig() *config.Config {
	cfg, err := config.ReadConfig()
	suite.Require().NoError(err, "Failed to read config file")
	return cfg
}

func (suite *SetTestSuite) TestSetCommand_ThreeDigitVersion() {
	suite.createTestFiles("v0.0.1")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"1.2.3"})
	suite.Require().NoError(err, "set command should succeed")

	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.2.3", strings.TrimSpace(string(content)), "VERSION should contain '1.2.3'")

	// DotNet should not be enabled for 3-digit version
	cfg := suite.readConfig()
	suite.False(cfg.DotNet, "DotNet mode should not be enabled for 3-digit version")
}

func (suite *SetTestSuite) TestSetCommand_FourDigitVersion() {
	suite.createTestFiles("v0.0.1")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"1.2.3.4"})
	suite.Require().NoError(err, "set command should succeed")

	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.2.3.4", strings.TrimSpace(string(content)), "VERSION should contain '1.2.3.4'")

	// DotNet should be enabled for 4-digit version
	cfg := suite.readConfig()
	suite.True(cfg.DotNet, "DotNet mode should be enabled for 4-digit version")
}

func (suite *SetTestSuite) TestSetCommand_WithPrefix() {
	suite.createTestFiles("0.0.1")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"release-2.0.0"})
	suite.Require().NoError(err, "set command should succeed")

	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("release-2.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'release-2.0.0'")

	// Config prefix should be updated
	cfg := suite.readConfig()
	suite.Equal("release-", cfg.Prefix, "Config prefix should be updated to 'release-'")
}

func (suite *SetTestSuite) TestSetCommand_WithVPrefix() {
	suite.createTestFiles("0.0.1")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"v3.0.0"})
	suite.Require().NoError(err, "set command should succeed")

	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v3.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'v3.0.0'")

	cfg := suite.readConfig()
	suite.Equal("v", cfg.Prefix, "Config prefix should be updated to 'v'")
}

func (suite *SetTestSuite) TestSetCommand_FourDigitWithPrefix() {
	suite.createTestFiles("0.0.1")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"v1.0.0.0"})
	suite.Require().NoError(err, "set command should succeed")

	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("v1.0.0.0", strings.TrimSpace(string(content)), "VERSION should contain 'v1.0.0.0'")

	cfg := suite.readConfig()
	suite.Equal("v", cfg.Prefix, "Config prefix should be 'v'")
	suite.True(cfg.DotNet, "DotNet mode should be enabled for 4-digit version")
}

func (suite *SetTestSuite) TestSetCommand_InvalidVersion() {
	suite.createTestFiles("v0.0.1")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"invalid"})
	suite.Error(err, "set command should fail for invalid version")
}

func (suite *SetTestSuite) TestSetCommand_NoArgs() {
	suite.createTestFiles("v0.0.1")

	err := SetCmd.Args(SetCmd, []string{})
	suite.Error(err, "set command should fail without arguments")
}

func (suite *SetTestSuite) TestSetCommand_ClearsPrefix() {
	suite.createTestFiles("v1.0.0")

	var stdout bytes.Buffer
	SetCmd.SetOut(&stdout)
	err := SetCmd.RunE(SetCmd, []string{"2.0.0"})
	suite.Require().NoError(err, "set command should succeed")

	content, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("2.0.0", strings.TrimSpace(string(content)), "VERSION should contain '2.0.0' without prefix")

	cfg := suite.readConfig()
	suite.Equal("", cfg.Prefix, "Config prefix should be empty when version has no prefix")
}

// TestSetTestSuite runs the set test suite
func TestSetTestSuite(t *testing.T) {
	suite.Run(t, new(SetTestSuite))
}
