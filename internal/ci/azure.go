package ci

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// AzureFormatter outputs variables for Azure DevOps
// Uses ##vso[task.setvariable] format
type AzureFormatter struct{}

// Name returns the formatter name
func (f *AzureFormatter) Name() string {
	return "azure"
}

// Format returns Azure DevOps variable commands
func (f *AzureFormatter) Format(vars map[string]string) string {
	var sb strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vars[k]
		// Azure DevOps format: ##vso[task.setvariable variable=NAME]VALUE
		sb.WriteString(fmt.Sprintf("##vso[task.setvariable variable=%s]%s\n", k, v))
	}

	return sb.String()
}

// Write writes the Azure DevOps commands to the writer
func (f *AzureFormatter) Write(vars map[string]string, w io.Writer) error {
	_, err := w.Write([]byte(f.Format(vars)))
	return err
}
