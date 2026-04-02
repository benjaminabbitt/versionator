package config

import (
	"os"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose: reading and writing configuration
// files, and round-trip persistence of config values.
// =============================================================================

// TestReadConfig_ValidConfigFile verifies that ReadConfig correctly parses
// a complete configuration file with all fields specified.
//
// Why: This is the primary use case - loading a user's configuration file.
// If this fails, the entire config system is broken.
//
// What: Given a valid YAML config file with custom prefix, templates, hash length,
// and logging output, ReadConfig should return a Config struct with all values
// correctly populated.
func TestReadConfig_ValidConfigFile(t *testing.T) {
	// Precondition: A valid config file exists in the working directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

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

	// Action: Read the config file
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading valid config, got: %v", err)
	}

	// Expected: All values match the file contents
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

// TestWriteConfig_Success verifies that WriteConfig correctly creates a config
// file with the expected content and format.
//
// Why: Writing config is essential for the `config init` command and programmatic
// config updates. The file must be readable by future ReadConfig calls.
//
// What: Given a Config struct with various values, WriteConfig should create
// a properly formatted YAML file with a header comment.
func TestWriteConfig_Success(t *testing.T) {
	// Precondition: Working directory is writable
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

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

	// Action: Write the config
	err := WriteConfig(config)
	if err != nil {
		t.Fatalf("Expected no error writing config, got: %v", err)
	}

	// Expected: File exists with proper content
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

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

// TestWriteConfig_ReadBack verifies that a config can be written and read back
// with all values preserved (round-trip persistence).
//
// Why: This validates the serialization/deserialization cycle. Any field that
// fails round-trip would silently lose user configuration.
//
// What: Write a config with specific values, read it back, and verify all
// values match the original.
func TestWriteConfig_ReadBack(t *testing.T) {
	// Precondition: Empty working directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

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

	// Action: Write then read
	err := WriteConfig(originalConfig)
	if err != nil {
		t.Fatalf("Expected no error writing config, got: %v", err)
	}

	readConfig, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading config back, got: %v", err)
	}

	// Expected: All values match
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

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows: default values, stability settings,
// custom variables, and different configuration options.
// =============================================================================

// TestReadConfig_DefaultConfig verifies that when no config file exists,
// ReadConfig returns sensible defaults.
//
// Why: Users should be able to run versionator without creating a config file.
// Good defaults enable zero-config usage for common cases.
//
// What: With no config file present, ReadConfig should return defaults:
// prefix "v", hash length 12, console logging output.
func TestReadConfig_DefaultConfig(t *testing.T) {
	// Precondition: No config file exists
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Read config when file doesn't exist
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error when config file doesn't exist, got: %v", err)
	}

	// Expected: Default values are applied
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

// TestReadConfig_DefaultStability verifies default stability settings for
// prerelease and metadata (both should be false for CD workflow).
//
// Why: Stability controls whether values are persisted to VERSION file or
// generated dynamically. Wrong defaults would break CD pipelines.
//
// What: Without explicit stability settings, both prerelease.stable and
// metadata.stable should default to false, and templates should be empty.
func TestReadConfig_DefaultStability(t *testing.T) {
	// Precondition: No config file exists
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Read default config
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error when config file doesn't exist, got: %v", err)
	}

	// Expected: Stability defaults to false (CD workflow)
	if config.PreRelease.Stable {
		t.Error("Expected default prerelease stability to be false")
	}
	if config.Metadata.Stable {
		t.Error("Expected default metadata stability to be false")
	}
	if config.PreRelease.Template != "" {
		t.Errorf("Expected default prerelease template to be empty, got '%s'", config.PreRelease.Template)
	}
	if config.Metadata.Template != "" {
		t.Errorf("Expected default metadata template to be empty, got '%s'", config.Metadata.Template)
	}
}

// TestReadConfig_StabilityFromFile verifies that explicit stability settings
// in the config file are correctly parsed.
//
// Why: Users need to be able to opt into traditional release workflow with
// stable=true, overriding the CD-focused defaults.
//
// What: A config file with prerelease.stable=true and metadata.stable=false
// should produce a Config with those exact values.
func TestReadConfig_StabilityFromFile(t *testing.T) {
	// Precondition: Config file with explicit stability settings
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `prerelease:
  template: "alpha"
  stable: true
metadata:
  template: "{{ShortHash}}"
  stable: false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Action: Read config with explicit stability
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading config, got: %v", err)
	}

	// Expected: Stability values match file
	if !config.PreRelease.Stable {
		t.Error("Expected prerelease stability to be true")
	}
	if config.Metadata.Stable {
		t.Error("Expected metadata stability to be false")
	}
}

// TestReadConfig_StabilityTrueForBoth verifies that both prerelease and
// metadata can be set to stable=true.
//
// Why: Some workflows require both prerelease and metadata to be persisted
// in the VERSION file (traditional release workflow).
//
// What: Config with both stable=true should result in both being true.
func TestReadConfig_StabilityTrueForBoth(t *testing.T) {
	// Precondition: Config with both stable=true
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `prerelease:
  template: "rc.1"
  stable: true
metadata:
  template: "build123"
  stable: true
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Action: Read config
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading config, got: %v", err)
	}

	// Expected: Both stable values are true
	if !config.PreRelease.Stable {
		t.Error("Expected prerelease stability to be true")
	}
	if !config.Metadata.Stable {
		t.Error("Expected metadata stability to be true")
	}
}

