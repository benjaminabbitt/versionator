package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
)

// =============================================================================
// TRACKED-ONLY DIRTINESS (dirty.ignoreUntracked / --ignore-untracked)
// Every dirty-derived template variable ({{Dirty}}, {{UncommittedChanges}},
// {{DateTimeDirty}}, and the ci DIRTY variable) reads one count, so all of them
// follow the setting together.
// =============================================================================

const dirtyProbe = "[{{UncommittedChanges}}|{{Dirty}}|{{DateTimeDirty}}]"

// newUntrackedOnlyRepo makes a git repository whose only change is an
// untracked file, with the given .versionator.yaml committed (or none).
func newUntrackedOnlyRepo(t *testing.T, configYAML string) {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	_ = os.WriteFile("VERSION", []byte("1.2.3\n"), 0644)
	files := []string{"VERSION"}
	if configYAML != "" {
		_ = os.WriteFile(".versionator.yaml", []byte(configYAML), 0644)
		files = append(files, ".versionator.yaml")
	}
	run := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-q")
	run("add", files[0])
	if len(files) > 1 {
		run("add", files[1])
	}
	run("-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-q", "-m", "seed")
	_ = os.WriteFile("scratch.txt", []byte("untracked\n"), 0644)
}

// runDirtyProbe renders the dirty variables through the real command line.
func runDirtyProbe(t *testing.T, extraArgs ...string) string {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	args := append(extraArgs, "output", "version", "-t", dirtyProbe)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
		// cobra keeps flag values between Execute calls in one process.
		if f := rootCmd.PersistentFlags().Lookup("ignore-untracked"); f != nil {
			_ = f.Value.Set("false")
			f.Changed = false
		}
		vcs.SetIgnoreUntracked(false)
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("output version: %v", err)
	}
	return buf.String()
}

func TestDirtyVariables_DefaultUntrackedOnly_AreDirty(t *testing.T) {
	newUntrackedOnlyRepo(t, "prefix: \"\"\n")

	got := runDirtyProbe(t)

	// UncommittedChanges=1, Dirty=dirty, DateTimeDirty=.<timestamp>
	if !bytes.HasPrefix([]byte(got), []byte("[1|dirty|.")) {
		t.Errorf("default mode: expected the untracked file to make the tree dirty, got %q", got)
	}
}

func TestDirtyVariables_ConfigIgnoreUntracked_UntrackedOnlyIsClean(t *testing.T) {
	newUntrackedOnlyRepo(t, "prefix: \"\"\ndirty:\n  ignoreUntracked: true\n")

	got := runDirtyProbe(t)

	if got != "[0||]\n" {
		t.Errorf("dirty.ignoreUntracked: expected every dirty variable clean, got %q", got)
	}
}

func TestDirtyVariables_FlagIgnoreUntracked_UntrackedOnlyIsClean(t *testing.T) {
	newUntrackedOnlyRepo(t, "prefix: \"\"\n")

	got := runDirtyProbe(t, "--ignore-untracked")

	if got != "[0||]\n" {
		t.Errorf("--ignore-untracked: expected every dirty variable clean, got %q", got)
	}
}

func TestDirtyVariables_FlagFalseOverridesConfig_UntrackedCounts(t *testing.T) {
	newUntrackedOnlyRepo(t, "prefix: \"\"\ndirty:\n  ignoreUntracked: true\n")

	got := runDirtyProbe(t, "--ignore-untracked=false")

	if !bytes.HasPrefix([]byte(got), []byte("[1|dirty|.")) {
		t.Errorf("--ignore-untracked=false must override the config, got %q", got)
	}
}

func TestDirtyVariables_ConfigIgnoreUntracked_ModifiedTrackedIsDirty(t *testing.T) {
	newUntrackedOnlyRepo(t, "prefix: \"\"\ndirty:\n  ignoreUntracked: true\n")
	_ = os.WriteFile("VERSION", []byte("1.2.3\n\n"), 0644)

	got := runDirtyProbe(t)

	if !bytes.HasPrefix([]byte(got), []byte("[1|dirty|.")) {
		t.Errorf("dirty.ignoreUntracked: a modified tracked file must still be dirty, got %q", got)
	}
}

func TestCIDirty_ConfigIgnoreUntracked_UntrackedOnlyIsFalse(t *testing.T) {
	newUntrackedOnlyRepo(t, "prefix: \"\"\ndirty:\n  ignoreUntracked: true\n")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	// cobra keeps flag values between Execute calls; name every ci flag.
	rootCmd.SetArgs([]string{"output", "ci", "--format=shell", "--output=", "--prefix="})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
		vcs.SetIgnoreUntracked(false)
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("output ci: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte(`DIRTY="false"`)) {
		t.Errorf("ci DIRTY must follow dirty.ignoreUntracked, got:\n%s", buf.String())
	}
}
