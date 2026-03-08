package ci

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// JenkinsFormatter outputs variables in properties file format
// Can be used with EnvInject plugin or properties file
type JenkinsFormatter struct{}

// Name returns the formatter name
func (f *JenkinsFormatter) Name() string {
	return "jenkins"
}

// Format returns properties file format (KEY=value)
func (f *JenkinsFormatter) Format(vars map[string]string) string {
	var sb strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vars[k]
		// Properties files: escape backslashes and special chars
		v = strings.ReplaceAll(v, "\\", "\\\\")
		v = strings.ReplaceAll(v, "\n", "\\n")
		v = strings.ReplaceAll(v, "\r", "\\r")
		v = strings.ReplaceAll(v, "\t", "\\t")
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	return sb.String()
}

// Write writes the properties format to the writer
func (f *JenkinsFormatter) Write(vars map[string]string, w io.Writer) error {
	_, err := w.Write([]byte(f.Format(vars)))
	return err
}
