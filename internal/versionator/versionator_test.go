package versionator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPreReleaseTemplate_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// No config file exists - should return defaults (empty template)
	template, err := GetPreReleaseTemplate()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Default template should be empty
	if template != "" {
		t.Errorf("Expected empty default template, got '%s'", template)
	}
}

func TestGetPreReleaseTemplate_WithConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create config file with pre-release template
	configContent := `prerelease:
  template: "alpha-{{CommitsSinceTag}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	template, err := GetPreReleaseTemplate()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if template != "alpha-{{CommitsSinceTag}}" {
		t.Errorf("Expected template 'alpha-{{CommitsSinceTag}}', got '%s'", template)
	}
}

func TestGetMetadataTemplate_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// No config file exists - should return defaults (empty template)
	template, err := GetMetadataTemplate()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Default template should be empty
	if template != "" {
		t.Errorf("Expected empty default template, got '%s'", template)
	}
}

func TestGetMetadataTemplate_WithConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create config file with metadata template
	configContent := `metadata:
  template: "{{ShortHash}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	template, err := GetMetadataTemplate()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if template != "{{ShortHash}}" {
		t.Errorf("Expected template '{{ShortHash}}', got '%s'", template)
	}
}

func TestRenderPreRelease_EmptyTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create VERSION file
	err := os.WriteFile("VERSION", []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// No config file - empty template
	result, err := RenderPreRelease()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Empty template should return empty string
	if result != "" {
		t.Errorf("Expected empty result for empty template, got '%s'", result)
	}
}

func TestRenderPreRelease_WithTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create VERSION file
	err := os.WriteFile("VERSION", []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create config with template using version data
	configContent := `prerelease:
  template: "alpha-{{Major}}"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	result, err := RenderPreRelease()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "alpha-1" {
		t.Errorf("Expected 'alpha-1', got '%s'", result)
	}
}

func TestRenderPreRelease_ConfigReadError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create invalid config file (malformed YAML)
	err := os.WriteFile(".versionator.yaml", []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err = RenderPreRelease()
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestRenderPreRelease_VersionLoadError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create config with non-empty template (to exercise version.Load path)
	configContent := `prerelease:
  template: "alpha-{{Major}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create unreadable VERSION file to trigger Load error
	versionPath := filepath.Join(tempDir, "VERSION")
	err = os.WriteFile(versionPath, []byte("1.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Make VERSION file a directory to cause read error
	os.Remove(versionPath)
	err = os.Mkdir(versionPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create VERSION directory: %v", err)
	}

	// This should fail when trying to load version
	_, err = RenderPreRelease()
	if err == nil {
		t.Error("Expected error when VERSION is a directory, got nil")
	}
}

func TestRenderMetadata_EmptyTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create VERSION file
	err := os.WriteFile("VERSION", []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// No config file - empty template
	result, err := RenderMetadata()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Empty template should return empty string
	if result != "" {
		t.Errorf("Expected empty result for empty template, got '%s'", result)
	}
}

func TestRenderMetadata_WithTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create VERSION file
	err := os.WriteFile("VERSION", []byte("2.5.9"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create config with template using version data
	configContent := `metadata:
  template: "build-{{Minor}}"
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	result, err := RenderMetadata()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "build-5" {
		t.Errorf("Expected 'build-5', got '%s'", result)
	}
}

func TestRenderMetadata_ConfigReadError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create invalid config file (malformed YAML)
	err := os.WriteFile(".versionator.yaml", []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err = RenderMetadata()
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestRenderMetadata_VersionLoadError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create config with non-empty template (to exercise version.Load path)
	configContent := `metadata:
  template: "build-{{Major}}"
`
	err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create VERSION file then make it a directory to cause error
	versionPath := filepath.Join(tempDir, "VERSION")
	err = os.Mkdir(versionPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create VERSION directory: %v", err)
	}

	// This should fail when trying to load version
	_, err = RenderMetadata()
	if err == nil {
		t.Error("Expected error when VERSION is a directory, got nil")
	}
}
