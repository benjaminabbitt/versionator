package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
)

// =============================================================================
// ONE TEMPLATE DATA
// The file updates of every version change, the pre-save check, and release's
// tag, branch, message and file updates render with one builder, so the check
// validates exactly what release will write: the release-only variables
// ({{ModulePath}}, {{ModuleName}}, {{BranchPrefix}}) and the effective
// {{Prefix}} (the VERSION file's, else "v", as release tags) included.
// =============================================================================

// newModuleRepo makes a git repository with a module in tools/, holding its
// own VERSION and .versionator.yaml, and changes into it.
func newModuleRepo(t *testing.T, version, configYAML string) {
	t.Helper()
	root := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	module := filepath.Join(root, "tools")
	_ = os.MkdirAll(module, 0755)
	_ = os.WriteFile(filepath.Join(module, "VERSION"), []byte(version+"\n"), 0644)
	_ = os.WriteFile(filepath.Join(module, ".versionator.yaml"), []byte(configYAML), 0644)
	c := exec.Command("git", "init", "-q")
	c.Dir = root
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	if err := os.Chdir(module); err != nil {
		t.Fatalf("chdir: %v", err)
	}
}

func TestTemplateData_UpdatesSeeReleaseVariables(t *testing.T) {
	newModuleRepo(t, "1.0.0", `prefix: ""
updates:
  - file: info.json
    path: v
    template: "{{ModulePath}}|{{ModuleName}}|{{BranchPrefix}}|{{Prefix}}{{MajorMinorPatch}}"
`)
	_ = os.WriteFile("info.json", []byte(`{"v": ""}`), 0644)

	if err := runVersionator(t, "set", "1.2.3"); err != nil {
		t.Fatalf("set: %v", err)
	}

	b, _ := os.ReadFile("info.json")
	if string(b) != `{"v": "tools/|tools|release/|v1.2.3"}` {
		t.Errorf("got %s", b)
	}
}

// The check renders with the same data: {{ModulePath}} makes this Cargo
// version "tools/1.2.3", which Cargo refuses, so set is refused before VERSION
// is written.
func TestTemplateData_CheckSeesReleaseVariables(t *testing.T) {
	newModuleRepo(t, "1.0.0", `prefix: ""
updates:
  - file: Cargo.toml
    path: package.version
    template: "{{ModulePath}}{{MajorMinorPatch}}"
`)
	_ = os.WriteFile("Cargo.toml", []byte("[package]\nversion = \"1.0.0\"\n"), 0644)

	if err := runVersionator(t, "set", "1.2.3"); err == nil {
		t.Fatal("expected the check to refuse tools/1.2.3")
	}
	assertVersionFile(t, "1.0.0\n")
}
