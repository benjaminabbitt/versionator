package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Tests using mock repository to verify GitVersionControlSystem behavior
// without needing a real git repository. These tests validate that the VCS
// layer correctly translates between go-git operations and the VCS interface.

// =============================================================================
// CORE FUNCTIONALITY - Happy path tests for primary VCS operations
// =============================================================================

// TestMock_GetVCSIdentifier_Success validates that commit hash truncation works correctly.
//
// Why: The VCS identifier (short hash) is used throughout the system for build metadata
// and display purposes. Incorrect truncation could cause version collisions or display issues.
//
// What: Given a repository with a known HEAD commit, when GetVCSIdentifier is called with
// a length of 7, then it returns the first 7 characters of the commit hash.
func TestMock_GetVCSIdentifier_Success(t *testing.T) {
	// Precondition: Repository has a HEAD pointing to a known commit
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now())

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Request VCS identifier with 7 character length
	hash, err := vcs.GetVCSIdentifier(7)

	// Expected: Returns first 7 characters of commit hash without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "abc123d" {
		t.Errorf("expected 'abc123d', got '%s'", hash)
	}
}

// TestMock_IsWorkingDirectoryClean_Clean validates clean working directory detection.
//
// Why: Clean working directory status is critical for release workflows - releases should
// typically only be created from clean working directories to ensure reproducibility.
//
// What: Given a repository with no uncommitted changes, when IsWorkingDirectoryClean is
// called, then it returns true.
func TestMock_IsWorkingDirectoryClean_Clean(t *testing.T) {
	// Precondition: Real git repo with initial commit and no dirty files
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	v := NewGitVCS(DefaultRepositoryOpener)
	v.repoRoot = h.dir

	// Action: Check if working directory is clean
	clean, err := v.IsWorkingDirectoryClean()

	// Expected: Returns true without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !clean {
		t.Error("expected clean working directory")
	}
}

// TestMock_CreateTag_Success validates successful tag creation.
//
// Why: Tag creation is the primary mechanism for marking releases in version control.
// Failure to create tags correctly could result in lost release markers.
//
// What: Given a repository with a valid HEAD commit, when CreateTag is called with a
// tag name and message, then the tag is created successfully.
func TestMock_CreateTag_Success(t *testing.T) {
	// Precondition: Repository has a valid HEAD commit
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test Author", "test@test.com", time.Now())
	mock.CreateTagRef = plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0"), commitHash)

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Create a new tag
	err := vcs.CreateTag("v1.0.0", "Release 1.0.0")

	// Expected: Tag creation succeeds without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestMock_TagExists_Found validates that existing tags are correctly detected.
//
// Why: Tag existence checks prevent duplicate tag creation and are used to verify
// that version tags exist before operations that depend on them.
//
// What: Given a repository with a tag "v1.0.0", when TagExists is called for "v1.0.0",
// then it returns true.
func TestMock_TagExists_Found(t *testing.T) {
	// Precondition: Repository has tag v1.0.0
	tagHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	mock := NewMockRepository()
	mock.TagRefs = []plumbing.Reference{
		MakeTestTag("v1.0.0", tagHash),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Check if tag exists
	exists, err := vcs.TagExists("v1.0.0")

	// Expected: Returns true without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected tag to exist")
	}
}

// TestMock_CreateBranch_Success validates successful branch creation.
//
// Why: Branch creation is used for release branch workflows where releases are
// prepared on dedicated branches before merging.
//
// What: Given a repository with a valid HEAD, when CreateBranch is called, then
// a new branch reference is created pointing to HEAD.
func TestMock_CreateBranch_Success(t *testing.T) {
	// Precondition: Repository has a valid HEAD commit
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Create a new branch
	err := vcs.CreateBranch("release/v1.0.0")

	// Expected: Branch is created with correct name pointing to HEAD
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.SetRefCalled {
		t.Error("expected SetReference to be called")
	}
	if mock.SetRefArg.Name().Short() != "release/v1.0.0" {
		t.Errorf("expected branch name 'release/v1.0.0', got '%s'", mock.SetRefArg.Name().Short())
	}
}

// TestMock_BranchExists_Found validates that existing branches are correctly detected.
//
// Why: Branch existence checks prevent duplicate branch creation and verify branch
// availability before checkout or merge operations.
//
// What: Given a repository with branches "main" and "feature/test", when BranchExists
// is called for "feature/test", then it returns true.
func TestMock_BranchExists_Found(t *testing.T) {
	// Precondition: Repository has multiple branches including feature/test
	branchHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	mock := NewMockRepository()
	mock.BranchRefs = []plumbing.Reference{
		MakeTestBranch("main", branchHash),
		MakeTestBranch("feature/test", branchHash),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Check if branch exists
	exists, err := vcs.BranchExists("feature/test")

	// Expected: Returns true without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected branch to exist")
	}
}

// TestMock_GetBranchName_OnBranch validates branch name retrieval when on a branch.
//
// Why: The current branch name is used for CI environment detection, branch-based
// versioning strategies, and workflow decisions.
//
// What: Given HEAD points to branch "main", when GetBranchName is called, then it
// returns "main".
func TestMock_GetBranchName_OnBranch(t *testing.T) {
	// Precondition: HEAD points to a branch reference (not detached)
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test", "Test", "test@test.com", time.Now())

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get current branch name
	branch, err := vcs.GetBranchName()

	// Expected: Returns branch name without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("expected 'main', got '%s'", branch)
	}
}

