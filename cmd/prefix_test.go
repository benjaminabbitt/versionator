package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"versionator/internal/config"
)

func TestPrefixEnableCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		expectedPrefix string
		expectError    bool
	}{
		{
			name: "enable prefix with default config",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.2.3",
			expectedPrefix: "v",
			expectError:    false,
		},
		{
			name: "enable prefix when already enabled",
			initialConfig: &config.Config{
				Prefix: "release-",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "2.0.0",
			expectedPrefix: "v",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create VERSION file
			err = os.WriteFile("VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"prefix", "enable"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify config was updated
				cfg, err := config.ReadConfig()
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPrefix, cfg.Prefix)

				// Verify output
				output := stdout.String()
				assert.Contains(t, output, "Version prefix enabled with default value 'v'")
				assert.Contains(t, output, "Current version: v"+tt.initialVersion)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPrefixDisableCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		expectError    bool
	}{
		{
			name: "disable prefix when enabled",
			initialConfig: &config.Config{
				Prefix: "v",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.2.3",
			expectError:    false,
		},
		{
			name: "disable prefix when already disabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "2.0.0",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create VERSION file
			err = os.WriteFile("VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"prefix", "disable"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify config was updated
				cfg, err := config.ReadConfig()
				require.NoError(t, err)
				assert.Equal(t, "", cfg.Prefix)

				// Verify output
				output := stdout.String()
				assert.Contains(t, output, "Version prefix disabled")
				assert.Contains(t, output, "Current version: "+tt.initialVersion)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPrefixSetCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		initialConfig  *config.Config
		initialVersion string
		expectedPrefix string
		expectError    bool
		errorContains  string
	}{
		{
			name: "set custom prefix",
			args: []string{"prefix", "set", "release-"},
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.2.3",
			expectedPrefix: "release-",
			expectError:    false,
		},
		{
			name: "set empty prefix",
			args: []string{"prefix", "set", ""},
			initialConfig: &config.Config{
				Prefix: "v",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "2.0.0",
			expectedPrefix: "",
			expectError:    false,
		},
		{
			name: "set prefix with special characters",
			args: []string{"prefix", "set", "v2.0-"},
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.0.0",
			expectedPrefix: "v2.0-",
			expectError:    false,
		},
		{
			name: "set prefix without argument",
			args: []string{"prefix", "set"},
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.0.0",
			expectedPrefix: "",
			expectError:    true,
			errorContains:  "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create VERSION file
			err = os.WriteFile("VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify config was updated
				cfg, err := config.ReadConfig()
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPrefix, cfg.Prefix)

				// Verify output
				output := stdout.String()
				if tt.expectedPrefix == "" {
					assert.Contains(t, output, "Version prefix disabled (set to empty)")
					assert.Contains(t, output, "Current version: "+tt.initialVersion)
				} else {
					assert.Contains(t, output, "Version prefix set to: "+tt.expectedPrefix)
					assert.Contains(t, output, "Current version: "+tt.expectedPrefix+tt.initialVersion)
				}
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPrefixStatusCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		expectError    bool
	}{
		{
			name: "status with prefix enabled",
			initialConfig: &config.Config{
				Prefix: "v",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.2.3",
			expectError:    false,
		},
		{
			name: "status with prefix disabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "2.0.0",
			expectError:    false,
		},
		{
			name: "status with custom prefix",
			initialConfig: &config.Config{
				Prefix: "release-",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "3.1.0",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create VERSION file
			err = os.WriteFile("VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"prefix", "status"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify output
				output := stdout.String()
				if tt.initialConfig.Prefix == "" {
					assert.Contains(t, output, "Version prefix: DISABLED")
					assert.Contains(t, output, "Current version: "+tt.initialVersion)
				} else {
					assert.Contains(t, output, "Version prefix: ENABLED")
					assert.Contains(t, output, "Prefix value: "+tt.initialConfig.Prefix)
					assert.Contains(t, output, "Current version: "+tt.initialConfig.Prefix+tt.initialVersion)
				}
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPrefixCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "prefix help",
			args: []string{"prefix", "--help"},
		},
		{
			name: "prefix enable help",
			args: []string{"prefix", "enable", "--help"},
		},
		{
			name: "prefix disable help",
			args: []string{"prefix", "disable", "--help"},
		},
		{
			name: "prefix set help",
			args: []string{"prefix", "set", "--help"},
		},
		{
			name: "prefix status help",
			args: []string{"prefix", "status", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "Usage:")

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPrefixCommandConfigErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupFunc   func(tempDir string) error
		expectError bool
	}{
		{
			name: "missing config file",
			args: []string{"prefix", "enable"},
			setupFunc: func(tempDir string) error {
				// Create VERSION file but no config file
				return os.WriteFile("VERSION", []byte("1.0.0"), 0644)
			},
			expectError: false, // Should create default config
		},
		{
			name: "invalid config file",
			args: []string{"prefix", "status"},
			setupFunc: func(tempDir string) error {
				// Create invalid YAML config
				err := os.WriteFile("VERSION", []byte("1.0.0"), 0644)
				if err != nil {
					return err
				}
				return os.WriteFile(".versionator.yaml", []byte("invalid: yaml: content: ["), 0644)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Setup test environment
			err = tt.setupFunc(tempDir)
			require.NoError(t, err)

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			rootCmd.SetArgs(nil)
		})
	}
}
