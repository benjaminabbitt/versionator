package cmd

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"versionator/internal/app"
	"versionator/internal/config"
	"versionator/internal/vcs"
	"versionator/internal/vcs/mock"
	"versionator/internal/version"
	"versionator/internal/versionator"
)

// SuffixTestSuite contains common constants and test data
type SuffixTestSuite struct {
	suite.Suite
	
	// Common test constants extracted from individual tests
	DefaultVersion         string
	StandardHashLength     int
	DefaultGitHash         string
	AlternateGitHash       string
	TestRepositoryPath     string
	
	// Standard configuration templates
	StandardConfig         *config.Config
	EnabledConfig          *config.Config
	CustomHashLengthConfig *config.Config
}

// SetupSuite runs once before all tests
func (s *SuffixTestSuite) SetupSuite() {
	// Initialize common constants
	s.DefaultVersion = "1.2.3"
	s.StandardHashLength = 7
	s.DefaultGitHash = "abc1234"
	s.AlternateGitHash = "def5678"
	s.TestRepositoryPath = "/tmp"
	
	// Standard disabled configuration
	s.StandardConfig = &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: false,
			Type:    "git",
			Git:     config.GitConfig{HashLength: s.StandardHashLength},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	
	// Standard enabled configuration
	s.EnabledConfig = &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: true,
			Type:    "git",
			Git:     config.GitConfig{HashLength: s.StandardHashLength},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	
	// Custom hash length configuration
	s.CustomHashLengthConfig = &config.Config{
		Prefix: "",
		Suffix: config.SuffixConfig{
			Enabled: true,
			Type:    "git",
			Git:     config.GitConfig{HashLength: 8},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
}

// createTestAppWithMockVCS creates a test app with mock VCS
func (s *SuffixTestSuite) createTestAppWithMockVCS(ctrl *gomock.Controller, isRepo bool, hashLength int, hash string) (afero.Fs, *app.App, *mock.MockVersionControlSystem) {
	fs := afero.NewMemMapFs()
	
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(isRepo).AnyTimes()
	
	if isRepo {
		mockVCS.EXPECT().GetRepositoryRoot().Return(s.TestRepositoryPath, nil).AnyTimes()
		if hash != "" {
			mockVCS.EXPECT().GetVCSIdentifier(hashLength).Return(hash, nil).AnyTimes()
		}
	}
	
	testApp := &app.App{
		ConfigManager:  config.NewConfigManager(fs),
		VersionManager: version.NewVersion(fs, ".", mockVCS),
		Versionator:    versionator.NewVersionator(fs, mockVCS),
		VCS:            mockVCS,
		FileSystem:     fs,
	}
	
	return fs, testApp, mockVCS
}

// createTestAppWithoutVCS creates a test app without VCS
func (s *SuffixTestSuite) createTestAppWithoutVCS() (afero.Fs, *app.App) {
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

// createTestFiles creates VERSION and config files
func (s *SuffixTestSuite) createTestFiles(fs afero.Fs, version string, cfg *config.Config) {
	err := afero.WriteFile(fs, "VERSION", []byte(version), 0644)
	s.Require().NoError(err)
	
	configData, err := yaml.Marshal(cfg)
	s.Require().NoError(err)
	err = afero.WriteFile(fs, ".versionator.yaml", configData, 0644)
	s.Require().NoError(err)
}

func (s *SuffixTestSuite) TestSuffixEnableCommand() {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		setupVCS       func(ctrl *gomock.Controller) *mock.MockVersionControlSystem
		expectError    bool
	}{
		{
			name:           "enable suffix with git repository",
			initialConfig:  s.StandardConfig,
			initialVersion: s.DefaultVersion,
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return(s.TestRepositoryPath, nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(s.StandardHashLength).Return(s.DefaultGitHash, nil).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
		{
			name:           "enable suffix without git repository",
			initialConfig:  s.StandardConfig,
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
			name:           "enable suffix when already enabled",
			initialConfig:  s.CustomHashLengthConfig,
			initialVersion: "1.5.0",
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return(s.TestRepositoryPath, nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(8).Return(s.AlternateGitHash, nil).AnyTimes()
				return mockVCS
			},
			expectError: false,
		},
	}

		for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup gomock
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			// Setup VCS mock and create test app
			mockVCS := tt.setupVCS(ctrl)
			fs, testApp, _ := s.createTestAppWithMockVCS(ctrl, mockVCS.IsRepository(), s.StandardHashLength, s.DefaultGitHash)

			// Replace global app instance
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			// Create test files
			s.createTestFiles(fs, tt.initialVersion, tt.initialConfig)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"suffix", "enable"})

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err, "Command should execute successfully")

				// Verify config was updated
				cfg, err := testApp.ReadConfig()
				s.NoError(err, "Should be able to read config")
				s.True(cfg.Suffix.Enabled, "Suffix should be enabled")
				s.Equal("git", cfg.Suffix.Type, "Suffix type should be git")

				// Verify output
				output := stdout.String()
				s.Contains(output, "Git hash suffix enabled", "Should show enabled message")
				s.Contains(output, "Current version:", "Should show current version")
			}
		})
	}
}

