package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
)

func TestSuffixEnableCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		setupVCS       func(ctrl *gomock.Controller) *mock.MockVersionControlSystem
		expectError    bool
	}{
		{
			name: "enable suffix with git repository",
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
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return("/tmp", nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
		{
			name: "enable suffix without git repository",
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
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(false).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
		{
			name: "enable suffix when already enabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: true,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 8},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.5.0",
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return("/tmp", nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(7).Return("def5678", nil).AnyTimes()
				return mockVCS
			},
			expectError: false,
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

			// Setup gomock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Setup VCS mock
			mockVCS := tt.setupVCS(ctrl)
			vcs.RegisterVCS(mockVCS)
			defer vcs.UnregisterVCS("git")

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
			rootCmd.SetArgs([]string{"suffix", "enable"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify config was updated
				cfg, err := config.ReadConfig()
				require.NoError(t, err)
				assert.True(t, cfg.Suffix.Enabled)
				assert.Equal(t, "git", cfg.Suffix.Type)

				// Verify output
				output := stdout.String()
				assert.Contains(t, output, "Git hash suffix enabled")
				assert.Contains(t, output, "Current version:")
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestSuffixDisableCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		expectError    bool
	}{
		{
			name: "disable suffix when enabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: true,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.2.3",
			expectError:    false,
		},
		{
			name: "disable suffix when already disabled",
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
			rootCmd.SetArgs([]string{"suffix", "disable"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify config was updated
				cfg, err := config.ReadConfig()
				require.NoError(t, err)
				assert.False(t, cfg.Suffix.Enabled)

				// Verify output
				output := stdout.String()
				assert.Contains(t, output, "Git hash suffix disabled")
				assert.Contains(t, output, "Current version: "+tt.initialVersion)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestSuffixStatusCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		setupVCS       func(ctrl *gomock.Controller) *mock.MockVersionControlSystem
		expectError    bool
	}{
		{
			name: "status with suffix enabled and in git repo",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: true,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.2.3",
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return("/tmp", nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil).Times(1)
				return mockVCS
			},
			expectError: false,
		},
		{
			name: "status with suffix enabled but not in git repo",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: true,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 8},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "2.0.0",
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(false).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
		{
			name: "status with suffix disabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "3.1.0",
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				// No VCS expectations needed when suffix is disabled
				return nil
			},
			expectError: false,
		},
		{
			name: "status with git error",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: true,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			initialVersion: "1.0.0",
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return("/tmp", nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(7).Return("", os.ErrPermission).Times(1)
				return mockVCS
			},
			expectError: false,
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

			// Setup gomock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Setup VCS mock if provided
			if tt.setupVCS != nil {
				mockVCS := tt.setupVCS(ctrl)
				if mockVCS != nil {
					vcs.RegisterVCS(mockVCS)
					defer vcs.UnregisterVCS("git")
				}
			}

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
			rootCmd.SetArgs([]string{"suffix", "status"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify output
				output := stdout.String()
				if tt.initialConfig.Suffix.Enabled && tt.initialConfig.Suffix.Type == "git" {
					assert.Contains(t, output, "Git hash suffix: ENABLED")
					assert.Contains(t, output, "Hash length:")
				} else {
					assert.Contains(t, output, "Git hash suffix: DISABLED")
				}
				assert.Contains(t, output, "Current version:")
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestSuffixConfigureCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		expectError    bool
	}{
		{
			name: "configure with suffix enabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: true,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 7},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			expectError: false,
		},
		{
			name: "configure with suffix disabled",
			initialConfig: &config.Config{
				Prefix: "",
				Suffix: config.SuffixConfig{
					Enabled: false,
					Type:    "git",
					Git:     config.GitConfig{HashLength: 8},
				},
				Logging: config.LoggingConfig{Output: "console"},
			},
			expectError: false,
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

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"suffix", "configure"})

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify output
				output := stdout.String()
				assert.Contains(t, output, "Current configuration:")
				assert.Contains(t, output, "Git hash suffix enabled:")
				assert.Contains(t, output, "Suffix type:")
				assert.Contains(t, output, "Git hash length:")
				assert.Contains(t, output, "Configuration is stored in .versionator.yaml")
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestSuffixCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "suffix help",
			args: []string{"suffix", "--help"},
		},
		{
			name: "suffix enable help",
			args: []string{"suffix", "enable", "--help"},
		},
		{
			name: "suffix disable help",
			args: []string{"suffix", "disable", "--help"},
		},
		{
			name: "suffix status help",
			args: []string{"suffix", "status", "--help"},
		},
		{
			name: "suffix configure help",
			args: []string{"suffix", "configure", "--help"},
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

func TestSuffixCommandConfigErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupFunc   func(tempDir string) error
		expectError bool
	}{
		{
			name: "missing config file",
			args: []string{"suffix", "enable"},
			setupFunc: func(tempDir string) error {
				// Create VERSION file but no config file
				return os.WriteFile("VERSION", []byte("1.0.0"), 0644)
			},
			expectError: false, // Should create default config
		},
		{
			name: "invalid config file",
			args: []string{"suffix", "status"},
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