package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// =============================================================================
// TRACKED-ONLY DIRTINESS
// With SetIgnoreUntracked(true), GetUncommittedChanges counts what
// `git describe --dirty` counts: staged or unstaged changes to tracked files.
// Untracked files are left out. The default (false) keeps counting them.
// =============================================================================

// newTrackedRepo makes a repository with one committed file, test.txt, and
// returns a git VCS pointed at it.
func newTrackedRepo(t *testing.T) (*TestHelper, *GitVersionControlSystem) {
	t.Helper()
	h := NewTestHelper(t)
	t.Cleanup(h.Cleanup)
	h.CreateCommit("initial commit")
	v := NewGitVCS(DefaultRepositoryOpener)
	v.repoRoot = h.dir
	return h, v
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func uncommitted(t *testing.T, v *GitVersionControlSystem) int {
	t.Helper()
	n, err := v.GetUncommittedChanges()
	if err != nil {
		t.Fatalf("GetUncommittedChanges: %v", err)
	}
	return n
}

func TestGetUncommittedChanges_DefaultUntrackedOnly_CountsUntracked(t *testing.T) {
	h, v := newTrackedRepo(t)
	writeFile(t, filepath.Join(h.dir, "scratch.txt"), "untracked")

	if got := uncommitted(t, v); got != 1 {
		t.Errorf("default mode: expected the untracked file to count (1), got %d", got)
	}
}

func TestGetUncommittedChanges_IgnoreUntrackedUntrackedOnly_ReturnsZero(t *testing.T) {
	h, v := newTrackedRepo(t)
	v.SetIgnoreUntracked(true)
	writeFile(t, filepath.Join(h.dir, "scratch.txt"), "untracked")
	writeFile(t, filepath.Join(h.dir, "another.log"), "untracked")

	if got := uncommitted(t, v); got != 0 {
		t.Errorf("tracked-only mode: an untracked-only tree must be clean, got %d", got)
	}
}

func TestGetUncommittedChanges_IgnoreUntrackedModifiedTracked_Counts(t *testing.T) {
	h, v := newTrackedRepo(t)
	v.SetIgnoreUntracked(true)
	writeFile(t, filepath.Join(h.dir, "test.txt"), "modified, unstaged")
	writeFile(t, filepath.Join(h.dir, "scratch.txt"), "untracked")

	if got := uncommitted(t, v); got != 1 {
		t.Errorf("tracked-only mode: the modified tracked file alone must count (1), got %d", got)
	}
}

func TestGetUncommittedChanges_IgnoreUntrackedStagedNewFile_Counts(t *testing.T) {
	h, v := newTrackedRepo(t)
	v.SetIgnoreUntracked(true)
	writeFile(t, filepath.Join(h.dir, "added.txt"), "staged")
	if out, err := exec.Command("git", "-C", h.dir, "add", "added.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, out)
	}

	if got := uncommitted(t, v); got != 1 {
		t.Errorf("tracked-only mode: a staged file must count (1), got %d", got)
	}
}

func TestGetUncommittedChanges_IgnoreUntrackedStagedModification_Counts(t *testing.T) {
	h, v := newTrackedRepo(t)
	v.SetIgnoreUntracked(true)
	writeFile(t, filepath.Join(h.dir, "test.txt"), "modified and staged")
	if out, err := exec.Command("git", "-C", h.dir, "add", "test.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, out)
	}

	if got := uncommitted(t, v); got != 1 {
		t.Errorf("tracked-only mode: a staged modification must count (1), got %d", got)
	}
}

func TestSetIgnoreUntracked_False_RestoresDefault(t *testing.T) {
	h, v := newTrackedRepo(t)
	writeFile(t, filepath.Join(h.dir, "scratch.txt"), "untracked")
	v.SetIgnoreUntracked(true)
	v.SetIgnoreUntracked(false)

	if got := uncommitted(t, v); got != 1 {
		t.Errorf("after SetIgnoreUntracked(false) the untracked file must count again (1), got %d", got)
	}
}

// The release guard decides which dirty files block a release from
// GetDirtyFiles. The dirtiness setting is for the template variables only, so
// GetDirtyFiles keeps listing untracked files in either mode.
func TestGetDirtyFiles_IgnoreUntracked_StillListsUntracked(t *testing.T) {
	h, v := newTrackedRepo(t)
	v.SetIgnoreUntracked(true)
	writeFile(t, filepath.Join(h.dir, "scratch.txt"), "untracked")

	files, err := v.GetDirtyFiles()
	if err != nil {
		t.Fatalf("GetDirtyFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "scratch.txt" {
		t.Errorf("expected [scratch.txt], got %v", files)
	}
}