// TestWriteConfig_ReadBack_WithStability verifies that stability settings
// survive the write/read cycle.
//
// Why: Stability is a critical setting that affects version output behavior.
// Loss of this setting during persistence would cause unexpected behavior.
//
// What: Write config with specific stability settings, read back, verify match.
func TestWriteConfig_ReadBack_WithStability(t *testing.T) {
	// Precondition: Empty directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	originalConfig := &Config{
		Prefix: "v",
		PreRelease: PreReleaseConfig{
			Template: "alpha-{{CommitsSinceTag}}",
			Stable:   true,
		},
		Metadata: MetadataConfig{
			Template: "{{ShortHash}}",
			Stable:   false,
			Git: GitConfig{
				HashLength: 12,
			},
		},
	}

	// Action: Write and read back
	err := WriteConfig(originalConfig)
	if err != nil {
		t.Fatalf("Expected no error writing config, got: %v", err)
	}

	readConfig, err := ReadConfig()
	if err != nil {
		t.Fatalf("Expected no error reading config back, got: %v", err)
	}

	// Expected: Stability values preserved
	if readConfig.PreRelease.Stable != originalConfig.PreRelease.Stable {
		t.Errorf("PreRelease stability mismatch: expected %v, got %v", originalConfig.PreRelease.Stable, readConfig.PreRelease.Stable)
	}
	if readConfig.Metadata.Stable != originalConfig.Metadata.Stable {
		t.Errorf("Metadata stability mismatch: expected %v, got %v", originalConfig.Metadata.Stable, readConfig.Metadata.Stable)
	}
}

// TestReadConfig_ReleaseDefaults verifies default values for release configuration.
//
// Why: Release config controls branch creation behavior during tagging.
// Wrong defaults could cause unexpected branch creation or naming.
//
// What: Without explicit release config, createBranch should be true and
// branchPrefix should be "release/".
func TestReadConfig_ReleaseDefaults(t *testing.T) {
	// Precondition: No config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Read default config
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig() returned unexpected error: %v", err)
	}

	// Expected: Release defaults
	if !config.Release.CreateBranch {
		t.Error("Expected release.createBranch to be true by default")
	}
	if config.Release.BranchPrefix != "release/" {
		t.Errorf("Expected release.branchPrefix 'release/', got '%s'", config.Release.BranchPrefix)
	}
}

// TestReadConfig_BranchVersioningDefaults verifies default values for
// branch versioning configuration.
//
// Why: Branch versioning is opt-in. Wrong defaults could enable it unexpectedly
// or use incorrect main branch patterns.
//
// What: Without explicit config, branch versioning should be disabled,
// use standard main branches, and have "replace" mode.
func TestReadConfig_BranchVersioningDefaults(t *testing.T) {
	// Precondition: No config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Read default config
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig() returned unexpected error: %v", err)
	}

	// Expected: Branch versioning defaults
	if config.BranchVersioning.Enabled {
		t.Error("Expected branch versioning to be disabled by default")
	}
	if len(config.BranchVersioning.MainBranches) != 3 {
		t.Errorf("Expected 3 default main branches, got %d", len(config.BranchVersioning.MainBranches))
	}
	if config.BranchVersioning.Mode != "replace" {
		t.Errorf("Expected default mode 'replace', got '%s'", config.BranchVersioning.Mode)
	}
	if config.BranchVersioning.PrereleaseTemplate != "{{EscapedBranchName}}-{{CommitsSinceTag}}" {
		t.Errorf("Unexpected default prerelease template: %s", config.BranchVersioning.PrereleaseTemplate)
	}
}

