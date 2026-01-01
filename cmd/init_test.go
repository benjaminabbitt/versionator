package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/filesystem"
	fstesting "github.com/benjaminabbitt/versionator/internal/filesystem/testing"
	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/stretchr/testify/suite"
)

// InitTestSuite defines the test suite for init command tests
type InitTestSuite struct {
	suite.Suite
	memFs     *fstesting.MemFs
	fsCleanup func()
	cwd       string
}

// SetupTest runs before each test
func (suite *InitTestSuite) SetupTest() {
	// Get current working directory for absolute paths
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	suite.cwd = cwd

	// Set up in-memory filesystem
	suite.memFs, suite.fsCleanup = fstesting.SetupTestFs()

	// Reset command state
	suite.resetInitCommand()
}

// TearDownTest runs after each test
func (suite *InitTestSuite) TearDownTest() {
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Restore real filesystem
	if suite.fsCleanup != nil {
		suite.fsCleanup()
	}
}

// absPath returns absolute path for a filename relative to cwd
func (suite *InitTestSuite) absPath(filename string) string {
	return filepath.Join(suite.cwd, filename)
}

// resetInitCommand resets command state between tests
func (suite *InitTestSuite) resetInitCommand() {
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)

	// Reset init command flags
	initCmd.Flags().Set("go", "false")

	// Reset verbosity
	logging.ResetVerbosity()
	verboseCount = 0
}

func (suite *InitTestSuite) TestInitCommand_CreatesVersionFile() {
	// Verify VERSION doesn't exist
	_, err := filesystem.AppFs.Stat(suite.absPath("VERSION"))
	suite.True(os.IsNotExist(err), "VERSION file should not exist before init")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify VERSION was created
	data, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "VERSION file should exist after init")

	content := strings.TrimSpace(string(data))
	suite.True(content == "0.0.0" || content == "v0.0.0",
		"VERSION should be 0.0.0 or v0.0.0, got %s", content)
}

func (suite *InitTestSuite) TestInitCommand_CreatesConfigFile() {
	// Verify config doesn't exist
	_, err := filesystem.AppFs.Stat(suite.absPath(".versionator.yaml"))
	suite.True(os.IsNotExist(err), ".versionator.yaml should not exist before init")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify config was created
	data, err := filesystem.AppFs.ReadFile(suite.absPath(".versionator.yaml"))
	suite.Require().NoError(err, ".versionator.yaml should exist after init")

	content := string(data)
	suite.Contains(content, "prefix:", "config should contain prefix setting")
	suite.Contains(content, "prerelease:", "config should contain prerelease setting")
}

func (suite *InitTestSuite) TestInitCommand_DoesNotOverwriteExistingVersion() {
	// Create existing VERSION file with custom content
	existingVersion := "1.2.3\n"
	err := filesystem.AppFs.WriteFile(suite.absPath("VERSION"), []byte(existingVersion), 0644)
	suite.Require().NoError(err, "failed to create existing VERSION")

	// Also create config so that init doesn't create it (which would trigger re-render)
	existingConfig := "prefix: \"\"\nprerelease:\n  elements: []\nmetadata:\n  elements: []\n  git:\n    hashLength: 7\nlogging:\n  output: console\n"
	err = filesystem.AppFs.WriteFile(suite.absPath(".versionator.yaml"), []byte(existingConfig), 0644)
	suite.Require().NoError(err, "failed to create existing config")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify VERSION was NOT overwritten
	data, err := filesystem.AppFs.ReadFile(suite.absPath("VERSION"))
	suite.Require().NoError(err, "failed to read VERSION")
	suite.Equal(existingVersion, string(data), "VERSION should not be overwritten")
}

func (suite *InitTestSuite) TestInitCommand_DoesNotOverwriteExistingConfig() {
	// Create existing config file
	existingConfig := "# custom config\nprefix: \"release-\"\n"
	err := filesystem.AppFs.WriteFile(suite.absPath(".versionator.yaml"), []byte(existingConfig), 0644)
	suite.Require().NoError(err, "failed to create existing config")

	// Execute init command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify config was NOT overwritten
	data, err := filesystem.AppFs.ReadFile(suite.absPath(".versionator.yaml"))
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
	err := filesystem.AppFs.WriteFile(suite.absPath("VERSION"), []byte("1.0.0\n"), 0644)
	suite.Require().NoError(err)
	err = filesystem.AppFs.WriteFile(suite.absPath(".versionator.yaml"), []byte("prefix: \"v\"\n"), 0644)
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

func (suite *InitTestSuite) TestInitCommand_GoFlag_EnablesPrerelease() {
	// Execute init command with --go flag
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init", "--go"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify config has prerelease template set
	data, err := filesystem.AppFs.ReadFile(suite.absPath(".versionator.yaml"))
	suite.Require().NoError(err, "failed to read config")

	content := string(data)
	// Should have prerelease elements with CommitsSinceTag for Go pseudo-versions
	suite.Contains(content, "elements:", "config should contain elements setting")
	suite.Contains(content, "CommitsSinceTag", "Go config should use CommitsSinceTag in prerelease")

	output := stdout.String()
	suite.Contains(output, "Go", "output should mention Go mode")
}

func (suite *InitTestSuite) TestInitCommand_NoGoFlag_DefaultPrerelease() {
	// Execute init command without --go flag
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	suite.Require().NoError(err, "init command should succeed")

	// Verify config exists with default settings
	data, err := filesystem.AppFs.ReadFile(suite.absPath(".versionator.yaml"))
	suite.Require().NoError(err, "failed to read config")

	content := string(data)
	suite.Contains(content, "prerelease:", "config should contain prerelease section")
}

// TestInitTestSuite runs the init test suite
func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}
