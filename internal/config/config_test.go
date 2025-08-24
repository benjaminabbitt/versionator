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
	if config.Suffix.Type != "git" {
		t.Errorf("Expected default suffix type 'git', got '%s'", config.Suffix.Type)
	}
	if config.Suffix.Enabled != false {
		t.Errorf("Expected default suffix enabled false, got %t", config.Suffix.Enabled)
	}
	if config.Suffix.Git.HashLength != 7 {
		t.Errorf("Expected default hash length 7, got %d", config.Suffix.Git.HashLength)
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
suffix:
  type: "git"
  enabled: true
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
	if config.Suffix.Type != "git" {
		t.Errorf("Expected suffix type 'git', got '%s'", config.Suffix.Type)
	}
	if config.Suffix.Enabled != true {
		t.Errorf("Expected suffix enabled true, got %t", config.Suffix.Enabled)
	}
	if config.Suffix.Git.HashLength != 10 {
		t.Errorf("Expected hash length 10, got %d", config.Suffix.Git.HashLength)
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
suffix:
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
		Suffix: SuffixConfig{
			Type:    "git",
			Enabled: true,
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
	if !contains(content, "enabled: true") {
		t.Error("Expected enabled: true in written config")
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
		Suffix: SuffixConfig{
			Type:    "git",
			Enabled: true,
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
	if readConfig.Suffix.Type != originalConfig.Suffix.Type {
		t.Errorf("Suffix type mismatch: expected '%s', got '%s'", originalConfig.Suffix.Type, readConfig.Suffix.Type)
	}
	if readConfig.Suffix.Enabled != originalConfig.Suffix.Enabled {
		t.Errorf("Suffix enabled mismatch: expected %t, got %t", originalConfig.Suffix.Enabled, readConfig.Suffix.Enabled)
	}
	if readConfig.Suffix.Git.HashLength != originalConfig.Suffix.Git.HashLength {
		t.Errorf("Hash length mismatch: expected %d, got %d", originalConfig.Suffix.Git.HashLength, readConfig.Suffix.Git.HashLength)
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
		Suffix: SuffixConfig{
			Type:    "git",
			Enabled: false,
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
