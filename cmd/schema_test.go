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
	rootCmd.SetArgs([]string{"support", "schema"})
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

func TestSchema_IncludesOutputCommand(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
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

	// Find output command and verify it has version subcommand
	for _, cmd := range schema.Commands {
		if cmd.Name == "output" {
			if cmd.Short == "" {
				t.Error("output command should have short description")
			}
			// Check for version subcommand
			hasVersion := false
			for _, sub := range cmd.Subcommands {
				if sub.Name == "version" {
					hasVersion = true
					break
				}
			}
			if !hasVersion {
				t.Error("output command should have version subcommand")
			}
			return
		}
	}

	t.Error("output command not found in schema")
}

func TestSchema_IncludesSubcommands(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
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

	// Find bump command and verify it has subcommands
	for _, cmd := range schema.Commands {
		if cmd.Name == "bump" {
			if len(cmd.Subcommands) == 0 {
				t.Error("bump command should have subcommands (major, minor, patch)")
			}
			// Check for major subcommand
			hasMajor := false
			for _, sub := range cmd.Subcommands {
				if sub.Name == "major" {
					hasMajor = true
					break
				}
			}
			if !hasMajor {
				t.Error("bump command should have major subcommand")
			}
			return
		}
	}

	t.Error("bump command not found in schema")
}

func TestSchema_IncludesGlobalFlags(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
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
	rootCmd.SetArgs([]string{"support", "schema"})
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
	rootCmd.SetArgs([]string{"support", "schema"})
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

	// Find output command and then emit subcommand - it should have flags
	for _, cmd := range schema.Commands {
		if cmd.Name == "output" {
			for _, sub := range cmd.Subcommands {
				if sub.Name == "emit" {
					if len(sub.Flags) == 0 {
						t.Error("emit command should have flags")
					}
					return
				}
			}
			t.Error("emit subcommand not found under output")
			return
		}
	}

	t.Error("output command not found in schema")
}
