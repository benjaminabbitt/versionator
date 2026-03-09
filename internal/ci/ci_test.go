package ci

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the CI package: detecting CI
// environments and formatting version variables for output.
// =============================================================================

// TestDetect_GitHubActions_ReturnsGitHubEnvironment validates that when running
// in GitHub Actions (indicated by the GITHUB_ACTIONS=true environment variable),
// the Detect function correctly identifies the CI environment.
//
// Why: GitHub Actions is one of the most common CI environments. Correct
// detection enables the appropriate output format for GitHub Actions workflows.
//
// What: When GITHUB_ACTIONS=true is set, Detect should return EnvGitHubActions.
func TestDetect_GitHubActions_ReturnsGitHubEnvironment(t *testing.T) {
	// Precondition: GITHUB_ACTIONS environment variable is set to "true"
	os.Setenv("GITHUB_ACTIONS", "true")
	defer os.Unsetenv("GITHUB_ACTIONS")

	// Action: Detect the CI environment
	env := Detect()

	// Expected: The environment is identified as GitHub Actions
	if env != EnvGitHubActions {
		t.Errorf("expected EnvGitHubActions, got %v", env)
	}
}

// TestGetFormatter_ReturnsCorrectFormatterForEachEnvironment validates that
// GetFormatter maps each CI environment to its corresponding formatter type.
//
// Why: Each CI platform has different requirements for setting environment
// variables. The correct formatter ensures variables are exported in a format
// that the CI platform can consume.
//
// What: Each Environment constant should map to a formatter with the expected name.
func TestGetFormatter_ReturnsCorrectFormatterForEachEnvironment(t *testing.T) {
	// Precondition: All Environment constants and their expected formatter names
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
		// Action: Get the formatter for this environment
		formatter := GetFormatter(tt.env)

		// Expected: The formatter's name matches the expected name
		if formatter.Name() != tt.expected {
			t.Errorf("GetFormatter(%v).Name() = %s, want %s", tt.env, formatter.Name(), tt.expected)
		}
	}
}

