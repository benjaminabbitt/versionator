package cmd

import (
	"bytes"
	"os"
	"testing"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/version"
	"versionator/internal/versionator"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Helper functions for DRY test setup

// createPrefixTestApp creates a fresh filesystem and test app instance
func createPrefixTestApp() (afero.Fs, *app.App) {
	fs := afero.NewMemMapFs()
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", nil),
		Versionator:    versionator.NewVersionator(fs, nil),
		VCS:            nil,
		FileSystem:     fs,
	}
	return fs, testApp
}

// replacePrefixAppInstance replaces the global app instance and returns a restore function
func replacePrefixAppInstance(testApp *app.App) func() {
	originalApp := appInstance
	appInstance = testApp
	return func() {
		appInstance = originalApp
	}
}

// setupPrefixTestEnvironment sets up isolated test environment with proper cleanup
func setupPrefixTestEnvironment(t *testing.T) func() {
	// Save original state
	origDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	
	// Create temporary directory and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")
	
	// Return cleanup function
	return func() {
		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		
		// Restore original directory
		if origDir != "" {
			err := os.Chdir(origDir)
			if err != nil {
				t.Logf("Warning: Failed to restore original directory: %v", err)
			}
		}
	}
}

// createPrefixTestFiles creates test files needed for prefix tests
func createPrefixTestFiles(t *testing.T, fs afero.Fs, version string, cfg *config.Config) {
	// Create VERSION file
	err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
	require.NoError(t, err, "Failed to create VERSION file")

	// Create config file
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err, "Failed to marshal config")
	err = afero.WriteFile(fs, ".versionator.yaml", configData, 0644)
	require.NoError(t, err, "Failed to create config file")
}

func TestPrefixEnableCommand_DefaultConfig(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with default config (prefix disabled)
	cfg := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix enable command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "enable"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix enable command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix enabled with default value 'v'", "Should show prefix enabled message")

	// Verify config file was updated
	configData, err := afero.ReadFile(fs, ".versionator.yaml")
	require.NoError(t, err, "Should be able to read config file")

	var updatedCfg config.Config
	err = yaml.Unmarshal(configData, &updatedCfg)
	require.NoError(t, err, "Should be able to parse config file")
	require.Equal(t, "v", updatedCfg.Prefix, "Config should have prefix set to 'v'")
}

func TestPrefixEnableCommand_WhenAlreadyEnabled(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with prefix already enabled
	cfg := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix enable command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "enable"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix enable command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix enabled with default value 'v'", "Should show already enabled message")
}

func TestPrefixDisableCommand_WhenEnabled(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with prefix enabled
	cfg := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix disable command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "disable"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix disable command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix disabled", "Should show prefix disabled message")

	// Verify config file was updated
	configData, err := afero.ReadFile(fs, ".versionator.yaml")
	require.NoError(t, err, "Should be able to read config file")

	var updatedCfg config.Config
	err = yaml.Unmarshal(configData, &updatedCfg)
	require.NoError(t, err, "Should be able to parse config file")
	require.Equal(t, "", updatedCfg.Prefix, "Config should have empty prefix")
}

func TestPrefixDisableCommand_WhenAlreadyDisabled(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with prefix already disabled
	cfg := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix disable command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "disable"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix disable command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix disabled", "Should show already disabled message")
}

func TestPrefixSetCommand_CustomPrefix(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with default config
	cfg := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix set command with custom prefix
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "set", "release-"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix set command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix set to: release-", "Should show prefix set message")

	// Verify config file was updated
	configData, err := afero.ReadFile(fs, ".versionator.yaml")
	require.NoError(t, err, "Should be able to read config file")

	var updatedCfg config.Config
	err = yaml.Unmarshal(configData, &updatedCfg)
	require.NoError(t, err, "Should be able to parse config file")
	require.Equal(t, "release-", updatedCfg.Prefix, "Config should have custom prefix")
}

func TestPrefixSetCommand_EmptyPrefix(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with prefix enabled
	cfg := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix set command with empty prefix
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "set", ""})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix set command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix disabled (set to empty)", "Should show prefix disabled message")

	// Verify config file was updated
	configData, err := afero.ReadFile(fs, ".versionator.yaml")
	require.NoError(t, err, "Should be able to read config file")

	var updatedCfg config.Config
	err = yaml.Unmarshal(configData, &updatedCfg)
	require.NoError(t, err, "Should be able to parse config file")
	require.Equal(t, "", updatedCfg.Prefix, "Config should have empty prefix")
}

func TestPrefixSetCommand_SpecialCharacters(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with default config
	cfg := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix set command with special characters
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "set", "v@#$-"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix set command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix set to: v@#$-", "Should show prefix set message")

	// Verify config file was updated
	configData, err := afero.ReadFile(fs, ".versionator.yaml")
	require.NoError(t, err, "Should be able to read config file")

	var updatedCfg config.Config
	err = yaml.Unmarshal(configData, &updatedCfg)
	require.NoError(t, err, "Should be able to parse config file")
	require.Equal(t, "v@#$-", updatedCfg.Prefix, "Config should have special character prefix")
}

func TestPrefixSetCommand_NoArgument(t *testing.T) {
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Execute the prefix set command without argument - should fail
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"prefix", "set"})

	err := rootCmd.Execute()
	require.Error(t, err, "prefix set command should fail without argument")
}

func TestPrefixStatusCommand_Enabled(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with prefix enabled
	cfg := &config.Config{
		Prefix: "v",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix status command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "status"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix status command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix: ENABLED", "Should show prefix enabled")
	require.Contains(t, output, "Prefix value: v", "Should show prefix value")
}

func TestPrefixStatusCommand_Disabled(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with prefix disabled
	cfg := &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix status command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "status"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix status command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix: DISABLED", "Should show prefix disabled")
}

func TestPrefixStatusCommand_CustomPrefix(t *testing.T) {
	defer setupPrefixTestEnvironment(t)()

	fs, testApp := createPrefixTestApp()
	defer replacePrefixAppInstance(testApp)()

	// Create test files with custom prefix
	cfg := &config.Config{
		Prefix: "release-",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git: config.GitConfig{
				HashLength: 7,
			},
		},
		Logging: config.LoggingConfig{
			Output: "console",
		},
	}
	createPrefixTestFiles(t, fs, "1.2.3", cfg)

	// Execute the prefix status command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"prefix", "status"})

	err := rootCmd.Execute()
	require.NoError(t, err, "prefix status command should succeed")

	// Check output
	output := buf.String()
	require.Contains(t, output, "Version prefix: ENABLED", "Should show prefix enabled")
	require.Contains(t, output, "Prefix value: release-", "Should show custom prefix value")
}