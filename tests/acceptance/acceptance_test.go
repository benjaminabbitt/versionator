package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// testContext holds state between step definitions
type testContext struct {
	workDir      string
	output       string
	exitCode     int
	versionator  string // path to versionator binary
	originalDir  string
}

// Singleton for current test context
var ctx *testContext

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     "~@slow", // Skip slow tests by default
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestSlowFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow tests in short mode")
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     "@slow",
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run slow feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	// Before each scenario
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return setupTestContext(ctx)
	})

	// After each scenario
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return teardownTestContext(ctx)
	})

	// Background steps
	sc.Step(`^a clean git repository$`, aCleanGitRepository)
	sc.Step(`^versionator is installed$`, versionatorIsInstalled)
	sc.Step(`^a VERSION\.json file with version "([^"]*)"$`, aVersionJSONFileWithVersion)
	sc.Step(`^a VERSION\.json file with prefix "([^"]*)" and version "([^"]*)"$`, aVersionJSONFileWithPrefixAndVersion)
	sc.Step(`^a VERSION\.json file with prefix "([^"]*)", version "([^"]*)" and prerelease "([^"]*)"$`, aVersionJSONFileWithPrefixVersionAndPrerelease)
	sc.Step(`^a VERSION\.json file with prefix "([^"]*)", version "([^"]*)" and metadata "([^"]*)"$`, aVersionJSONFileWithPrefixVersionAndMetadata)
	sc.Step(`^a VERSION\.json file with version "([^"]*)" and custom variable "([^"]*)" set to "([^"]*)"$`, aVersionJSONFileWithCustomVariable)
	sc.Step(`^a committed file "([^"]*)" with content "([^"]*)"$`, aCommittedFileWithContent)
	sc.Step(`^a file "([^"]*)" with content "([^"]*)"$`, aFileWithContent)
	sc.Step(`^a template file "([^"]*)" with content "([^"]*)"$`, aFileWithContent) // Same implementation
	sc.Step(`^a config file with prerelease enabled and template "([^"]*)"$`, aConfigFileWithPrereleaseTemplate)
	sc.Step(`^a config file with:$`, aConfigFileWithDocString)

	// Action steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^I commit a file "([^"]*)" with content "([^"]*)"$`, iCommitAFileWithContent)
	sc.Step(`^I commit the VERSION\.json changes$`, iCommitTheVersionJSONChanges)
	sc.Step(`^I create (\d+) commits with message prefix "([^"]*)"$`, iCreateCommitsWithMessagePrefix)

	// Assertion steps
	sc.Step(`^the output should be "([^"]*)"$`, theOutputShouldBe)
	sc.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)
	sc.Step(`^the output should contain '([^']*)'$`, theOutputShouldContain) // Single-quoted variant
	sc.Step(`^the output should match pattern "([^"]*)"$`, theOutputShouldMatchPattern)
	sc.Step(`^the exit code should be (\d+)$`, theExitCodeShouldBe)
	sc.Step(`^the exit code should not be (\d+)$`, theExitCodeShouldNotBe)
	sc.Step(`^a git tag "([^"]*)" should exist$`, aGitTagShouldExist)
	sc.Step(`^the tag "([^"]*)" should point to HEAD$`, theTagShouldPointToHEAD)
	sc.Step(`^the tag "([^"]*)" should have message "([^"]*)"$`, theTagShouldHaveMessage)
	sc.Step(`^the tag "([^"]*)" should be (\d+) commits ahead of "([^"]*)"$`, theTagShouldBeCommitsAheadOf)
	sc.Step(`^the VERSION\.json should have version "([^"]*)"$`, theVersionJSONShouldHaveVersion)
	sc.Step(`^the VERSION\.json should have prefix "([^"]*)"$`, theVersionJSONShouldHavePrefix)
	sc.Step(`^the file "([^"]*)" should exist$`, theFileShouldExist)
	sc.Step(`^the file "([^"]*)" should contain "([^"]*)"$`, theFileShouldContain)
	sc.Step(`^the file "([^"]*)" should contain '([^']*)'$`, theFileShouldContain) // Single-quoted variant
}

