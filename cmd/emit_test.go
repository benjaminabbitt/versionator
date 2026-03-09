package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// TEST HELPERS
// =============================================================================

// captureStdout captures stdout output from a function.
// Used by emit tests to verify command output without modifying global state.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the emit command: listing and
// validating supported output formats for version emission.
// =============================================================================

// TestEmit_SupportedFormats_ReturnsAllExpectedFormats verifies that the emit
// package advertises all language formats that users depend on.
//
// Why: Users rely on emit supporting their target language. If a format is
// accidentally removed from the supported list, builds would fail silently
// or produce errors. This test acts as a contract for supported formats.
//
// What: Checks that every expected format (python, json, yaml, go, c, etc.)
// appears in the SupportedFormats() return value.
func TestEmit_SupportedFormats_ReturnsAllExpectedFormats(t *testing.T) {
	// Precondition: emit package is properly initialized
	formats := emit.SupportedFormats()

	// Action: verify each expected format is present
	expectedFormats := []string{
		"python", "json", "yaml", "go", "c", "c-header", "cpp", "cpp-header",
		"js", "ts", "java", "kotlin", "csharp", "php", "swift", "ruby", "rust",
	}

	// Expected: all formats are found in the supported list
	for _, expected := range expectedFormats {
		found := false
		for _, f := range formats {
			if f == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected format %s to be supported", expected)
		}
	}
}

