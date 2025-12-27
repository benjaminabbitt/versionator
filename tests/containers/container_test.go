package containers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

const imagePrefix = "versionator-test"

// testContext holds state between step definitions
type testContext struct {
	containerName   string
	containerOutput string
	exitCode        int
	projectRoot     string
}

// Singleton for current test context
var ctx *testContext

func TestContainerFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping container tests in short mode")
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     "@slow", // Container tests are all slow
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run container feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	// Before each scenario
	sc.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		return setupTestContext(c)
	})

	// Step definitions
	sc.Step(`^I run the "([^"]*)" container test$`, iRunTheContainerTest)
	sc.Step(`^the container should exit successfully$`, theContainerShouldExitSuccessfully)
	sc.Step(`^the container should exit with code (\d+)$`, theContainerShouldExitWithCode)
	sc.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)
	sc.Step(`^the output should not contain "([^"]*)"$`, theOutputShouldNotContain)
}

func setupTestContext(c context.Context) (context.Context, error) {
	ctx = &testContext{}

	// Find project root
	root, err := findProjectRoot()
	if err != nil {
		return c, fmt.Errorf("failed to find project root: %w", err)
	}
	ctx.projectRoot = root

	// Ensure versionator-builder image exists
	if err := ensureBuilderImage(); err != nil {
		return c, fmt.Errorf("failed to ensure builder image: %w", err)
	}

	return c, nil
}

func findProjectRoot() (string, error) {
	// Try environment variable first
	if root := os.Getenv("VERSIONATOR_PROJECT_ROOT"); root != "" {
		return root, nil
	}

	// Try to find by looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (go.mod)")
}

func ensureBuilderImage() error {
	// Check if image exists
	cmd := exec.Command("docker", "images", "-q", "versionator-builder:latest")
	out, err := cmd.Output()
	if err == nil && len(bytes.TrimSpace(out)) > 0 {
		return nil // Image exists
	}

	// Build the builder image
	fmt.Println("Building versionator-builder image...")
	cmd = exec.Command("docker", "build",
		"-t", "versionator-builder:latest",
		"-f", filepath.Join(ctx.projectRoot, "tests/containers/images/versionator-builder.Dockerfile"),
		ctx.projectRoot)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func iRunTheContainerTest(containerName string) error {
	ctx.containerName = containerName
	imageName := fmt.Sprintf("%s-%s:latest", imagePrefix, containerName)
	dockerfilePath := filepath.Join(ctx.projectRoot, "tests/containers/images", containerName+".Dockerfile")

	// Build the container
	fmt.Printf("Building %s...\n", imageName)
	buildCmd := exec.Command("docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		ctx.projectRoot)

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build container %s: %w\nOutput:\n%s", containerName, err, buildOutput.String())
	}

	// Run the container
	fmt.Printf("Running %s...\n", imageName)
	runCmd := exec.Command("docker", "run", "--rm", imageName)

	var runOutput bytes.Buffer
	runCmd.Stdout = &runOutput
	runCmd.Stderr = &runOutput

	err := runCmd.Run()
	ctx.containerOutput = runOutput.String()

	if exitErr, ok := err.(*exec.ExitError); ok {
		ctx.exitCode = exitErr.ExitCode()
	} else if err != nil {
		ctx.exitCode = 1
	} else {
		ctx.exitCode = 0
	}

	return nil
}

func theContainerShouldExitSuccessfully() error {
	if ctx.exitCode != 0 {
		return fmt.Errorf("expected container to exit with code 0, got %d\nOutput:\n%s",
			ctx.exitCode, ctx.containerOutput)
	}
	return nil
}

func theContainerShouldExitWithCode(expected int) error {
	if ctx.exitCode != expected {
		return fmt.Errorf("expected container to exit with code %d, got %d\nOutput:\n%s",
			expected, ctx.exitCode, ctx.containerOutput)
	}
	return nil
}

func theOutputShouldContain(substring string) error {
	if !strings.Contains(ctx.containerOutput, substring) {
		return fmt.Errorf("expected output to contain %q\nActual output:\n%s",
			substring, ctx.containerOutput)
	}
	return nil
}

func theOutputShouldNotContain(substring string) error {
	if strings.Contains(ctx.containerOutput, substring) {
		return fmt.Errorf("expected output NOT to contain %q\nActual output:\n%s",
			substring, ctx.containerOutput)
	}
	return nil
}
