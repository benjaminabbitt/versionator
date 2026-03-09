package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// resetPrereleaseFlags resets global flag state that persists between tests.
// This must be called at the start of each test to avoid test interference.
func resetPrereleaseFlags() {
	prereleaseForceFlag = false
}

// =============================================================================
// CORE FUNCTIONALITY
// Tests for the primary happy path scenarios of prerelease commands
// =============================================================================

// TestPrereleaseSetCommand_WhenStableTrue validates that the prerelease set command
// correctly updates the VERSION file when stability mode is enabled.
//
// Why: This is the core happy path for setting prerelease values. When stable mode
// is true, the VERSION file should be treated as the source of truth and directly
// modified.
//
// What: Given a config with stable=true and a VERSION file, when running
// "config prerelease set alpha", the VERSION file should be updated to include
// the "alpha" prerelease identifier.
func TestPrereleaseSetCommand_WhenStableTrue(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: Set up temp directory with VERSION file and stable=true config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "set", "alpha"})

	// Action: Execute the prerelease set command
	err = rootCmd.Execute()

	// Expected: Command succeeds and VERSION file has prerelease set to "alpha"
	assert.NoError(t, err)

	v, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "alpha", v.PreRelease)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrereleaseClearCommand_WhenStableTrue validates that the prerelease clear
// command removes the prerelease identifier from the VERSION file when stable mode
// is enabled.
//
// Why: Users need the ability to clear prerelease identifiers when preparing for
// a stable release. This is a core operation in the release workflow.
//
// What: Given a VERSION file with "1.0.0-alpha" and stable=true config, when
// running "config prerelease clear", the VERSION file should be updated to "1.0.0"
// with no prerelease identifier.
func TestPrereleaseClearCommand_WhenStableTrue(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: Set up temp directory with VERSION file containing prerelease
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0-alpha\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "alpha",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "clear"})

	// Action: Execute the prerelease clear command
	err = rootCmd.Execute()

	// Expected: Command succeeds and VERSION file has no prerelease
	assert.NoError(t, err)

	v, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", v.PreRelease)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows and configuration combinations
// =============================================================================