// TestEmit_IsValidFormat_DistinguishesValidFromInvalid verifies the format
// validation function correctly accepts known formats and rejects unknown ones.
//
// Why: Invalid format names passed to emit would cause confusing errors or
// silent failures. Early validation with clear error messages improves UX.
// This test ensures the validation boundary is correctly implemented.
//
// What: Tests that all known formats return true from IsValidFormat, while
// unknown format strings (including empty string) return false.
func TestEmit_IsValidFormat_DistinguishesValidFromInvalid(t *testing.T) {
	// Precondition: emit package has a defined set of valid formats

	// Action: test all valid formats
	validFormats := []string{
		"python", "json", "yaml", "go", "c", "c-header", "cpp", "cpp-header",
		"js", "ts", "java", "kotlin", "csharp", "php", "swift", "ruby", "rust",
	}

	// Expected: each valid format returns true
	for _, format := range validFormats {
		if !emit.IsValidFormat(format) {
			t.Errorf("expected %s to be a valid format", format)
		}
	}

	// Action: test invalid formats
	invalidFormats := []string{"invalid", "perl", "lua", ""}

	// Expected: each invalid format returns false
	for _, format := range invalidFormats {
		if emit.IsValidFormat(format) {
			t.Errorf("expected %s to be an invalid format", format)
		}
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for stability flags that control whether prerelease/metadata come from
// the VERSION file (stable=true) or are rendered from templates (stable=false).
// =============================================================================

// TestEmit_StabilityFalse_RendersTemplatesForPrereleaseAndMetadata verifies
// that when both prerelease and metadata have stable=false, their templates
// are rendered and included in the output.
//
// Why: The stability flag is a core feature allowing CI/CD pipelines to inject
// dynamic values (commit SHA, build number) into versions. If templates aren't
// rendered when stable=false, builds would have incorrect version strings.
//
// What: Creates a VERSION file with base version "1.2.3" (no prerelease or
// metadata), configures stable=false with templates "alpha" and "build99",
// then verifies emit outputs "1.2.3-alpha+build99".
func TestEmit_StabilityFalse_RendersTemplatesForPrereleaseAndMetadata(t *testing.T) {
	// Precondition: temporary directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Precondition: VERSION file contains only base version (no prerelease/metadata)
	err = os.WriteFile("VERSION", []byte("1.2.3\n"), 0644)
	require.NoError(t, err)

	// Precondition: config has stable=false for both prerelease and metadata
	cfg := &config.Config{
		Prefix: "v",
		PreRelease: config.PreReleaseConfig{
			Template: "alpha",
			Stable:   false,
		},
		Metadata: config.MetadataConfig{
			Template: "build99",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: execute emit with template that includes prerelease and metadata
	rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"})

	output := captureStdout(func() {
		err = rootCmd.Execute()
	})

	// Expected: templates are rendered, producing "1.2.3-alpha+build99"
	assert.NoError(t, err)
	assert.Contains(t, output, "1.2.3-alpha+build99")

	rootCmd.SetArgs(nil)
}

// TestEmit_StabilityTrue_UsesPrereleaseAndMetadataFromVersionFile verifies
// that when both prerelease and metadata have stable=true, values are taken
// from the VERSION file and templates are ignored.
//
// Why: For release builds, the VERSION file is the source of truth. Users
// manually set the prerelease (e.g., "-rc.1") and metadata when ready for
// release. If templates overrode these values, releases would be incorrect.
//
// What: Creates a VERSION file with "1.2.3-beta+build42", configures
// stable=true (with templates that would produce different values), then
// verifies emit outputs the VERSION file values unchanged.
func TestEmit_StabilityTrue_UsesPrereleaseAndMetadataFromVersionFile(t *testing.T) {
	// Precondition: temporary directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Precondition: VERSION file contains full version with prerelease and metadata
	err = os.WriteFile("VERSION", []byte("1.2.3-beta+build42\n"), 0644)
	require.NoError(t, err)

	// Precondition: config has stable=true with templates that differ from VERSION
	cfg := &config.Config{
		Prefix: "",
		PreRelease: config.PreReleaseConfig{
			Template: "alpha",  // This should be ignored when stable=true
			Stable:   true,
		},
		Metadata: config.MetadataConfig{
			Template: "build99",  // This should be ignored when stable=true
			Stable:   true,
		},
	}
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: execute emit with template
	rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"})

	output := captureStdout(func() {
		err = rootCmd.Execute()
	})

	// Expected: VERSION file values used, templates ignored, outputs "1.2.3-beta+build42"
	assert.NoError(t, err)
	assert.Contains(t, output, "1.2.3-beta+build42")

	rootCmd.SetArgs(nil)
}

// TestEmit_MixedStability_CombinesVersionFileAndTemplateValues verifies that
// prerelease and metadata stability flags operate independently.
//
// Why: Complex workflows may need prerelease from VERSION (for release
// candidates) while metadata is dynamic (for build traceability). This test
// ensures the flags don't interfere with each other.
//
// What: Creates VERSION with "2.0.0-rc.1" (prerelease, no metadata), sets
// prerelease.stable=true and metadata.stable=false with template "dynamic-meta",
// verifies output is "2.0.0-rc.1+dynamic-meta".
func TestEmit_MixedStability_CombinesVersionFileAndTemplateValues(t *testing.T) {
	// Precondition: temporary directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Precondition: VERSION file has prerelease but no metadata
	err = os.WriteFile("VERSION", []byte("2.0.0-rc.1\n"), 0644)
	require.NoError(t, err)

	// Precondition: prerelease stable (from VERSION), metadata dynamic (from template)
	cfg := &config.Config{
		Prefix: "",
		PreRelease: config.PreReleaseConfig{
			Template: "should-be-ignored",
			Stable:   true,
		},
		Metadata: config.MetadataConfig{
			Template: "dynamic-meta",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: execute emit with template
	rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"})

	output := captureStdout(func() {
		err = rootCmd.Execute()
	})

	// Expected: prerelease from VERSION (rc.1), metadata from template (dynamic-meta)
	assert.NoError(t, err)
	assert.Contains(t, output, "2.0.0-rc.1+dynamic-meta")

	rootCmd.SetArgs(nil)
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and unusual but valid configurations.
// =============================================================================

// TestEmit_EmptyTemplatesWithStabilityFalse_ProducesCleanVersion verifies that
// empty template strings with stable=false result in no prerelease or metadata
// being appended to the version.
//
// Why: Empty templates are a valid way to produce clean versions without
// prerelease or metadata even when stable=false. This ensures empty strings
// don't cause errors or produce malformed output like trailing dashes.
//
// What: Creates VERSION with "1.0.0", configures empty templates with
// stable=false, verifies output is exactly "1.0.0" with no dash or plus.
func TestEmit_EmptyTemplatesWithStabilityFalse_ProducesCleanVersion(t *testing.T) {
	// Precondition: temporary directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Precondition: VERSION file with base version only
	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	// Precondition: config with empty templates and stable=false
	cfg := &config.Config{
		Prefix: "",
		PreRelease: config.PreReleaseConfig{
			Template: "",
			Stable:   false,
		},
		Metadata: config.MetadataConfig{
			Template: "",
			Stable:   false,
		},
	}
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: execute emit with template
	rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"})

	output := captureStdout(func() {
		err = rootCmd.Execute()
	})

	// Expected: clean version "1.0.0" with no prerelease dash or metadata plus
	assert.NoError(t, err)
	assert.Contains(t, output, "1.0.0")
	assert.NotContains(t, output, "-")
	assert.NotContains(t, output, "+")

	rootCmd.SetArgs(nil)
}

// =============================================================================
// EMIT DUMP TESTS
// =============================================================================

// TestEmitDump_ValidFormat_OutputsTemplate verifies that the emit dump command
// outputs the embedded template for a valid format.
//
// Why: Users need to customize templates, which starts by dumping the default.
//
// What: Run "output emit dump python", verify output contains template content.
func TestEmitDump_ValidFormat_OutputsTemplate(t *testing.T) {
	// Precondition: temp directory (no VERSION needed for dump)
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Action: Execute "output emit dump python"
	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "dump", "python"})
		err = rootCmd.Execute()
	})

	// Expected: Command succeeds and outputs Python template content
	assert.NoError(t, err)
	// Python templates contain __version__ variable assignment
	assert.Contains(t, output, "__version__")

	rootCmd.SetArgs(nil)
}

// TestEmitDump_InvalidFormat_ReturnsError verifies that the emit dump command
// fails with an error for an unknown format.
//
// Why: Users need clear feedback when requesting an invalid format.
//
// What: Run "output emit dump invalid", verify error is returned.
func TestEmitDump_InvalidFormat_ReturnsError(t *testing.T) {
	// Precondition: temp directory
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Action: Execute "output emit dump invalidformat"
	rootCmd.SetArgs([]string{"output", "emit", "dump", "invalidformat"})
	err = rootCmd.Execute()

	// Expected: Command fails with unsupported format error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")

	rootCmd.SetArgs(nil)
}

// TestEmitDump_WithOutput_WritesToFile verifies that --output flag writes
// template to file instead of stdout.
//
// Why: Users need to save templates to files for customization.
//
// What: Run "output emit dump json --output template.json", verify file created.
func TestEmitDump_WithOutput_WritesToFile(t *testing.T) {
	// Precondition: temp directory
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Action: Execute "output emit dump json --output template.json"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"output", "emit", "dump", "json", "--output", "template.json"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and file is created
	assert.NoError(t, err)

	// Verify file exists and has content
	content, err := os.ReadFile("template.json")
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)

	// Reset the output flag
	dumpOutput = ""
}

// =============================================================================
// EMIT FLAG TESTS
// =============================================================================

// TestEmit_WithPrefixFlag_OverridesVersionPrefix verifies that --prefix flag
// overrides the prefix from VERSION file.
func TestEmit_WithPrefixFlag_OverridesVersionPrefix(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template", "{{Prefix}}{{MajorMinorPatch}}", "--prefix=v"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "v1.0.0", output)
	rootCmd.SetArgs(nil)
	emitPrefixOverride = ""
}

// TestEmit_WithPrefixFlagDefault_UsesDefaultPrefix verifies that --prefix without
// a value uses default "v".
func TestEmit_WithPrefixFlagDefault_UsesDefaultPrefix(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template", "{{Prefix}}{{MajorMinorPatch}}", "--prefix"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "v1.0.0", output)
	rootCmd.SetArgs(nil)
	emitPrefixOverride = ""
}

// TestEmit_WithPrereleaseFlag_RendersOverrideTemplate verifies that --prerelease=value
// overrides the config template.
func TestEmit_WithPrereleaseFlag_RendersOverrideTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	cfg := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "alpha",
			Stable:   true,
		},
	}
	configData, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(".versionator.yaml", configData, 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{PreReleaseWithDash}}", "--prerelease=beta"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "1.0.0-beta", output)
	rootCmd.SetArgs(nil)
	emitPrereleaseTemplate = ""
}

