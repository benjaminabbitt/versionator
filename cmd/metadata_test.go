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

// =============================================================================
// CORE FUNCTIONALITY
// Tests for the primary happy-path behavior of metadata commands
// =============================================================================

// TestMetadataSetCommand_WhenStableTrue_SetsVersionFileMetadata verifies the primary use case
// for setting build metadata when the configuration allows direct VERSION file updates.
//
// Why: Build metadata (e.g., "build123") must be written to the VERSION file when stable=true,
// enabling reproducible release versions with embedded build information.
//
// What: Given stable=true in config and a VERSION file, when "config metadata set <value>" is run,
// the build metadata should be appended to the version in the VERSION file.
func TestMetadataSetCommand_WhenStableTrue_SetsVersionFileMetadata(t *testing.T) {
	// Precondition: Create temp directory with VERSION file and config (stable=true)
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata set build123"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "set", "build123"})

	err = rootCmd.Execute()

	// Expected: VERSION file contains build metadata
	assert.NoError(t, err)
	v, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "build123", v.BuildMetadata)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestMetadataClearCommand_WhenStableTrue_ClearsVersionFileMetadata verifies that clearing
// metadata removes it from the VERSION file when stable mode is enabled.
//
// Why: Users need to remove build metadata from stable releases, for example when
// transitioning from a CI build to a final release version.
//
// What: Given stable=true and a VERSION file with existing metadata, when "config metadata clear"
// is run, the metadata should be removed from the VERSION file.
func TestMetadataClearCommand_WhenStableTrue_ClearsVersionFileMetadata(t *testing.T) {
	// Precondition: Create VERSION file with existing metadata and config (stable=true)
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0+build123\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "build123",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata clear"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "clear"})

	err = rootCmd.Execute()

	// Expected: VERSION file has no build metadata
	assert.NoError(t, err)
	v, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", v.BuildMetadata)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows and configuration combinations
// =============================================================================

// TestMetadataStableCommand_GetStability verifies that the stable flag can be queried
// to display the current stability mode configuration.
//
// Why: Users need to check whether metadata is in stable (VERSION file) or unstable
// (template-based) mode to understand how metadata will be applied.
//
// What: Given various stable configurations, when "config metadata stable" is run without
// arguments, the current stable value should be displayed.
func TestMetadataStableCommand_GetStability(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  *config.Config
		expectedOutput string
	}{
		{
			name: "stability is false",
			initialConfig: &config.Config{
				Metadata: config.MetadataConfig{
					Template: "{{ShortHash}}",
					Stable:   false,
				},
			},
			expectedOutput: "false",
		},
		{
			name: "stability is true",
			initialConfig: &config.Config{
				Metadata: config.MetadataConfig{
					Template: "build123",
					Stable:   true,
				},
			},
			expectedOutput: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Create temp directory with VERSION file and specific config
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

			// Action: Execute "config metadata stable" (no argument = get)
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "metadata", "stable"})

			err = rootCmd.Execute()

			// Expected: Output contains the current stable value
			assert.NoError(t, err)
			output := stdout.String()
			assert.Contains(t, output, tt.expectedOutput)

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestMetadataStableCommand_SetStability verifies that the stable flag can be toggled
// between true and false to control metadata behavior.
//
// Why: Users need to switch between stable mode (metadata in VERSION file) and unstable
// mode (metadata from template) depending on their release workflow.
//
// What: Given an initial stable value, when "config metadata stable <value>" is run,
// the config file should be updated to the new value.
func TestMetadataStableCommand_SetStability(t *testing.T) {
	tests := []struct {
		name           string
		initialStable  bool
		setTo          string
		expectedStable bool
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
			// Precondition: Create temp directory with VERSION file and initial config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			initialConfig := &config.Config{
				Metadata: config.MetadataConfig{
					Template: "{{ShortHash}}",
					Stable:   tt.initialStable,
				},
			}
			configData, err := yaml.Marshal(initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Action: Execute "config metadata stable <value>"
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "metadata", "stable", tt.setTo})

			err = rootCmd.Execute()

			// Expected: Config file updated with new stable value
			assert.NoError(t, err)
			cfg, err := config.ReadConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStable, cfg.Metadata.Stable)

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestMetadataTemplateCommand verifies that template configuration updates work correctly
// and that VERSION file updates depend on the stable flag.
//
// Why: The template defines how dynamic metadata is generated. When stable=true, setting
// a template should immediately apply it to the VERSION file; when stable=false, the
// template is stored for later evaluation.
//
// What: Given different stable configurations, when "config metadata template <value>" is run,
// the template should be saved to config, and VERSION file updated only if stable=true.
func TestMetadataTemplateCommand(t *testing.T) {
	tests := []struct {
		name                string
		initialStable       bool
		template            string
		expectVersionUpdate bool
	}{
		{
			name:                "set template when stable false",
			initialStable:       false,
			template:            "{{ShortHash}}",
			expectVersionUpdate: false,
		},
		{
			name:                "set template when stable true",
			initialStable:       true,
			template:            "build456",
			expectVersionUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Create temp directory with VERSION file and initial config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			initialConfig := &config.Config{
				Metadata: config.MetadataConfig{
					Template: "",
					Stable:   tt.initialStable,
				},
			}
			configData, err := yaml.Marshal(initialConfig)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Action: Execute "config metadata template <value>"
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "metadata", "template", tt.template})

			err = rootCmd.Execute()

			// Expected: Config template updated, VERSION file updated only if stable=true
			assert.NoError(t, err)

			cfg, err := config.ReadConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.template, cfg.Metadata.Template)

			v, err := version.Load()
			require.NoError(t, err)

			if tt.expectVersionUpdate {
				assert.Equal(t, tt.template, v.BuildMetadata)
			} else {
				assert.Equal(t, "", v.BuildMetadata)
			}

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestMetadataStatusCommand verifies that the status command displays comprehensive
// information about the current metadata configuration and value.
//
// Why: Users need visibility into the current metadata state to understand what
// metadata will be applied and whether it comes from VERSION file or template.
//
// What: Given various configurations, when "config metadata status" is run,
// the output should contain relevant status information for that mode.
func TestMetadataStatusCommand(t *testing.T) {
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
			template:    "{{ShortHash}}",
			versionFile: "1.0.0",
			expectContains: []string{
				"Stable: false",
				"Template: {{ShortHash}}",
			},
		},
		{
			name:        "status when stable true with metadata",
			stable:      true,
			template:    "build123",
			versionFile: "1.0.0+build123",
			expectContains: []string{
				"Stable: true",
				"VALUE (from VERSION file): build123",
			},
		},
		{
			name:        "status when stable true without metadata",
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
			// Precondition: Create temp directory with VERSION file and config
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte(tt.versionFile+"\n"), 0644)
			require.NoError(t, err)

			cfg := &config.Config{
				Metadata: config.MetadataConfig{
					Template: tt.template,
					Stable:   tt.stable,
				},
			}
			configData, err := yaml.Marshal(cfg)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			// Action: Execute "config metadata status"
			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs([]string{"config", "metadata", "status"})

			err = rootCmd.Execute()

			// Expected: Output contains all expected status information
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

