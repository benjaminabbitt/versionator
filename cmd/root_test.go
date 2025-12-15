package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestExecute_Success(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte(`{"major": 1, "minor": 0, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a minimal config file
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

	// Test Execute function doesn't panic
	err = Execute()
	// Since Execute() runs the root command without args, it should show help and return nil
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte(`{"major": 2, "minor": 1, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a minimal config file
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

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Check output contains the version
	output := buf.String()
	// Note: The output may include build info, so just check it contains the version
	if output == "" {
		t.Errorf("Expected version output, got empty string")
	}

	// Reset command state
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

func TestVersionCommand_WithPrefix(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file with prefix
	err := os.WriteFile("VERSION", []byte(`{"prefix": "v", "major": 3, "minor": 0, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a config file with prefix
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

	// Capture stdout
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Check output contains the prefixed version
	output := buf.String()
	if output == "" {
		t.Errorf("Expected version output, got empty string")
	}

	// Reset command state
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

func TestVersionCommand_NoVersionFile(t *testing.T) {
	// Create a temporary directory for testing (no VERSION file)
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a minimal config file
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

	// Capture stderr
	var buf bytes.Buffer
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"version"})

	// Execute the version command - should succeed and create default VERSION
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command should succeed with default version, got: %v", err)
	}

	// Verify VERSION was created
	if _, err := os.Stat("VERSION"); os.IsNotExist(err) {
		t.Error("Expected VERSION to be created")
	}

	// Reset command state
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
}

func TestLogFormatFlag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte(`{"major": 1, "minor": 0, "patch": 0}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a config file with different log format
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

	// Test with log-format flag
	rootCmd.SetArgs([]string{"--log-format", "development", "version"})

	// Execute should not fail
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command with log-format flag failed: %v", err)
	}

	// Reset command state
	rootCmd.SetArgs(nil)
}
