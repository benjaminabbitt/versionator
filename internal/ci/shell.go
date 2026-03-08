package ci

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ShellFormatter outputs shell export statements
type ShellFormatter struct{}

// Name returns the formatter name
func (f *ShellFormatter) Name() string {
	return "shell"
}

// Format returns shell export statements
func (f *ShellFormatter) Format(vars map[string]string) string {
	var sb strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vars[k]
		// Escape double quotes in value
		escaped := strings.ReplaceAll(v, "\"", "\\\"")
		sb.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, escaped))
	}

	return sb.String()
}

// Write writes the shell exports to the writer
func (f *ShellFormatter) Write(vars map[string]string, w io.Writer) error {
	_, err := w.Write([]byte(f.Format(vars)))
	return err
}