// TestEmit_WithPrereleaseFlagDefault_UsesConfigTemplate verifies that --prerelease
// without a value uses the config template.
func TestEmit_WithPrereleaseFlagDefault_UsesConfigTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	cfg := &config.Config{
		PreRelease: config.PreReleaseConfig{
			Template: "rc1",
			Stable:   true,
		},
	}
	configData, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(".versionator.yaml", configData, 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{PreReleaseWithDash}}", "--prerelease"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "1.0.0-rc1", output)
	rootCmd.SetArgs(nil)
	emitPrereleaseTemplate = ""
}

// TestEmit_WithMetadataFlag_RendersOverrideTemplate verifies that --metadata=value
// overrides the config template.
func TestEmit_WithMetadataFlag_RendersOverrideTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "build1",
			Stable:   true,
		},
	}
	configData, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(".versionator.yaml", configData, 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{MetadataWithPlus}}", "--metadata=build999"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "1.0.0+build999", output)
	rootCmd.SetArgs(nil)
	emitMetadataTemplate = ""
}

// TestEmit_WithMetadataFlagDefault_UsesConfigTemplate verifies that --metadata
// without a value uses the config template.
func TestEmit_WithMetadataFlagDefault_UsesConfigTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Template: "build42",
			Stable:   true,
		},
	}
	configData, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(".versionator.yaml", configData, 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template", "{{MajorMinorPatch}}{{MetadataWithPlus}}", "--metadata"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "1.0.0+build42", output)
	rootCmd.SetArgs(nil)
	emitMetadataTemplate = ""
}