// TestDefaultConfigYAML verifies that the default config YAML is well-formed
// and contains expected documentation and default values.
//
// Why: This YAML is used by `config init` to generate a starting config file.
// It must be valid YAML and include helpful documentation for users.
//
// What: The output should contain section headers, all major config fields,
// template variable documentation, and correct default values.
func TestDefaultConfigYAML(t *testing.T) {
	// Action: Get default YAML
	yaml := DefaultConfigYAML()

	// Expected: Non-empty and contains key sections
	if yaml == "" {
		t.Fatal("DefaultConfigYAML() returned empty string")
	}

	expectedSections := []string{
		"# Versionator Configuration",
		"prefix:",
		"prerelease:",
		"metadata:",
		"release:",
		"logging:",
		"AVAILABLE TEMPLATE VARIABLES",
		"{{Major}}",
		"{{ShortHash}}",
		"{{CommitsSinceTag}}",
	}

	for _, section := range expectedSections {
		if !contains(yaml, section) {
			t.Errorf("DefaultConfigYAML() missing expected section: %s", section)
		}
	}

	expectedDefaults := []string{
		`prefix: "v"`,
		`template: ""`,
		`stable: false`,
		`hashLength: 12`,
		`createBranch: true`,
		`branchPrefix: "release/"`,
		`output: "console"`,
	}

	for _, def := range expectedDefaults {
		if !contains(yaml, def) {
			t.Errorf("DefaultConfigYAML() missing expected default: %s", def)
		}
	}
}

// =============================================================================
// CUSTOM VARIABLES
// Tests for custom template variable management (set, get, delete, list).
// =============================================================================

// TestSetCustom_Success verifies that custom variables can be set and persisted
// to the config file.
//
// Why: Custom variables enable users to define their own template variables
// for specialized versioning needs (e.g., product names, team identifiers).
//
// What: Setting a custom variable should persist it to the config file,
// retrievable via subsequent ReadConfig calls.
func TestSetCustom_Success(t *testing.T) {
	// Precondition: Empty directory (no config yet)
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Set a custom variable
	err := SetCustom("MyVar", "my-value")
	if err != nil {
		t.Fatalf("SetCustom() returned unexpected error: %v", err)
	}

	// Expected: Value is persisted
	cfg, err := ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig() returned unexpected error: %v", err)
	}

	if cfg.Custom == nil {
		t.Fatal("Custom map should not be nil after SetCustom")
	}

	if val, ok := cfg.Custom["MyVar"]; !ok {
		t.Error("Expected custom variable 'MyVar' to exist")
	} else if val != "my-value" {
		t.Errorf("Expected custom variable value 'my-value', got '%s'", val)
	}
}

// TestSetCustom_OverwriteExisting verifies that setting a custom variable
// with an existing key overwrites the previous value.
//
// Why: Users need to be able to update custom variables without deleting first.
// Standard map semantics apply: last write wins.
//
// What: Setting a key twice should result in the second value being stored.
func TestSetCustom_OverwriteExisting(t *testing.T) {
	// Precondition: Empty directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Set initial value then overwrite
	err := SetCustom("MyVar", "initial")
	if err != nil {
		t.Fatalf("First SetCustom() returned unexpected error: %v", err)
	}

	err = SetCustom("MyVar", "updated")
	if err != nil {
		t.Fatalf("Second SetCustom() returned unexpected error: %v", err)
	}

	// Expected: Updated value is stored
	cfg, err := ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig() returned unexpected error: %v", err)
	}

	if val := cfg.Custom["MyVar"]; val != "updated" {
		t.Errorf("Expected updated value 'updated', got '%s'", val)
	}
}

// TestGetCustom_ExistingKey verifies that existing custom variables can be retrieved.
//
// Why: GetCustom is the primary way to check if a custom variable exists
// and retrieve its value programmatically.
//
// What: After setting a variable, GetCustom should return the value with exists=true.
func TestGetCustom_ExistingKey(t *testing.T) {
	// Precondition: A custom variable has been set
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := SetCustom("TestKey", "test-value")
	if err != nil {
		t.Fatalf("SetCustom() returned unexpected error: %v", err)
	}

	// Action: Get the value
	value, exists, err := GetCustom("TestKey")

	// Expected: Value returned with exists=true
	if err != nil {
		t.Fatalf("GetCustom() returned unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected custom variable 'TestKey' to exist")
	}
	if value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", value)
	}
}