// TestMock_GetCommitDate_Success validates commit date retrieval.
//
// Why: Commit dates are used in version metadata for build timestamps and
// determining commit ordering in changelog generation.
//
// What: Given a commit with a specific date, when GetCommitDate is called,
// then it returns the exact commit timestamp.
func TestMock_GetCommitDate_Success(t *testing.T) {
	// Precondition: Repository has HEAD commit with known date
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)
	expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test", "Test", "test@test.com", expectedTime)

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commit date
	date, err := vcs.GetCommitDate()

	// Expected: Returns exact commit timestamp
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !date.Equal(expectedTime) {
		t.Errorf("expected %v, got %v", expectedTime, date)
	}
}

// TestMock_GetCommitAuthor_Success validates commit author name retrieval.
//
// Why: Author information is used in changelog generation and audit trails
// for tracking who made specific changes.
//
// What: Given a commit by "John Doe", when GetCommitAuthor is called, then
// it returns "John Doe".
func TestMock_GetCommitAuthor_Success(t *testing.T) {
	// Precondition: Repository has HEAD commit with known author
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test", "John Doe", "john@example.com", time.Now())

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commit author
	author, err := vcs.GetCommitAuthor()

	// Expected: Returns author name without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if author != "John Doe" {
		t.Errorf("expected 'John Doe', got '%s'", author)
	}
}

// TestMock_GetCommitAuthorEmail_Success validates commit author email retrieval.
//
// Why: Author email is used for notifications, attribution, and matching
// commits to user accounts in collaborative workflows.
//
// What: Given a commit with email "john@example.com", when GetCommitAuthorEmail
// is called, then it returns "john@example.com".
func TestMock_GetCommitAuthorEmail_Success(t *testing.T) {
	// Precondition: Repository has HEAD commit with known author email
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test", "John Doe", "john@example.com", time.Now())

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commit author email
	email, err := vcs.GetCommitAuthorEmail()

	// Expected: Returns author email without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email != "john@example.com" {
		t.Errorf("expected 'john@example.com', got '%s'", email)
	}
}

// TestMock_CommitFiles_Success validates file staging and commit creation.
//
// Why: CommitFiles is used to create version bump commits with updated version
// files. Failure would leave the repository in an inconsistent state.
//
// What: Given a list of files to commit, when CommitFiles is called, then all
// files are staged and a new commit is created with the provided message.
func TestMock_CommitFiles_Success(t *testing.T) {
	// Precondition: Repository has valid HEAD and worktree
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	newCommitHash := plumbing.NewHash("def456789012345678901234567890deadbeef")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "previous", "Test", "test@test.com", time.Now())
	wt := NewMockWorktree()
	wt.CommitHash = newCommitHash
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Commit multiple files
	err := vcs.CommitFiles([]string{"file1.txt", "file2.txt"}, "Test commit")

	// Expected: All files staged, commit created with correct message
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wt.AddCalls) != 2 {
		t.Errorf("expected 2 Add calls, got %d", len(wt.AddCalls))
	}
	if len(wt.CommitCalls) != 1 {
		t.Errorf("expected 1 Commit call, got %d", len(wt.CommitCalls))
	}
	if wt.CommitCalls[0].Msg != "Test commit" {
		t.Errorf("expected message 'Test commit', got '%s'", wt.CommitCalls[0].Msg)
	}
}

// TestMock_AmendCommit_Success validates amending an existing commit.
//
// Why: AmendCommit is used to add forgotten files to a version bump commit
// or fix commit content without creating additional commits.
//
// What: Given an existing commit, when AmendCommit is called with files, then
// the files are staged and the commit is amended with the original message.
func TestMock_AmendCommit_Success(t *testing.T) {
	// Precondition: Repository has an existing HEAD commit with known message
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "original message", "Test", "test@test.com", time.Now())
	wt := NewMockWorktree()
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Amend commit with additional files
	err := vcs.AmendCommit([]string{"file1.txt"})

	// Expected: Commit is amended with Amend option and original message preserved
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wt.CommitCalls) != 1 {
		t.Errorf("expected 1 Commit call, got %d", len(wt.CommitCalls))
	}
	if wt.CommitCalls[0].Msg != "original message" {
		t.Errorf("expected message 'original message', got '%s'", wt.CommitCalls[0].Msg)
	}
	if !wt.CommitCalls[0].Opts.Amend {
		t.Error("expected Amend option to be true")
	}
}

// TestMock_GetHooksPath_Success validates git hooks directory path construction.
//
// Why: The hooks path is needed to install git hooks for pre-commit validation
// and other automated checks.
//
// What: Given a repository root, when GetHooksPath is called, then it returns
// the path to the .git/hooks directory.
func TestMock_GetHooksPath_Success(t *testing.T) {
	// Precondition: VCS has known repository root
	vcs := NewGitVCSDefault()
	vcs.repoRoot = "/fake/repo"

	// Action: Get hooks path
	path, err := vcs.GetHooksPath()

	// Expected: Returns correct hooks directory path
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/fake/repo/.git/hooks" {
		t.Errorf("expected '/fake/repo/.git/hooks', got '%s'", path)
	}
}

// =============================================================================
// KEY VARIATIONS - Important alternate flows and states
// =============================================================================

// TestMock_IsWorkingDirectoryClean_Dirty validates dirty working directory detection.
//
// Why: Detecting dirty state prevents accidental releases from uncommitted work
// and ensures version reproducibility.
//
// What: Given a repository with an untracked file, when IsWorkingDirectoryClean
// is called, then it returns false.
func TestMock_IsWorkingDirectoryClean_Dirty(t *testing.T) {
	// Precondition: Real git repo with a modified tracked file
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	// Modify the tracked file to make it dirty
	os.WriteFile(filepath.Join(h.dir, "test.txt"), []byte("modified"), 0644)

	v := NewGitVCS(DefaultRepositoryOpener)
	v.repoRoot = h.dir

	// Action: Check if working directory is clean
	clean, err := v.IsWorkingDirectoryClean()

	// Expected: Returns false without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clean {
		t.Error("expected dirty working directory")
	}
}