func setupTestContext(c context.Context) (context.Context, error) {
	ctx = &testContext{}

	// Save original directory
	var err error
	ctx.originalDir, err = os.Getwd()
	if err != nil {
		return c, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create temp directory
	ctx.workDir, err = os.MkdirTemp("", "versionator-test-*")
	if err != nil {
		return c, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Change to temp directory
	if err := os.Chdir(ctx.workDir); err != nil {
		return c, fmt.Errorf("failed to change to temp directory: %w", err)
	}

	// Find versionator binary - look in project root first
	ctx.versionator = findVersionatorBinary()

	return c, nil
}

func teardownTestContext(c context.Context) (context.Context, error) {
	if ctx == nil {
		return c, nil
	}

	// Change back to original directory
	if ctx.originalDir != "" {
		os.Chdir(ctx.originalDir)
	}

	// Remove temp directory
	if ctx.workDir != "" {
		os.RemoveAll(ctx.workDir)
	}

	ctx = nil
	return c, nil
}

func findVersionatorBinary() string {
	// First try the project's built binary
	projectBinary := filepath.Join(os.Getenv("VERSIONATOR_PROJECT_ROOT"), "versionator")
	if _, err := os.Stat(projectBinary); err == nil {
		return projectBinary
	}

	// Try go install location
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	goBinary := filepath.Join(gopath, "bin", "versionator")
	if _, err := os.Stat(goBinary); err == nil {
		return goBinary
	}

	// Fall back to PATH
	if path, err := exec.LookPath("versionator"); err == nil {
		return path
	}

	// Last resort - use "go run" with the module
	return "go run github.com/benjaminabbitt/versionator"
}

// Background steps

func aCleanGitRepository() error {
	// Initialize git repository
	if err := runCommand("git", "init"); err != nil {
		return err
	}
	if err := runCommand("git", "config", "user.email", "test@example.com"); err != nil {
		return err
	}
	if err := runCommand("git", "config", "user.name", "Test User"); err != nil {
		return err
	}
	return nil
}

func versionatorIsInstalled() error {
	if ctx.versionator == "" {
		return fmt.Errorf("versionator binary not found")
	}
	return nil
}

func aVersionJSONFileWithVersion(version string) error {
	if err := writeVersionJSON("", version, "", ""); err != nil {
		return err
	}
	// Commit VERSION.json to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION.json"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION.json")
}

func aVersionJSONFileWithPrefixAndVersion(prefix, version string) error {
	if err := writeVersionJSON(prefix, version, "", ""); err != nil {
		return err
	}
	// Commit VERSION.json to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION.json"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION.json")
}

func aVersionJSONFileWithPrefixVersionAndPrerelease(prefix, version, prerelease string) error {
	if err := writeVersionJSON(prefix, version, prerelease, ""); err != nil {
		return err
	}
	// Commit VERSION.json to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION.json"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION.json")
}

func aVersionJSONFileWithPrefixVersionAndMetadata(prefix, version, metadata string) error {
	if err := writeVersionJSON(prefix, version, "", metadata); err != nil {
		return err
	}
	// Commit VERSION.json to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION.json"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION.json")
}

func aVersionJSONFileWithCustomVariable(version, key, value string) error {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	data := map[string]interface{}{
		"major": major,
		"minor": minor,
		"patch": patch,
		"custom": map[string]string{
			key: value,
		},
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile("VERSION.json", content, 0644); err != nil {
		return err
	}
	// Commit VERSION.json to ensure clean working directory
	if err := runCommand("git", "add", "VERSION.json"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION.json")
}

func aCommittedFileWithContent(filename, content string) error {
	if err := aFileWithContent(filename, content); err != nil {
		return err
	}
	if err := runCommand("git", "add", filename); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", fmt.Sprintf("Add %s", filename))
}

func aFileWithContent(filename, content string) error {
	// Create parent directories if needed
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(filename, []byte(content), 0644)
}

func aConfigFileWithPrereleaseTemplate(template string) error {
	config := fmt.Sprintf(`prerelease:
  enabled: true
  template: "%s"
`, template)
	return os.WriteFile(".versionator.yaml", []byte(config), 0644)
}

func aConfigFileWithDocString(doc *godog.DocString) error {
	return os.WriteFile(".versionator.yaml", []byte(doc.Content), 0644)
}

// Action steps

func iRun(command string) error {
	// Parse command with proper quote handling
	parts, err := parseCommand(command)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Replace versionator with actual binary path
	if parts[0] == "versionator" {
		if strings.HasPrefix(ctx.versionator, "go run") {
			// Using go run
			parts = append(strings.Fields(ctx.versionator), parts[1:]...)
		} else {
			parts[0] = ctx.versionator
		}
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = ctx.workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	ctx.output = strings.TrimSpace(stdout.String())
	if ctx.output == "" {
		ctx.output = strings.TrimSpace(stderr.String())
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		ctx.exitCode = exitErr.ExitCode()
	} else if err != nil {
		ctx.exitCode = 1
	} else {
		ctx.exitCode = 0
	}

	return nil // Don't return error for non-zero exit - let assertions check it
}

// parseCommand parses a shell-like command string, handling quoted arguments
func parseCommand(command string) ([]string, error) {
	var parts []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(command); i++ {
		c := command[i]

		switch {
		case c == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
		case c == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
		case c == ' ' && !inSingleQuote && !inDoubleQuote:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unclosed quote in command: %s", command)
	}

	return parts, nil
}

func iCommitAFileWithContent(filename, content string) error {
	if err := aFileWithContent(filename, content); err != nil {
		return err
	}
	if err := runCommand("git", "add", filename); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", fmt.Sprintf("Update %s", filename))
}

func iCommitTheVersionJSONChanges() error {
	// Add VERSION.json to staging
	if err := runCommand("git", "add", "VERSION.json"); err != nil {
		return err
	}
	// Try to commit - allow failure if nothing to commit
	cmd := exec.Command("git", "commit", "-m", "Update VERSION.json")
	cmd.Dir = ctx.workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's just "nothing to commit"
		if strings.Contains(string(output), "nothing to commit") {
			return nil // This is OK - no changes needed
		}
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, output)
	}
	return nil
}

func iCreateCommitsWithMessagePrefix(count int, prefix string) error {
	for i := 1; i <= count; i++ {
		filename := fmt.Sprintf("file_%d.txt", i)
		if err := os.WriteFile(filename, []byte(fmt.Sprintf("Content %d", i)), 0644); err != nil {
			return err
		}
		if err := runCommand("git", "add", filename); err != nil {
			return err
		}
		if err := runCommand("git", "commit", "-m", fmt.Sprintf("%s commit %d", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}

// Assertion steps

func theOutputShouldBe(expected string) error {
	if ctx.output != expected {
		return fmt.Errorf("expected output %q, got %q", expected, ctx.output)
	}
	return nil
}

func theOutputShouldContain(substring string) error {
	if !strings.Contains(ctx.output, substring) {
		return fmt.Errorf("expected output to contain %q, got %q", substring, ctx.output)
	}
	return nil
}

func theOutputShouldMatchPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	if !re.MatchString(ctx.output) {
		return fmt.Errorf("expected output to match pattern %q, got %q", pattern, ctx.output)
	}
	return nil
}

func theExitCodeShouldBe(expected int) error {
	if ctx.exitCode != expected {
		return fmt.Errorf("expected exit code %d, got %d (output: %s)", expected, ctx.exitCode, ctx.output)
	}
	return nil
}

func theExitCodeShouldNotBe(notExpected int) error {
	if ctx.exitCode == notExpected {
		return fmt.Errorf("expected exit code not to be %d", notExpected)
	}
	return nil
}

func aGitTagShouldExist(tag string) error {
	cmd := exec.Command("git", "tag", "-l", tag)
	cmd.Dir = ctx.workDir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}
	if strings.TrimSpace(string(out)) != tag {
		return fmt.Errorf("tag %q does not exist", tag)
	}
	return nil
}

func theTagShouldPointToHEAD(tag string) error {
	tagCmd := exec.Command("git", "rev-parse", tag+"^{}")
	tagCmd.Dir = ctx.workDir
	tagCommit, err := tagCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get tag commit: %w", err)
	}
	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = ctx.workDir
	headCommit, err := headCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get HEAD commit: %w", err)
	}
	if strings.TrimSpace(string(tagCommit)) != strings.TrimSpace(string(headCommit)) {
		return fmt.Errorf("tag %q does not point to HEAD", tag)
	}
	return nil
}

func theTagShouldHaveMessage(tag, message string) error {
	cmd := exec.Command("git", "tag", "-l", "-n", tag)
	cmd.Dir = ctx.workDir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get tag message: %w", err)
	}
	if !strings.Contains(string(out), message) {
		return fmt.Errorf("tag %q message does not contain %q, got %q", tag, message, string(out))
	}
	return nil
}

func theTagShouldBeCommitsAheadOf(tag string, count int, baseTag string) error {
	cmd := exec.Command("git", "rev-list", "--count", baseTag+".."+tag)
	cmd.Dir = ctx.workDir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to count commits: %w", err)
	}
	actual, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return fmt.Errorf("failed to parse commit count: %w", err)
	}
	if actual != count {
		return fmt.Errorf("expected %d commits between %s and %s, got %d", count, baseTag, tag, actual)
	}
	return nil
}