// TestGetCustom_NonExistingKey verifies that getting a non-existent key
// returns exists=false without error.
//
// Why: Callers need to distinguish between "key not found" and "error occurred".
// Missing keys are normal, not errors.
//
// What: GetCustom for a non-existent key should return empty value, exists=false, no error.
func TestGetCustom_NonExistingKey(t *testing.T) {
	// Precondition: No config file or custom variables
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Get non-existent key
	value, exists, err := GetCustom("NonExistent")

	// Expected: No error, exists=false, empty value
	if err != nil {
		t.Fatalf("GetCustom() returned unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected exists=false for non-existent key")
	}
	if value != "" {
		t.Errorf("Expected empty value for non-existent key, got '%s'", value)
	}
}

// TestGetAllCustom_WithVariables verifies that all custom variables are returned.
//
// Why: Users need to list all custom variables to see what is configured.
//
// What: After setting multiple variables, GetAllCustom should return all of them.
func TestGetAllCustom_WithVariables(t *testing.T) {
	// Precondition: Multiple custom variables set
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = SetCustom("Var1", "value1")
	_ = SetCustom("Var2", "value2")
	_ = SetCustom("Var3", "value3")

	// Action: Get all custom variables
	customs, err := GetAllCustom()

	// Expected: All three variables returned
	if err != nil {
		t.Fatalf("GetAllCustom() returned unexpected error: %v", err)
	}

	if len(customs) != 3 {
		t.Errorf("Expected 3 custom variables, got %d", len(customs))
	}

	expected := map[string]string{
		"Var1": "value1",
		"Var2": "value2",
		"Var3": "value3",
	}

	for k, v := range expected {
		if customs[k] != v {
			t.Errorf("Expected customs[%q] = %q, got %q", k, v, customs[k])
		}
	}
}

// TestGetAllCustom_Empty verifies that an empty map is returned when no custom vars exist.
//
// Why: Callers should get a usable empty map, not nil, to avoid nil pointer errors.
//
// What: With no custom variables, GetAllCustom should return non-nil empty map.
func TestGetAllCustom_Empty(t *testing.T) {
	// Precondition: No config file
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Get all custom variables
	customs, err := GetAllCustom()

	// Expected: Non-nil empty map
	if err != nil {
		t.Fatalf("GetAllCustom() returned unexpected error: %v", err)
	}

	if customs == nil {
		t.Error("Expected non-nil map, got nil")
	}
	if len(customs) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(customs))
	}
}

// TestDeleteCustom_ExistingKey verifies that custom variables can be deleted.
//
// Why: Users need to remove custom variables they no longer need.
//
// What: After deleting a variable, it should no longer exist in the config.
func TestDeleteCustom_ExistingKey(t *testing.T) {
	// Precondition: A custom variable exists
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := SetCustom("ToDelete", "value")
	if err != nil {
		t.Fatalf("SetCustom() returned unexpected error: %v", err)
	}

	// Action: Delete the variable
	err = DeleteCustom("ToDelete")
	if err != nil {
		t.Fatalf("DeleteCustom() returned unexpected error: %v", err)
	}

	// Expected: Variable no longer exists
	_, exists, err := GetCustom("ToDelete")
	if err != nil {
		t.Fatalf("GetCustom() returned unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected custom variable to be deleted")
	}
}

// TestDeleteCustom_NonExistingKey verifies that deleting a non-existent key
// does not cause an error (idempotent operation).
//
// Why: Idempotent delete simplifies code - callers don't need to check existence first.
//
// What: Deleting a key that doesn't exist should succeed silently.
func TestDeleteCustom_NonExistingKey(t *testing.T) {
	// Precondition: No custom variables
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Delete non-existent key
	err := DeleteCustom("NonExistent")

	// Expected: No error
	if err != nil {
		t.Errorf("DeleteCustom() for non-existent key should not error, got: %v", err)
	}
}

