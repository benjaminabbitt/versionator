package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/benjaminabbitt/versionator/internal/buildinfo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Schema types for JSON output

// CLISchema is the root schema describing the CLI
type CLISchema struct {
	Schema            string             `json:"$schema"`
	Name              string             `json:"name"`
	Version           string             `json:"version"`
	Description       string             `json:"description"`
	Generated         string             `json:"generated"`
	Commands          []CommandSchema    `json:"commands"`
	GlobalFlags       []FlagSchema       `json:"globalFlags"`
	TemplateVariables TemplateVarsSchema `json:"templateVariables"`
}

// CommandSchema describes a CLI command
type CommandSchema struct {
	Name        string          `json:"name"`
	Path        []string        `json:"path"`
	Short       string          `json:"short"`
	Long        string          `json:"long,omitempty"`
	Usage       string          `json:"usage"`
	Aliases     []string        `json:"aliases,omitempty"`
	Flags       []FlagSchema    `json:"flags,omitempty"`
	Subcommands []CommandSchema `json:"subcommands,omitempty"`
}

// FlagSchema describes a CLI flag
type FlagSchema struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

// TemplateVarsSchema groups template variables by category
type TemplateVarsSchema struct {
	VersionComponents []TemplateVarSchema `json:"versionComponents"`
	PreRelease        []TemplateVarSchema `json:"preRelease"`
	Metadata          []TemplateVarSchema `json:"metadata"`
	VCS               []TemplateVarSchema `json:"vcs"`
	CommitInfo        []TemplateVarSchema `json:"commitInfo"`
	BuildTimestamps   []TemplateVarSchema `json:"buildTimestamps"`
}

// TemplateVarSchema describes a template variable
type TemplateVarSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

var schemaOutput string

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate machine-readable CLI schema",
	Long: `Generate a JSON schema describing all versionator commands, flags, and options.

This schema is designed for:
- AI assistants to understand available commands
- IDE plugins for intelligent completion
- Documentation generators
- CI/CD tooling integration

The schema is generated from the actual command tree, ensuring accuracy.`,
	RunE: runSchema,
}

