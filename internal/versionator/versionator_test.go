package versionator

import (
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the versionator package:
// retrieving templates from config and rendering them with version data.
// =============================================================================

// TestGetPreReleaseTemplate_WithConfigFile_ReturnsConfiguredTemplate validates
// that the pre-release template is correctly read from the configuration file.
//
// Why: The primary use case is reading a user-configured pre-release template.
// This ensures the config parsing and template extraction work correctly.
//
// What: Given a valid config file with a prerelease template configured,
// GetPreReleaseTemplate should return that exact template string.
func TestGetPreReleaseTemplate_WithConfigFile_ReturnsConfiguredTemplate(t *testing.T) {
	// Precondition: Working directory with a valid .versionator.yaml config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `prerelease:
  template: "alpha-{{CommitsSinceTag}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Retrieve the pre-release template
	template, err := GetPreReleaseTemplate()

	// Expected: The configured template is returned without error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if template != "alpha-{{CommitsSinceTag}}" {
		t.Errorf("Expected template 'alpha-{{CommitsSinceTag}}', got '%s'", template)
	}
}

// TestGetMetadataTemplate_WithConfigFile_ReturnsConfiguredTemplate validates
// that the metadata template is correctly read from the configuration file.
//
// Why: The primary use case is reading a user-configured metadata template.
// This ensures the config parsing and template extraction work correctly.
//
// What: Given a valid config file with a metadata template configured,
// GetMetadataTemplate should return that exact template string.
func TestGetMetadataTemplate_WithConfigFile_ReturnsConfiguredTemplate(t *testing.T) {
	// Precondition: Working directory with a valid .versionator.yaml config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `metadata:
  template: "{{ShortHash}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Retrieve the metadata template
	template, err := GetMetadataTemplate()

	// Expected: The configured template is returned without error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if template != "{{ShortHash}}" {
		t.Errorf("Expected template '{{ShortHash}}', got '%s'", template)
	}
}

// TestRenderPreRelease_WithTemplate_RendersVersionData validates that
// the pre-release template is rendered with actual version data.
//
// Why: This is the core rendering functionality - taking a template and
// substituting version variables. Users depend on this for CI/CD versioning.
//
// What: Given a config with a template using {{Major}} and a VERSION file
// containing "1.2.3", RenderPreRelease should return "alpha-1".
func TestRenderPreRelease_WithTemplate_RendersVersionData(t *testing.T) {
	// Precondition: Working directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `prerelease:
  template: "alpha-{{Major}}"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Render the pre-release template
	result, err := RenderPreRelease()

	// Expected: Template variables are substituted with version data
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "alpha-1" {
		t.Errorf("Expected 'alpha-1', got '%s'", result)
	}
}