// TestDeleteCustom_PreservesOtherKeys verifies that deleting one key
// does not affect other custom variables.
//
// Why: Deletion must be surgical - only the specified key should be removed.
//
// What: After deleting one key, other keys should still exist with correct values.
func TestDeleteCustom_PreservesOtherKeys(t *testing.T) {
	// Precondition: Multiple custom variables
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	_ = SetCustom("Keep1", "value1")
	_ = SetCustom("ToDelete", "delete-me")
	_ = SetCustom("Keep2", "value2")

	// Action: Delete one variable
	err := DeleteCustom("ToDelete")
	if err != nil {
		t.Fatalf("DeleteCustom() returned unexpected error: %v", err)
	}

	// Expected: Other variables preserved
	customs, _ := GetAllCustom()
	if len(customs) != 2 {
		t.Errorf("Expected 2 remaining custom variables, got %d", len(customs))
	}
	if customs["Keep1"] != "value1" {
		t.Errorf("Expected Keep1=value1, got %s", customs["Keep1"])
	}
	if customs["Keep2"] != "value2" {
		t.Errorf("Expected Keep2=value2, got %s", customs["Keep2"])
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes: invalid YAML, permission errors,
// validation failures.
// =============================================================================

// TestReadConfig_InvalidYAML verifies that malformed YAML produces a clear error.
//
// Why: Users may hand-edit config files and introduce syntax errors.
// Clear error messages help them fix the problem.
//
// What: A config file with invalid YAML syntax should cause ReadConfig to
// return an error mentioning "parse".
func TestReadConfig_InvalidYAML(t *testing.T) {
	// Precondition: Config file with invalid YAML
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

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

	// Action: Try to read invalid config
	_, err = ReadConfig()

	// Expected: Error with parse message
	if err == nil {
		t.Error("Expected error when reading invalid YAML, got nil")
	}
	if !contains(err.Error(), "failed to parse config file") {
		t.Errorf("Expected parse error message, got: %v", err)
	}
}

// TestReadConfig_PermissionError verifies that unreadable files produce clear errors.
//
// Why: Permission issues can occur in production. Users need clear error messages
// to diagnose access problems.
//
// What: A config file with no read permissions should cause ReadConfig to
// return a "failed to read" error.
func TestReadConfig_PermissionError(t *testing.T) {
	// Precondition: Config file exists but is not readable
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.WriteFile(configFile, []byte("prefix: test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	err = os.Chmod(configFile, 0000)
	if err != nil {
		t.Skip("Cannot change file permissions on this system")
	}
	defer func() { _ = os.Chmod(configFile, 0644) }()

	// Action: Try to read unreadable config
	_, err = ReadConfig()

	// Expected: Error about reading
	if err == nil {
		t.Skip("Permission test not supported on this system")
	}
	if err != nil && !contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected read error message, got: %v", err)
	}
}

// TestWriteConfig_PermissionError verifies that unwritable directories produce errors.
//
// Why: Permission issues can prevent config writes. Clear errors help diagnose.
//
// What: Attempting to write config in a read-only directory should fail.
func TestWriteConfig_PermissionError(t *testing.T) {
	// Precondition: Read-only directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	err := os.Chmod(tempDir, 0444)
	if err != nil {
		t.Skip("Cannot change directory permissions on this system")
	}
	defer func() { _ = os.Chmod(tempDir, 0755) }()

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

	// Action: Try to write in read-only directory
	err = WriteConfig(config)

	// Expected: Error (or skip if permissions not supported)
	if err == nil {
		t.Skip("Permission test not supported on this system")
	}
}

// TestWriteConfig_InvalidTemplateRejected verifies that WriteConfig validates
// templates before writing and rejects invalid ones.
//
// Why: Writing an invalid template to config would cause runtime failures.
// Validation at write time prevents corrupted configs.
//
// What: A config with an invalid Mustache template should fail validation
// and NOT create a config file.
func TestWriteConfig_InvalidTemplateRejected(t *testing.T) {
	// Precondition: Empty directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	config := &Config{
		Prefix: "v",
		PreRelease: PreReleaseConfig{
			Template: "{{unclosed tag",
		},
	}

	// Action: Try to write invalid config
	err := WriteConfig(config)

	// Expected: Error about invalid config, file not created
	if err == nil {
		t.Error("WriteConfig() expected error for invalid template, got nil")
	}
	if !contains(err.Error(), "invalid config") {
		t.Errorf("Expected error to mention invalid config, got: %v", err)
	}

	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("Config file should not be created when validation fails")
	}
}

// TestConfig_Validate_InvalidPreReleaseTemplate verifies that validation
// catches invalid prerelease templates.
//
// Why: Templates are executed at runtime. Invalid syntax would cause failures.
// Early validation provides better user experience.
//
// What: A Config with unclosed Mustache tag in prerelease template should
// fail validation with clear error message.
func TestConfig_Validate_InvalidPreReleaseTemplate(t *testing.T) {
	// Precondition: Config with invalid prerelease template
	config := &Config{
		PreRelease: PreReleaseConfig{
			Template: "{{unclosed",
		},
	}

	// Action: Validate config
	err := config.Validate()

	// Expected: Error mentioning prerelease template
	if err == nil {
		t.Error("Config.Validate() expected error for invalid prerelease template, got nil")
	}
	if !contains(err.Error(), "prerelease template") {
		t.Errorf("Expected error to mention prerelease template, got: %v", err)
	}
}

// TestConfig_Validate_InvalidMetadataTemplate verifies that validation
// catches invalid metadata templates.
//
// Why: Same as prerelease - invalid metadata templates would fail at runtime.
//
// What: A Config with invalid Mustache section in metadata template should
// fail validation with clear error message.
func TestConfig_Validate_InvalidMetadataTemplate(t *testing.T) {
	// Precondition: Config with invalid metadata template
	config := &Config{
		Metadata: MetadataConfig{
			Template: "{{#section}}",
		},
	}

	// Action: Validate config
	err := config.Validate()

	// Expected: Error mentioning metadata template
	if err == nil {
		t.Error("Config.Validate() expected error for invalid metadata template, got nil")
	}
	if !contains(err.Error(), "metadata template") {
		t.Errorf("Expected error to mention metadata template, got: %v", err)
	}
}

// TestConfig_Validate_BranchVersioningMode verifies validation of the
// branch versioning mode field (must be empty, "replace", or "append").
//
// Why: Invalid mode values would cause undefined behavior during versioning.
// Validation ensures only valid modes are accepted.
//
// What: Empty, "replace", and "append" should pass; other values should fail.
func TestConfig_Validate_BranchVersioningMode(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		expectErr bool
	}{
		{name: "empty mode is valid", mode: "", expectErr: false},
		{name: "replace mode is valid", mode: "replace", expectErr: false},
		{name: "append mode is valid", mode: "append", expectErr: false},
		{name: "invalid mode rejected", mode: "invalid", expectErr: true},
		{name: "prepend mode rejected", mode: "prepend", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Config with specific mode
			config := &Config{
				BranchVersioning: BranchVersioningConfig{
					Mode: tt.mode,
				},
			}

			// Action: Validate
			err := config.Validate()

			// Expected: Error or no error based on mode validity
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for mode %q, got nil", tt.mode)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for mode %q: %v", tt.mode, err)
			}
			if tt.expectErr && err != nil && !contains(err.Error(), "branch versioning mode") {
				t.Errorf("Expected error about branch versioning mode, got: %v", err)
			}
		})
	}
}

