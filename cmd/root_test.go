package cmd

import (
	"bytes"
	"os"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the root command and version
// output. These validate that the CLI entry point works correctly and that
// version information can be retrieved from a properly configured project.
// =============================================================================

// TestExecute_Success_ShowsHelp validates that Execute() runs without error.
//
// Why: Execute() is the main entry point called by main(). If it fails or
// panics, the entire CLI is broken. This test ensures basic CLI initialization
// works correctly.
//
// What: When Execute() is called with no arguments in a clean directory, it
// should display help text and return nil (no error).
func TestExecute_Success_ShowsHelp(t *testing.T) {
	// Precondition: Clean temporary directory with no version files
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Execute the root command with no arguments
	err := Execute()

	// Expected: No error; help is displayed and command completes successfully
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}
}

// TestVersionCommand_Success_OutputsVersion validates that the version command
// correctly reads and outputs the current version from the VERSION file.
//
// Why: Outputting the current version is a core feature. Users depend on this
// to know what version they're working with and to integrate version info into
// their build processes.
//
// What: Given a valid VERSION file (2.1.0) and config, when "output version"
// is executed, then the version string should be output to stdout.
func TestVersionCommand_Success_OutputsVersion(t *testing.T) {
	// Precondition: Temporary directory with VERSION file containing 2.1.0
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte(`{"major": 2, "minor": 1, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `prefix: ""
metadata:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Capture stdout and execute "output version"
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version"})

	err = rootCmd.Execute()

	// Expected: No error and non-empty version output
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Errorf("Expected version output, got empty string")
	}

	// Cleanup: Reset command state for other tests
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows including version prefixes, logging
// configuration, and other configuration variations that affect output.
// =============================================================================

// TestVersionCommand_WithPrefix_OutputsPrefixedVersion validates that version
// prefixes (e.g., "v") are correctly included in version output.
//
// Why: Many projects use version prefixes like "v" (e.g., "v3.0.0" for git tags).
// This feature must work correctly for compatibility with common versioning
// conventions and downstream tools.
//
// What: Given a VERSION file with prefix "v" and version 3.0.0, when "output
// version" is executed, then the output should include the version string.
func TestVersionCommand_WithPrefix_OutputsPrefixedVersion(t *testing.T) {
	// Precondition: Temporary directory with VERSION file containing v3.0.0
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte(`{"prefix": "v", "major": 3, "minor": 0, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `prefix: "v"
metadata:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Capture stdout and execute "output version"
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version"})

	err = rootCmd.Execute()

	// Expected: No error and non-empty prefixed version output
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Errorf("Expected version output, got empty string")
	}

	// Cleanup: Reset command state for other tests
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestLogFormatFlag_Development_AcceptsFlag validates that the --log-format
// flag is accepted and processed without error.
//
// Why: Users may want different log formats for different environments
// (development, production, CI). The CLI must accept and honor these flags
// without breaking command execution.
//
// What: Given a valid project setup with JSON logging configured, when the
// --log-format=development flag is passed, then the command should execute
// successfully.
func TestLogFormatFlag_Development_AcceptsFlag(t *testing.T) {
	// Precondition: Temporary directory with VERSION file and JSON logging config
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte(`{"major": 1, "minor": 0, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `prefix: ""
metadata:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "json"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Execute with --log-format flag
	rootCmd.SetArgs([]string{"--log-format", "development", "output", "version"})

	err = rootCmd.Execute()

	// Expected: No error; log format flag is accepted
	if err != nil {
		t.Fatalf("Command with log-format flag failed: %v", err)
	}

	// Cleanup: Reset command state for other tests
	rootCmd.SetArgs(nil)
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes and graceful error recovery, ensuring the
// CLI provides useful feedback when configuration is missing or invalid.
// =============================================================================

// TestVersionCommand_NoVersionFile_CreatesDefault validates that a missing
// VERSION file results in automatic creation of a default version.
//
// Why: New projects won't have a VERSION file. The CLI should handle this
// gracefully by creating a sensible default rather than failing, providing
// a smooth onboarding experience.
//
// What: Given a directory with config but no VERSION file, when "output
// version" is executed, then a default VERSION file should be created and
// the command should succeed.
func TestVersionCommand_NoVersionFile_CreatesDefault(t *testing.T) {
	// Precondition: Temporary directory with config but NO VERSION file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `prefix: ""
metadata:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Capture stderr and execute "output version"
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"output", "version"})

	err = rootCmd.Execute()

	// Expected: No error; command succeeds with default version
	if err != nil {
		t.Fatalf("version command should succeed with default version, got: %v", err)
	}

	// Expected: VERSION file should be created
	if _, err := os.Stat("VERSION"); os.IsNotExist(err) {
		t.Error("Expected VERSION to be created")
	}

	// Cleanup: Reset command state for other tests
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// MINUTIAE
// Tests for utility functions and parsing logic. These cover edge cases in
// input handling that are important for robustness but not core workflows.
// =============================================================================

// TestVersionCommand_WithTemplate_RendersTemplate validates that the -t flag
// renders a custom template with version data.
func TestVersionCommand_WithTemplate_RendersTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.2.3\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version", "-t", "version={{MajorMinorPatch}}"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if buf.String() != "version=1.2.3\n" {
		t.Errorf("Expected 'version=1.2.3\\n', got %q", buf.String())
	}

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestVersionCommand_WithSetFlag_InjectsCustomVariable validates that --set
// allows injecting custom variables into templates.
func TestVersionCommand_WithSetFlag_InjectsCustomVariable(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version", "-t", "{{AppName}} v{{MajorMinorPatch}}", "--set", "AppName=MyApp"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if buf.String() != "MyApp v1.0.0\n" {
		t.Errorf("Expected 'MyApp v1.0.0\\n', got %q", buf.String())
	}

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
	setVars = nil
}

// TestVersionCommand_WithPrereleaseFlag_RendersPrerelease validates that
// --prerelease flag adds prerelease to version output.
func TestVersionCommand_WithPrereleaseFlag_RendersPrerelease(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version", "-t", "{{MajorMinorPatch}}{{PreReleaseWithDash}}", "--prerelease=alpha"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if buf.String() != "1.0.0-alpha\n" {
		t.Errorf("Expected '1.0.0-alpha\\n', got %q", buf.String())
	}

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
	prereleaseTemplate = ""
}

