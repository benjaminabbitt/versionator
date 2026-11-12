package version

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
)

// =============================================================================
// CORE FUNCTIONALITY
// ModulePath/ModuleName locate the active VERSION file within the repository so
// release templates can reproduce Go's submodule tag format (dir/vX.Y.Z).
// =============================================================================

// registerMockRepo registers a mock VCS rooted at root and returns a cleanup func.
func registerMockRepo(t *testing.T, root string) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock-git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(root, nil).AnyTimes()
	vcs.RegisterVCS(mockVCS)
	t.Cleanup(func() {
		vcs.UnregisterVCS("mock-git")
		ctrl.Finish()
	})
}

// Validates that a VERSION file at the repository root contributes no path
// segment. A root module's Go tag is plain "v1.2.3", so ModulePath must be
// empty rather than "." or "/".
func TestModulePath_VersionFileAtRepoRoot_ReturnsEmpty(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION file lives at the repository root
	if err := os.WriteFile(versionFile, []byte("1.2.3"), 0644); err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}
	registerMockRepo(t, tempDir)

	// Action: Derive the module path
	got, err := ModulePath()

	// Expected: No path segment
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if got != "" {
		t.Errorf("Expected empty module path at repo root, got %q", got)
	}
}

// Validates that a VERSION file in a subdirectory yields the repo-relative
// directory with a trailing slash.
//
// Why: Go requires a submodule tag to be the module directory followed by the
// version ("internal/tools/v1.2.3"). The trailing slash lets one template,
// "{{ModulePath}}{{Prefix}}{{MajorMinorPatch}}", render correctly both here and
// at the repository root, where ModulePath contributes nothing.
func TestModulePath_VersionFileInSubdirectory_ReturnsSlashTerminatedPath(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	// Precondition: VERSION file lives two directories below the repo root
	moduleDir := filepath.Join(tempDir, "internal", "tools")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, versionFile), []byte("1.2.3"), 0644); err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}
	_ = os.Chdir(moduleDir)
	registerMockRepo(t, tempDir)

	// Action: Derive the module path
	got, err := ModulePath()

	// Expected: Slash-separated, repo-relative, trailing slash
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if got != "internal/tools/" {
		t.Errorf("Expected 'internal/tools/', got %q", got)
	}
}

// Validates that ModuleName returns the bare directory name, for component-style
// tag conventions such as "tools-v1.2.3" used outside the Go ecosystem.
func TestModuleName_VersionFileInSubdirectory_ReturnsBaseName(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	// Precondition: VERSION file lives in internal/tools
	moduleDir := filepath.Join(tempDir, "internal", "tools")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, versionFile), []byte("1.2.3"), 0644); err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}
	_ = os.Chdir(moduleDir)
	registerMockRepo(t, tempDir)

	// Action: Derive the module name
	got, err := ModuleName()

	// Expected: Base directory name only
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if got != "tools" {
		t.Errorf("Expected 'tools', got %q", got)
	}
}

// =============================================================================
// EDGE CASES
// =============================================================================

// Validates that ModulePath is empty outside a repository. Without a repo root
// there is nothing to be relative to, and a tag template must still render.
func TestModulePath_NoRepository_ReturnsEmpty(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: No VCS registered for this directory
	vcs.UnregisterVCS("git")
	defer func() { vcs.RegisterVCS(gitVCS.NewGitVCSDefault()) }()

	if err := os.WriteFile(versionFile, []byte("1.2.3"), 0644); err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Derive the module path
	got, err := ModulePath()

	// Expected: Empty, no error
	if err != nil {
		t.Fatalf("Expected no error outside a repository, got: %v", err)
	}
	if got != "" {
		t.Errorf("Expected empty module path outside a repository, got %q", got)
	}
}
