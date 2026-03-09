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
// TEST HELPERS
// =============================================================================

func createPrefixVersionString(ver string) string {
	return createPrefixVersionStringWithPrefix(ver, "")
}

func createPrefixVersionStringWithPrefix(ver, prefix string) string {
	return prefix + ver
}

// =============================================================================
// CORE FUNCTIONALITY
// =============================================================================
// Tests for the primary use cases of the prefix command: enabling, disabling,
// setting, and checking prefix status.

// TestPrefixEnableCommand_DefaultConfig_EnablesVPrefix validates that the enable
// command correctly enables the version prefix.
//
// Why: The prefix enable command is the primary way users add a "v" prefix to
// their version strings (e.g., "v1.2.3"). This is a common convention in semantic
// versioning, especially for Git tags.
//
// What: Given a VERSION file without a prefix and default configuration, when
// the user runs "config prefix enable", the prefix should be set to "v" and the
// VERSION file should be updated accordingly.
func TestPrefixEnableCommand_DefaultConfig_EnablesVPrefix(t *testing.T) {
	// Precondition: Create isolated test environment with VERSION file and config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.2.3"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the enable command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "enable"})

	err = rootCmd.Execute()

	// Expected: Prefix should be "v" and output should confirm the change
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "v", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix enabled with value 'v'")
	assert.Contains(t, output, "Current version: v"+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixDisableCommand_PrefixEnabled_RemovesPrefix validates that the disable
// command correctly removes the version prefix.
//
// Why: Users may need to remove the prefix if their downstream tooling expects
// bare version numbers (e.g., "1.2.3" instead of "v1.2.3"). This ensures the
// disable command properly strips the prefix.
//
// What: Given a VERSION file with a "v" prefix, when the user runs
// "config prefix disable", the prefix should be removed and the VERSION file
// should contain only the numeric version.
func TestPrefixDisableCommand_PrefixEnabled_RemovesPrefix(t *testing.T) {
	// Precondition: Create isolated test environment with prefixed VERSION file
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "v",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.2.3"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the disable command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "disable"})

	err = rootCmd.Execute()

	// Expected: Prefix should be empty and output should confirm the change
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix disabled")
	assert.Contains(t, output, "Current version: "+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixSetCommand_ValidPrefix_SetsPrefix validates that the set command
// correctly sets a specific prefix value.
//
// Why: The set command provides explicit control over the prefix value, allowing
// users to specify exactly what prefix they want. This is the most flexible way
// to configure the prefix.
//
// What: Given a VERSION file without a prefix, when the user runs
// "config prefix set v", the prefix should be set to "v" and the VERSION file
// should be updated to include it.
func TestPrefixSetCommand_ValidPrefix_SetsPrefix(t *testing.T) {
	// Precondition: Create isolated test environment with VERSION file
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.2.3"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the set command with "v" prefix
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "set", "v"})

	err = rootCmd.Execute()

	// Expected: Prefix should be "v" and output should confirm the change
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "v", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix set to: v")
	assert.Contains(t, output, "Current version: v"+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixStatusCommand_PrefixEnabled_ShowsEnabledStatus validates that the
// status command correctly displays prefix information when enabled.
//
// Why: Users need to be able to check the current prefix configuration without
// modifying it. The status command provides read-only visibility into the
// current state.
//
// What: Given a VERSION file with a "v" prefix, when the user runs
// "config prefix status", the output should indicate the prefix is enabled
// and display the current prefix value and full version string.
func TestPrefixStatusCommand_PrefixEnabled_ShowsEnabledStatus(t *testing.T) {
	// Precondition: Create isolated test environment with prefixed VERSION file
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "v",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.2.3"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionStringWithPrefix(initialVersion, initialConfig.Prefix)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the status command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "status"})

	err = rootCmd.Execute()

	// Expected: Output should show prefix is enabled with its value
	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Prefix: ENABLED")
	assert.Contains(t, output, "Value: v")
	assert.Contains(t, output, "Current version: v"+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// KEY VARIATIONS
// =============================================================================
// Tests for important alternate flows and configuration variations.

// TestPrefixEnableCommand_ConfigHasPrefix_UsesConfigValue validates that enable
// respects an existing prefix value in the configuration.
//
// Why: When a user has already configured a specific prefix value, the enable
// command should use that value rather than overwriting it with a default.
// This preserves user preferences.
//
// What: Given a configuration with prefix already set to "v", when the user
// runs "config prefix enable", the command should use the configured "v" prefix
// rather than assuming a default.
func TestPrefixEnableCommand_ConfigHasPrefix_UsesConfigValue(t *testing.T) {
	// Precondition: Create isolated test environment with pre-configured prefix
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "v",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "2.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the enable command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "enable"})

	err = rootCmd.Execute()

	// Expected: Should use the configured "v" prefix
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "v", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix enabled with value 'v'")
	assert.Contains(t, output, "Current version: v"+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixSetCommand_UppercaseV_SetsUppercasePrefix validates that the set
// command supports uppercase "V" as a valid prefix.
//
// Why: While lowercase "v" is more common, some projects use uppercase "V"
// for their version prefixes. The command should support both variants.
//
// What: Given a VERSION file without a prefix, when the user runs
// "config prefix set V", the prefix should be set to uppercase "V".
func TestPrefixSetCommand_UppercaseV_SetsUppercasePrefix(t *testing.T) {
	// Precondition: Create isolated test environment
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the set command with "V" prefix
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "set", "V"})

	err = rootCmd.Execute()

	// Expected: Prefix should be uppercase "V"
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "V", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix set to: V")
	assert.Contains(t, output, "Current version: V"+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixStatusCommand_PrefixDisabled_ShowsDisabledStatus validates that the
// status command correctly displays information when prefix is disabled.
//
// Why: Users need to distinguish between an enabled and disabled prefix state.
// The status command should clearly indicate when no prefix is configured.
//
// What: Given a VERSION file without a prefix, when the user runs
// "config prefix status", the output should indicate the prefix is disabled.
func TestPrefixStatusCommand_PrefixDisabled_ShowsDisabledStatus(t *testing.T) {
	// Precondition: Create isolated test environment without prefix
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "2.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionStringWithPrefix(initialVersion, "")+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the status command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "status"})

	err = rootCmd.Execute()

	// Expected: Output should show prefix is disabled
	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Prefix: DISABLED")
	assert.Contains(t, output, "Current version: "+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixStatusCommand_UppercaseVPrefix_ShowsCorrectValue validates that
// status correctly displays uppercase "V" prefix.
//
// Why: The status command must accurately reflect whatever prefix value is
// configured, including case sensitivity. Users should see exactly what they set.
//
// What: Given a VERSION file with uppercase "V" prefix, when the user runs
// "config prefix status", the output should show "V" as the prefix value.
func TestPrefixStatusCommand_UppercaseVPrefix_ShowsCorrectValue(t *testing.T) {
	// Precondition: Create isolated test environment with uppercase V prefix
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "V",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "3.1.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionStringWithPrefix(initialVersion, "V")+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the status command
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "status"})

	err = rootCmd.Execute()

	// Expected: Output should show uppercase "V" prefix
	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Prefix: ENABLED")
	assert.Contains(t, output, "Value: V")
	assert.Contains(t, output, "Current version: V"+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// ERROR HANDLING
// =============================================================================
// Tests for expected failure modes and error conditions.

// TestPrefixSetCommand_MissingArgument_ReturnsError validates that the set
// command requires a prefix argument.
//
// Why: The set command must have a value to set. Without an argument, the
// command cannot determine what prefix the user wants. This ensures proper
// CLI argument validation.
//
// What: When the user runs "config prefix set" without providing a prefix value,
// the command should return an error indicating that an argument is required.
func TestPrefixSetCommand_MissingArgument_ReturnsError(t *testing.T) {
	// Precondition: Create isolated test environment
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the set command without argument
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "set"})

	err = rootCmd.Execute()

	// Expected: Should return an error about missing argument
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixSetCommand_InvalidPrefix_ReturnsError validates that the set command
// rejects invalid prefix values.
//
// Why: Version prefixes are conventionally limited to "v" or "V". Allowing
// arbitrary strings like "release-" would break compatibility with most
// tooling that expects standard semver prefixes. This ensures data integrity.
//
// What: When the user runs "config prefix set release-", the command should
// return an error indicating that only "v" or "V" prefixes are allowed.
func TestPrefixSetCommand_InvalidPrefix_ReturnsError(t *testing.T) {
	// Precondition: Create isolated test environment
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "1.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the set command with invalid prefix
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "set", "release-"})

	err = rootCmd.Execute()

	// Expected: Should return an error about invalid prefix
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only 'v' or 'V' allowed")

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// EDGE CASES
// =============================================================================
// Tests for boundary conditions and unusual but valid scenarios.

// TestPrefixSetCommand_EmptyString_DisablesPrefix validates that setting an
// empty prefix effectively disables the prefix.
//
// Why: Setting the prefix to an empty string is a valid way to disable the
// prefix. This provides an alternative to the "disable" command and ensures
// consistency in the prefix configuration behavior.
//
// What: When the user runs "config prefix set ''", the prefix should be cleared
// and the output should indicate the prefix has been disabled.
func TestPrefixSetCommand_EmptyString_DisablesPrefix(t *testing.T) {
	// Precondition: Create isolated test environment with existing prefix
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "v",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "2.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the set command with empty string
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "set", ""})

	err = rootCmd.Execute()

	// Expected: Prefix should be empty, output should indicate disabled
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix disabled (set to empty)")
	assert.Contains(t, output, "Current version: "+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixDisableCommand_AlreadyDisabled_Idempotent validates that disabling
// an already disabled prefix is a no-op.
//
// Why: Idempotent operations are important for scripting and automation. Users
// should be able to run "disable" without worrying about the current state,
// and the result should always be a disabled prefix.
//
// What: Given a VERSION file without a prefix (already disabled), when the user
// runs "config prefix disable", the command should succeed without error and
// the prefix should remain disabled.
func TestPrefixDisableCommand_AlreadyDisabled_Idempotent(t *testing.T) {
	// Precondition: Create isolated test environment with no prefix
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Prefix: "",
		Metadata: config.MetadataConfig{
			Git: config.GitConfig{HashLength: 7},
		},
		Logging: config.LoggingConfig{Output: "console"},
	}
	initialVersion := "2.0.0"

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString(initialVersion)+"\n"), 0644)
	require.NoError(t, err)

	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute the disable command when already disabled
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "prefix", "disable"})

	err = rootCmd.Execute()

	// Expected: Should succeed and prefix should still be empty
	assert.NoError(t, err)

	vd, err := version.Load()
	require.NoError(t, err)
	assert.Equal(t, "", vd.Prefix)

	output := stdout.String()
	assert.Contains(t, output, "Version prefix disabled")
	assert.Contains(t, output, "Current version: "+initialVersion)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestPrefixEnableCommand_MissingConfigFile_WorksWithVersionFile validates that
// the enable command works even without a config file present.
//
// Why: The prefix commands primarily operate on the VERSION file. Users should
// be able to manage prefixes even in a minimal setup without a full config file.
// This ensures graceful degradation.
//
// What: Given a VERSION file but no .versionator.yaml config file, when the user
// runs "config prefix enable", the command should succeed using default values.
func TestPrefixEnableCommand_MissingConfigFile_WorksWithVersionFile(t *testing.T) {
	// Precondition: Create isolated test environment with only VERSION file
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte(createPrefixVersionString("1.0.0")+"\n"), 0644)
	require.NoError(t, err)

	// Action: Execute the enable command without config file
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "prefix", "enable"})

	err = rootCmd.Execute()

	// Expected: Should work since prefix commands read from VERSION
	assert.NoError(t, err)

	// Cleanup
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// MINUTIAE
// =============================================================================
// Tests for obscure scenarios, help output, and other low-priority validations.

// TestPrefixCommandHelp_AllSubcommands_DisplayUsage validates that all prefix
// subcommands display proper help text.
//
// Why: Users rely on --help output to understand command usage. Each subcommand
// should provide clear usage instructions. This ensures the CLI is self-documenting.
//
// What: When the user runs any prefix command with --help, the output should
// contain a "Usage:" section with relevant information.
func TestPrefixCommandHelp_AllSubcommands_DisplayUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "prefix help",
			args: []string{"config", "prefix", "--help"},
		},
		{
			name: "prefix enable help",
			args: []string{"config", "prefix", "enable", "--help"},
		},
		{
			name: "prefix disable help",
			args: []string{"config", "prefix", "disable", "--help"},
		},
		{
			name: "prefix set help",
			args: []string{"config", "prefix", "set", "--help"},
		},
		{
			name: "prefix status help",
			args: []string{"config", "prefix", "status", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Execute the command with --help flag
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			// Expected: No error and output contains "Usage:"
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "Usage:")

			// Cleanup
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}