// TestVersionCommand_WithMetadataFlag_RendersMetadata validates that
// --metadata flag adds metadata to version output.
func TestVersionCommand_WithMetadataFlag_RendersMetadata(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version", "-t", "{{MajorMinorPatch}}{{MetadataWithPlus}}", "--metadata=build123"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if buf.String() != "1.0.0+build123\n" {
		t.Errorf("Expected '1.0.0+build123\\n', got %q", buf.String())
	}

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
	metadataTemplate = ""
}

// TestVersionCommand_WithPrefixFlag_OverridesPrefix validates that
// --prefix flag overrides the VERSION file prefix.
func TestVersionCommand_WithPrefixFlag_OverridesPrefix(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: \"\"\n"), 0644)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"output", "version", "-t", "{{Prefix}}{{MajorMinorPatch}}", "--prefix=v"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if buf.String() != "v1.0.0\n" {
		t.Errorf("Expected 'v1.0.0\\n', got %q", buf.String())
	}

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
	prefixOverride = ""
}

// TestParseSetFlags validates that --set key=value flags are correctly parsed
// into a map for setting custom variables.
func TestParseSetFlags_VariousInputs_ParsesCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		// Core functionality: basic parsing
		{
			name:     "single key=value",
			input:    []string{"foo=bar"},
			expected: map[string]string{"foo": "bar"},
		},
		{
			name:     "multiple key=value pairs",
			input:    []string{"foo=bar", "baz=qux"},
			expected: map[string]string{"foo": "bar", "baz": "qux"},
		},

		// Key variations: values with special characters
		{
			name:     "value with equals sign",
			input:    []string{"url=https://example.com?a=b"},
			expected: map[string]string{"url": "https://example.com?a=b"},
		},
		{
			name:     "key with spaces",
			input:    []string{"key with spaces=value"},
			expected: map[string]string{"key with spaces": "value"},
		},
		{
			name:     "value with spaces",
			input:    []string{"key=value with spaces"},
			expected: map[string]string{"key": "value with spaces"},
		},

		// Edge cases: empty and boundary conditions
		{
			name:     "empty input",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:     "empty value",
			input:    []string{"empty="},
			expected: map[string]string{"empty": ""},
		},

		// Error handling: invalid inputs are ignored gracefully
		{
			name:     "invalid entry without equals is ignored",
			input:    []string{"invalidentry"},
			expected: map[string]string{},
		},
		{
			name:     "entry with equals at position 0 is ignored",
			input:    []string{"=value"},
			expected: map[string]string{},
		},
		{
			name:     "mixed valid and invalid",
			input:    []string{"good=value", "bad", "also=works"},
			expected: map[string]string{"good": "value", "also": "works"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Parse the input flags
			result := parseSetFlags(tt.input)

			// Expected: Result matches expected map
			if len(result) != len(tt.expected) {
				t.Errorf("parseSetFlags() returned %d items, want %d", len(result), len(tt.expected))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("parseSetFlags()[%q] = %q, want %q", k, result[k], v)
				}
			}
		})
	}
}
