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
// CORE FUNCTIONALITY
// Tests for the primary use cases of custom variable commands
// =============================================================================

// TestCustomSetCommand_SetsNewVariable verifies that the custom set command
// creates a new custom variable in the config file.
//
// Why: Custom variables allow users to define project-specific template values
// that can be used in version output templates.
//
// What: Run "config custom set AppName MyApp", verify config contains the variable.
func TestCustomSetCommand_SetsNewVariable(t *testing.T) {
	// Precondition: temp directory with config file
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create minimal config
	initialConfig := &config.Config{}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config custom set AppName MyApp"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "custom", "set", "AppName", "MyApp"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and config contains the custom variable
	assert.NoError(t, err)

	value, ok, err := config.GetCustom("AppName")
	require.NoError(t, err)
	assert.True(t, ok, "Custom variable should exist")
	assert.Equal(t, "MyApp", value)

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestCustomGetCommand_RetrievesExistingVariable verifies that the custom get
// command returns the value of an existing custom variable.
//
// Why: Users need to verify what values are currently set for custom variables.
//
// What: Set a custom variable, then run "config custom get", verify output.
func TestCustomGetCommand_RetrievesExistingVariable(t *testing.T) {
	// Precondition: temp directory with config containing custom variable
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Custom: map[string]string{
			"BuildEnv": "production",
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config custom get BuildEnv"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "custom", "get", "BuildEnv"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and outputs the value
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "production")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestCustomListCommand_ListsAllVariables verifies that the custom list command
// displays all custom variables.
//
// Why: Users need visibility into all configured custom variables.
//
// What: Configure multiple custom variables, run "config custom list", verify all appear.
func TestCustomListCommand_ListsAllVariables(t *testing.T) {
	// Precondition: temp directory with config containing multiple custom variables
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Custom: map[string]string{
			"AppName":  "MyApplication",
			"BuildEnv": "staging",
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config custom list"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "custom", "list"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and output contains both variables
	assert.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "AppName")
	assert.Contains(t, output, "MyApplication")
	assert.Contains(t, output, "BuildEnv")
	assert.Contains(t, output, "staging")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestCustomDeleteCommand_RemovesVariable verifies that the custom delete command
// removes a custom variable from the config.
//
// Why: Users need to remove obsolete custom variables from their configuration.
//
// What: Set a custom variable, run "config custom delete", verify it's removed.
func TestCustomDeleteCommand_RemovesVariable(t *testing.T) {
	// Precondition: temp directory with config containing custom variable
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{
		Custom: map[string]string{
			"ToDelete": "value",
		},
	}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config custom delete ToDelete"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "custom", "delete", "ToDelete"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and variable is removed
	assert.NoError(t, err)

	_, ok, err := config.GetCustom("ToDelete")
	require.NoError(t, err)
	assert.False(t, ok, "Custom variable should be deleted")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// EDGE CASES
// =============================================================================

// TestCustomListCommand_EmptyList_ShowsMessage verifies that listing with no
// custom variables shows an appropriate message.
//
// Why: Users should get clear feedback when no custom variables are defined.
//
// What: Run "config custom list" with no custom variables, verify message.
func TestCustomListCommand_EmptyList_ShowsMessage(t *testing.T) {
	// Precondition: temp directory with config but no custom variables
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config custom list"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "custom", "list"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and shows "no custom values" message
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "No custom values")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// =============================================================================
// ERROR HANDLING
// =============================================================================

// TestCustomGetCommand_NonExistentKey_ReturnsError verifies that getting a
// non-existent custom variable returns an error.
//
// Why: Users need clear feedback when requesting a variable that doesn't exist.
//
// What: Run "config custom get NonExistent", verify error is returned.
func TestCustomGetCommand_NonExistentKey_ReturnsError(t *testing.T) {
	// Precondition: temp directory with empty config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	initialConfig := &config.Config{}
	configData, err := yaml.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(".versionator.yaml", configData, 0644)
	require.NoError(t, err)

	// Action: Execute "config custom get NonExistent"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "custom", "get", "NonExistent"})

	err = rootCmd.Execute()

	// Expected: Command fails with key not found error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NonExistent")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}