// TestMock_TagExists_NotFound validates non-existent tag detection.
//
// Why: Correctly detecting missing tags allows the system to create new version
// tags without conflicts.
//
// What: Given a repository with no tags, when TagExists is called for "v1.0.0",
// then it returns false.
func TestMock_TagExists_NotFound(t *testing.T) {
	// Precondition: Repository has no tags
	mock := NewMockRepository()
	mock.TagRefs = []plumbing.Reference{}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Check if tag exists
	exists, err := vcs.TagExists("v1.0.0")

	// Expected: Returns false without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected tag to not exist")
	}
}

// TestMock_BranchExists_NotFound validates non-existent branch detection.
//
// Why: Correctly detecting missing branches allows the system to create new
// release branches without conflicts.
//
// What: Given a repository with only "main" branch, when BranchExists is called
// for "feature/nonexistent", then it returns false.
func TestMock_BranchExists_NotFound(t *testing.T) {
	// Precondition: Repository has only main branch
	branchHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	mock := NewMockRepository()
	mock.BranchRefs = []plumbing.Reference{
		MakeTestBranch("main", branchHash),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Check if nonexistent branch exists
	exists, err := vcs.BranchExists("feature/nonexistent")

	// Expected: Returns false without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected branch to not exist")
	}
}

// TestMock_GetBranchName_DetachedHead validates behavior when HEAD is detached.
//
// Why: Detached HEAD is common in CI environments during tag checkouts. The system
// must handle this gracefully rather than failing.
//
// What: Given a detached HEAD (pointing directly to a commit, not a branch), when
// GetBranchName is called, then it returns an empty string.
func TestMock_GetBranchName_DetachedHead(t *testing.T) {
	// Precondition: HEAD points directly to commit (detached state)
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get branch name in detached state
	branch, err := vcs.GetBranchName()

	// Expected: Returns empty string (no branch) without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "" {
		t.Errorf("expected empty string for detached HEAD, got '%s'", branch)
	}
}

// TestMock_GetUncommittedChanges_Zero validates zero change count detection.
//
// Why: Zero uncommitted changes indicates a clean state suitable for releases.
//
// What: Given a repository with no uncommitted changes, when GetUncommittedChanges
// is called, then it returns 0.
func TestMock_GetUncommittedChanges_Zero(t *testing.T) {
	// Precondition: Real git repo with clean state
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	v := NewGitVCS(DefaultRepositoryOpener)
	v.repoRoot = h.dir

	// Action: Get count of uncommitted changes
	count, err := v.GetUncommittedChanges()

	// Expected: Returns 0 without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// TestMock_GetUncommittedChanges_Multiple validates counting multiple changes.
//
// Why: Accurate change counting is used to warn users about uncommitted work
// and for CI status checks.
//
// What: Given a repository with 3 changed files, when GetUncommittedChanges
// is called, then it returns 3.
func TestMock_GetUncommittedChanges_Multiple(t *testing.T) {
	// Precondition: Real git repo with multiple dirty files
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Create multiple dirty files
	os.WriteFile(filepath.Join(h.dir, "test.txt"), []byte("modified"), 0644)
	os.WriteFile(filepath.Join(h.dir, "file2.txt"), []byte("new"), 0644)
	exec.Command("git", "-C", h.dir, "add", "file2.txt").Run()

	v := NewGitVCS(DefaultRepositoryOpener)
	v.repoRoot = h.dir

	// Action: Get count of uncommitted changes
	count, err := v.GetUncommittedChanges()

	// Expected: Returns count of changed files
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2, got %d", count)
	}
}

// TestMock_GetDirtyFiles_Success validates dirty file list retrieval.
//
// Why: Knowing which files are dirty helps users understand what needs to be
// committed before a release.
//
// What: Given a repository with 2 dirty files, when GetDirtyFiles is called,
// then it returns a list of 2 files.
func TestMock_GetDirtyFiles_Success(t *testing.T) {
	// Precondition: Real git repo with multiple dirty files
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Modify tracked file + stage a new file
	os.WriteFile(filepath.Join(h.dir, "test.txt"), []byte("modified"), 0644)
	os.WriteFile(filepath.Join(h.dir, "new.txt"), []byte("added"), 0644)
	exec.Command("git", "-C", h.dir, "add", "new.txt").Run()

	v := NewGitVCS(DefaultRepositoryOpener)
	v.repoRoot = h.dir

	// Action: Get list of dirty files
	files, err := v.GetDirtyFiles()

	// Expected: Returns list containing dirty files
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) < 2 {
		t.Errorf("expected at least 2 files, got %d", len(files))
	}
}

