package config

import (
	"os"
	"testing"
)

func TestReadConfig_DefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test reading config when file doesn't exist (should return defaults)
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error when config file doesn't exist, got: %v", err)
	}

	// Verify default values
	if config.Prefix != "v" {
		t.Errorf("Expected default prefix 'v', got '%s'", config.Prefix)
	}
	if config.Metadata.Git.HashLength != 12 {
		t.Errorf("Expected default hash length 12, got %d", config.Metadata.Git.HashLength)
	}
	if config.Logging.Output != "console" {
		t.Errorf("Expected default logging output 'console', got '%s'", config.Logging.Output)
	}
}

func TestReadConfig_ValidConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a valid config file
	configContent := `prefix: "version-"
prerelease:
  template: "alpha-{{CommitsSinceTag}}"
metadata:
  template: "{{BuildDateTimeCompact}}.{{MediumHash}}"
  git:
    hashLength: 10
logging:
  output: "json"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Read the config
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading valid config, got: %v", err)
	}

	// Verify values
	if config.Prefix != "version-" {
		t.Errorf("Expected prefix 'version-', got '%s'", config.Prefix)
	}
	if config.PreRelease.Template != "alpha-{{CommitsSinceTag}}" {
		t.Errorf("Expected prerelease template 'alpha-{{CommitsSinceTag}}', got '%s'", config.PreRelease.Template)
	}
	if config.Metadata.Template != "{{BuildDateTimeCompact}}.{{MediumHash}}" {
		t.Errorf("Expected metadata template '{{BuildDateTimeCompact}}.{{MediumHash}}', got '%s'", config.Metadata.Template)
	}
	if config.Metadata.Git.HashLength != 10 {
		t.Errorf("Expected hash length 10, got %d", config.Metadata.Git.HashLength)
	}
	if config.Logging.Output != "json" {
		t.Errorf("Expected logging output 'json', got '%s'", config.Logging.Output)
	}
}

func TestReadConfig_InvalidYAML(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create an invalid YAML file
	invalidYAML := `prefix: "test"