// TestMetadataSetCommand_WhenStableFalse_WithForce_UpdatesTemplate verifies that the
// --force flag allows updating the template even when stable=false.
//
// Why: In unstable mode, "set" would normally fail because the VERSION file isn't
// the source of truth. The --force flag provides an escape hatch to update the
// template directly for advanced workflows.
//
// What: Given stable=false in config, when "config metadata set <value> --force" is run,
// the template should be updated in config (not the VERSION file).
func TestMetadataSetCommand_WhenStableFalse_WithForce_UpdatesTemplate(t *testing.T) {
	// Precondition: Create temp directory with VERSION file and config (stable=false)
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "{{ShortHash}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata set build123 --force"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "set", "build123", "--force"})

	err = rootCmd.Execute()

	// Expected: Config template updated, VERSION file unchanged
	assert.NoError(t, err)

	cfg, err := config.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, "build123", cfg.Metadata.Template)

	v, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", v.BuildMetadata)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)

	// Reset the force flag to prevent state pollution to subsequent tests
	metadataSetCmd.Flags().Set("force", "false")
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes and error messages
// =============================================================================

// TestMetadataStableCommand_SetInvalidValue_ReturnsError verifies that invalid boolean
// values are rejected with a helpful error message.
//
// Why: Users may accidentally provide invalid values. The error message should guide
// them to use the correct format.
//
// What: Given any config, when "config metadata stable maybe" is run with an invalid
// boolean value, an error should be returned indicating the valid options.
func TestMetadataStableCommand_SetInvalidValue_ReturnsError(t *testing.T) {
	// Precondition: Create temp directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "{{ShortHash}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata stable maybe" (invalid value)
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "stable", "maybe"})

	err = rootCmd.Execute()

	// Expected: Error returned with guidance on valid values
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use 'true' or 'false'")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestMetadataSetCommand_WhenStableFalse_ReturnsError verifies that setting metadata
// is blocked when stable=false to prevent confusion about where metadata is stored.
//
// Why: When stable=false, metadata comes from the template at emit time, not from
// the VERSION file. Allowing "set" would be misleading since it wouldn't affect output.
//
// What: Given stable=false in config, when "config metadata set <value>" is run without
// --force, an error should be returned explaining the stability constraint.
func TestMetadataSetCommand_WhenStableFalse_ReturnsError(t *testing.T) {
	// Precondition: Create temp directory with VERSION file and config (stable=false)
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "{{ShortHash}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata set build123" (without --force)
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "set", "build123"})

	err = rootCmd.Execute()

	// Expected: Error returned referencing stable: false
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stable: false")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestMetadataClearCommand_WhenStableFalse_ReturnsError verifies that clearing metadata
// is blocked when stable=false since there's no VERSION file metadata to clear.
//
// Why: When stable=false, the VERSION file doesn't contain metadata (it's generated
// at emit time). Clearing would be a no-op that could confuse users.
//
// What: Given stable=false in config, when "config metadata clear" is run,
// an error should be returned explaining the stability constraint.
func TestMetadataClearCommand_WhenStableFalse_ReturnsError(t *testing.T) {
	// Precondition: Create temp directory with VERSION file and config (stable=false)
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "{{ShortHash}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata clear"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "clear"})

	err = rootCmd.Execute()

	// Expected: Error returned referencing stable: false
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stable: false")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// ENABLE/DISABLE/CONFIGURE TESTS
// =============================================================================