// TestMock_GetCommitMessagesSinceTag_WithCommits validates message collection with commits.
//
// Why: Commit messages since the last tag are used for changelog generation and
// release notes.
//
// What: Given 2 commits after the last tag, when GetCommitMessagesSinceTag is called,
// then it returns 2 commit messages.
func TestMock_GetCommitMessagesSinceTag_WithCommits(t *testing.T) {
	// Precondition: Repository has 2 commits after the tagged commit
	taggedHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	commit1Hash := plumbing.NewHash("111111111111111111111111111111111111dead")
	commit2Hash := plumbing.NewHash("222222222222222222222222222222222222dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commit2Hash)
	tagRef := MakeTestTag("v1.0.0", taggedHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[taggedHash] = MakeTestCommit(taggedHash, "tagged commit", "Test", "test@test.com", time.Now())
	mock.Commits[commit1Hash] = MakeTestCommit(commit1Hash, "first after tag", "Test", "test@test.com", time.Now())
	mock.Commits[commit2Hash] = MakeTestCommit(commit2Hash, "second after tag", "Test", "test@test.com", time.Now())
	mock.TagRefs = []plumbing.Reference{tagRef}
	mock.LogCommits = []*object.Commit{
		MakeTestCommit(commit2Hash, "second after tag", "Test", "test@test.com", time.Now()),
		MakeTestCommit(commit1Hash, "first after tag", "Test", "test@test.com", time.Now()),
		MakeTestCommit(taggedHash, "tagged commit", "Test", "test@test.com", time.Now()),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commit messages since tag
	messages, err := vcs.GetCommitMessagesSinceTag()

	// Expected: Returns 2 messages (excluding tagged commit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

// TestMock_GetTemplateVariables_WithShortHash validates template variable generation.
//
// Why: Template variables allow users to customize version output with VCS-specific
// prefixes for different systems (git vs sha prefix conventions).
//
// What: Given a short hash "abc1234", when GetTemplateVariables is called, then it
// returns variables with both git and sha prefixed versions.
func TestMock_GetTemplateVariables_WithShortHash(t *testing.T) {
	// Precondition: Input contains a short hash value
	vcs := NewGitVCSDefault()

	// Action: Get template variables with short hash
	vars := vcs.GetTemplateVariables(map[string]string{"ShortHash": "abc1234"})

	// Expected: Returns both prefixed versions
	if vars["GitShortHash"] != "git.abc1234" {
		t.Errorf("expected 'git.abc1234', got '%s'", vars["GitShortHash"])
	}
	if vars["ShaShortHash"] != "sha.abc1234" {
		t.Errorf("expected 'sha.abc1234', got '%s'", vars["ShaShortHash"])
	}
}

// =============================================================================
// ERROR HANDLING - Expected failure modes
// =============================================================================

// TestMock_GetVCSIdentifier_HeadError validates error propagation when HEAD fails.
//
// Why: HEAD resolution errors indicate repository corruption or invalid state.
// The error must propagate to allow proper error handling upstream.
//
// What: Given a repository where HEAD resolution fails, when GetVCSIdentifier
// is called, then it returns an error.
func TestMock_GetVCSIdentifier_HeadError(t *testing.T) {
	// Precondition: Repository HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get VCS identifier
	_, err := vcs.GetVCSIdentifier(7)

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_GetVCSIdentifier_CommitError validates error when commit lookup fails.
//
// Why: If HEAD points to a non-existent commit, this indicates repository corruption
// that must be reported rather than masked.
//
// What: Given a repository where commit lookup fails, when GetVCSIdentifier is called,
// then it returns an error.
func TestMock_GetVCSIdentifier_CommitError(t *testing.T) {
	// Precondition: HEAD exists but commit lookup fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.CommitErr = errors.New("commit not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get VCS identifier
	_, err := vcs.GetVCSIdentifier(7)

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when commit lookup fails")
	}
}

// TestMock_IsWorkingDirectoryClean_WorktreeError validates worktree error handling.
//
// Why: Worktree errors indicate the repository is in an invalid state (e.g., bare
// repository) and must be reported.
//
// What: Given a repository where worktree retrieval fails, when IsWorkingDirectoryClean
// is called, then it returns an error.
func TestMock_IsWorkingDirectoryClean_WorktreeError(t *testing.T) {
	// Precondition: Worktree retrieval fails
	mock := NewMockRepository()
	mock.WorktreeErr = errors.New("worktree error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to check working directory status
	_, err := vcs.IsWorkingDirectoryClean()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when worktree fails")
	}
}

// TestMock_IsWorkingDirectoryClean_StatusError validates status error handling.
//
// Why: Status command errors can occur due to filesystem issues or corrupted index.
// These must be reported rather than defaulting to clean/dirty.
//
// What: Given a repository where status command fails, when IsWorkingDirectoryClean
// is called, then it returns an error.
func TestMock_IsWorkingDirectoryClean_StatusError(t *testing.T) {
	// Precondition: Worktree exists but status fails
	mock := NewMockRepository()
	wt := NewMockWorktree()
	wt.StatusErr = errors.New("status error")
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to check working directory status
	_, err := vcs.IsWorkingDirectoryClean()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when status fails")
	}
}

// TestMock_CreateTag_HeadError validates HEAD error handling during tag creation.
//
// Why: Tags are created pointing to HEAD. If HEAD cannot be resolved, tag creation
// must fail rather than creating a tag pointing to an invalid commit.
//
// What: Given a repository where HEAD resolution fails, when CreateTag is called,
// then it returns an error.
func TestMock_CreateTag_HeadError(t *testing.T) {
	// Precondition: HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to create tag
	err := vcs.CreateTag("v1.0.0", "Release 1.0.0")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_CreateTag_CreateError validates tag creation error handling.
//
// Why: Tag creation can fail due to existing tags, permissions, or other issues.
// These failures must be reported to prevent false success assumptions.
//
// What: Given a repository where tag creation fails, when CreateTag is called,
// then it returns an error.
func TestMock_CreateTag_CreateError(t *testing.T) {
	// Precondition: HEAD exists but tag creation fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now())
	mock.CreateTagErr = errors.New("tag already exists")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to create tag
	err := vcs.CreateTag("v1.0.0", "Release 1.0.0")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when tag creation fails")
	}
}

// TestMock_TagExists_TagsError validates tags enumeration error handling.
//
// Why: Tag enumeration errors indicate repository corruption or access issues.
// These must be reported rather than returning false (tag not found).
//
// What: Given a repository where tag enumeration fails, when TagExists is called,
// then it returns an error.
func TestMock_TagExists_TagsError(t *testing.T) {
	// Precondition: Tags enumeration fails
	mock := NewMockRepository()
	mock.TagsErr = errors.New("tags error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to check tag existence
	_, err := vcs.TagExists("v1.0.0")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when tags fails")
	}
}

// TestMock_CreateBranch_HeadError validates HEAD error handling during branch creation.
//
// Why: Branches are created pointing to HEAD. If HEAD cannot be resolved, branch
// creation must fail rather than creating an invalid branch reference.
//
// What: Given a repository where HEAD resolution fails, when CreateBranch is called,
// then it returns an error.
func TestMock_CreateBranch_HeadError(t *testing.T) {
	// Precondition: HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to create branch
	err := vcs.CreateBranch("release/v1.0.0")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_CreateBranch_SetRefError validates reference creation error handling.
//
// Why: SetReference can fail due to permissions, locking, or other issues. These
// failures must be reported to prevent false success assumptions.
//
// What: Given a repository where SetReference fails, when CreateBranch is called,
// then it returns an error.
func TestMock_CreateBranch_SetRefError(t *testing.T) {
	// Precondition: HEAD exists but SetReference fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.SetRefErr = errors.New("reference error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to create branch
	err := vcs.CreateBranch("release/v1.0.0")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when SetReference fails")
	}
}

// TestMock_BranchExists_BranchesError validates branch enumeration error handling.
//
// Why: Branch enumeration errors indicate repository corruption or access issues.
// These must be reported rather than returning false (branch not found).
//
// What: Given a repository where branch enumeration fails, when BranchExists is
// called, then it returns an error.
func TestMock_BranchExists_BranchesError(t *testing.T) {
	// Precondition: Branches enumeration fails
	mock := NewMockRepository()
	mock.BranchesErr = errors.New("branches error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to check branch existence
	_, err := vcs.BranchExists("main")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when branches fails")
	}
}

// TestMock_GetBranchName_HeadError validates HEAD error handling in branch name retrieval.
//
// Why: Branch name depends on HEAD resolution. If HEAD cannot be resolved, the error
// must propagate rather than returning an empty string (which means detached HEAD).
//
// What: Given a repository where HEAD resolution fails, when GetBranchName is called,
// then it returns an error.
func TestMock_GetBranchName_HeadError(t *testing.T) {
	// Precondition: HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get branch name
	_, err := vcs.GetBranchName()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_CommitFiles_AddError validates file staging error handling.
//
// Why: If file staging fails (e.g., file not found, permissions), the commit operation
// must fail rather than creating a commit without the intended files.
//
// What: Given a repository where file addition fails, when CommitFiles is called,
// then it returns an error.
func TestMock_CommitFiles_AddError(t *testing.T) {
	// Precondition: Repository exists but Add operation fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "previous", "Test", "test@test.com", time.Now())
	wt := NewMockWorktree()
	wt.AddErr = errors.New("add error")
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to commit files
	err := vcs.CommitFiles([]string{"file1.txt"}, "Test commit")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when Add fails")
	}
}

// TestMock_CommitFiles_CommitError validates commit creation error handling.
//
// Why: Commit creation can fail due to hooks, permissions, or other issues. These
// failures must be reported to prevent false success assumptions.
//
// What: Given a repository where commit creation fails, when CommitFiles is called,
// then it returns an error.
func TestMock_CommitFiles_CommitError(t *testing.T) {
	// Precondition: Files staged but commit creation fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "previous", "Test", "test@test.com", time.Now())
	wt := NewMockWorktree()
	wt.CommitErr = errors.New("commit error")
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to commit files
	err := vcs.CommitFiles([]string{"file1.txt"}, "Test commit")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when Commit fails")
	}
}

// TestMock_RepositoryOpenerError validates repository open error handling.
//
// Why: If the repository cannot be opened (not a git repo, permissions, corruption),
// all VCS operations must fail gracefully with a descriptive error.
//
// What: Given a repository opener that fails, when any VCS operation is called,
// then it returns an error.
func TestMock_RepositoryOpenerError(t *testing.T) {
	// Precondition: Repository opener fails
	openErr := errors.New("failed to open repository")
	vcs := NewGitVCS(ErrorRepositoryOpener(openErr))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt any VCS operation
	_, err := vcs.GetVCSIdentifier(7)

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when repository opener fails")
	}
}

// =============================================================================
// EDGE CASES - Boundary conditions and special states
// =============================================================================

// TestMock_GetLastTag_NoTags validates behavior when repository has no tags.
//
// Why: A repository with no tags is a valid initial state. The system must handle
// this gracefully for initial version setup.
//
// What: Given a repository with no tags, when GetLastTag is called, then it returns
// an empty string (no error).
func TestMock_GetLastTag_NoTags(t *testing.T) {
	// Precondition: Repository has commits but no tags
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now())
	mock.TagRefs = []plumbing.Reference{}
	mock.LogCommits = []*object.Commit{
		MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now()),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get last tag
	tag, err := vcs.GetLastTag()

	// Expected: Returns empty string without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "" {
		t.Errorf("expected empty string, got '%s'", tag)
	}
}