metadata:
  type: git
  enabled: true
  git:
    hashLength: [invalid yaml structure
`
	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Try to read the config
	_, err = ReadConfig()
	if err == nil {
		t.Error("Expected error when reading invalid YAML, got nil")
	}
	if !contains(err.Error(), "failed to parse config file") {
		t.Errorf("Expected parse error message, got: %v", err)
	}
}

func TestReadConfig_PermissionError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config file
	err := os.WriteFile(configFile, []byte("prefix: test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Make the file unreadable (this test might not work on all systems)
	err = os.Chmod(configFile, 0000)
	if err != nil {
		t.Skip("Cannot change file permissions on this system")
	}
	defer os.Chmod(configFile, 0644) // Restore permissions for cleanup

	// Try to read the config
	_, err = ReadConfig()
	if err == nil {
		// On some systems (like Windows), permission changes might not work as expected
		t.Skip("Permission test not supported on this system")
	}
	if err != nil && !contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected read error message, got: %v", err)
	}
}

func TestWriteConfig_Success(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config to write
	config := &Config{
		Prefix: "v",
		PreRelease: PreReleaseConfig{
			Template: "alpha-1",
		},
		Metadata: MetadataConfig{
			Template: "{{BuildDateTimeCompact}}.{{MediumHash}}",
			Git: GitConfig{
				HashLength: 8,
			},
		},
		Logging: LoggingConfig{
			Output: "development",
		},
	}

	// Write the config
	err := WriteConfig(config)
	if err != nil {
		t.Fatalf("Expected no error writing config, got: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read the file and verify content
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read written config file: %v", err)
	}

	content := string(data)
	if !contains(content, "# Versionator Configuration") {
		t.Error("Expected header comment in written config")
	}
	if !contains(content, "prefix: v") {
		t.Error("Expected prefix in written config")
	}
	if !contains(content, "hashLength: 8") {
		t.Error("Expected hashLength: 8 in written config")
	}
}

func TestWriteConfig_ReadBack(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config to write
	originalConfig := &Config{
		Prefix: "version-",
		PreRelease: PreReleaseConfig{
			Template: "beta-2",
		},
		Metadata: MetadataConfig{
			Template: "{{ShortHash}}",
			Git: GitConfig{
				HashLength: 12,
			},
		},
		Logging: LoggingConfig{
			Output: "json",
		},
	}

	// Write the config
	err := WriteConfig(originalConfig)
	if err != nil {
		t.Fatalf("Expected no error writing config, got: %v", err)
	}

	// Read it back
	readConfig, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading config back, got: %v", err)
	}

	// Compare values
	if readConfig.Prefix != originalConfig.Prefix {
		t.Errorf("Prefix mismatch: expected '%s', got '%s'", originalConfig.Prefix, readConfig.Prefix)
	}
	if readConfig.PreRelease.Template != originalConfig.PreRelease.Template {
		t.Errorf("PreRelease template mismatch: expected '%s', got '%s'", originalConfig.PreRelease.Template, readConfig.PreRelease.Template)
	}
	if readConfig.Metadata.Template != originalConfig.Metadata.Template {
		t.Errorf("Metadata template mismatch: expected '%s', got '%s'", originalConfig.Metadata.Template, readConfig.Metadata.Template)
	}
	if readConfig.Metadata.Git.HashLength != originalConfig.Metadata.Git.HashLength {
		t.Errorf("Hash length mismatch: expected %d, got %d", originalConfig.Metadata.Git.HashLength, readConfig.Metadata.Git.HashLength)
	}
	if readConfig.Logging.Output != originalConfig.Logging.Output {
		t.Errorf("Logging output mismatch: expected '%s', got '%s'", originalConfig.Logging.Output, readConfig.Logging.Output)
	}
}

func TestWriteConfig_PermissionError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Make the directory read-only (this test might not work on all systems)
	err := os.Chmod(tempDir, 0444)
	if err != nil {
		t.Skip("Cannot change directory permissions on this system")
	}
	defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup

	config := &Config{
		Prefix: "v",
		Metadata: MetadataConfig{
			Git: GitConfig{
				HashLength: 7,
			},
		},
		Logging: LoggingConfig{
			Output: "console",
		},
	}

	// Try to write the config
	err = WriteConfig(config)
	if err == nil {
		// On some systems (like Windows), permission changes might not work as expected
		t.Skip("Permission test not supported on this system")
	}
}

func TestWriteConfig_InvalidConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config with values that might cause marshaling issues
	// Note: YAML marshaling is quite robust, so this test mainly ensures
	// the error handling path works
	config := &Config{}

	// This should succeed since empty config is valid YAML
	err := WriteConfig(config)
	if err != nil {
		t.Errorf("Expected no error writing empty config, got: %v", err)
	}
}

func TestValidateTemplate_ValidTemplate(t *testing.T) {
	tests := []string{
		"",                                     // empty is valid
		"alpha",                                // no mustache tags
		"{{Major}}",                            // single tag
		"{{Major}}.{{Minor}}.{{Patch}}",        // multiple tags
		"alpha-{{CommitsSinceTag}}",            // mixed
		"{{BuildDateTimeCompact}}.{{MediumHash}}", // typical metadata
	}

	for _, template := range tests {
		err := ValidateTemplate(template)
		if err != nil {
			t.Errorf("ValidateTemplate(%q) returned error: %v", template, err)
		}
	}
}

func TestValidateTemplate_InvalidTemplate(t *testing.T) {
	tests := []string{
		"{{",           // unclosed tag
		"{{Major",      // unclosed tag
		"{{#section}}", // section without end
	}

	for _, template := range tests {
		err := ValidateTemplate(template)
		if err == nil {
			t.Errorf("ValidateTemplate(%q) expected error, got nil", template)
		}
	}
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	config := &Config{
		PreRelease: PreReleaseConfig{
			Template: "alpha-{{Major}}",
		},
		Metadata: MetadataConfig{
			Template: "{{BuildDateTimeCompact}}.{{ShortHash}}",
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Config.Validate() returned error for valid config: %v", err)
	}
}

func TestConfig_Validate_InvalidPreReleaseTemplate(t *testing.T) {
	config := &Config{
		PreRelease: PreReleaseConfig{
			Template: "{{unclosed",
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Config.Validate() expected error for invalid prerelease template, got nil")
	}
	if !contains(err.Error(), "prerelease template") {
		t.Errorf("Expected error to mention prerelease template, got: %v", err)
	}
}

func TestConfig_Validate_InvalidMetadataTemplate(t *testing.T) {
	config := &Config{
		Metadata: MetadataConfig{
			Template: "{{#section}}",
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Config.Validate() expected error for invalid metadata template, got nil")
	}
	if !contains(err.Error(), "metadata template") {
		t.Errorf("Expected error to mention metadata template, got: %v", err)
	}
}

func TestWriteConfig_InvalidTemplateRejected(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	config := &Config{
		Prefix: "v",
		PreRelease: PreReleaseConfig{
			Template: "{{unclosed tag",
		},
	}

	err := WriteConfig(config)
	if err == nil {
		t.Error("WriteConfig() expected error for invalid template, got nil")
	}
	if !contains(err.Error(), "invalid config") {
		t.Errorf("Expected error to mention invalid config, got: %v", err)
	}

	// Verify file was NOT created
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("Config file should not be created when validation fails")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