func (s *SuffixTestSuite) TestSuffixDisableCommand() {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		expectError    bool
	}{
		{
			name:           "disable suffix when enabled",
			initialConfig:  s.EnabledConfig,
			initialVersion: s.DefaultVersion,
			expectError:    false,
		},
		{
			name:           "disable suffix when already disabled",
			initialConfig:  s.StandardConfig,
			initialVersion: "2.0.0",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create test app without VCS
			fs, testApp := s.createTestAppWithoutVCS()

			// Replace global app instance
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			// Create test files
			s.createTestFiles(fs, tt.initialVersion, tt.initialConfig)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"suffix", "disable"})

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err, "Disable command should execute successfully")

				// Verify config was updated
				cfg, err := testApp.ReadConfig()
				s.NoError(err, "Should be able to read config")
				s.False(cfg.Suffix.Enabled, "Suffix should be disabled")

				// Verify output
				output := stdout.String()
				s.Contains(output, "Git hash suffix disabled", "Should show disabled message")
				s.Contains(output, "Current version: "+tt.initialVersion, "Should show current version")
			}
		})
	}
}

func (s *SuffixTestSuite) TestSuffixStatusCommand() {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		initialVersion string
		setupVCS       func(ctrl *gomock.Controller) *mock.MockVersionControlSystem
		expectEnabled  bool
	}{
		{
			name:           "status when enabled with git repo",
			initialConfig:  s.EnabledConfig,
			initialVersion: s.DefaultVersion,
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
				mockVCS.EXPECT().GetRepositoryRoot().Return(s.TestRepositoryPath, nil).AnyTimes()
				mockVCS.EXPECT().GetVCSIdentifier(s.StandardHashLength).Return(s.DefaultGitHash, nil).AnyTimes()
				return mockVCS
			},
			expectEnabled: true,
		},
		{
			name:           "status when disabled",
			initialConfig:  s.StandardConfig,
			initialVersion: s.DefaultVersion,
			setupVCS:       nil,
			expectEnabled:  false,
		},
		{
			name:           "status when enabled without git repo",
			initialConfig:  s.EnabledConfig,
			initialVersion: s.DefaultVersion,
			setupVCS: func(ctrl *gomock.Controller) *mock.MockVersionControlSystem {
				mockVCS := mock.NewMockVersionControlSystem(ctrl)
				mockVCS.EXPECT().Name().Return("git").AnyTimes()
				mockVCS.EXPECT().IsRepository().Return(false).AnyTimes()
				return mockVCS
			},
			expectEnabled: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var fs afero.Fs
			var testApp *app.App
			var ctrl *gomock.Controller

			if tt.setupVCS != nil {
				// Setup with mock VCS
				ctrl = gomock.NewController(s.T())
				defer ctrl.Finish()
				
				mockVCS := tt.setupVCS(ctrl)
				fs, testApp, _ = s.createTestAppWithMockVCS(ctrl, mockVCS.IsRepository(), s.StandardHashLength, s.DefaultGitHash)
			} else {
				// Setup without VCS
				fs, testApp = s.createTestAppWithoutVCS()
			}

			// Replace global app instance
			originalApp := appInstance
			appInstance = testApp
			defer func() {
				appInstance = originalApp
				// Reset command state
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			}()

			// Create test files
			s.createTestFiles(fs, tt.initialVersion, tt.initialConfig)

			// Capture output
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"suffix", "status"})

			// Execute command
			err := rootCmd.Execute()
			s.NoError(err, "Status command should execute successfully")

			// Verify output
			output := stdout.String()
			if tt.expectEnabled {
				s.Contains(output, "Git hash suffix: ENABLED", "Should show suffix enabled")
			} else {
				s.Contains(output, "Git hash suffix: DISABLED", "Should show suffix disabled")
			}
			s.Contains(output, "Current version:", "Should show current version")
		})
	}
}

