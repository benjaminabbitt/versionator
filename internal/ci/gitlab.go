package ci

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// GitLabFormatter outputs variables for GitLab CI
// Uses dotenv format for artifacts or echo for job-level variables
type GitLabFormatter struct{}

// Name returns the formatter name
func (f *GitLabFormatter) Name() string {
	return "gitlab"
}

// Format returns dotenv format output
func (f *GitLabFormatter) Format(vars map[string]string) string {
	var sb strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vars[k]
		// Quote values that contain special characters
		if strings.ContainsAny(v, " \t\n\"'") {
			v = fmt.Sprintf("\"%s\"", strings.ReplaceAll(v, "\"", "\\\""))
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	return sb.String()
}

// Write writes the dotenv format to the writer
func (f *GitLabFormatter) Write(vars map[string]string, w io.Writer) error {
	_, err := w.Write([]byte(f.Format(vars)))
	return err
}
