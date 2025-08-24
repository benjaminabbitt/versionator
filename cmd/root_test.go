package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

// RootTestSuite provides a test suite for root command functionality
type RootTestSuite struct {
	suite.Suite
	originalDir string
	tempDir     string
}

// SetupTest runs before each test
func (suite *RootTestSuite) SetupTest() {
	var err error
	suite.originalDir, err = os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")

	suite.tempDir = suite.T().TempDir()
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err, "Failed to change to temp directory")
}

// TearDownTest runs after each test
func (suite *RootTestSuite) TearDownTest() {
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

// createTestFiles creates test files in the temp directory
func (suite *RootTestSuite) createTestFiles(version string, prefix string) {
	// Create VERSION file
	err := afero.WriteFile(afero.NewOsFs(), "VERSION", []byte(version), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Create config file
	configContent := `prefix: "` + prefix + `"
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = afero.WriteFile(afero.NewOsFs(), ".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")
}

func (suite *RootTestSuite) TestExecute_Success() {
	// Create test files
	suite.createTestFiles("1.0.0", "")

	// Test Execute function doesn't panic
	err := Execute()
	// Since Execute() runs the root command without args, it should show help and return nil
	suite.NoError(err, "Execute() should not return error")
}

func (suite *RootTestSuite) TestVersionCommand() {
	// Create test files
	suite.createTestFiles("2.1.0", "")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "version command should succeed")

	// Check output contains the version
	output := buf.String()
	suite.Equal("2.1.0\n", output, "Version command should output correct version")
}

func (suite *RootTestSuite) TestVersionCommand_WithPrefix() {
	// Create test files with prefix
	suite.createTestFiles("3.0.0", "v")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err := rootCmd.Execute()
	suite.Require().NoError(err, "version command should succeed")

	// Check output contains the prefixed version
	output := buf.String()
	suite.Equal("v3.0.0\n", output, "Version command should output prefixed version")
}

func (suite *RootTestSuite) TestVersionCommand_NoVersionFile() {
	// Create only config file (no VERSION file)
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := afero.WriteFile(afero.NewOsFs(), ".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err = rootCmd.Execute()
	suite.Require().NoError(err, "version command should succeed with default version")

	// Check output contains the default version
	output := buf.String()
	suite.Equal("0.0.0\n", output, "Version command should output default version when no VERSION file exists")
}

func (suite *RootTestSuite) TestLogFormatFlag() {
	testCases := []struct {
		name   string
		flag   string
		format string
	}{
		{
			name:   "console format",
			flag:   "--log-format=console",
			format: "console",
		},
		{
			name:   "json format",
			flag:   "--log-format=json",
			format: "json",
		},
		{
			name:   "development format",
			flag:   "--log-format=development",
			format: "development",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create test files
			suite.createTestFiles("1.0.0", "")

			// Capture stdout
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs([]string{tc.flag, "version"})

			// Execute the command with log format flag
			err := rootCmd.Execute()
			suite.NoError(err, "Command with log format flag should succeed")

			// Check that version is still output correctly
			output := buf.String()
			suite.Equal("1.0.0\n", output, "Version should be output correctly regardless of log format")
		})
	}
}

// TestRootTestSuite runs the root test suite
func TestRootTestSuite(t *testing.T) {
	suite.Run(t, new(RootTestSuite))
}