// TestMock_GetLastTagCommit_NoTags validates behavior when repository has no tags.
//
// Why: When there are no tags, there is no tagged commit. The system must handle
// this without error for initial version setup workflows.
//
// What: Given a repository with no tags, when GetLastTagCommit is called, then it
// returns an empty string.
func TestMock_GetLastTagCommit_NoTags(t *testing.T) {
	// Precondition: Repository has commits but no tags
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now())
	mock.TagRefs = []plumbing.Reference{}
	mock.LogCommits = []*object.Commit{
		MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now()),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get last tag commit
	commit, err := vcs.GetLastTagCommit()

	// Expected: Returns empty string without error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commit != "" {
		t.Errorf("expected empty string, got '%s'", commit)
	}
}

// TestMock_GetCommitsSinceTag_NoTagsReturnsNegative validates behavior with no tags.
//
// Why: A sentinel value (-1) distinguishes "no tags exist" from "0 commits since tag"
// which have different semantic meanings for version calculation.
//
// What: Given a repository with no tags, when GetCommitsSinceTag is called, then it
// returns -1 to indicate no tags exist.
func TestMock_GetCommitsSinceTag_NoTagsReturnsNegative(t *testing.T) {
	// Precondition: Repository has commits but no tags
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now())
	mock.TagRefs = []plumbing.Reference{}
	mock.LogCommits = []*object.Commit{
		MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now()),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commits since tag
	count, err := vcs.GetCommitsSinceTag()

	// Expected: Returns -1 (sentinel for no tags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != -1 {
		t.Errorf("expected -1 (no tags), got %d", count)
	}
}