func (s *SuffixTestSuite) TestSuffixConfigureCommand() {
	s.Run("configure command", func() {
		// Create test app without VCS
		fs, testApp := s.createTestAppWithoutVCS()

		// Replace global app instance
		originalApp := appInstance
		appInstance = testApp
		defer func() {
			appInstance = originalApp
		}()

		// Create test files
		s.createTestFiles(fs, s.DefaultVersion, s.StandardConfig)

		// Capture output
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"suffix", "configure"})

		// Execute command
		err := rootCmd.Execute()
		s.NoError(err, "Command should execute successfully")

		// Debug output if test fails
		output := stdout.String()
		if !s.Contains(output, "Current configuration:", "Should contain configuration header") {
			s.T().Logf("Actual output: %q", output)
			s.T().Logf("Stderr output: %q", stderr.String())
		}
		s.Contains(output, "Git hash suffix enabled: false", "Should show suffix disabled")
		s.Contains(output, "Suffix type: git", "Should show suffix type")
		s.Contains(output, "Git hash length: 7", "Should show hash length")
		s.Contains(output, "Configuration is stored in .versionator.yaml", "Should show config file location")

		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})
}

func (s *SuffixTestSuite) TestSuffixCommandHelp() {
	testCases := []struct {
		name string
		args []string
	}{
		{"suffix help", []string{"suffix", "--help"}},
		{"suffix enable help", []string{"suffix", "enable", "--help"}},
		{"suffix disable help", []string{"suffix", "disable", "--help"}},
		{"suffix status help", []string{"suffix", "status", "--help"}},
		{"suffix configure help", []string{"suffix", "configure", "--help"}},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tc.args)

			err := rootCmd.Execute()
			s.NoError(err)

			output := buf.String()
			s.Contains(output, "Usage:")

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func (s *SuffixTestSuite) TestSuffixCommandConfigErrors() {
	s.Run("missing config file", func() {
		// Unregister VCS to prevent interference
		vcs.UnregisterVCS("git")
		defer vcs.UnregisterVCS("git")

		// Create test app without VCS
		fs, testApp := s.createTestAppWithoutVCS()

		// Replace global app instance
		originalApp := appInstance
		appInstance = testApp
		defer func() {
			appInstance = originalApp
		}()

		// Create only VERSION file, no config file
		err := afero.WriteFile(fs, "VERSION", []byte("1.0.0"), 0644)
		s.Require().NoError(err)

		// Capture output
		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"suffix", "enable"})

		// Execute command - should create default config and succeed
		err = rootCmd.Execute()
		s.NoError(err)
		s.Contains(stdout.String(), "Git hash suffix enabled")

		// Reset command state
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})
}

// TestSuffixTestSuite runs the test suite
func TestSuffixTestSuite(t *testing.T) {
	suite.Run(t, new(SuffixTestSuite))
}