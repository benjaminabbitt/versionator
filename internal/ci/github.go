package ci

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// GitHubFormatter outputs variables for GitHub Actions
// Writes to $GITHUB_OUTPUT and $GITHUB_ENV files
type GitHubFormatter struct{}

// Name returns the formatter name
func (f *GitHubFormatter) Name() string {
	return "github"
}

// Format returns the formatted output (for display purposes)
func (f *GitHubFormatter) Format(vars map[string]string) string {
	var sb strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString("# GitHub Actions Output Variables\n")
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, vars[k]))
	}

	return sb.String()
}

// Write writes variables to GITHUB_OUTPUT and GITHUB_ENV files
func (f *GitHubFormatter) Write(vars map[string]string, w io.Writer) error {
	// Get the output file paths from environment
	outputFile := os.Getenv("GITHUB_OUTPUT")
	envFile := os.Getenv("GITHUB_ENV")

	if outputFile == "" && envFile == "" {
		// Not in GitHub Actions, just write to the provided writer
		_, err := w.Write([]byte(f.Format(vars)))
		return err
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write to GITHUB_OUTPUT (for step outputs)
	if outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open GITHUB_OUTPUT: %w", err)
		}
		defer file.Close()

		for _, k := range keys {
			// Convert to lowercase for output names (convention)
			outputName := strings.ToLower(strings.ReplaceAll(k, "_", "-"))
			if _, err := fmt.Fprintf(file, "%s=%s\n", outputName, vars[k]); err != nil {
				return fmt.Errorf("failed to write to GITHUB_OUTPUT: %w", err)
			}
		}
	}

	// Write to GITHUB_ENV (for environment variables)
	if envFile != "" {
		file, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open GITHUB_ENV: %w", err)
		}
		defer file.Close()

		for _, k := range keys {
			if _, err := fmt.Fprintf(file, "%s=%s\n", k, vars[k]); err != nil {
				return fmt.Errorf("failed to write to GITHUB_ENV: %w", err)
			}
		}
	}

	// Also write summary to the provided writer
	_, err := fmt.Fprintf(w, "Set %d variables in GitHub Actions\n", len(vars))
	return err
}