// TestConfig_Validate_BranchVersioningTemplate verifies validation of the
// branch versioning prerelease template.
//
// Why: Invalid templates in branch versioning would cause failures during
// version generation for feature branches.
//
// What: Valid templates and empty string should pass; invalid syntax should fail.
func TestConfig_Validate_BranchVersioningTemplate(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		expectErr bool
	}{
		{name: "empty template is valid", template: "", expectErr: false},
		{name: "valid template", template: "{{BranchName}}-{{CommitsSinceTag}}", expectErr: false},
		{name: "invalid template rejected", template: "{{unclosed", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Config with specific template
			config := &Config{
				BranchVersioning: BranchVersioningConfig{
					PrereleaseTemplate: tt.template,
				},
			}

			// Action: Validate
			err := config.Validate()

			// Expected: Error or no error based on template validity
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for template %q, got nil", tt.template)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for template %q: %v", tt.template, err)
			}
		})
	}
}

// TestSetCustom_EmptyKey verifies that empty keys are rejected.
//
// Why: Empty keys would create invalid template variables that cannot be
// referenced in Mustache templates.
//
// What: SetCustom with empty string key should return an error about empty key.
func TestSetCustom_EmptyKey(t *testing.T) {
	// Precondition: Empty directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Try to set empty key
	err := SetCustom("", "some-value")

	// Expected: Error about empty key
	if err == nil {
		t.Error("SetCustom() with empty key should return error")
	}
	if !contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected error about empty key, got: %v", err)
	}
}

