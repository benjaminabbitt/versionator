package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"versionator/internal/config"
)

// PrefixTestSuite provides a test suite for prefix commands
type PrefixTestSuite struct {
	suite.Suite
	originalDir string
	tempDir     string
	fs          afero.Fs
}

// SetupTest runs before each test
func (suite *PrefixTestSuite) SetupTest() {
	// Create an in-memory filesystem for testing
	suite.fs = afero.NewMemMapFs()
	
	var err error
	suite.originalDir, err = os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")

	suite.tempDir = suite.T().TempDir()
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err, "Failed to change to temp directory")
}

// TearDownTest runs after each test
func (suite *PrefixTestSuite) TearDownTest() {
	// Reset command state
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Change back to original directory
	if suite.originalDir != "" {
		err := os.Chdir(suite.originalDir)
		suite.Require().NoError(err, "Failed to restore original directory")
	}
}

// createTestFiles creates test files with the given version and config
func (suite *PrefixTestSuite) createTestFiles(version string, cfg *config.Config) {
	// Create VERSION file
	err := afero.WriteFile(suite.fs, "VERSION", []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create config file
	configData, err := yaml.Marshal(cfg)
	suite.Require().NoError(err, "Failed to marshal config")
	err = afero.WriteFile(suite.fs, ".versionator.yaml", configData, 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

func (suite *PrefixTestSuite) TestPrefixEnableCommand_DefaultConfig() {
	initialConfig := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("1.2.3", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "enable"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix enable should succeed")

	// Verify config was updated
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("v", cfg.Prefix, "Prefix should be set to 'v'")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix enabled with default value 'v'")
	suite.Contains(output, "Current version: v1.2.3")
}

func (suite *PrefixTestSuite) TestPrefixEnableCommand_WhenAlreadyEnabled() {
	initialConfig := &config.Config{
		Prefix: "release-",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("2.0.0", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "enable"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix enable should succeed")

	// Verify config was updated to default 'v'
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("v", cfg.Prefix, "Prefix should be reset to 'v'")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix enabled with default value 'v'")
	suite.Contains(output, "Current version: v2.0.0")
}

func (suite *PrefixTestSuite) TestPrefixDisableCommand_WhenEnabled() {
	initialConfig := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("1.2.3", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "disable"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix disable should succeed")

	// Verify config was updated
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("", cfg.Prefix, "Prefix should be empty")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix disabled")
	suite.Contains(output, "Current version: 1.2.3")
}

func (suite *PrefixTestSuite) TestPrefixDisableCommand_WhenAlreadyDisabled() {
	initialConfig := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("2.0.0", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "disable"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix disable should succeed")

	// Verify config remains empty
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("", cfg.Prefix, "Prefix should remain empty")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix disabled")
	suite.Contains(output, "Current version: 2.0.0")
}

func (suite *PrefixTestSuite) TestPrefixSetCommand_CustomPrefix() {
	initialConfig := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("1.2.3", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "set", "release-"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix set should succeed")

	// Verify config was updated
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("release-", cfg.Prefix, "Prefix should be 'release-'")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix set to: release-")
	suite.Contains(output, "Current version: release-1.2.3")
}

func (suite *PrefixTestSuite) TestPrefixSetCommand_EmptyPrefix() {
	initialConfig := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("2.0.0", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "set", ""})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix set should succeed")

	// Verify config was updated
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("", cfg.Prefix, "Prefix should be empty")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix disabled (set to empty)")
	suite.Contains(output, "Current version: 2.0.0")
}

func (suite *PrefixTestSuite) TestPrefixSetCommand_SpecialCharacters() {
	initialConfig := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("1.0.0", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "set", "v2.0-"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix set should succeed")

	// Verify config was updated
	cm := config.NewConfigManager(suite.fs)
	cfg, err := cm.ReadConfig()
	suite.Require().NoError(err, "Should be able to read config")
	suite.Equal("v2.0-", cfg.Prefix, "Prefix should be 'v2.0-'")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix set to: v2.0-")
	suite.Contains(output, "Current version: v2.0-1.0.0")
}

func (suite *PrefixTestSuite) TestPrefixSetCommand_NoArgument() {
	initialConfig := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("1.0.0", initialConfig)

	// Capture output
	var stderr bytes.Buffer
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"prefix", "set"})

	// Execute command
	err := rootCmd.Execute()
	suite.Error(err, "prefix set without argument should fail")
	suite.Contains(err.Error(), "accepts 1 arg(s), received 0")
}

func (suite *PrefixTestSuite) TestPrefixStatusCommand_Enabled() {
	initialConfig := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("1.2.3", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "status"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix status should succeed")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix: ENABLED")
	suite.Contains(output, "Prefix value: v")
	suite.Contains(output, "Current version: v1.2.3")
}

func (suite *PrefixTestSuite) TestPrefixStatusCommand_Disabled() {
	initialConfig := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("2.0.0", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "status"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix status should succeed")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix: DISABLED")
	suite.Contains(output, "Current version: 2.0.0")
}

func (suite *PrefixTestSuite) TestPrefixStatusCommand_CustomPrefix() {
	initialConfig := &config.Config{
		Prefix: "release-",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}

	// Create test files
	suite.createTestFiles("3.1.0", initialConfig)

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"prefix", "status"})

	// Execute command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "prefix status should succeed")

	// Verify output
	output := stdout.String()
	suite.Contains(output, "Version prefix: ENABLED")
	suite.Contains(output, "Prefix value: release-")
	suite.Contains(output, "Current version: release-3.1.0")
}

// TestPrefixTestSuite runs the prefix test suite
func TestPrefixTestSuite(t *testing.T) {
	suite.Run(t, new(PrefixTestSuite))
}