package ci

import (
	"fmt"
	"io"
)

// Formatter defines the interface for CI-specific output
type Formatter interface {
	// Name returns the formatter name
	Name() string

	// Format returns the formatted output as a string
	Format(vars map[string]string) string

	// Write writes the variables to the appropriate destination
	// For most CIs, this writes to specific files or stdout
	Write(vars map[string]string, w io.Writer) error
}

// GetFormatter returns the appropriate formatter for the environment
func GetFormatter(env Environment) Formatter {
	switch env {
	case EnvGitHubActions:
		return &GitHubFormatter{}
	case EnvGitLabCI:
		return &GitLabFormatter{}
	case EnvAzureDevOps:
		return &AzureFormatter{}
	case EnvCircleCI:
		return &CircleCIFormatter{}
	case EnvJenkins:
		return &JenkinsFormatter{}
	default:
		return &ShellFormatter{}
	}
}

// GetFormatterByName returns a formatter by name
func GetFormatterByName(name string) (Formatter, error) {
	switch name {
	case "github":
		return &GitHubFormatter{}, nil
	case "gitlab":
		return &GitLabFormatter{}, nil
	case "azure":
		return &AzureFormatter{}, nil
	case "circleci":
		return &CircleCIFormatter{}, nil
	case "jenkins":
		return &JenkinsFormatter{}, nil
	case "shell":
		return &ShellFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown formatter: %s", name)
	}
}

// AvailableFormatters returns the list of available formatter names
func AvailableFormatters() []string {
	return []string{"github", "gitlab", "azure", "circleci", "jenkins", "shell"}
}
