package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// TEMPLATE HANDLER TESTS
// Tests for the consolidated template handling logic used by both
// prerelease and metadata template commands.
// =============================================================================

// TestTemplateCommand_ShowsCurrentTemplate verifies that calling the template
// command without arguments displays the current template configuration.
func TestTemplateCommand_ShowsCurrentTemplate(t *testing.T) {
	tests := []struct {
		name     string
		command  []string
		template string
		stable   bool
		wantOut  []string
	}{
		{
			name:     "prerelease template with stable true",
			command:  []string{"config", "prerelease", "template"},
			template: "alpha-{{CommitsSinceTag}}",
			stable:   true,
			wantOut:  []string{"Stable: true", "Template: alpha-{{CommitsSinceTag}}"},
		},
		{
			name:     "prerelease template with stable false",
			command:  []string{"config", "prerelease", "template"},
			template: "beta",
			stable:   false,
			wantOut:  []string{"Stable: false", "Template: beta"},
		},
		{
			name:     "metadata template with stable true",
			command:  []string{"config", "metadata", "template"},
			template: "build-{{ShortHash}}",
			stable:   true,
			wantOut:  []string{"Stable: true", "Template: build-{{ShortHash}}"},
		},
		{
			name:     "metadata template with stable false",
			command:  []string{"config", "metadata", "template"},
			template: "{{BuildDateTimeCompact}}",
			stable:   false,
			wantOut:  []string{"Stable: false", "Template: {{BuildDateTimeCompact}}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			// Create config based on whether it's prerelease or metadata
			var cfg *config.Config
			if tt.command[1] == "prerelease" {
				cfg = &config.Config{
					PreRelease: config.PreReleaseConfig{
						Template: tt.template,
						Stable:   tt.stable,
					},
				}
			} else {
				cfg = &config.Config{
					Metadata: config.MetadataConfig{
						Template: tt.template,
						Stable:   tt.stable,
					},
				}
			}
			configData, err := yaml.Marshal(cfg)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs(tt.command)

			err = rootCmd.Execute()

			assert.NoError(t, err)
			output := stdout.String()
			for _, want := range tt.wantOut {
				assert.Contains(t, output, want)
			}

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestTemplateCommand_SetsTemplateInDynamicMode verifies that setting a template
// in dynamic mode (stable=false) saves to config but not VERSION file.
func TestTemplateCommand_SetsTemplateInDynamicMode(t *testing.T) {
	tests := []struct {
		name        string
		command     []string
		newTemplate string
		configField string
	}{
		{
			name:        "prerelease dynamic mode",
			command:     []string{"config", "prerelease", "template", "rc-{{CommitsSinceTag}}"},
			newTemplate: "rc-{{CommitsSinceTag}}",
			configField: "prerelease",
		},
		{
			name:        "metadata dynamic mode",
			command:     []string{"config", "metadata", "template", "{{BuildDateTimeCompact}}"},
			newTemplate: "{{BuildDateTimeCompact}}",
			configField: "metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			// Create config with stable=false
			var cfg *config.Config
			if tt.configField == "prerelease" {
				cfg = &config.Config{
					PreRelease: config.PreReleaseConfig{
						Template: "",
						Stable:   false,
					},
				}
			} else {
				cfg = &config.Config{
					Metadata: config.MetadataConfig{
						Template: "",
						Stable:   false,
					},
				}
			}
			configData, err := yaml.Marshal(cfg)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs(tt.command)

			err = rootCmd.Execute()

			assert.NoError(t, err)
			assert.Contains(t, stdout.String(), "Template will be rendered at output time")

			// Verify template was saved to config
			savedCfg, err := config.ReadConfig()
			require.NoError(t, err)
			if tt.configField == "prerelease" {
				assert.Equal(t, tt.newTemplate, savedCfg.PreRelease.Template)
			} else {
				assert.Equal(t, tt.newTemplate, savedCfg.Metadata.Template)
			}

			// Verify VERSION file unchanged
			versionContent, err := os.ReadFile("VERSION")
			require.NoError(t, err)
			assert.Equal(t, "1.0.0\n", string(versionContent))

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestTemplateCommand_SetsTemplateInStableMode verifies that setting a template
// in stable mode (stable=true) renders and writes to VERSION file.
func TestTemplateCommand_SetsTemplateInStableMode(t *testing.T) {
	tests := []struct {
		name        string
		command     []string
		newTemplate string
		configField string
		wantVersion string
	}{
		{
			name:        "prerelease stable mode",
			command:     []string{"config", "prerelease", "template", "alpha"},
			newTemplate: "alpha",
			configField: "prerelease",
			wantVersion: "1.0.0-alpha\n",
		},
		{
			name:        "metadata stable mode",
			command:     []string{"config", "metadata", "template", "build123"},
			newTemplate: "build123",
			configField: "metadata",
			wantVersion: "1.0.0+build123\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
			require.NoError(t, err)

			// Create config with stable=true
			var cfg *config.Config
			if tt.configField == "prerelease" {
				cfg = &config.Config{
					PreRelease: config.PreReleaseConfig{
						Template: "",
						Stable:   true,
					},
				}
			} else {
				cfg = &config.Config{
					Metadata: config.MetadataConfig{
						Template: "",
						Stable:   true,
					},
				}
			}
			configData, err := yaml.Marshal(cfg)
			require.NoError(t, err)
			err = os.WriteFile(".versionator.yaml", configData, 0644)
			require.NoError(t, err)

			var stdout bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetArgs(tt.command)

			err = rootCmd.Execute()

			assert.NoError(t, err)
			assert.Contains(t, stdout.String(), "set to:")

			// Verify VERSION file was updated
			versionContent, err := os.ReadFile("VERSION")
			require.NoError(t, err)
			assert.Equal(t, tt.wantVersion, string(versionContent))

			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

// TestTemplateCommand_ShowsRenderedValue verifies that when showing a template,
// the rendered value is also displayed if rendering succeeds.
func TestTemplateCommand_ShowsRenderedValue(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.2.3\n"), 0644)
	require.NoError(t, err)

	// Template that uses version data - Major will render to "1"
	cfg := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "build-{{Major}}",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prerelease", "template"})

	err = rootCmd.Execute()

	assert.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Template: build-{{Major}}")
	assert.Contains(t, output, "Rendered value: build-1")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}