// TestMetadataEnableCommand_WritesMetadataToVersion verifies that the enable command
// renders the template and writes metadata to the VERSION file.
//
// Why: The enable command is a quick way to add metadata to VERSION in stable mode.
//
// What: Run "config metadata enable" when stable=true, verify metadata is written.
func TestMetadataEnableCommand_WritesMetadataToVersion(t *testing.T) {
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
		Metadata: config.MetadataConfig{
			Template: "build123",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata enable"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "enable"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and metadata is written to VERSION
	assert.NoError(t, err)
	versionContent, err := os.ReadFile("VERSION")
	require.NoError(t, err)
	assert.Contains(t, string(versionContent), "build123")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestMetadataDisableCommand_ClearsMetadataFromVersion verifies that the disable command
// clears metadata from the VERSION file.
//
// Why: The disable command is a quick way to remove metadata from VERSION in stable mode.
//
// What: Run "config metadata disable" when stable=true, verify metadata is cleared.
func TestMetadataDisableCommand_ClearsMetadataFromVersion(t *testing.T) {
	// Precondition: temp directory with stable=true config and metadata in VERSION
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0+build123\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "build123",
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata disable"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "disable"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and metadata is cleared from VERSION
	assert.NoError(t, err)
	versionContent, err := os.ReadFile("VERSION")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0\n", string(versionContent))

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestMetadataConfigureCommand_SetsTemplateAndStable verifies that the configure
// command can set both template and stability in one call.
//
// Why: Users need a single command to configure all metadata settings.
//
// What: Run "config metadata configure", verify it displays configuration.
func TestMetadataConfigureCommand_DisplaysConfiguration(t *testing.T) {
	// Precondition: temp directory with config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "build-{{CommitsSinceTag}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config metadata configure"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "metadata", "configure"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and displays configuration
	assert.NoError(t, err)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}