// TestMock_GetCommitMessagesSinceTag_NoCommitsSinceTag validates zero commits after tag.
//
// Why: If HEAD is exactly at a tagged commit, there are zero commits since that tag.
// This is a valid state after a release.
//
// What: Given HEAD is at a tagged commit, when GetCommitMessagesSinceTag is called,
// then it returns an empty list.
func TestMock_GetCommitMessagesSinceTag_NoCommitsSinceTag(t *testing.T) {
	// Precondition: HEAD points to the tagged commit (no commits after tag)
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)
	tagRef := MakeTestTag("v1.0.0", commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "tagged commit", "Test", "test@test.com", time.Now())
	mock.TagRefs = []plumbing.Reference{tagRef}
	mock.LogCommits = []*object.Commit{
		MakeTestCommit(commitHash, "tagged commit", "Test", "test@test.com", time.Now()),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commit messages since tag
	messages, err := vcs.GetCommitMessagesSinceTag()

	// Expected: Returns empty list
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

// TestMock_GetCommitMessagesSinceTag_NoTags validates behavior when no tags exist.
//
// Why: Without tags, there is no reference point for "since tag". The system must
// handle this gracefully rather than returning all commits.
//
// What: Given a repository with no tags, when GetCommitMessagesSinceTag is called,
// then it returns an empty list.
func TestMock_GetCommitMessagesSinceTag_NoTags(t *testing.T) {
	// Precondition: Repository has commits but no tags
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now())
	mock.TagRefs = []plumbing.Reference{}
	mock.LogCommits = []*object.Commit{
		MakeTestCommit(commitHash, "test commit", "Test", "test@test.com", time.Now()),
	}

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Get commit messages since tag
	messages, err := vcs.GetCommitMessagesSinceTag()

	// Expected: Returns empty list (no reference point)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

// TestMock_GetTemplateVariables_EmptyShortHash validates nil return for missing hash.
//
// Why: If no short hash is available, the template variables are meaningless. Returning
// nil allows callers to skip template processing.
//
// What: Given an empty input map, when GetTemplateVariables is called, then it returns nil.
func TestMock_GetTemplateVariables_EmptyShortHash(t *testing.T) {
	// Precondition: Input map has no ShortHash key
	vcs := NewGitVCSDefault()

	// Action: Get template variables without short hash
	vars := vcs.GetTemplateVariables(map[string]string{})

	// Expected: Returns nil
	if vars != nil {
		t.Error("expected nil when ShortHash is empty")
	}
}

// =============================================================================
// MINUTIAE - Obscure scenarios and implementation details
// =============================================================================

// TestMock_Types_ReturnsVCSAndTemplateProvider validates plugin type registration.
//
// Why: The Types() method is used by the plugin system to identify what capabilities
// this plugin provides. Incorrect types would break plugin discovery.
//
// What: When Types() is called, then it returns both VCS and TemplateProvider types.
func TestMock_Types_ReturnsVCSAndTemplateProvider(t *testing.T) {
	// Precondition: Default VCS instance
	vcs := NewGitVCSDefault()

	// Action: Get plugin types
	types := vcs.Types()

	// Expected: Returns 2 plugin types (VCS and TemplateProvider)
	if len(types) != 2 {
		t.Errorf("expected 2 plugin types, got %d", len(types))
	}
}

// =============================================================================
// PUSH OPERATIONS - Tests for PushTag and PushBranch
// =============================================================================

// TestMock_PushTag_RepoRootError validates error handling when repository root fails.
//
// Why: PushTag depends on GetRepositoryRoot to determine the working directory.
// If the repository root cannot be determined, the push must fail gracefully.
//
// What: Given a VCS where GetRepositoryRoot fails, when PushTag is called,
// then it returns the error from GetRepositoryRoot.
func TestMock_PushTag_RepoRootError(t *testing.T) {
	// Precondition: VCS has no repository root set (empty string triggers error)
	vcs := NewGitVCSDefault()
	vcs.repoRoot = ""

	// Action: Attempt to push tag
	err := vcs.PushTag("v1.0.0")

	// Expected: Returns error about repository root
	if err == nil {
		t.Error("expected error when repository root is empty")
	}
}

// TestMock_PushBranch_RepoRootError validates error handling when repository root fails.
//
// Why: PushBranch depends on GetRepositoryRoot to determine the working directory.
// If the repository root cannot be determined, the push must fail gracefully.
//
// What: Given a VCS where GetRepositoryRoot fails, when PushBranch is called,
// then it returns the error from GetRepositoryRoot.
func TestMock_PushBranch_RepoRootError(t *testing.T) {
	// Precondition: VCS has no repository root set (empty string triggers error)
	vcs := NewGitVCSDefault()
	vcs.repoRoot = ""

	// Action: Attempt to push branch
	err := vcs.PushBranch("release/v1.0.0")

	// Expected: Returns error about repository root
	if err == nil {
		t.Error("expected error when repository root is empty")
	}
}

// =============================================================================
// AMEND COMMIT ERROR PATHS
// =============================================================================

// TestMock_AmendCommit_WorktreeError validates error handling when worktree fails.
//
// Why: AmendCommit requires a valid worktree to stage files. If worktree retrieval
// fails, the operation must fail rather than corrupting repository state.
//
// What: Given a repository where worktree retrieval fails, when AmendCommit is called,
// then it returns an error.
func TestMock_AmendCommit_WorktreeError(t *testing.T) {
	// Precondition: Repository exists but worktree fails
	mock := NewMockRepository()
	mock.WorktreeErr = errors.New("worktree error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to amend commit
	err := vcs.AmendCommit([]string{"file.txt"})

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when worktree fails")
	}
}

// TestMock_AmendCommit_HeadError validates error handling when HEAD resolution fails.
//
// Why: AmendCommit needs to read the last commit message from HEAD. If HEAD cannot
// be resolved, the amend operation must fail.
//
// What: Given a repository where HEAD resolution fails, when AmendCommit is called,
// then it returns an error.
func TestMock_AmendCommit_HeadError(t *testing.T) {
	// Precondition: Repository has worktree but HEAD fails
	mock := NewMockRepository()
	mock.MockWorktree = NewMockWorktree()
	mock.HeadErr = errors.New("HEAD not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to amend commit
	err := vcs.AmendCommit([]string{"file.txt"})

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_AmendCommit_CommitObjectError validates error handling when commit lookup fails.
//
// Why: AmendCommit needs to read the last commit to preserve its message. If the
// commit object cannot be retrieved, the amend operation must fail.
//
// What: Given a repository where commit lookup fails, when AmendCommit is called,
// then it returns an error.
func TestMock_AmendCommit_CommitObjectError(t *testing.T) {
	// Precondition: Repository has worktree and HEAD but commit lookup fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.MockWorktree = NewMockWorktree()
	mock.CommitErr = errors.New("commit not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to amend commit
	err := vcs.AmendCommit([]string{"file.txt"})

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when commit lookup fails")
	}
}

// TestMock_AmendCommit_CommitError validates error handling when amend fails.
//
// Why: The commit operation itself can fail due to hooks, permissions, or other issues.
// These failures must be reported to prevent false success assumptions.
//
// What: Given a repository where the commit operation fails, when AmendCommit is called,
// then it returns an error.
func TestMock_AmendCommit_CommitError(t *testing.T) {
	// Precondition: All setup succeeds but commit operation fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.Commits[commitHash] = MakeTestCommit(commitHash, "original", "Test", "test@test.com", time.Now())
	wt := NewMockWorktree()
	wt.CommitErr = errors.New("commit failed")
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to amend commit
	err := vcs.AmendCommit([]string{"file.txt"})

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when commit fails")
	}
}

// =============================================================================
// COMMIT DATE ERROR PATHS
// =============================================================================

// TestMock_GetCommitDate_HeadError validates error handling when HEAD resolution fails.
//
// Why: GetCommitDate needs to resolve HEAD to find the current commit. If HEAD
// cannot be resolved, the error must propagate.
//
// What: Given a repository where HEAD resolution fails, when GetCommitDate is called,
// then it returns an error.
func TestMock_GetCommitDate_HeadError(t *testing.T) {
	// Precondition: HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get commit date
	_, err := vcs.GetCommitDate()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_GetCommitDate_CommitError validates error handling when commit lookup fails.
//
// Why: GetCommitDate needs to retrieve the commit object to read its timestamp.
// If commit lookup fails, the error must propagate.
//
// What: Given a repository where commit lookup fails, when GetCommitDate is called,
// then it returns an error.
func TestMock_GetCommitDate_CommitError(t *testing.T) {
	// Precondition: HEAD exists but commit lookup fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.CommitErr = errors.New("commit not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get commit date
	_, err := vcs.GetCommitDate()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when commit lookup fails")
	}
}

// =============================================================================
// COMMIT AUTHOR ERROR PATHS
// =============================================================================

// TestMock_GetCommitAuthor_HeadError validates error handling when HEAD resolution fails.
//
// Why: GetCommitAuthor needs to resolve HEAD to find the current commit. If HEAD
// cannot be resolved, the error must propagate.
//
// What: Given a repository where HEAD resolution fails, when GetCommitAuthor is called,
// then it returns an error.
func TestMock_GetCommitAuthor_HeadError(t *testing.T) {
	// Precondition: HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get commit author
	_, err := vcs.GetCommitAuthor()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_GetCommitAuthor_CommitError validates error handling when commit lookup fails.
//
// Why: GetCommitAuthor needs to retrieve the commit object to read author info.
// If commit lookup fails, the error must propagate.
//
// What: Given a repository where commit lookup fails, when GetCommitAuthor is called,
// then it returns an error.
func TestMock_GetCommitAuthor_CommitError(t *testing.T) {
	// Precondition: HEAD exists but commit lookup fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.CommitErr = errors.New("commit not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get commit author
	_, err := vcs.GetCommitAuthor()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when commit lookup fails")
	}
}

// TestMock_GetCommitAuthorEmail_HeadError validates error handling when HEAD fails.
//
// Why: GetCommitAuthorEmail needs to resolve HEAD to find the current commit.
// If HEAD cannot be resolved, the error must propagate.
//
// What: Given a repository where HEAD resolution fails, when GetCommitAuthorEmail
// is called, then it returns an error.
func TestMock_GetCommitAuthorEmail_HeadError(t *testing.T) {
	// Precondition: HEAD resolution fails
	mock := NewMockRepository()
	mock.HeadErr = errors.New("HEAD not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get commit author email
	_, err := vcs.GetCommitAuthorEmail()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when HEAD fails")
	}
}

// TestMock_GetCommitAuthorEmail_CommitError validates error handling when commit fails.
//
// Why: GetCommitAuthorEmail needs to retrieve the commit object to read email info.
// If commit lookup fails, the error must propagate.
//
// What: Given a repository where commit lookup fails, when GetCommitAuthorEmail
// is called, then it returns an error.
func TestMock_GetCommitAuthorEmail_CommitError(t *testing.T) {
	// Precondition: HEAD exists but commit lookup fails
	commitHash := plumbing.NewHash("abc123def456789012345678901234567890dead")
	headRef := plumbing.NewHashReference(plumbing.HEAD, commitHash)

	mock := NewMockRepository()
	mock.HeadRef = headRef
	mock.CommitErr = errors.New("commit not found")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get commit author email
	_, err := vcs.GetCommitAuthorEmail()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when commit lookup fails")
	}
}

// =============================================================================
// HOOKS PATH ERROR PATHS
// =============================================================================

// TestMock_GetHooksPath_RepoRootError validates error handling when repository root fails.
//
// Why: GetHooksPath needs the repository root to construct the hooks path.
// If the root cannot be determined, the error must propagate.
//
// What: Given a VCS where repository opening fails, when GetHooksPath is called,
// then it returns an error.
func TestMock_GetHooksPath_RepoRootError(t *testing.T) {
	// Precondition: Change to a directory that's not a git repository
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Create VCS with empty root - it will try to find .git and fail
	vcs := NewGitVCSDefault()
	vcs.repoRoot = ""

	// Action: Attempt to get hooks path
	_, err := vcs.GetHooksPath()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when repository root cannot be found")
	}
}