// TestEmit_WithTemplateFile_ReadsFromFile verifies that --template-file reads
// template content from a file.
func TestEmit_WithTemplateFile_ReadsFromFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("2.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)
	_ = os.WriteFile("my-template.txt", []byte("Version is {{MajorMinorPatch}}"), 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "--template-file", "my-template.txt"})
		_ = rootCmd.Execute()
	})

	assert.Equal(t, "Version is 2.0.0", output)
	rootCmd.SetArgs(nil)
	emitTemplateFile = ""
}

// TestEmit_WithOutputFile_WritesToFile verifies that --output writes to file.
func TestEmit_WithOutputFile_WritesToFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.2.3\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "json", "--output", "version.json"})
		_ = rootCmd.Execute()
	})

	// Stdout should show success message
	assert.Contains(t, output, "written to version.json")

	// File should exist with content
	content, err := os.ReadFile("version.json")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "1.2.3")

	rootCmd.SetArgs(nil)
	emitOutput = ""
}

// TestEmit_InvalidFormat_ReturnsError verifies that an unknown format returns an error.
func TestEmit_InvalidFormat_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Reset all emit flags
	emitOutput = ""
	emitTemplate = ""
	emitTemplateFile = ""
	emitPrereleaseTemplate = ""
	emitMetadataTemplate = ""
	emitPrefixOverride = ""

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	rootCmd.SetArgs([]string{"output", "emit", "unknownformat"})
	err := rootCmd.Execute()

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "unsupported format")
	}

	rootCmd.SetArgs(nil)
}

// TestEmit_NoFormatOrTemplate_ReturnsError verifies that emit without format or
// template returns an error.
func TestEmit_NoFormatOrTemplate_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Reset all emit flags
	emitOutput = ""
	emitTemplate = ""
	emitTemplateFile = ""
	emitPrereleaseTemplate = ""
	emitMetadataTemplate = ""
	emitPrefixOverride = ""

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	rootCmd.SetArgs([]string{"output", "emit"})
	err := rootCmd.Execute()

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "format argument required")
	}

	rootCmd.SetArgs(nil)
}

// TestEmit_WithValidFormat_OutputsContent verifies that a valid format outputs
// the rendered template.
func TestEmit_WithValidFormat_OutputsContent(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Reset all emit flags
	emitOutput = ""
	emitTemplate = ""
	emitTemplateFile = ""
	emitPrereleaseTemplate = ""
	emitMetadataTemplate = ""
	emitPrefixOverride = ""

	_ = os.WriteFile("VERSION", []byte("3.2.1\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	output := captureStdout(func() {
		rootCmd.SetArgs([]string{"output", "emit", "json"})
		_ = rootCmd.Execute()
	})

	// JSON output should contain version info
	assert.Contains(t, output, "3.2.1")

	rootCmd.SetArgs(nil)
}