// TestRenderMetadata_WithTemplate_RendersVersionData validates that
// the metadata template is rendered with actual version data.
//
// Why: This is the core rendering functionality for build metadata.
// Users depend on this for embedding version info in builds.
//
// What: Given a config with a template using {{Minor}} and a VERSION file
// containing "2.5.9", RenderMetadata should return "build-5".
func TestRenderMetadata_WithTemplate_RendersVersionData(t *testing.T) {
	// Precondition: Working directory with VERSION file and config
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte("2.5.9"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `metadata:
  template: "build-{{Minor}}"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Render the metadata template
	result, err := RenderMetadata()

	// Expected: Template variables are substituted with version data
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "build-5" {
		t.Errorf("Expected 'build-5', got '%s'", result)
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows: missing config files, which represent
// the default/unconfigured state that many users will encounter.
// =============================================================================

// TestGetPreReleaseTemplate_NoConfigFile_ReturnsEmptyDefault validates that
// a missing config file results in an empty default template.
//
// Why: Users may run versionator before creating a config file. The tool
// should degrade gracefully rather than error, returning empty templates.
//
// What: Given no .versionator.yaml file exists, GetPreReleaseTemplate
// should return an empty string with no error.
func TestGetPreReleaseTemplate_NoConfigFile_ReturnsEmptyDefault(t *testing.T) {
	// Precondition: Working directory with no config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Retrieve the pre-release template without config
	template, err := GetPreReleaseTemplate()

	// Expected: Empty template returned (templates must be explicitly configured)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if template != "" {
		t.Errorf("Expected default template to be empty, got '%s'", template)
	}
}

// TestGetMetadataTemplate_NoConfigFile_ReturnsEmptyDefault validates that
// a missing config file results in an empty default template.
//
// Why: Users may run versionator before creating a config file. The tool
// should degrade gracefully rather than error, returning empty templates.
//
// What: Given no .versionator.yaml file exists, GetMetadataTemplate
// should return an empty string with no error.
func TestGetMetadataTemplate_NoConfigFile_ReturnsEmptyDefault(t *testing.T) {
	// Precondition: Working directory with no config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Retrieve the metadata template without config
	template, err := GetMetadataTemplate()

	// Expected: Empty template returned (templates must be explicitly configured)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if template != "" {
		t.Errorf("Expected default template to be empty, got '%s'", template)
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes: malformed config files and version
// file access errors. These validate proper error propagation.
// =============================================================================

// TestRenderPreRelease_ConfigReadError_ReturnsError validates that config
// parsing errors are properly propagated to the caller.
//
// Why: Users need clear feedback when their config file is malformed.
// Silent failures or unclear errors would make debugging difficult.
//
// What: Given a .versionator.yaml with invalid YAML syntax,
// RenderPreRelease should return an error.
func TestRenderPreRelease_ConfigReadError_ReturnsError(t *testing.T) {
	// Precondition: Working directory with malformed config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile(".versionator.yaml", []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Attempt to render with invalid config
	_, err = RenderPreRelease()

	// Expected: Error returned for malformed YAML
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

// TestRenderPreRelease_VersionLoadError_ReturnsError validates that version
// file access errors are properly propagated to the caller.
//
// Why: When the VERSION file cannot be read (permissions, directory instead
// of file, etc.), the user needs a clear error rather than silent failure.
//
// What: Given a valid config but VERSION path pointing to a directory,
// RenderPreRelease should return an error.
func TestRenderPreRelease_VersionLoadError_ReturnsError(t *testing.T) {
	// Precondition: Valid config but VERSION is a directory (unreadable as file)
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `prerelease:
  template: "alpha-{{Major}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	versionPath := filepath.Join(tempDir, "VERSION")
	err = os.Mkdir(versionPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create VERSION directory: %v", err)
	}

	// Action: Attempt to render when VERSION cannot be read
	_, err = RenderPreRelease()

	// Expected: Error returned for unreadable VERSION
	if err == nil {
		t.Error("Expected error when VERSION is a directory, got nil")
	}
}

// TestRenderMetadata_ConfigReadError_ReturnsError validates that config
// parsing errors are properly propagated to the caller.
//
// Why: Users need clear feedback when their config file is malformed.
// Silent failures or unclear errors would make debugging difficult.
//
// What: Given a .versionator.yaml with invalid YAML syntax,
// RenderMetadata should return an error.
func TestRenderMetadata_ConfigReadError_ReturnsError(t *testing.T) {
	// Precondition: Working directory with malformed config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile(".versionator.yaml", []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Attempt to render with invalid config
	_, err = RenderMetadata()

	// Expected: Error returned for malformed YAML
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

// TestRenderMetadata_VersionLoadError_ReturnsError validates that version
// file access errors are properly propagated to the caller.
//
// Why: When the VERSION file cannot be read (permissions, directory instead
// of file, etc.), the user needs a clear error rather than silent failure.
//
// What: Given a valid config but VERSION path pointing to a directory,
// RenderMetadata should return an error.
func TestRenderMetadata_VersionLoadError_ReturnsError(t *testing.T) {
	// Precondition: Valid config but VERSION is a directory (unreadable as file)
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `metadata:
  template: "build-{{Major}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	versionPath := filepath.Join(tempDir, "VERSION")
	err = os.Mkdir(versionPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create VERSION directory: %v", err)
	}

	// Action: Attempt to render when VERSION cannot be read
	_, err = RenderMetadata()

	// Expected: Error returned for unreadable VERSION
	if err == nil {
		t.Error("Expected error when VERSION is a directory, got nil")
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions: explicitly empty templates. These validate
// the early-return optimization when no template is configured.
// =============================================================================

// TestRenderPreRelease_EmptyTemplate_ReturnsEmptyString validates that an
// explicitly empty template returns empty string without attempting to
// load version data.
//
// Why: When the template is empty, there is nothing to render. The function
// should short-circuit and return empty without accessing the VERSION file.
// This is an optimization and also handles the "disabled" case gracefully.
//
// What: Given a config with prerelease.template set to "", RenderPreRelease
// should return an empty string without error.
func TestRenderPreRelease_EmptyTemplate_ReturnsEmptyString(t *testing.T) {
	// Precondition: Config with explicitly empty template, VERSION file exists
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `prerelease:
  template: ""
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Render with empty template
	result, err := RenderPreRelease()

	// Expected: Empty result returned (no template to render)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result for empty template, got '%s'", result)
	}
}

// TestRenderMetadata_EmptyTemplate_ReturnsEmptyString validates that an
// explicitly empty template returns empty string without attempting to
// load version data.
//
// Why: When the template is empty, there is nothing to render. The function
// should short-circuit and return empty without accessing the VERSION file.
// This is an optimization and also handles the "disabled" case gracefully.
//
// What: Given a config with metadata.template set to "", RenderMetadata
// should return an empty string without error.
func TestRenderMetadata_EmptyTemplate_ReturnsEmptyString(t *testing.T) {
	// Precondition: Config with explicitly empty template, VERSION file exists
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile("VERSION", []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	configContent := `metadata:
  template: ""
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Action: Render with empty template
	result, err := RenderMetadata()

	// Expected: Empty result returned (no template to render)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result for empty template, got '%s'", result)
	}
}