// =============================================================================
// GET DIRTY FILES ERROR PATHS
// =============================================================================

// TestMock_GetDirtyFiles_WorktreeError validates error handling when worktree fails.
//
// Why: GetDirtyFiles needs a valid worktree to check status. If worktree retrieval
// fails, the error must propagate.
//
// What: Given a repository where worktree retrieval fails, when GetDirtyFiles is called,
// then it returns an error.
func TestMock_GetDirtyFiles_WorktreeError(t *testing.T) {
	// Precondition: Worktree retrieval fails
	mock := NewMockRepository()
	mock.WorktreeErr = errors.New("worktree error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get dirty files
	_, err := vcs.GetDirtyFiles()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when worktree fails")
	}
}

// TestMock_GetDirtyFiles_StatusError validates error handling when status fails.
//
// Why: GetDirtyFiles needs to query worktree status. If status command fails,
// the error must propagate.
//
// What: Given a repository where status command fails, when GetDirtyFiles is called,
// then it returns an error.
func TestMock_GetDirtyFiles_StatusError(t *testing.T) {
	// Precondition: Worktree exists but status fails
	mock := NewMockRepository()
	wt := NewMockWorktree()
	wt.StatusErr = errors.New("status error")
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get dirty files
	_, err := vcs.GetDirtyFiles()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when status fails")
	}
}

