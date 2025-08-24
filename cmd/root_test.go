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
	err := os.WriteFile("VERSION", []byte("1.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a minimal config file
	configContent := `prefix: ""
suffix:
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
	err := os.WriteFile("VERSION", []byte("2.1.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a minimal config file
	configContent := `prefix: ""
suffix:
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
	if output != "2.1.0\n" {
		t.Errorf("Expected '2.1.0\\n', got '%s'", output)
	}
}

func TestVersionCommand_WithPrefix(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte("3.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a config file with prefix
	configContent := `prefix: "v"
suffix:
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
	if output != "v3.0.0\n" {
		t.Errorf("Expected 'v3.0.0\\n', got '%s'", output)
	}
}

func TestVersionCommand_NoVersionFile(t *testing.T) {
	// Create a temporary directory for testing (no VERSION file)
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a minimal config file
	configContent := `prefix: ""
suffix:
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

	// Execute the version command - should fail
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected version command to fail when no VERSION file exists")
	}
}

func TestLogFormatFlag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file
	err := os.WriteFile("VERSION", []byte("1.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Create a config file with different log format
	configContent := `prefix: ""
suffix:
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
}