func runSchema(cmd *cobra.Command, args []string) error {
	schema := buildSchema(cmd.Root())

	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("error generating schema: %w", err)
	}

	if schemaOutput != "" {
		if err := os.WriteFile(schemaOutput, output, FilePermission); err != nil {
			return fmt.Errorf("error writing schema to %s: %w", schemaOutput, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Schema written to %s\n", schemaOutput)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(output))
	return nil
}

func init() {
	schemaCmd.Flags().StringVarP(&schemaOutput, "output", "o", "", "Output file path (default: stdout)")
	supportCmd.AddCommand(schemaCmd)
}

func buildSchema(root *cobra.Command) CLISchema {
	return CLISchema{
		Schema:            "https://json-schema.org/draft/2020-12/schema",
		Name:              root.Name(),
		Version:           buildinfo.Version,
		Description:       root.Short,
		Generated:         time.Now().UTC().Format(time.RFC3339),
		Commands:          buildCommandSchemas(root.Commands(), []string{}),
		GlobalFlags:       buildFlagSchemas(root.PersistentFlags()),
		TemplateVariables: buildTemplateVarsSchema(),
	}
}

func buildCommandSchemas(commands []*cobra.Command, parentPath []string) []CommandSchema {
	var schemas []CommandSchema

	for _, cmd := range commands {
		// Skip internal commands
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "schema" {
			continue
		}

		path := append([]string{}, parentPath...)
		path = append(path, cmd.Name())

		schema := CommandSchema{
			Name:        cmd.Name(),
			Path:        path,
			Short:       cmd.Short,
			Long:        cmd.Long,
			Usage:       cmd.UseLine(),
			Aliases:     cmd.Aliases,
			Flags:       buildFlagSchemas(cmd.LocalFlags()),
			Subcommands: buildCommandSchemas(cmd.Commands(), path),
		}

		schemas = append(schemas, schema)
	}

	return schemas
}

func buildFlagSchemas(flags *pflag.FlagSet) []FlagSchema {
	var schemas []FlagSchema

	flags.VisitAll(func(f *pflag.Flag) {
		// Skip help flag as it's always present
		if f.Name == "help" {
			return
		}

		schema := FlagSchema{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		}
		schemas = append(schemas, schema)
	})

	return schemas
}

func buildTemplateVarsSchema() TemplateVarsSchema {
	return TemplateVarsSchema{
		VersionComponents: []TemplateVarSchema{
			{Name: "Major", Description: "Major version number", Example: "1"},
			{Name: "Minor", Description: "Minor version number", Example: "2"},
			{Name: "Patch", Description: "Patch version number", Example: "3"},
			{Name: "MajorMinorPatch", Description: "Core version: Major.Minor.Patch", Example: "1.2.3"},
			{Name: "MajorMinor", Description: "Major.Minor", Example: "1.2"},
			{Name: "Prefix", Description: "Version prefix", Example: "v"},
		},
		PreRelease: []TemplateVarSchema{
			{Name: "PreRelease", Description: "Rendered pre-release identifier", Example: "alpha-5"},
			{Name: "PreReleaseWithDash", Description: "Pre-release with leading dash", Example: "-alpha-5"},
			{Name: "PreReleaseLabel", Description: "Label part of pre-release", Example: "alpha"},
			{Name: "PreReleaseNumber", Description: "Number part of pre-release", Example: "5"},
		},
		Metadata: []TemplateVarSchema{
			{Name: "Metadata", Description: "Rendered build metadata", Example: "20241211.abc1234"},
			{Name: "MetadataWithPlus", Description: "Metadata with leading plus", Example: "+20241211.abc1234"},
		},
		VCS: []TemplateVarSchema{
			{Name: "Hash", Description: "Full commit hash (40 chars)", Example: "abc1234def5678..."},
			{Name: "ShortHash", Description: "Short commit hash (7 chars)", Example: "abc1234"},
			{Name: "MediumHash", Description: "Medium commit hash (12 chars)", Example: "abc1234def01"},
			{Name: "BranchName", Description: "Current branch name", Example: "feature/foo"},
			{Name: "EscapedBranchName", Description: "Branch with slashes replaced", Example: "feature-foo"},
			{Name: "CommitsSinceTag", Description: "Commits since last tag", Example: "42"},
			{Name: "BuildNumber", Description: "Alias for CommitsSinceTag", Example: "42"},
			{Name: "BuildNumberPadded", Description: "Padded to 4 digits", Example: "0042"},
			{Name: "UncommittedChanges", Description: "Count of uncommitted files", Example: "3"},
			{Name: "Dirty", Description: "'dirty' if uncommitted changes exist", Example: "dirty"},
			{Name: "VersionSourceHash", Description: "Hash of commit that last tag points to", Example: "def5678"},
		},
		CommitInfo: []TemplateVarSchema{
			{Name: "CommitAuthor", Description: "Commit author name", Example: "John Doe"},
			{Name: "CommitAuthorEmail", Description: "Commit author email", Example: "john@example.com"},
			{Name: "CommitDate", Description: "ISO 8601 commit date", Example: "2024-01-15T10:30:00Z"},
			{Name: "CommitDateCompact", Description: "Compact commit date", Example: "20240115103045"},
			{Name: "CommitDateShort", Description: "Date only", Example: "2024-01-15"},
			{Name: "CommitYear", Description: "Commit year", Example: "2024"},
			{Name: "CommitMonth", Description: "Commit month (zero-padded)", Example: "01"},
			{Name: "CommitDay", Description: "Commit day (zero-padded)", Example: "15"},
		},
		BuildTimestamps: []TemplateVarSchema{
			{Name: "BuildDateTimeUTC", Description: "ISO 8601 build time", Example: "2024-01-15T10:30:00Z"},
			{Name: "BuildDateTimeCompact", Description: "Compact build time", Example: "20240115103045"},
			{Name: "BuildDateUTC", Description: "Build date only", Example: "2024-01-15"},
			{Name: "BuildYear", Description: "Build year", Example: "2024"},
			{Name: "BuildMonth", Description: "Build month (zero-padded)", Example: "01"},
			{Name: "BuildDay", Description: "Build day (zero-padded)", Example: "15"},
		},
	}
}
