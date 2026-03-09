package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests verifying the primary happy-path behavior of shell completion generation.
// =============================================================================

// TestCompletion_Bash_GeneratesOutput validates that the completion command
// generates a valid Bash completion script.
//
// Why: Users sourcing the completion script in Bash need a correctly formatted
// script that includes the application name and the Bash-specific completion
// functions. Without this, tab completion would not work in Bash shells.
//
// What: Given the completion command is invoked with "bash" argument,
// the output should contain "versionator" (the command name) and
// "__start_versionator" (the Bash completion function entry point).
func TestCompletion_Bash_GeneratesOutput(t *testing.T) {
	// Precondition: Set up command output capture
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "bash"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the completion command for Bash
	err := rootCmd.Execute()

	// Expected: Command succeeds and output contains Bash completion markers
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "versionator") {
		t.Errorf("expected bash completion script to contain 'versionator', got: %s", truncate(output, 200))
	}
	if !strings.Contains(output, "__start_versionator") {
		t.Errorf("expected bash completion script to contain completion function, got: %s", truncate(output, 200))
	}
}

// TestCompletion_Zsh_GeneratesOutput validates that the completion command
// generates a valid Zsh completion script.
//
// Why: Zsh users need a completion script that follows Zsh conventions,
// starting with the #compdef directive. This directive tells Zsh which
// command the completion script is for.
//
// What: Given the completion command is invoked with "zsh" argument,
// the output should contain "#compdef" (the Zsh completion directive).
func TestCompletion_Zsh_GeneratesOutput(t *testing.T) {
	// Precondition: Set up command output capture
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "zsh"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the completion command for Zsh
	err := rootCmd.Execute()

	// Expected: Command succeeds and output contains Zsh completion directive
	if err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#compdef") {
		t.Errorf("expected zsh completion script with #compdef, got: %s", truncate(output, 200))
	}
}

// TestCompletion_Fish_GeneratesOutput validates that the completion command
// generates a valid Fish shell completion script.
//
// Why: Fish shell users need completions in Fish's native format, which uses
// the "complete -c <command>" syntax to define completions. Without the
// correct format, Fish would not provide tab completion for versionator.
//
// What: Given the completion command is invoked with "fish" argument,
// the output should contain "complete -c versionator" (Fish completion syntax).
func TestCompletion_Fish_GeneratesOutput(t *testing.T) {
	// Precondition: Set up command output capture
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "fish"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the completion command for Fish
	err := rootCmd.Execute()

	// Expected: Command succeeds and output contains Fish completion syntax
	if err != nil {
		t.Fatalf("completion fish failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "complete -c versionator") {
		t.Errorf("expected fish completion script, got: %s", truncate(output, 200))
	}
}

// TestCompletion_PowerShell_GeneratesOutput validates that the completion command
// generates a valid PowerShell completion script.
//
// Why: Windows users running PowerShell need completions that use PowerShell's
// Register-ArgumentCompleter cmdlet. This is the standard way to add tab
// completion for commands in PowerShell.
//
// What: Given the completion command is invoked with "powershell" argument,
// the output should contain "Register-ArgumentCompleter" (PowerShell cmdlet).
func TestCompletion_PowerShell_GeneratesOutput(t *testing.T) {
	// Precondition: Set up command output capture
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "powershell"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the completion command for PowerShell
	err := rootCmd.Execute()

	// Expected: Command succeeds and output contains PowerShell completion cmdlet
	if err != nil {
		t.Fatalf("completion powershell failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("expected powershell completion script, got: %s", truncate(output, 200))
	}
}

// =============================================================================
// ERROR HANDLING
// Tests verifying proper error responses for invalid inputs.
// =============================================================================

// TestCompletion_InvalidShell_ReturnsError validates that the completion command
// rejects unsupported shell names with an appropriate error.
//
// Why: Users might mistype a shell name or try to use an unsupported shell.
// The command should fail gracefully with a clear error rather than producing
// invalid output or crashing.
//
// What: Given the completion command is invoked with an invalid shell name
// "invalid", the command should return an error.
func TestCompletion_InvalidShell_ReturnsError(t *testing.T) {
	// Precondition: Set up command output capture
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "invalid"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the completion command with invalid shell name
	err := rootCmd.Execute()

	// Expected: Command returns an error
	if err == nil {
		t.Error("expected error for invalid shell")
	}
}

// TestCompletion_NoArgs_ReturnsError validates that the completion command
// requires a shell argument and fails when none is provided.
//
// Why: The completion command cannot generate output without knowing which
// shell format to use. Users need clear feedback that a shell type is required.
//
// What: Given the completion command is invoked without specifying a shell,
// the command should return an error.
func TestCompletion_NoArgs_ReturnsError(t *testing.T) {
	// Precondition: Set up command output capture
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the completion command without shell argument
	err := rootCmd.Execute()

	// Expected: Command returns an error
	if err == nil {
		t.Error("expected error when no shell specified")
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// truncate shortens a string to the specified maximum length, appending "..."
// if truncation occurs. Used for readable error messages in test output.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