// TestSetCustom_InvalidKey verifies that keys with invalid characters are rejected.
//
// Why: Keys must be valid Mustache variable names. Invalid characters would
// cause template parsing failures.
//
// What: Keys starting with numbers, underscores, or containing special
// characters should be rejected.
func TestSetCustom_InvalidKey(t *testing.T) {
	// Precondition: Empty directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	invalidKeys := []string{
		"123start",    // starts with number
		"_underscore", // starts with underscore
		"has-dash",    // contains dash
		"has.dot",     // contains dot
		"has space",   // contains space
	}

	for _, key := range invalidKeys {
		// Action: Try to set invalid key
		err := SetCustom(key, "value")

		// Expected: Error about invalid key
		if err == nil {
			t.Errorf("SetCustom(%q) should return error for invalid key", key)
		}
		if !contains(err.Error(), "invalid custom variable key") {
			t.Errorf("Expected error about invalid key for %q, got: %v", key, err)
		}
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions: empty config, partial config, validation
// of special inputs.
// =============================================================================

// TestWriteConfig_EmptyConfig verifies that an empty config struct can be
// written successfully.
//
// Why: Edge case - users might create minimal configs. Empty struct should
// serialize without error.
//
// What: Writing an empty Config{} should succeed.
func TestWriteConfig_EmptyConfig(t *testing.T) {
	// Precondition: Empty directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Action: Write empty config
	config := &Config{}
	err := WriteConfig(config)

	// Expected: Success (empty config is valid YAML)
	if err != nil {
		t.Errorf("Expected no error writing empty config, got: %v", err)
	}
}

// TestGetCustom_EmptyCustomMap verifies behavior when config exists but has
// no custom section.
//
// Why: Partial configs are common - users may not use custom variables.
// GetCustom should handle nil custom map gracefully.
//
// What: Config with no custom section should return exists=false, no error.
func TestGetCustom_EmptyCustomMap(t *testing.T) {
	// Precondition: Config exists but has no custom section
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	configContent := `prefix: "v"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Action: Get custom variable
	value, exists, err := GetCustom("SomeKey")

	// Expected: No error, exists=false
	if err != nil {
		t.Fatalf("GetCustom() returned unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected exists=false when custom map is nil")
	}
	if value != "" {
		t.Errorf("Expected empty value, got '%s'", value)
	}
}

// TestConfig_Validate_ValidConfig verifies that a well-formed config passes validation.
//
// Why: Positive test - ensure valid configs are not rejected incorrectly.
//
// What: A config with valid templates should pass validation without error.
func TestConfig_Validate_ValidConfig(t *testing.T) {
	// Precondition: Config with valid templates
	config := &Config{
		PreRelease: PreReleaseConfig{
			Template: "alpha-{{Major}}",
		},
		Metadata: MetadataConfig{
			Template: "{{BuildDateTimeCompact}}.{{ShortHash}}",
		},
	}

	// Action: Validate
	err := config.Validate()

	// Expected: No error
	if err != nil {
		t.Errorf("Config.Validate() returned error for valid config: %v", err)
	}
}

// TestValidateTemplate_ValidTemplate verifies that valid Mustache templates
// pass validation.
//
// Why: ValidateTemplate is the foundation of config validation. Must accept
// all valid template syntaxes.
//
// What: Empty templates, plain strings, single tags, and multiple tags
// should all be valid.
func TestValidateTemplate_ValidTemplate(t *testing.T) {
	tests := []string{
		"",                                        // empty is valid
		"alpha",                                   // no mustache tags
		"{{Major}}",                               // single tag
		"{{Major}}.{{Minor}}.{{Patch}}",           // multiple tags
		"alpha-{{CommitsSinceTag}}",               // mixed
		"{{BuildDateTimeCompact}}.{{MediumHash}}", // typical metadata
	}

	for _, template := range tests {
		// Action: Validate template
		err := ValidateTemplate(template)

		// Expected: No error
		if err != nil {
			t.Errorf("ValidateTemplate(%q) returned error: %v", template, err)
		}
	}
}

// TestValidateTemplate_InvalidTemplate verifies that invalid Mustache templates
// are rejected.
//
// Why: Invalid templates would cause runtime errors. Early validation is essential.
//
// What: Unclosed tags and unclosed sections should be rejected.
func TestValidateTemplate_InvalidTemplate(t *testing.T) {
	tests := []string{
		"{{",           // unclosed tag
		"{{Major",      // unclosed tag
		"{{#section}}", // section without end
	}

	for _, template := range tests {
		// Action: Validate template
		err := ValidateTemplate(template)

		// Expected: Error
		if err == nil {
			t.Errorf("ValidateTemplate(%q) expected error, got nil", template)
		}
	}
}

// TestConfig_Validate_UpdatesRequiredFields verifies that updates entries
// must have file, path, and template fields.
//
// Why: Updates without these fields would fail during file update operations.
//
// What: Missing file, path, or template should each produce validation errors.
func TestConfig_Validate_UpdatesRequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		update    UpdateConfig
		expectErr string
	}{
		{
			name:      "missing file",
			update:    UpdateConfig{Path: "version", Template: "{{Major}}"},
			expectErr: "file is required",
		},
		{
			name:      "missing path",
			update:    UpdateConfig{File: "package.json", Template: "{{Major}}"},
			expectErr: "path is required",
		},
		{
			name:      "missing template",
			update:    UpdateConfig{File: "package.json", Path: "version"},
			expectErr: "template is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Updates: []UpdateConfig{tt.update},
			}

			err := config.Validate()

			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if err != nil && !contains(err.Error(), tt.expectErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectErr, err)
			}
		})
	}
}

// TestConfig_Validate_UpdatesInvalidTemplate verifies that updates with
// invalid templates are rejected.
//
// Why: Invalid templates in updates would fail when processing file updates.
//
// What: An update with unclosed mustache tag should fail validation.
func TestConfig_Validate_UpdatesInvalidTemplate(t *testing.T) {
	config := &Config{
		Updates: []UpdateConfig{
			{
				File:     "package.json",
				Path:     "version",
				Template: "{{unclosed",
			},
		},
	}

	err := config.Validate()

	if err == nil {
		t.Error("Expected error for invalid update template, got nil")
	}
	if err != nil && !contains(err.Error(), "updates[0] template") {
		t.Errorf("Expected error about updates[0] template, got: %v", err)
	}
}

// TestConfig_Validate_UpdatesInvalidFormat verifies that updates with
// invalid format values are rejected.
//
// Why: Only json, yaml, and toml formats are supported. Invalid formats
// would cause file parsing failures.
//
// What: Invalid format value should fail; valid formats should pass.
func TestConfig_Validate_UpdatesInvalidFormat(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		expectErr bool
	}{
		{name: "empty format is valid", format: "", expectErr: false},
		{name: "json format is valid", format: "json", expectErr: false},
		{name: "yaml format is valid", format: "yaml", expectErr: false},
		{name: "toml format is valid", format: "toml", expectErr: false},
		{name: "xml format is invalid", format: "xml", expectErr: true},
		{name: "ini format is invalid", format: "ini", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Updates: []UpdateConfig{
					{
						File:     "config.json",
						Path:     "version",
						Template: "{{Major}}.{{Minor}}.{{Patch}}",
						Format:   tt.format,
					},
				},
			}

			err := config.Validate()

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for format %q, got nil", tt.format)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for format %q: %v", tt.format, err)
			}
			if tt.expectErr && err != nil && !contains(err.Error(), "format must be") {
				t.Errorf("Expected error about format, got: %v", err)
			}
		})
	}
}

// TestConfig_Validate_ValidUpdates verifies that well-formed updates pass validation.
//
// Why: Positive test - ensure valid updates are not incorrectly rejected.
//
// What: Updates with all required fields and valid values should pass.
func TestConfig_Validate_ValidUpdates(t *testing.T) {
	config := &Config{
		Updates: []UpdateConfig{
			{
				File:     "package.json",
				Path:     "version",
				Template: "{{Major}}.{{Minor}}.{{Patch}}",
				Format:   "json",
			},
			{
				File:     "pyproject.toml",
				Path:     "tool.poetry.version",
				Template: "{{Major}}.{{Minor}}.{{Patch}}",
				Format:   "toml",
			},
		},
	}

	err := config.Validate()

	if err != nil {
		t.Errorf("Expected no error for valid updates, got: %v", err)
	}
}

// =============================================================================
// MINUTIAE
// Tests for obscure scenarios and internal helper functions.
// =============================================================================

// TestIsValidTemplateKey verifies the validation rules for custom variable keys.
//
// Why: Keys must start with a letter, contain only alphanumeric characters
// and underscores, and cannot be empty. This prevents injection of invalid
// Mustache template variables.
//
// What: Valid keys (letter start, alphanumeric, underscores) should pass;
// invalid keys (number start, dashes, dots, spaces, empty) should fail.
func TestIsValidTemplateKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		// Valid keys - alphanumeric starting with letter
		{name: "simple lowercase", key: "foo", expected: true},
		{name: "simple uppercase", key: "FOO", expected: true},
		{name: "mixed case", key: "FooBar", expected: true},
		{name: "with numbers", key: "foo123", expected: true},
		{name: "with underscore", key: "foo_bar", expected: true},
		{name: "single letter", key: "x", expected: true},
		{name: "uppercase with underscore", key: "MY_VAR", expected: true},
		{name: "complex valid", key: "MyVar_123_test", expected: true},

		// Invalid keys - must start with letter
		{name: "starts with number", key: "123foo", expected: false},
		{name: "starts with underscore", key: "_foo", expected: false},
		{name: "starts with dash", key: "-foo", expected: false},

		// Invalid keys - invalid characters
		{name: "contains dash", key: "foo-bar", expected: false},
		{name: "contains dot", key: "foo.bar", expected: false},
		{name: "contains space", key: "foo bar", expected: false},
		{name: "contains special char", key: "foo@bar", expected: false},

		// Edge cases
		{name: "empty string", key: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Validate key
			result := isValidTemplateKey(tt.key)

			// Expected: Match expected result
			if result != tt.expected {
				t.Errorf("isValidTemplateKey(%q) = %v, expected %v", tt.key, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// contains checks if a string contains a substring.
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
