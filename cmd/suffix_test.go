package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/vcs/mock"
	"versionator/internal/version"
	"versionator/internal/versionator"
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
				mockVCS.EXPECT().GetVCSIdentifier(8).Return("def5678", nil).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create in-memory filesystem
			fs := afero.NewMemMapFs()
			
			// Setup gomock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Setup VCS mock
			mockVCS := tt.setupVCS(ctrl)
			
			// Create test app instance with in-memory filesystem and injected VCS
			testApp := &app.App{
				ConfigManager:  config.NewConfigManager(fs),
				VersionManager: version.NewVersion(fs, ".", mockVCS),
				Versionator:    versionator.NewVersionator(fs, mockVCS),
				VCS:            mockVCS,
				FileSystem:     fs,
			}
			
			// Replace global app instance for this test
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
			}()

			// Create VERSION file
			err := afero.WriteFile(fs, "VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = afero.WriteFile(fs, ".versionator.yaml", configData, 0644)
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
				cfg, err := testApp.ReadConfig()
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
			// Create in-memory filesystem
			fs := afero.NewMemMapFs()
			
			// Create test app instance with in-memory filesystem
			testApp := &app.App{
				ConfigManager:  config.NewConfigManager(fs),
				VersionManager: version.NewVersion(fs, ".", nil),
				Versionator:    versionator.NewVersionator(fs, nil),
				VCS:            nil,
				FileSystem:     fs,
			}
			
			// Replace global app instance for this test
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
			}()

			// Create VERSION file
			err := afero.WriteFile(fs, "VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = afero.WriteFile(fs, ".versionator.yaml", configData, 0644)
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
				cfg, err := testApp.ReadConfig()
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
				mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil).AnyTimes()
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
				// Create mock VCS but no expectations needed when suffix is disabled
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().IsRepository().Return(false).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return("/tmp", nil).AnyTimes()
				return mockVCS
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
				mockVCS.EXPECT().GetVCSIdentifier(7).Return("", os.ErrPermission).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create in-memory filesystem
			fs := afero.NewMemMapFs()

			// Setup gomock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Setup VCS mock if provided and create app with it
			var testApp *app.App
			if tt.setupVCS != nil {
				mockVCS := tt.setupVCS(ctrl)
				testApp = &app.App{
					ConfigManager:  config.NewConfigManager(fs),
					VersionManager: version.NewVersion(fs, ".", mockVCS),
					Versionator:    versionator.NewVersionator(fs, mockVCS),
					VCS:            mockVCS,
					FileSystem:     fs,
				}
			} else {
				testApp = &app.App{
					ConfigManager:  config.NewConfigManager(fs),
					VersionManager: version.NewVersion(fs, ".", nil),
					Versionator:    versionator.NewVersionator(fs, nil),
					VCS:            nil,
					FileSystem:     fs,
				}
			}

			// Replace global app instance for this test
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
			}()
			
			// Create VERSION file
			err := afero.WriteFile(fs, "VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = afero.WriteFile(fs, ".versionator.yaml", configData, 0644)
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
			// Create in-memory filesystem
			fs := afero.NewMemMapFs()

			// Create initial config file
			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = afero.WriteFile(fs, ".versionator.yaml", configData, 0644)
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
		setupFunc   func(fs afero.Fs) error
		expectError bool
	}{
		{
			name: "missing config file",
			args: []string{"suffix", "enable"},
			setupFunc: func(fs afero.Fs) error {
				// Create VERSION file but no config file
				return afero.WriteFile(fs, "VERSION", []byte("1.0.0"), 0644)
			},
			expectError: false, // Should create default config
		},
		{
			name: "invalid config file",
			args: []string{"suffix", "status"},
			setupFunc: func(fs afero.Fs) error {
				// Create invalid YAML config
				err := afero.WriteFile(fs, "VERSION", []byte("1.0.0"), 0644)
				if err != nil {
					return err
				}
				return afero.WriteFile(fs, ".versionator.yaml", []byte("invalid yaml content\n  [unclosed bracket\n    bad: indentation"), 0644)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unregister any VCS to prevent interference
			vcs.UnregisterVCS("git")
			defer func() {
				// Clean up any registered VCS after test
				vcs.UnregisterVCS("git")
			}()

			// Create in-memory filesystem
			fs := afero.NewMemMapFs()
			
			// Create test app instance with in-memory filesystem
			testApp := &app.App{
				ConfigManager:  config.NewConfigManager(fs),
				VersionManager: version.NewVersion(fs, ".", nil),
				Versionator:    versionator.NewVersionator(fs, nil),
				VCS:            nil,
				FileSystem:     fs,
			}
			
			// Replace global app instance for this test
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
			}()

			// Setup test environment
			err := tt.setupFunc(fs)
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