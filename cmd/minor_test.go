package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

// MinorTestSuite provides a test suite for minor version commands
type MinorTestSuite struct {
	suite.Suite
	originalDir string
	tempDir     string
}

// SetupTest runs before each test
func (suite *MinorTestSuite) SetupTest() {
	var err error
	suite.originalDir, err = os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")

	suite.tempDir = suite.T().TempDir()
	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err, "Failed to change to temp directory")
	
	// Unregister Git VCS to prevent interference with tests
	vcs.UnregisterVCS("git")
}

// TearDownTest runs after each test
func (suite *MinorTestSuite) TearDownTest() {
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


func (suite *MinorTestSuite) TestMinorIncrementCommand() {
	// Create fresh filesystem for this test
	fs := afero.NewMemMapFs()
	
	// Create test app instance with fresh filesystem
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", nil),
		Versionator:    versionator.NewVersionator(fs, nil),
		VCS:            nil,
		FileSystem:     fs,
	}
	
	// Replace global app instance for command execution
	originalApp := appInstance
	appInstance = testApp
	defer func() {
		appInstance = originalApp
	}()

	// Create config file
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Create VERSION file
	err = afero.WriteFile(fs, "VERSION", []byte("1.2.3"), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Execute the minor increment command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"minor", "increment"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "minor increment command should succeed")
	
	// Reset command state
	rootCmd.SetArgs([]string{})

	// Verify VERSION file was updated correctly
	content, err := afero.ReadFile(fs, "VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.3.0", strings.TrimSpace(string(content)), "VERSION file should contain '1.3.0'")
}

func (suite *MinorTestSuite) TestMinorIncrementCommand_Aliases() {
	testCases := []string{"inc", "+"}

	for _, alias := range testCases {
		suite.Run("alias_"+alias, func() {
			// Create fresh filesystem for this test
			fs := afero.NewMemMapFs()
			
			// Create test app instance with fresh filesystem
			testApp := &app.App{
				ConfigManager:  config.NewConfigManager(fs),
				VersionManager: version.NewVersion(fs, ".", nil),
				Versionator:    versionator.NewVersionator(fs, nil),
				VCS:            nil,
				FileSystem:     fs,
			}
			
			// Replace global app instance for command execution
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
			}()

			// Create config file
			configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
			err := afero.WriteFile(fs, ".versionator.yaml", []byte(configContent), 0644)
			suite.Require().NoError(err, "Failed to create config file")

			// Create VERSION file
			err = afero.WriteFile(fs, "VERSION", []byte("0.5.7"), 0644)
			suite.Require().NoError(err, "Failed to create VERSION file")

			// Execute the minor increment command with alias
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"minor", alias})
			err = rootCmd.Execute()
			suite.Require().NoError(err, "minor %s command should succeed", alias)
			
			// Reset command state
			rootCmd.SetArgs([]string{})

			// Verify VERSION file was updated correctly
			content, err := afero.ReadFile(fs, "VERSION")
			suite.Require().NoError(err, "Should be able to read VERSION file")
			suite.Equal("0.6.0", strings.TrimSpace(string(content)), "VERSION file should contain '0.6.0'")
		})
	}
}

func (suite *MinorTestSuite) TestMinorDecrementCommand() {
	// Create fresh filesystem for this test
	fs := afero.NewMemMapFs()
	
	// Create test app instance with fresh filesystem
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", nil),
		Versionator:    versionator.NewVersionator(fs, nil),
		VCS:            nil,
		FileSystem:     fs,
	}
	
	// Replace global app instance for command execution
	originalApp := appInstance
	appInstance = testApp
	defer func() {
		appInstance = originalApp
	}()

	// Create config file
	configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Create VERSION file
	err = afero.WriteFile(fs, "VERSION", []byte("1.3.5"), 0644)
	suite.Require().NoError(err, "Failed to create VERSION file")

	// Execute the minor decrement command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"minor", "decrement"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "minor decrement command should succeed")
	
	// Reset command state
	rootCmd.SetArgs([]string{})

	// Verify VERSION file was updated correctly
	content, err := afero.ReadFile(fs, "VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("1.2.0", strings.TrimSpace(string(content)), "VERSION file should contain '1.2.0'")
}

func (suite *MinorTestSuite) TestMinorDecrementCommand_Aliases() {
	testCases := []string{"dec"}

	for _, alias := range testCases {
		suite.Run("alias_"+alias, func() {
			// Create fresh filesystem for this test
			fs := afero.NewMemMapFs()
			
			// Create test app instance with fresh filesystem
			testApp := &app.App{
				ConfigManager:  config.NewConfigManager(fs),
				VersionManager: version.NewVersion(fs, ".", nil),
				Versionator:    versionator.NewVersionator(fs, nil),
				VCS:            nil,
				FileSystem:     fs,
			}
			
			// Replace global app instance for command execution
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
			}()

			// Create config file
			configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
			err := afero.WriteFile(fs, ".versionator.yaml", []byte(configContent), 0644)
			suite.Require().NoError(err, "Failed to create config file")

			// Create VERSION file
			err = afero.WriteFile(fs, "VERSION", []byte("2.5.1"), 0644)
			suite.Require().NoError(err, "Failed to create VERSION file")

			// Execute the minor decrement command with alias
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"minor", alias})
			err = rootCmd.Execute()
			suite.Require().NoError(err, "minor %s command should succeed", alias)
			
			// Reset command state
			rootCmd.SetArgs([]string{})

			// Verify VERSION file was updated correctly
			content, err := afero.ReadFile(fs, "VERSION")
			suite.Require().NoError(err, "Should be able to read VERSION file")
			suite.Equal("2.4.0", strings.TrimSpace(string(content)), "VERSION file should contain '2.4.0'")
		})
	}
}

func (suite *MinorTestSuite) TestMinorIncrementCommand_NoVersionFile() {
	// Create fresh filesystem for this test
	fs := afero.NewMemMapFs()
	
	// Create test app instance with fresh filesystem
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", nil),
		Versionator:    versionator.NewVersionator(fs, nil),
		VCS:            nil,
		FileSystem:     fs,
	}
	
	// Replace global app instance for command execution
	originalApp := appInstance
	appInstance = testApp
	defer func() {
		appInstance = originalApp
	}()

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
	err := afero.WriteFile(fs, ".versionator.yaml", []byte(configContent), 0644)
	suite.Require().NoError(err, "Failed to create config file")

	// Execute the minor increment command - should succeed with default version
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"minor", "increment"})
	err = rootCmd.Execute()
	suite.Require().NoError(err, "minor increment command should succeed with default version")
	
	// Reset command state
	rootCmd.SetArgs([]string{})

	// Verify VERSION file was created and updated correctly
	content, err := afero.ReadFile(fs, "VERSION")
	suite.Require().NoError(err, "Should be able to read VERSION file")
	suite.Equal("0.1.0", strings.TrimSpace(string(content)), "VERSION file should contain '0.1.0'")
}

func (suite *MinorTestSuite) TestMinorCommandHelp() {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "minor help",
			args: []string{"minor", "--help"},
		},
		{
			name: "minor increment help",
			args: []string{"minor", "increment", "--help"},
		},
		{
			name: "minor decrement help",
			args: []string{"minor", "decrement", "--help"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tc.args)

			err := rootCmd.Execute()
			suite.NoError(err, "Help command should succeed")

			output := buf.String()
			suite.Contains(output, "Usage:", "Help output should contain usage information")
		})
	}
}

// TestMinorTestSuite runs the minor test suite
func TestMinorTestSuite(t *testing.T) {
	suite.Run(t, new(MinorTestSuite))
}