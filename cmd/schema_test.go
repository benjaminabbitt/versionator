package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests that validate the primary happy-path behavior of the schema command.
// =============================================================================

// TestSchema_GeneratesValidJSON_ReturnsWellFormedSchema validates that the
// schema command produces valid, parseable JSON output.
//
// Why: The schema command is the foundation for tooling integrations (IDE
// plugins, documentation generators, shell completions). Invalid JSON output
// would break all downstream consumers.
//
// What: Executes the "support schema" command and verifies the output is valid
// JSON that deserializes into a CLISchema struct with essential fields populated.
func TestSchema_GeneratesValidJSON_ReturnsWellFormedSchema(t *testing.T) {
	// Precondition: Configure rootCmd to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the schema command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	output := buf.String()

	// Expected: Output is valid JSON with required fields
	var schema CLISchema
	if err := json.Unmarshal([]byte(output), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v\nOutput: %s", err, truncate(output, 500))
	}

	if schema.Name != "versionator" {
		t.Errorf("expected name 'versionator', got %q", schema.Name)
	}

	if len(schema.Commands) == 0 {
		t.Error("expected at least one command in schema")
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests that validate important alternate flows and schema content completeness.
// =============================================================================

// TestSchema_IncludesSubcommands_BumpHasMajorMinorPatch validates that nested
// subcommand hierarchies are properly represented in the schema output.
//
// Why: The CLI uses nested subcommands (e.g., "bump major", "bump minor"). If
// subcommands are missing from the schema, documentation and completion tools
// will present an incomplete view of the CLI's capabilities.
//
// What: Executes the schema command and verifies that the "bump" command
// includes its "major" subcommand (representative of the full subcommand set).
func TestSchema_IncludesSubcommands_BumpHasMajorMinorPatch(t *testing.T) {
	// Precondition: Configure rootCmd to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the schema command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Expected: Bump command exists with major subcommand
	for _, cmd := range schema.Commands {
		if cmd.Name == "bump" {
			if len(cmd.Subcommands) == 0 {
				t.Error("bump command should have subcommands (major, minor, patch)")
			}
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

// TestSchema_IncludesOutputCommand_HasVersionSubcommand validates that the
// output command hierarchy is correctly represented with its subcommands.
//
// Why: The "output" command group contains critical functionality like "version"
// and "emit". Missing subcommands would leave users unable to discover these
// features through schema-based tools.
//
// What: Executes the schema command and verifies that the "output" command
// exists with a short description and includes the "version" subcommand.
func TestSchema_IncludesOutputCommand_HasVersionSubcommand(t *testing.T) {
	// Precondition: Configure rootCmd to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the schema command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Expected: Output command exists with version subcommand
	for _, cmd := range schema.Commands {
		if cmd.Name == "output" {
			if cmd.Short == "" {
				t.Error("output command should have short description")
			}
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

// TestSchema_IncludesGlobalFlags_LogFormatPresent validates that global flags
// are included in the schema output.
//
// Why: Global flags like "--log-format" apply across all commands. If they're
// missing from the schema, tooling won't be able to offer them in completions
// or document them alongside commands.
//
// What: Executes the schema command and verifies that the "log-format" global
// flag is present in the GlobalFlags array.
func TestSchema_IncludesGlobalFlags_LogFormatPresent(t *testing.T) {
	// Precondition: Configure rootCmd to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the schema command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Expected: log-format global flag exists
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

// TestSchema_IncludesTemplateVariables_VersionAndVCSPresent validates that
// template variable documentation is included in the schema.
//
// Why: Template variables are a key feature for customizing version output.
// Without them in the schema, users and tools cannot discover available
// placeholders like {{.Major}} or {{.CommitHash}}.
//
// What: Executes the schema command and verifies that VersionComponents and
// VCS template variable categories are populated, including the specific
// "Major" variable.
func TestSchema_IncludesTemplateVariables_VersionAndVCSPresent(t *testing.T) {
	// Precondition: Configure rootCmd to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the schema command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Expected: Template variable categories are populated
	if len(schema.TemplateVariables.VersionComponents) == 0 {
		t.Error("expected version component template variables")
	}

	if len(schema.TemplateVariables.VCS) == 0 {
		t.Error("expected VCS template variables")
	}

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

// TestSchema_WithOutputFlag_WritesToFile validates that the --output flag writes
// the schema to a file instead of stdout.
func TestSchema_WithOutputFlag_WritesToFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := tempDir + "/schema.json"

	// Reset schemaOutput flag
	schemaOutput = ""

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema", "--output", outputFile})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		schemaOutput = ""
	}()

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	// Verify stdout shows success message
	if !bytes.Contains(buf.Bytes(), []byte("Schema written to")) {
		t.Errorf("expected success message, got %q", buf.String())
	}

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("output file is not valid JSON: %v", err)
	}

	if schema.Name != "versionator" {
		t.Errorf("expected name 'versionator', got %q", schema.Name)
	}
}

// TestSchema_IncludesCommandFlags_EmitHasFlags validates that command-specific
// flags are included in the schema for individual commands.
//
// Why: Many commands have local flags (e.g., "emit --template"). If these flags
// are missing from the schema, tooling cannot provide accurate completions or
// documentation for command usage.
//
// What: Executes the schema command and verifies that the "output emit" command
// has flags defined in its schema representation.
func TestSchema_IncludesCommandFlags_EmitHasFlags(t *testing.T) {
	// Precondition: Configure rootCmd to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"support", "schema"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	// Action: Execute the schema command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var schema CLISchema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Expected: Output command's emit subcommand has flags
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