func theVersionJSONShouldHaveVersion(expected string) error {
	data, err := os.ReadFile("VERSION.json")
	if err != nil {
		return fmt.Errorf("failed to read VERSION.json: %w", err)
	}

	var v struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
		Patch int `json:"patch"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to parse VERSION.json: %w", err)
	}

	actual := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if actual != expected {
		return fmt.Errorf("expected version %q, got %q", expected, actual)
	}
	return nil
}

func theVersionJSONShouldHavePrefix(expected string) error {
	data, err := os.ReadFile("VERSION.json")
	if err != nil {
		return fmt.Errorf("failed to read VERSION.json: %w", err)
	}

	var v struct {
		Prefix string `json:"prefix"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to parse VERSION.json: %w", err)
	}

	if v.Prefix != expected {
		return fmt.Errorf("expected prefix %q, got %q", expected, v.Prefix)
	}
	return nil
}

func theFileShouldExist(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %q does not exist", filename)
	}
	return nil
}

func theFileShouldContain(filename, substring string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filename, err)
	}
	if !strings.Contains(string(data), substring) {
		return fmt.Errorf("file %q does not contain %q", filename, substring)
	}
	return nil
}

// Helper functions

func writeVersionJSON(prefix, version, prerelease, metadata string) error {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	data := map[string]interface{}{
		"major": major,
		"minor": minor,
		"patch": patch,
	}

	if prefix != "" {
		data["prefix"] = prefix
	}
	if prerelease != "" {
		data["prerelease"] = prerelease
	}
	if metadata != "" {
		data["metadata"] = metadata
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("VERSION.json", content, 0644)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = ctx.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command %q failed: %w\nOutput: %s", name, err, output)
	}
	return nil
}