// =============================================================================
// GET UNCOMMITTED CHANGES ERROR PATHS
// =============================================================================

// TestMock_GetUncommittedChanges_WorktreeError validates error when worktree fails.
//
// Why: GetUncommittedChanges needs a valid worktree to check status.
// If worktree retrieval fails, the error must propagate.
//
// What: Given a repository where worktree retrieval fails, when GetUncommittedChanges
// is called, then it returns an error.
func TestMock_GetUncommittedChanges_WorktreeError(t *testing.T) {
	// Precondition: Worktree retrieval fails
	mock := NewMockRepository()
	mock.WorktreeErr = errors.New("worktree error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get uncommitted changes count
	_, err := vcs.GetUncommittedChanges()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when worktree fails")
	}
}

// TestMock_GetUncommittedChanges_StatusError validates error when status fails.
//
// Why: GetUncommittedChanges needs to query worktree status. If status command fails,
// the error must propagate.
//
// What: Given a repository where status command fails, when GetUncommittedChanges
// is called, then it returns an error.
func TestMock_GetUncommittedChanges_StatusError(t *testing.T) {
	// Precondition: Worktree exists but status fails
	mock := NewMockRepository()
	wt := NewMockWorktree()
	wt.StatusErr = errors.New("status error")
	mock.MockWorktree = wt

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to get uncommitted changes count
	_, err := vcs.GetUncommittedChanges()

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when status fails")
	}
}

// =============================================================================
// COMMIT FILES ERROR PATHS
// =============================================================================

// TestMock_CommitFiles_WorktreeError validates error handling when worktree fails.
//
// Why: CommitFiles needs a valid worktree to stage files. If worktree retrieval
// fails, the operation must fail rather than silently skipping files.
//
// What: Given a repository where worktree retrieval fails, when CommitFiles is called,
// then it returns an error.
func TestMock_CommitFiles_WorktreeError(t *testing.T) {
	// Precondition: Worktree retrieval fails
	mock := NewMockRepository()
	mock.WorktreeErr = errors.New("worktree error")

	vcs := NewGitVCS(MockRepositoryOpener(mock))
	vcs.repoRoot = "/fake/path"

	// Action: Attempt to commit files
	err := vcs.CommitFiles([]string{"file.txt"}, "Test commit")

	// Expected: Returns error
	if err == nil {
		t.Error("expected error when worktree fails")
	}
}
