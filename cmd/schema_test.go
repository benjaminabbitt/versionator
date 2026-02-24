package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSchema_GeneratesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	output := buf.String()

	// Verify it's valid JSON
	var schema CLISchema
	if err := json.Unmarshal([]byte(output), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v\nOutput: %s", err, truncate(output, 500))
	}

	// Verify essential fields
	if schema.Name != "versionator" {
		t.Errorf("expected name 'versionator', got %q", schema.Name)
	}

	if len(schema.Commands) == 0 {
		t.Error("expected at least one command in schema")
	}
}

func TestSchema_IncludesVersionCommand(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	found := false
	for _, cmd := range schema.Commands {
		if cmd.Name == "version" {
			found = true
			if cmd.Short == "" {
				t.Error("version command should have short description")
			}
			break
		}
	}

	if !found {
		t.Error("version command not found in schema")
	}
}

func TestSchema_IncludesSubcommands(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Find major command and verify it has subcommands
	for _, cmd := range schema.Commands {
		if cmd.Name == "major" {
			if len(cmd.Subcommands) == 0 {
				t.Error("major command should have subcommands (increment, decrement)")
			}
			// Check for increment subcommand
			hasIncrement := false
			for _, sub := range cmd.Subcommands {
				if sub.Name == "increment" {
					hasIncrement = true
					break
				}
			}
			if !hasIncrement {
				t.Error("major command should have increment subcommand")
			}
			return
		}
	}

	t.Error("major command not found in schema")
}

func TestSchema_IncludesGlobalFlags(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Should include log-format global flag
	found := false
	for _, flag := range schema.GlobalFlags {
		if flag.Name == "log-format" {
			found = true
			break
		}
	}

	if !found {
		t.Error("log-format global flag not found in schema")
	}
}

func TestSchema_IncludesTemplateVariables(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(schema.TemplateVariables.VersionComponents) == 0 {
		t.Error("expected version component template variables")
	}

	if len(schema.TemplateVariables.VCS) == 0 {
		t.Error("expected VCS template variables")
	}

	// Check for specific well-known variables
	hasMajor := false
	for _, v := range schema.TemplateVariables.VersionComponents {
		if v.Name == "Major" {
			hasMajor = true
			break
		}
	}
	if !hasMajor {
		t.Error("expected Major in version component template variables")
	}
}

func TestSchema_IncludesCommandFlags(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Find emit command - it should have flags
	for _, cmd := range schema.Commands {
		if cmd.Name == "emit" {
			if len(cmd.Flags) == 0 {
				t.Error("emit command should have flags")
			}
			return
		}
	}

	t.Error("emit command not found in schema")
}
