package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletion_Bash_GeneratesOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "bash"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
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

func TestCompletion_Zsh_GeneratesOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "zsh"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#compdef") {
		t.Errorf("expected zsh completion script with #compdef, got: %s", truncate(output, 200))
	}
}

func TestCompletion_Fish_GeneratesOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "fish"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion fish failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "complete -c versionator") {
		t.Errorf("expected fish completion script, got: %s", truncate(output, 200))
	}
}

func TestCompletion_PowerShell_GeneratesOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "powershell"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("completion powershell failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("expected powershell completion script, got: %s", truncate(output, 200))
	}
}

func TestCompletion_InvalidShell_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion", "invalid"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid shell")
	}
}

func TestCompletion_NoArgs_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "completion"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when no shell specified")
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
