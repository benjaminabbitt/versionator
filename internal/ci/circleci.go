package ci

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// CircleCIFormatter outputs variables for CircleCI
// Appends to $BASH_ENV for environment persistence
type CircleCIFormatter struct{}

// Name returns the formatter name
func (f *CircleCIFormatter) Name() string {
	return "circleci"
}

// Format returns shell export statements (same as shell formatter)
func (f *CircleCIFormatter) Format(vars map[string]string) string {
	var sb strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vars[k]
		escaped := strings.ReplaceAll(v, "\"", "\\\"")
		sb.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, escaped))
	}

	return sb.String()
}

// Write appends to BASH_ENV if available, otherwise writes to the provided writer
func (f *CircleCIFormatter) Write(vars map[string]string, w io.Writer) error {
	bashEnv := os.Getenv("BASH_ENV")

	if bashEnv == "" {
		// Not in CircleCI, just write to the provided writer
		_, err := w.Write([]byte(f.Format(vars)))
		return err
	}

	// Append to BASH_ENV
	file, err := os.OpenFile(bashEnv, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open BASH_ENV: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(f.Format(vars)); err != nil {
		return fmt.Errorf("failed to write to BASH_ENV: %w", err)
	}

	// Also write summary to the provided writer
	_, err = fmt.Fprintf(w, "Appended %d variables to BASH_ENV\n", len(vars))
	return err
}
