package ci

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDetect_GitHubActions(t *testing.T) {
	os.Setenv("GITHUB_ACTIONS", "true")
	defer os.Unsetenv("GITHUB_ACTIONS")

	env := Detect()
	if env != EnvGitHubActions {
		t.Errorf("expected EnvGitHubActions, got %v", env)
	}
}

func TestDetect_GitLabCI(t *testing.T) {
	os.Setenv("GITLAB_CI", "true")
	defer os.Unsetenv("GITLAB_CI")

	env := Detect()
	if env != EnvGitLabCI {
		t.Errorf("expected EnvGitLabCI, got %v", env)
	}
}

func TestDetect_AzureDevOps(t *testing.T) {
	os.Setenv("TF_BUILD", "True")
	defer os.Unsetenv("TF_BUILD")

	env := Detect()
	if env != EnvAzureDevOps {
		t.Errorf("expected EnvAzureDevOps, got %v", env)
	}
}

func TestDetect_CircleCI(t *testing.T) {
	os.Setenv("CIRCLECI", "true")
	defer os.Unsetenv("CIRCLECI")

	env := Detect()
	if env != EnvCircleCI {
		t.Errorf("expected EnvCircleCI, got %v", env)
	}
}

func TestDetect_Jenkins(t *testing.T) {
	os.Setenv("JENKINS_URL", "http://jenkins.example.com")
	defer os.Unsetenv("JENKINS_URL")

	env := Detect()
	if env != EnvJenkins {
		t.Errorf("expected EnvJenkins, got %v", env)
	}
}

func TestDetect_None(t *testing.T) {
	// Clear all CI environment variables
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITLAB_CI")
	os.Unsetenv("TF_BUILD")
	os.Unsetenv("CIRCLECI")
	os.Unsetenv("JENKINS_URL")

	env := Detect()
	if env != EnvNone {
		t.Errorf("expected EnvNone, got %v", env)
	}
}

func TestShellFormatter_Format(t *testing.T) {
	formatter := &ShellFormatter{}
	vars := map[string]string{
		"VERSION":       "1.2.3",
		"VERSION_MAJOR": "1",
	}

	output := formatter.Format(vars)

	if !strings.Contains(output, `export VERSION="1.2.3"`) {
		t.Error("expected VERSION export")
	}
	if !strings.Contains(output, `export VERSION_MAJOR="1"`) {
		t.Error("expected VERSION_MAJOR export")
	}
}

func TestShellFormatter_EscapesQuotes(t *testing.T) {
	formatter := &ShellFormatter{}
	vars := map[string]string{
		"MESSAGE": `hello "world"`,
	}

	output := formatter.Format(vars)

	if !strings.Contains(output, `export MESSAGE="hello \"world\""`) {
		t.Errorf("expected escaped quotes, got: %s", output)
	}
}

func TestAzureFormatter_Format(t *testing.T) {
	formatter := &AzureFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	output := formatter.Format(vars)

	expected := "##vso[task.setvariable variable=VERSION]1.2.3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestGitLabFormatter_Format(t *testing.T) {
	formatter := &GitLabFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	output := formatter.Format(vars)

	expected := "VERSION=1.2.3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestJenkinsFormatter_Format(t *testing.T) {
	formatter := &JenkinsFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	output := formatter.Format(vars)

	expected := "VERSION=1.2.3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestGetFormatter_ReturnsCorrectType(t *testing.T) {
	tests := []struct {
		env      Environment
		expected string
	}{
		{EnvGitHubActions, "github"},
		{EnvGitLabCI, "gitlab"},
		{EnvAzureDevOps, "azure"},
		{EnvCircleCI, "circleci"},
		{EnvJenkins, "jenkins"},
		{EnvNone, "shell"},
		{EnvGeneric, "shell"},
	}

	for _, tt := range tests {
		formatter := GetFormatter(tt.env)
		if formatter.Name() != tt.expected {
			t.Errorf("GetFormatter(%v).Name() = %s, want %s", tt.env, formatter.Name(), tt.expected)
		}
	}
}

func TestGetFormatterByName_ValidNames(t *testing.T) {
	for _, name := range AvailableFormatters() {
		formatter, err := GetFormatterByName(name)
		if err != nil {
			t.Errorf("GetFormatterByName(%q) returned error: %v", name, err)
		}
		if formatter.Name() != name {
			t.Errorf("GetFormatterByName(%q).Name() = %q", name, formatter.Name())
		}
	}
}

func TestGetFormatterByName_InvalidName(t *testing.T) {
	_, err := GetFormatterByName("invalid")
	if err == nil {
		t.Error("expected error for invalid formatter name")
	}
}

func TestVariables_ToMap(t *testing.T) {
	vars := &Variables{
		Version:     "v1.2.3",
		Major:       "1",
		Minor:       "2",
		Patch:       "3",
		GitSHA:      "abc123",
		GitSHAShort: "abc1234",
		GitBranch:   "main",
	}

	result := vars.ToMap("")

	if result["VERSION"] != "v1.2.3" {
		t.Errorf("expected VERSION='v1.2.3', got %q", result["VERSION"])
	}
	if result["VERSION_MAJOR"] != "1" {
		t.Errorf("expected VERSION_MAJOR='1', got %q", result["VERSION_MAJOR"])
	}
}

func TestVariables_ToMapWithPrefix(t *testing.T) {
	vars := &Variables{
		Version: "v1.2.3",
	}

	result := vars.ToMap("MYAPP_")

	if _, ok := result["MYAPP_VERSION"]; !ok {
		t.Error("expected MYAPP_VERSION to be present")
	}
	if result["MYAPP_VERSION"] != "v1.2.3" {
		t.Errorf("expected MYAPP_VERSION='v1.2.3', got %q", result["MYAPP_VERSION"])
	}
}

func TestShellFormatter_Write(t *testing.T) {
	formatter := &ShellFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "export VERSION") {
		t.Errorf("expected export statement in output")
	}
}
