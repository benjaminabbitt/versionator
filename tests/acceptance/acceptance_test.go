package acceptance

import (
	"bytes"
	"context"
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
	sc.Step(`^a VERSION file with version "([^"]*)"$`, aVersionFileWithVersion)
	sc.Step(`^a VERSION file with prefix "([^"]*)" and version "([^"]*)"$`, aVersionFileWithPrefixAndVersion)
	sc.Step(`^a VERSION file with prefix "([^"]*)", version "([^"]*)" and prerelease "([^"]*)"$`, aVersionFileWithPrefixVersionAndPrerelease)
	sc.Step(`^a VERSION file with prefix "([^"]*)", version "([^"]*)" and metadata "([^"]*)"$`, aVersionFileWithPrefixVersionAndMetadata)
	sc.Step(`^a VERSION file with version "([^"]*)" and custom variable "([^"]*)" set to "([^"]*)"$`, aVersionFileWithCustomVariable)
	sc.Step(`^a committed file "([^"]*)" with content "([^"]*)"$`, aCommittedFileWithContent)
	sc.Step(`^a file "([^"]*)" with content "([^"]*)"$`, aFileWithContent)
	sc.Step(`^a template file "([^"]*)" with content "([^"]*)"$`, aFileWithContent) // Same implementation
	sc.Step(`^a config file with prerelease enabled and template "([^"]*)"$`, aConfigFileWithPrereleaseTemplate)
	sc.Step(`^a config file with:$`, aConfigFileWithDocString)

	// Action steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^I commit a file "([^"]*)" with content "([^"]*)"$`, iCommitAFileWithContent)
	sc.Step(`^I commit the VERSION changes$`, iCommitTheVersionChanges)
	sc.Step(`^I create (\d+) commits with message prefix "([^"]*)"$`, iCreateCommitsWithMessagePrefix)

	// Assertion steps
	sc.Step(`^the output should be "([^"]*)"$`, theOutputShouldBe)
	sc.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)
	sc.Step(`^the output should contain '([^']*)'$`, theOutputShouldContain)  // Single-quoted variant
	sc.Step(`^the output should contain ""([^"]*)""$`, theOutputShouldContain) // Double-quoted variant (for values with embedded quotes)
	sc.Step(`^the output should match pattern "([^"]*)"$`, theOutputShouldMatchPattern)
	sc.Step(`^the exit code should be (\d+)$`, theExitCodeShouldBe)
	sc.Step(`^the exit code should not be (\d+)$`, theExitCodeShouldNotBe)
	sc.Step(`^a git tag "([^"]*)" should exist$`, aGitTagShouldExist)
	sc.Step(`^the tag "([^"]*)" should point to HEAD$`, theTagShouldPointToHEAD)
	sc.Step(`^the tag "([^"]*)" should have message "([^"]*)"$`, theTagShouldHaveMessage)
	sc.Step(`^the tag "([^"]*)" should be (\d+) commits ahead of "([^"]*)"$`, theTagShouldBeCommitsAheadOf)
	sc.Step(`^the VERSION should have version "([^"]*)"$`, theVersionShouldHaveVersion)
	sc.Step(`^the VERSION should have prefix "([^"]*)"$`, theVersionShouldHavePrefix)
	sc.Step(`^the VERSION should have prerelease "([^"]*)"$`, theVersionShouldHavePrerelease)
	sc.Step(`^the VERSION should have metadata "([^"]*)"$`, theVersionShouldHaveMetadata)
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

	// Find versionator binary BEFORE changing to temp directory
	ctx.versionator = findVersionatorBinary()

	// Create temp directory
	ctx.workDir, err = os.MkdirTemp("", "versionator-test-*")
	if err != nil {
		return c, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Change to temp directory
	if err := os.Chdir(ctx.workDir); err != nil {
		return c, fmt.Errorf("failed to change to temp directory: %w", err)
	}

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
	// First try the project's built binary from env var
	if root := os.Getenv("VERSIONATOR_PROJECT_ROOT"); root != "" {
		projectBinary := filepath.Join(root, "versionator")
		if _, err := os.Stat(projectBinary); err == nil {
			return projectBinary
		}
	}

	// Try to find project root by going up from current directory
	// This works when running tests from the project directory
	if wd, err := os.Getwd(); err == nil {
		// Look for versionator binary in parent directories
		dir := wd
		for i := 0; i < 5; i++ { // Limit depth
			binary := filepath.Join(dir, "versionator")
			if _, err := os.Stat(binary); err == nil {
				return binary
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
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

func aVersionFileWithVersion(version string) error {
	if err := writeVersion("", version, "", ""); err != nil {
		return err
	}
	// Commit VERSION to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION")
}

func aVersionFileWithPrefixAndVersion(prefix, version string) error {
	if err := writeVersion(prefix, version, "", ""); err != nil {
		return err
	}
	// Commit VERSION to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION")
}

func aVersionFileWithPrefixVersionAndPrerelease(prefix, version, prerelease string) error {
	if err := writeVersion(prefix, version, prerelease, ""); err != nil {
		return err
	}
	// Commit VERSION to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION")
}

func aVersionFileWithPrefixVersionAndMetadata(prefix, version, metadata string) error {
	if err := writeVersion(prefix, version, "", metadata); err != nil {
		return err
	}
	// Commit VERSION to ensure clean working directory for commit commands
	if err := runCommand("git", "add", "VERSION"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION")
}

func aVersionFileWithCustomVariable(version, key, value string) error {
	// Custom variables are now stored in config file, not VERSION file
	// Write the VERSION file
	if err := writeVersion("", version, "", ""); err != nil {
		return err
	}

	// Write custom variable to config file
	configContent := fmt.Sprintf(`custom:
  %s: "%s"
`, key, value)
	if err := os.WriteFile(".versionator.yaml", []byte(configContent), 0644); err != nil {
		return err
	}

	// Commit VERSION to ensure clean working directory
	if err := runCommand("git", "add", "VERSION"); err != nil {
		return err
	}
	if err := runCommand("git", "add", ".versionator.yaml"); err != nil {
		return err
	}
	return runCommand("git", "commit", "-m", "Add VERSION and config")
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

func iCommitTheVersionChanges() error {
	// Add VERSION to staging
	if err := runCommand("git", "add", "VERSION"); err != nil {
		return err
	}
	// Try to commit - allow failure if nothing to commit
	cmd := exec.Command("git", "commit", "-m", "Update VERSION")
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

func theVersionShouldHaveVersion(expected string) error {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return fmt.Errorf("failed to read VERSION: %w", err)
	}

	versionStr := strings.TrimSpace(string(data))

	// Find first digit - everything after is the version (including pre-release/metadata)
	stripped := versionStr
	for i, c := range versionStr {
		if c >= '0' && c <= '9' {
			stripped = versionStr[i:]
			break
		}
	}

	// Extract just the core version part (before any - or +)
	versionPart := stripped
	if idx := strings.Index(versionPart, "-"); idx != -1 {
		versionPart = versionPart[:idx]
	}
	if idx := strings.Index(versionPart, "+"); idx != -1 {
		versionPart = versionPart[:idx]
	}

	if versionPart != expected {
		return fmt.Errorf("expected version %q, got %q (from VERSION: %q)", expected, versionPart, versionStr)
	}
	return nil
}

func theVersionShouldHavePrefix(expected string) error {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return fmt.Errorf("failed to read VERSION: %w", err)
	}

	versionStr := strings.TrimSpace(string(data))

	// Extract prefix as everything before the first digit
	prefix := ""
	for i, c := range versionStr {
		if c >= '0' && c <= '9' {
			prefix = versionStr[:i]
			break
		}
	}

	if prefix != expected {
		return fmt.Errorf("expected prefix %q, got %q (from VERSION: %q)", expected, prefix, versionStr)
	}
	return nil
}

func theVersionShouldHavePrerelease(expected string) error {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return fmt.Errorf("failed to read VERSION: %w", err)
	}

	versionStr := strings.TrimSpace(string(data))

	// Extract prerelease (after - but before +)
	prerelease := ""
	if idx := strings.Index(versionStr, "-"); idx != -1 {
		rest := versionStr[idx+1:]
		if plusIdx := strings.Index(rest, "+"); plusIdx != -1 {
			prerelease = rest[:plusIdx]
		} else {
			prerelease = rest
		}
	}

	if prerelease != expected {
		return fmt.Errorf("expected prerelease %q, got %q (from VERSION: %q)", expected, prerelease, versionStr)
	}
	return nil
}

func theVersionShouldHaveMetadata(expected string) error {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return fmt.Errorf("failed to read VERSION: %w", err)
	}

	versionStr := strings.TrimSpace(string(data))

	// Extract metadata (after +)
	metadata := ""
	if idx := strings.Index(versionStr, "+"); idx != -1 {
		metadata = versionStr[idx+1:]
	}

	if metadata != expected {
		return fmt.Errorf("expected metadata %q, got %q (from VERSION: %q)", expected, metadata, versionStr)
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

func writeVersion(prefix, version, prerelease, metadata string) error {
	// Build full version string: [prefix]major.minor.patch[-prerelease][+metadata]
	versionStr := prefix + version
	if prerelease != "" {
		versionStr += "-" + prerelease
	}
	if metadata != "" {
		versionStr += "+" + metadata
	}

	return os.WriteFile("VERSION", []byte(versionStr+"\n"), 0644)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = ctx.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command %q failed: %w\nOutput: %s", name, err, output)
	}
	return nil
}