// TestPrereleaseStableCommand_GetStability validates that the stable subcommand
// correctly retrieves and displays the current stability configuration value.
//
// Why: Users need to query the current stability mode to understand how
// prerelease values will be managed (stable mode vs template mode).
//
// What: Given various stability configurations, when running "config prerelease stable"
// without arguments, the command should output the current boolean value.
func TestPrereleaseStableCommand_GetStability(t *testing.T) {
	resetPrereleaseFlags()

	tests := []struct {
		name           string
		initialConfig  *config.Config
		expectedOutput string
	}{
		{
			name: "stability is false",
			initialConfig: &config.Config{
				PreRelease: config.PreReleaseConfig{
					Template: "build-{{CommitsSinceTag}}",
					Stable:   false,
				},
			},
			expectedOutput: "false",
		},
		{
			name: "stability is true",
			initialConfig: &config.Config{
				PreRelease: config.PreReleaseConfig{
					Template: "alpha",
					Stable:   true,
				},
			},
			expectedOutput: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Set up temp directory with VERSION file and config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			configData, err := yaml.Marshal(tt.initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "prerelease", "stable"})

			// Action: Execute the stable get command
			err = rootCmd.Execute()

			// Expected: Command succeeds and outputs the stability value
			assert.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, tt.expectedOutput)

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestPrereleaseStableCommand_SetStability validates that the stable subcommand
// correctly updates the stability configuration when a value is provided.
//
// Why: Users need to toggle between stable mode (VERSION file is source of truth)
// and template mode (prerelease is computed at emit time) based on their workflow.
//
// What: Given an initial stability value, when running "config prerelease stable <value>",
// the config file should be updated with the new boolean value.
func TestPrereleaseStableCommand_SetStability(t *testing.T) {
	resetPrereleaseFlags()

	tests := []struct {
		name           string
		initialStable  bool
		setTo          string
		expectedStable bool
		expectError    bool
		errorContains  string
	}{
		{
			name:           "set stable to true",
			initialStable:  false,
			setTo:          "true",
			expectedStable: true,
		},
		{
			name:           "set stable to false",
			initialStable:  true,
			setTo:          "false",
			expectedStable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Set up temp directory with VERSION file and config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			initialConfig := &config.Config{
				PreRelease: config.PreReleaseConfig{
					Template: "build-{{CommitsSinceTag}}",
					Stable:   tt.initialStable,
				},
			}
			configData, err := yaml.Marshal(initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "prerelease", "stable", tt.setTo})

			// Action: Execute the stable set command
			err = rootCmd.Execute()

			// Expected: Command succeeds and config is updated
			assert.NoError(t, err)

			cfg, err := config.ReadConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStable, cfg.PreRelease.Stable)

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestPrereleaseTemplateCommand validates that the template subcommand correctly
// updates the prerelease template in configuration and applies it appropriately
// based on stability mode.
//
// Why: The template command allows users to define dynamic prerelease identifiers
// using placeholders. Behavior differs based on stability mode - when stable=true
// the template is immediately rendered to the VERSION file.
//
// What: Given various stability configurations, when running "config prerelease template <value>",
// the config template should be updated, and the VERSION file should only be modified
// when stable=true.
func TestPrereleaseTemplateCommand(t *testing.T) {
	resetPrereleaseFlags()

	tests := []struct {
		name                string
		initialStable       bool
		template            string
		expectVersionUpdate bool
	}{
		{
			name:                "set template when stable false",
			initialStable:       false,
			template:            "alpha-{{CommitsSinceTag}}",
			expectVersionUpdate: false,
		},
		{
			name:                "set template when stable true",
			initialStable:       true,
			template:            "beta",
			expectVersionUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Set up temp directory with VERSION file and config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			initialConfig := &config.Config{
				PreRelease: config.PreReleaseConfig{
					Template: "",
					Stable:   tt.initialStable,
				},
			}
			configData, err := yaml.Marshal(initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "prerelease", "template", tt.template})

			// Action: Execute the template command
			err = rootCmd.Execute()

			// Expected: Command succeeds, config template is updated
			assert.NoError(t, err)

			cfg, err := config.ReadConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.template, cfg.PreRelease.Template)

			// Expected: VERSION file is only modified when stable=true
			v, err := version.Load()
			require.NoError(t, err)

			if tt.expectVersionUpdate {
				assert.Equal(t, tt.template, v.PreRelease)
			} else {
				assert.Equal(t, "", v.PreRelease)
			}

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestPrereleaseSetCommand_WhenStableFalse_WithForce validates that the --force
// flag allows the set command to update the template even when stable mode is disabled.
//
// Why: Users may need to override the stability check to update the template
// configuration without enabling stable mode. The --force flag provides this escape hatch.
//
// What: Given stable=false config, when running "config prerelease set alpha --force",
// the config template should be updated but the VERSION file should remain unchanged.
func TestPrereleaseSetCommand_WhenStableFalse_WithForce(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: Set up temp directory with VERSION file and stable=false config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "build-{{CommitsSinceTag}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "set", "alpha", "--force"})

	// Action: Execute the prerelease set command with --force
	err = rootCmd.Execute()

	// Expected: Command succeeds, config template is updated
	assert.NoError(t, err)

	cfg, err := config.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, "alpha", cfg.PreRelease.Template)

	// Expected: VERSION file is unchanged since stable=false
	v, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", v.PreRelease)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrereleaseStatusCommand validates that the status subcommand correctly
// displays the current prerelease configuration and values.
//
// Why: Users need visibility into the current prerelease configuration to
// understand how versions will be emitted and what values are currently set.
//
// What: Given various configurations, when running "config prerelease status",
// the output should include stability mode, template, and current VERSION file value.
func TestPrereleaseStatusCommand(t *testing.T) {
	resetPrereleaseFlags()

	tests := []struct {
		name           string
		stable         bool
		template       string
		versionFile    string
		expectContains []string
	}{
		{
			name:        "status when stable false",
			stable:      false,
			template:    "build-{{CommitsSinceTag}}",
			versionFile: "1.0.0",
			expectContains: []string{
				"Stable: false",
				"Template: build-{{CommitsSinceTag}}",
			},
		},
		{
			name:        "status when stable true with prerelease",
			stable:      true,
			template:    "alpha",
			versionFile: "1.0.0-alpha",
			expectContains: []string{
				"Stable: true",
				"VALUE (from VERSION file): alpha",
			},
		},
		{
			name:        "status when stable true without prerelease",
			stable:      true,
			template:    "",
			versionFile: "1.0.0",
			expectContains: []string{
				"Stable: true",
				"VALUE (from VERSION file): (none)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Set up temp directory with VERSION file and config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte(tt.versionFile+"\n"), 0644)
			require.NoError(t, err)

			cfg := &config.Config{
				PreRelease: config.PreReleaseConfig{
					Template: tt.template,
					Stable:   tt.stable,
				},
			}
			configData, err := yaml.Marshal(cfg)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "prerelease", "status"})

			// Action: Execute the status command
			err = rootCmd.Execute()

			// Expected: Command succeeds and output contains expected strings
			assert.NoError(t, err)

			output := stdout.String()
			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected)
			}

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes and error responses
// =============================================================================

// TestPrereleaseSetCommand_WhenStableFalse_Errors validates that the set command
// correctly rejects attempts to modify the VERSION file when stable mode is disabled.
//
// Why: When stable=false, the VERSION file should not be modified directly as
// prerelease values are computed dynamically at emit time. Attempting to set
// a static value would be inconsistent with the configured behavior.
//
// What: Given stable=false config, when running "config prerelease set alpha"
// without --force, the command should fail with an error explaining that
// stable mode must be enabled.
func TestPrereleaseSetCommand_WhenStableFalse_Errors(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: Set up temp directory with VERSION file and stable=false config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "build-{{CommitsSinceTag}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "set", "alpha"})

	// Action: Execute the prerelease set command
	err = rootCmd.Execute()

	// Expected: Command fails with error mentioning stable: false
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stable: false")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrereleaseClearCommand_WhenStableFalse_Errors validates that the clear
// command correctly rejects attempts to clear the VERSION file when stable mode
// is disabled.
//
// Why: When stable=false, the VERSION file's prerelease field is not used
// (prerelease is computed dynamically). Clearing it would be meaningless and
// could confuse users about the actual behavior.
//
// What: Given stable=false config, when running "config prerelease clear",
// the command should fail with an error explaining that stable mode must be enabled.
func TestPrereleaseClearCommand_WhenStableFalse_Errors(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: Set up temp directory with VERSION file and stable=false config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "build-{{CommitsSinceTag}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "clear"})

	// Action: Execute the prerelease clear command
	err = rootCmd.Execute()

	// Expected: Command fails with error mentioning stable: false
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stable: false")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrereleaseStableCommand_SetStability_InvalidValue validates that the stable
// subcommand correctly rejects invalid boolean values.
//
// Why: The stable flag only accepts "true" or "false" as valid values. Any other
// input should be rejected with a clear error message to prevent user confusion.
//
// What: Given any config, when running "config prerelease stable maybe" (invalid value),
// the command should fail with an error explaining valid values.
func TestPrereleaseStableCommand_SetStability_InvalidValue(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: Set up temp directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "build-{{CommitsSinceTag}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "stable", "maybe"})

	// Action: Execute the stable set command with invalid value
	err = rootCmd.Execute()

	// Expected: Command fails with error about invalid value
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use 'true' or 'false'")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// ENABLE/DISABLE TESTS
// =============================================================================

// TestPrereleaseEnableCommand_WritesPrereleaseToVersion verifies that the enable command
// renders the template and writes prerelease to the VERSION file.
//
// Why: The enable command is a quick way to add prerelease to VERSION in stable mode.
//
// What: Run "config prerelease enable" when stable=true, verify prerelease is written.
func TestPrereleaseEnableCommand_WritesPrereleaseToVersion(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: temp directory with stable=true config and template
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "alpha",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config prerelease enable"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "enable"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and prerelease is written to VERSION
	assert.NoError(t, err)
	versionContent, err := os.ReadFile("VERSION")
	require.NoError(t, err)
	assert.Contains(t, string(versionContent), "alpha")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrereleaseDisableCommand_ClearsPrereleaseFromVersion verifies that the disable command
// clears the prerelease from the VERSION file.
//
// Why: The disable command is a quick way to remove prerelease from VERSION in stable mode.
//
// What: Run "config prerelease disable" when stable=true, verify prerelease is cleared.
func TestPrereleaseDisableCommand_ClearsPrereleaseFromVersion(t *testing.T) {
	resetPrereleaseFlags()

	// Precondition: temp directory with stable=true config and prerelease in VERSION
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0-alpha\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "alpha",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config prerelease disable"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "disable"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and prerelease is cleared from VERSION
	assert.NoError(t, err)
	versionContent, err := os.ReadFile("VERSION")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0\n", string(versionContent))

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions (none identified for this module currently)
// =============================================================================

// =============================================================================
// MINUTIAE
// Tests for obscure scenarios (none identified for this module currently)
// =============================================================================