// TestShellFormatter_Format_ProducesExportStatements validates that the shell
// formatter produces valid shell export statements that can be sourced.
//
// Why: The shell formatter is the default formatter and is used when no CI
// environment is detected. Its output must be valid shell syntax.
//
// What: Given a map of variables, Format should produce "export KEY="value""
// statements for each entry.
func TestShellFormatter_Format_ProducesExportStatements(t *testing.T) {
	// Precondition: A shell formatter and a map of version variables
	formatter := &ShellFormatter{}
	vars := map[string]string{
		"VERSION":       "1.2.3",
		"VERSION_MAJOR": "1",
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: The output contains export statements for each variable
	if !strings.Contains(output, `export VERSION="1.2.3"`) {
		t.Error("expected VERSION export")
	}
	if !strings.Contains(output, `export VERSION_MAJOR="1"`) {
		t.Error("expected VERSION_MAJOR export")
	}
}

// TestVariables_ToMap_ConvertsStructToStandardVariableNames validates that
// the Variables struct is correctly converted to a map with standard CI
// variable names.
//
// Why: The Variables struct provides a typed way to work with version info
// internally, but CI systems need a flat map with standardized names.
//
// What: ToMap should convert each field to its corresponding variable name
// (e.g., Version -> VERSION, Major -> VERSION_MAJOR).
func TestVariables_ToMap_ConvertsStructToStandardVariableNames(t *testing.T) {
	// Precondition: A populated Variables struct
	vars := &Variables{
		Version:     "v1.2.3",
		Major:       "1",
		Minor:       "2",
		Patch:       "3",
		GitSHA:      "abc123",
		GitSHAShort: "abc1234",
		GitBranch:   "main",
	}

	// Action: Convert to a map without a prefix
	result := vars.ToMap("")

	// Expected: Each field maps to its standard variable name
	if result["VERSION"] != "v1.2.3" {
		t.Errorf("expected VERSION='v1.2.3', got %q", result["VERSION"])
	}
	if result["VERSION_MAJOR"] != "1" {
		t.Errorf("expected VERSION_MAJOR='1', got %q", result["VERSION_MAJOR"])
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for different CI environments and formatter types, verifying each
// platform's specific output format requirements.
// =============================================================================

// TestDetect_GitLabCI_ReturnsGitLabEnvironment validates detection of GitLab CI.
//
// Why: GitLab CI uses a different environment variable (GITLAB_CI) and has its
// own output format for setting variables in subsequent jobs.
//
// What: When GITLAB_CI=true is set, Detect should return EnvGitLabCI.
func TestDetect_GitLabCI_ReturnsGitLabEnvironment(t *testing.T) {
	// Precondition: GITLAB_CI environment variable is set to "true"
	os.Setenv("GITLAB_CI", "true")
	defer os.Unsetenv("GITLAB_CI")

	// Action: Detect the CI environment
	env := Detect()

	// Expected: The environment is identified as GitLab CI
	if env != EnvGitLabCI {
		t.Errorf("expected EnvGitLabCI, got %v", env)
	}
}

// TestDetect_AzureDevOps_ReturnsAzureEnvironment validates detection of Azure DevOps.
//
// Why: Azure DevOps uses TF_BUILD=True (note capital T) to indicate its environment.
//
// What: When TF_BUILD=True is set, Detect should return EnvAzureDevOps.
func TestDetect_AzureDevOps_ReturnsAzureEnvironment(t *testing.T) {
	// Precondition: TF_BUILD environment variable is set to "True"
	os.Setenv("TF_BUILD", "True")
	defer os.Unsetenv("TF_BUILD")

	// Action: Detect the CI environment
	env := Detect()

	// Expected: The environment is identified as Azure DevOps
	if env != EnvAzureDevOps {
		t.Errorf("expected EnvAzureDevOps, got %v", env)
	}
}

// TestDetect_CircleCI_ReturnsCircleCIEnvironment validates detection of CircleCI.
//
// Why: CircleCI uses CIRCLECI=true to indicate its environment.
//
// What: When CIRCLECI=true is set, Detect should return EnvCircleCI.
func TestDetect_CircleCI_ReturnsCircleCIEnvironment(t *testing.T) {
	// Precondition: CIRCLECI environment variable is set to "true"
	os.Setenv("CIRCLECI", "true")
	defer os.Unsetenv("CIRCLECI")

	// Action: Detect the CI environment
	env := Detect()

	// Expected: The environment is identified as CircleCI
	if env != EnvCircleCI {
		t.Errorf("expected EnvCircleCI, got %v", env)
	}
}

// TestDetect_Jenkins_ReturnsJenkinsEnvironment validates detection of Jenkins.
//
// Why: Jenkins uses JENKINS_URL (any non-empty value) to indicate its environment.
//
// What: When JENKINS_URL is set, Detect should return EnvJenkins.
func TestDetect_Jenkins_ReturnsJenkinsEnvironment(t *testing.T) {
	// Precondition: JENKINS_URL environment variable is set
	os.Setenv("JENKINS_URL", "http://jenkins.example.com")
	defer os.Unsetenv("JENKINS_URL")

	// Action: Detect the CI environment
	env := Detect()

	// Expected: The environment is identified as Jenkins
	if env != EnvJenkins {
		t.Errorf("expected EnvJenkins, got %v", env)
	}
}

// TestGitHubFormatter_Format_ProducesGitHubActionsFormat validates the GitHub
// Actions output format which includes a header comment and KEY=value lines.
//
// Why: GitHub Actions uses a specific format for step outputs. The format
// must be correct for workflows to consume the version variables.
//
// What: Output should include a header comment and KEY=value lines for each variable.
func TestGitHubFormatter_Format_ProducesGitHubActionsFormat(t *testing.T) {
	// Precondition: A GitHub formatter and version variables
	formatter := &GitHubFormatter{}
	vars := map[string]string{
		"VERSION":       "1.2.3",
		"VERSION_MAJOR": "1",
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Output has header comment and KEY=value format
	if !strings.Contains(output, "# GitHub Actions Output Variables") {
		t.Error("expected header comment in output")
	}
	if !strings.Contains(output, "VERSION=1.2.3") {
		t.Error("expected VERSION=1.2.3 in output")
	}
	if !strings.Contains(output, "VERSION_MAJOR=1") {
		t.Error("expected VERSION_MAJOR=1 in output")
	}
}

// TestAzureFormatter_Format_ProducesVSOCommandFormat validates the Azure DevOps
// output format which uses ##vso commands to set pipeline variables.
//
// Why: Azure DevOps requires variables to be set using logging commands with
// the ##vso[task.setvariable] syntax.
//
// What: Output should use the ##vso[task.setvariable variable=NAME]value format.
func TestAzureFormatter_Format_ProducesVSOCommandFormat(t *testing.T) {
	// Precondition: An Azure formatter and version variables
	formatter := &AzureFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Output uses Azure DevOps vso command format
	expected := "##vso[task.setvariable variable=VERSION]1.2.3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestGitLabFormatter_Format_ProducesDotenvFormat validates the GitLab CI
// output format which uses KEY=value lines for dotenv artifact.
//
// Why: GitLab CI uses dotenv reports to pass variables between jobs. The output
// must be valid dotenv format.
//
// What: Output should use simple KEY=value format.
func TestGitLabFormatter_Format_ProducesDotenvFormat(t *testing.T) {
	// Precondition: A GitLab formatter and version variables
	formatter := &GitLabFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Output uses dotenv format
	expected := "VERSION=1.2.3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestJenkinsFormatter_Format_ProducesPropertiesFormat validates the Jenkins
// output format which uses KEY=value lines for properties file.
//
// Why: Jenkins uses properties files to pass variables between stages. The
// output must be valid Java properties format.
//
// What: Output should use simple KEY=value format.
func TestJenkinsFormatter_Format_ProducesPropertiesFormat(t *testing.T) {
	// Precondition: A Jenkins formatter and version variables
	formatter := &JenkinsFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Output uses properties format
	expected := "VERSION=1.2.3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestCircleCIFormatter_Format_ProducesShellExportFormat validates the CircleCI
// output format which uses shell export statements.
//
// Why: CircleCI uses BASH_ENV to persist environment variables. The output
// must be valid shell syntax that can be sourced.
//
// What: Output should produce export KEY="value" statements.
func TestCircleCIFormatter_Format_ProducesShellExportFormat(t *testing.T) {
	// Precondition: A CircleCI formatter and version variables
	formatter := &CircleCIFormatter{}
	vars := map[string]string{
		"VERSION": "1.2.3",
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Output uses shell export format
	if !strings.Contains(output, `export VERSION="1.2.3"`) {
		t.Errorf("expected export statement, got: %s", output)
	}
}

// TestGetFormatterByName_ValidNames_ReturnsCorrectFormatter validates that
// GetFormatterByName returns the correct formatter for each valid name.
//
// Why: Users may explicitly specify a formatter by name via CLI flags.
// Each valid name must return the corresponding formatter.
//
// What: Each name from AvailableFormatters() should return a formatter
// whose Name() matches the requested name.
func TestGetFormatterByName_ValidNames_ReturnsCorrectFormatter(t *testing.T) {
	// Precondition: All available formatter names
	for _, name := range AvailableFormatters() {
		// Action: Get the formatter by name
		formatter, err := GetFormatterByName(name)

		// Expected: No error and the formatter's name matches
		if err != nil {
			t.Errorf("GetFormatterByName(%q) returned error: %v", name, err)
		}
		if formatter.Name() != name {
			t.Errorf("GetFormatterByName(%q).Name() = %q", name, formatter.Name())
		}
	}
}

// TestVariables_ToMapWithPrefix_AppliesPrefixToAllVariableNames validates that
// ToMap correctly applies a prefix to all generated variable names.
//
// Why: Some CI configurations require prefixed variables to avoid naming
// conflicts (e.g., MYAPP_VERSION instead of VERSION).
//
// What: When a prefix is provided, all variable names should include it.
func TestVariables_ToMapWithPrefix_AppliesPrefixToAllVariableNames(t *testing.T) {
	// Precondition: A Variables struct with a version
	vars := &Variables{
		Version: "v1.2.3",
	}

	// Action: Convert to a map with a prefix
	result := vars.ToMap("MYAPP_")

	// Expected: Variable names have the prefix
	if _, ok := result["MYAPP_VERSION"]; !ok {
		t.Error("expected MYAPP_VERSION to be present")
	}
	if result["MYAPP_VERSION"] != "v1.2.3" {
		t.Errorf("expected MYAPP_VERSION='v1.2.3', got %q", result["MYAPP_VERSION"])
	}
}

// TestEnvironment_String_ReturnsExpectedStringRepresentation validates that
// Environment constants can be converted to their string representation.
//
// Why: String representation is used for display, logging, and formatter selection.
//
// What: Each Environment constant should return its expected string value.
func TestEnvironment_String_ReturnsExpectedStringRepresentation(t *testing.T) {
	// Precondition: All Environment constants and their expected strings
	tests := []struct {
		env      Environment
		expected string
	}{
		{EnvNone, "none"},
		{EnvGitHubActions, "github"},
		{EnvGitLabCI, "gitlab"},
		{EnvAzureDevOps, "azure"},
		{EnvCircleCI, "circleci"},
		{EnvJenkins, "jenkins"},
		{EnvGeneric, "shell"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// Action: Get the string representation
			result := tt.env.String()

			// Expected: The string matches the expected value
			if result != tt.expected {
				t.Errorf("Environment.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// WRITE OPERATIONS
// Tests for the Write method of each formatter, which outputs variables to
// the appropriate destination (stdout, files, or CI-specific locations).
// =============================================================================

// TestShellFormatter_Write_WritesToProvidedWriter validates that the shell
// formatter writes its output to the provided io.Writer.
//
// Why: When not in a CI environment, output should go to the provided writer
// (typically stdout) for manual sourcing.
//
// What: Write should produce export statements in the provided buffer.
func TestShellFormatter_Write_WritesToProvidedWriter(t *testing.T) {
	// Precondition: A shell formatter and version variables
	formatter := &ShellFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write to a buffer
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: No error and output contains export statements
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "export VERSION") {
		t.Errorf("expected export statement in output")
	}
}

// TestGitHubFormatter_Write_NoEnvVars_WritesToProvidedWriter validates that
// when not running in GitHub Actions, output goes to the provided writer.
//
// Why: When GITHUB_OUTPUT and GITHUB_ENV are not set, the formatter should
// fall back to writing to the provided writer for manual use.
//
// What: When GitHub env vars are unset, Write should output to the buffer.
func TestGitHubFormatter_Write_NoEnvVars_WritesToProvidedWriter(t *testing.T) {
	// Precondition: GitHub environment variables are not set
	os.Unsetenv("GITHUB_OUTPUT")
	os.Unsetenv("GITHUB_ENV")

	formatter := &GitHubFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write to a buffer
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: No error and output contains the variable
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "VERSION=1.2.3") {
		t.Errorf("expected VERSION in output, got: %s", buf.String())
	}
}

// TestGitHubFormatter_Write_WithEnvVars_WritesToGitHubFiles validates that
// when running in GitHub Actions, variables are written to GITHUB_OUTPUT
// and GITHUB_ENV files.
//
// Why: GitHub Actions requires variables to be written to specific files
// (GITHUB_OUTPUT for step outputs, GITHUB_ENV for environment variables).
//
// What: When GitHub env vars are set, Write should write to those files.
func TestGitHubFormatter_Write_WithEnvVars_WritesToGitHubFiles(t *testing.T) {
	// Precondition: Temporary files simulating GitHub Actions environment
	tempDir := t.TempDir()
	outputFile := tempDir + "/github_output"
	envFile := tempDir + "/github_env"

	os.Setenv("GITHUB_OUTPUT", outputFile)
	os.Setenv("GITHUB_ENV", envFile)
	defer os.Unsetenv("GITHUB_OUTPUT")
	defer os.Unsetenv("GITHUB_ENV")

	formatter := &GitHubFormatter{}
	vars := map[string]string{
		"VERSION":       "1.2.3",
		"VERSION_MAJOR": "1",
	}

	// Action: Write the variables
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: Variables are written to both files
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read GITHUB_OUTPUT: %v", err)
	}
	if !strings.Contains(string(outputContent), "version=1.2.3") {
		t.Errorf("expected version=1.2.3 in GITHUB_OUTPUT, got: %s", outputContent)
	}

	envContent, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("failed to read GITHUB_ENV: %v", err)
	}
	if !strings.Contains(string(envContent), "VERSION=1.2.3") {
		t.Errorf("expected VERSION=1.2.3 in GITHUB_ENV, got: %s", envContent)
	}

	if !strings.Contains(buf.String(), "Set 2 variables") {
		t.Errorf("expected summary message, got: %s", buf.String())
	}
}

// TestCircleCIFormatter_Write_NoEnvVars_WritesToProvidedWriter validates that
// when BASH_ENV is not set, output goes to the provided writer.
//
// Why: When not in CircleCI, the formatter should output to the provided
// writer for manual use or debugging.
//
// What: When BASH_ENV is unset, Write should output to the buffer.
func TestCircleCIFormatter_Write_NoEnvVars_WritesToProvidedWriter(t *testing.T) {
	// Precondition: BASH_ENV is not set
	os.Unsetenv("BASH_ENV")

	formatter := &CircleCIFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write to a buffer
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: No error and output contains export statements
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "export VERSION") {
		t.Errorf("expected export statement, got: %s", buf.String())
	}
}

// TestCircleCIFormatter_Write_WithBashEnv_WritesToBashEnvFile validates that
// when running in CircleCI, variables are appended to the BASH_ENV file.
//
// Why: CircleCI persists environment variables by writing to BASH_ENV, which
// is sourced before each step.
//
// What: When BASH_ENV is set, Write should append to that file.
func TestCircleCIFormatter_Write_WithBashEnv_WritesToBashEnvFile(t *testing.T) {
	// Precondition: Temporary file simulating CircleCI BASH_ENV
	tempDir := t.TempDir()
	bashEnvFile := tempDir + "/bash_env"

	os.Setenv("BASH_ENV", bashEnvFile)
	defer os.Unsetenv("BASH_ENV")

	formatter := &CircleCIFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write the variables
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: Variables are written to BASH_ENV file
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(bashEnvFile)
	if err != nil {
		t.Fatalf("failed to read BASH_ENV: %v", err)
	}
	if !strings.Contains(string(content), "export VERSION") {
		t.Errorf("expected export in BASH_ENV, got: %s", content)
	}

	if !strings.Contains(buf.String(), "Appended 1 variables to BASH_ENV") {
		t.Errorf("expected summary message, got: %s", buf.String())
	}
}

// TestAzureFormatter_Write_WritesToProvidedWriter validates that the Azure
// DevOps formatter writes its output to the provided writer.
//
// Why: Azure DevOps uses logging commands that are parsed from stdout.
//
// What: Write should produce ##vso commands in the provided buffer.
func TestAzureFormatter_Write_WritesToProvidedWriter(t *testing.T) {
	// Precondition: An Azure formatter and version variables
	formatter := &AzureFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write to a buffer
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: No error and output uses Azure format
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "##vso[task.setvariable variable=VERSION]1.2.3") {
		t.Errorf("expected Azure DevOps format, got: %s", buf.String())
	}
}

// TestGitLabFormatter_Write_WritesToProvidedWriter validates that the GitLab
// CI formatter writes its output to the provided writer.
//
// Why: GitLab CI typically redirects output to a dotenv file. The Write
// method outputs to the provided writer for flexibility.
//
// What: Write should produce dotenv format in the provided buffer.
func TestGitLabFormatter_Write_WritesToProvidedWriter(t *testing.T) {
	// Precondition: A GitLab formatter and version variables
	formatter := &GitLabFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write to a buffer
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: No error and output uses dotenv format
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "VERSION=1.2.3") {
		t.Errorf("expected dotenv format, got: %s", buf.String())
	}
}

// TestJenkinsFormatter_Write_WritesToProvidedWriter validates that the Jenkins
// formatter writes its output to the provided writer.
//
// Why: Jenkins output is typically redirected to a properties file. The Write
// method outputs to the provided writer for flexibility.
//
// What: Write should produce properties format in the provided buffer.
func TestJenkinsFormatter_Write_WritesToProvidedWriter(t *testing.T) {
	// Precondition: A Jenkins formatter and version variables
	formatter := &JenkinsFormatter{}
	vars := map[string]string{"VERSION": "1.2.3"}

	// Action: Write to a buffer
	var buf bytes.Buffer
	err := formatter.Write(vars, &buf)

	// Expected: No error and output uses properties format
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "VERSION=1.2.3") {
		t.Errorf("expected properties format, got: %s", buf.String())
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes and error conditions.
// =============================================================================

// TestDetect_NoCIEnvironment_ReturnsEnvNone validates that when no CI
// environment variables are set, Detect returns EnvNone.
//
// Why: When running locally or in an unknown environment, the detection
// should gracefully fall back to EnvNone rather than failing.
//
// What: With no CI environment variables set, Detect returns EnvNone.
func TestDetect_NoCIEnvironment_ReturnsEnvNone(t *testing.T) {
	// Precondition: All CI environment variables are unset
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITLAB_CI")
	os.Unsetenv("TF_BUILD")
	os.Unsetenv("CIRCLECI")
	os.Unsetenv("JENKINS_URL")

	// Action: Detect the CI environment
	env := Detect()

	// Expected: EnvNone is returned
	if env != EnvNone {
		t.Errorf("expected EnvNone, got %v", env)
	}
}

// TestGetFormatterByName_InvalidName_ReturnsError validates that requesting
// an unknown formatter name returns an error.
//
// Why: Users may typo a formatter name or use an outdated name. The error
// should clearly indicate the problem.
//
// What: GetFormatterByName("invalid") should return an error.
func TestGetFormatterByName_InvalidName_ReturnsError(t *testing.T) {
	// Precondition: An invalid formatter name
	// Action: Request the formatter
	_, err := GetFormatterByName("invalid")

	// Expected: An error is returned
	if err == nil {
		t.Error("expected error for invalid formatter name")
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and special character handling.
// =============================================================================

// TestShellFormatter_Format_EscapesQuotesInValues validates that double quotes
// in variable values are properly escaped in shell export statements.
//
// Why: Unescaped quotes would break shell syntax and potentially cause
// command injection vulnerabilities.
//
// What: A value containing quotes should have them escaped with backslashes.
func TestShellFormatter_Format_EscapesQuotesInValues(t *testing.T) {
	// Precondition: A value containing double quotes
	formatter := &ShellFormatter{}
	vars := map[string]string{
		"MESSAGE": `hello "world"`,
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Quotes are escaped
	if !strings.Contains(output, `export MESSAGE="hello \"world\""`) {
		t.Errorf("expected escaped quotes, got: %s", output)
	}
}

// TestCircleCIFormatter_Format_EscapesQuotesInValues validates that the
// CircleCI formatter also escapes quotes in values.
//
// Why: CircleCI uses the same shell export format, so quotes must be escaped.
//
// What: A value containing quotes should have them escaped.
func TestCircleCIFormatter_Format_EscapesQuotesInValues(t *testing.T) {
	// Precondition: A value containing double quotes
	formatter := &CircleCIFormatter{}
	vars := map[string]string{
		"MESSAGE": `hello "world"`,
	}

	// Action: Format the variables
	output := formatter.Format(vars)

	// Expected: Quotes are escaped
	if !strings.Contains(output, `export MESSAGE="hello \"world\""`) {
		t.Errorf("expected escaped quotes, got: %s", output)
	}
}

// TestGitLabFormatter_Format_QuotesValuesWithSpecialCharacters validates that
// values containing special characters are properly quoted in dotenv format.
//
// Why: Dotenv format requires quoting for values with spaces, tabs, newlines,
// or quotes to be parsed correctly.
//
// What: Values with special characters should be quoted and escaped appropriately.
func TestGitLabFormatter_Format_QuotesValuesWithSpecialCharacters(t *testing.T) {
	// Precondition: A GitLab formatter
	formatter := &GitLabFormatter{}

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "space", value: "hello world", expected: `"hello world"`},
		{name: "tab", value: "hello\tworld", expected: `"hello	world"`},
		{name: "newline", value: "hello\nworld", expected: `"hello`},
		{name: "quotes", value: `say "hi"`, expected: `"say \"hi\""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]string{"KEY": tt.value}

			// Action: Format the variables
			output := formatter.Format(vars)

			// Expected: Value is properly quoted
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected %q in output, got: %s", tt.expected, output)
			}
		})
	}
}

// TestJenkinsFormatter_Format_EscapesPropertiesFileSpecialCharacters validates
// that special characters are escaped according to Java properties file format.
//
// Why: Java properties files require backslash escaping for backslashes,
// newlines, carriage returns, and tabs.
//
// What: Special characters should be escaped with backslash sequences.
func TestJenkinsFormatter_Format_EscapesPropertiesFileSpecialCharacters(t *testing.T) {
	// Precondition: A Jenkins formatter
	formatter := &JenkinsFormatter{}

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "backslash", value: `path\to\file`, expected: `path\\to\\file`},
		{name: "newline", value: "line1\nline2", expected: `line1\nline2`},
		{name: "carriage return", value: "line1\rline2", expected: `line1\rline2`},
		{name: "tab", value: "col1\tcol2", expected: `col1\tcol2`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]string{"KEY": tt.value}

			// Action: Format the variables
			output := formatter.Format(vars)

			// Expected: Special characters are escaped
			if !strings.Contains(output, "KEY="+tt.expected) {
				t.Errorf("expected KEY=%s in output, got: %s", tt.expected, output)
			}
		})
	}
}

// TestVariables_ToMap_AllFields_MapsEveryFieldCorrectly validates that every
// field in the Variables struct is mapped to its corresponding variable name.
//
// Why: Ensures complete coverage of all version-related fields that users
// may need in their CI pipelines.
//
// What: All 13 fields should appear in the output map with correct values.
func TestVariables_ToMap_AllFields_MapsEveryFieldCorrectly(t *testing.T) {
	// Precondition: A fully populated Variables struct
	vars := &Variables{
		Version:       "v1.2.3-alpha+build",
		VersionSemver: "1.2.3-alpha",
		VersionCore:   "1.2.3",
		Major:         "1",
		Minor:         "2",
		Patch:         "3",
		PreRelease:    "alpha",
		Metadata:      "build",
		GitSHA:        "abc123def456",
		GitSHAShort:   "abc123d",
		GitBranch:     "feature/test",
		BuildNumber:   "42",
		Dirty:         "false",
	}

	// Action: Convert to a map
	result := vars.ToMap("")

	// Expected: All fields are correctly mapped
	expected := map[string]string{
		"VERSION":            "v1.2.3-alpha+build",
		"VERSION_SEMVER":     "1.2.3-alpha",
		"VERSION_CORE":       "1.2.3",
		"VERSION_MAJOR":      "1",
		"VERSION_MINOR":      "2",
		"VERSION_PATCH":      "3",
		"VERSION_PRERELEASE": "alpha",
		"VERSION_METADATA":   "build",
		"GIT_SHA":            "abc123def456",
		"GIT_SHA_SHORT":      "abc123d",
		"GIT_BRANCH":         "feature/test",
		"BUILD_NUMBER":       "42",
		"DIRTY":              "false",
	}

	for k, v := range expected {
		if result[k] != v {
			t.Errorf("expected %s=%s, got %s=%s", k, v, k, result[k])
		}
	}
}

// TestVariableNames_ReturnsCorrectNamesWithAndWithoutPrefix validates that
// VariableNames returns the standard variable names, optionally prefixed.
//
// Why: Users need to know the exact variable names that will be set for
// use in their CI configurations.
//
// What: VariableNames should return a map from field name to variable name.
func TestVariableNames_ReturnsCorrectNamesWithAndWithoutPrefix(t *testing.T) {
	// Precondition: None (testing the function itself)

	// Action: Get variable names without prefix
	names := VariableNames("")

	// Expected: Standard names are returned
	if names["Version"] != "VERSION" {
		t.Errorf("expected VERSION, got %s", names["Version"])
	}
	if names["GitSHA"] != "GIT_SHA" {
		t.Errorf("expected GIT_SHA, got %s", names["GitSHA"])
	}

	// Action: Get variable names with prefix
	names = VariableNames("APP_")

	// Expected: Prefixed names are returned
	if names["Version"] != "APP_VERSION" {
		t.Errorf("expected APP_VERSION, got %s", names["Version"])
	}
	if names["GitSHA"] != "APP_GIT_SHA" {
		t.Errorf("expected APP_GIT_SHA, got %s", names["GitSHA"])
	}
}

// =============================================================================
// MINUTIAE
// Tests for obscure scenarios, implementation details, and consistency checks.
// =============================================================================

// TestAllFormatters_ProduceSortedOutput validates that all formatters produce
// consistent, alphabetically sorted output regardless of map iteration order.
//
// Why: Go maps have non-deterministic iteration order. Sorted output ensures
// reproducible builds and easier debugging.
//
// What: Running Format multiple times should produce identical output, and
// variables should appear in alphabetical order.
func TestAllFormatters_ProduceSortedOutput(t *testing.T) {
	// Precondition: Variables with names that would reveal sorting issues
	vars := map[string]string{
		"ZEBRA":  "z",
		"ALPHA":  "a",
		"MIDDLE": "m",
		"BETA":   "b",
	}

	formatters := []Formatter{
		&ShellFormatter{},
		&GitHubFormatter{},
		&GitLabFormatter{},
		&AzureFormatter{},
		&CircleCIFormatter{},
		&JenkinsFormatter{},
	}

	for _, formatter := range formatters {
		t.Run(formatter.Name(), func(t *testing.T) {
			// Action: Format the variables multiple times
			output := formatter.Format(vars)
			for i := 0; i < 3; i++ {
				if formatter.Format(vars) != output {
					t.Errorf("output not consistent across calls for %s", formatter.Name())
				}
			}

			// Expected: Variables appear in alphabetical order
			alphaIdx := strings.Index(output, "ALPHA")
			betaIdx := strings.Index(output, "BETA")
			middleIdx := strings.Index(output, "MIDDLE")
			zebraIdx := strings.Index(output, "ZEBRA")

			if alphaIdx >= betaIdx || betaIdx >= middleIdx || middleIdx >= zebraIdx {
				t.Errorf("output not sorted for %s: ALPHA@%d, BETA@%d, MIDDLE@%d, ZEBRA@%d",
					formatter.Name(), alphaIdx, betaIdx, middleIdx, zebraIdx)
			}
		})
	}
}

// TestAvailableFormatters_ContainsAllExpectedFormatters validates that the
// list of available formatters includes all supported CI platforms.
//
// Why: Users rely on this list to discover available formatters. Missing
// entries would prevent users from using certain formatters.
//
// What: AvailableFormatters should return all six supported formatter names.
func TestAvailableFormatters_ContainsAllExpectedFormatters(t *testing.T) {
	// Precondition: The expected list of formatter names
	expected := []string{"shell", "github", "gitlab", "azure", "circleci", "jenkins"}

	// Action: Get the available formatters
	formatters := AvailableFormatters()

	// Expected: The count matches
	if len(formatters) != len(expected) {
		t.Errorf("expected %d formatters, got %d", len(expected), len(formatters))
	}

	// Expected: All expected formatters are present
	for _, name := range expected {
		found := false
		for _, f := range formatters {
			if f == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected formatter %q not found in AvailableFormatters()", name)
		}
	}
}
