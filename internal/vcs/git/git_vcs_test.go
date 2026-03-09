package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// =============================================================================
// TEST HELPERS
// Test utilities for creating and managing temporary git repositories.
// =============================================================================

// TestHelper creates a temporary git repository for testing
type TestHelper struct {
	t       *testing.T
	dir     string
	repo    *git.Repository
	origDir string
}

// NewTestHelper creates a new test helper with a temporary git repository
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	// Create temp directory
	dir, err := os.MkdirTemp("", "git-vcs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repository
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(dir); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	return &TestHelper{
		t:       t,
		dir:     dir,
		repo:    repo,
		origDir: origDir,
	}
}

// Cleanup restores original directory and removes temp directory
func (h *TestHelper) Cleanup() {
	_ = os.Chdir(h.origDir)
	os.RemoveAll(h.dir)
}

// CreateCommit creates a commit in the test repository
func (h *TestHelper) CreateCommit(message string) {
	h.t.Helper()

	// Create a file to commit
	filename := filepath.Join(h.dir, "test.txt")
	content := []byte(message + "\n")
	if err := os.WriteFile(filename, content, 0644); err != nil {
		h.t.Fatalf("failed to create file: %v", err)
	}

	wt, err := h.repo.Worktree()
	if err != nil {
		h.t.Fatalf("failed to get worktree: %v", err)
	}

	if _, err := wt.Add("test.txt"); err != nil {
		h.t.Fatalf("failed to add file: %v", err)
	}

	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		h.t.Fatalf("failed to commit: %v", err)
	}
}

// CreateTag creates a tag in the test repository
func (h *TestHelper) CreateTag(name, message string) {
	h.t.Helper()

	head, err := h.repo.Head()
	if err != nil {
		h.t.Fatalf("failed to get HEAD: %v", err)
	}

	_, err = h.repo.CreateTag(name, head.Hash(), &git.CreateTagOptions{
		Message: message,
		Tagger: &object.Signature{
			Name:  "Test Tagger",
			Email: "tagger@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		h.t.Fatalf("failed to create tag: %v", err)
	}
}

// CreateLightweightTag creates a lightweight tag (no annotation)
func (h *TestHelper) CreateLightweightTag(name string) {
	h.t.Helper()

	head, err := h.repo.Head()
	if err != nil {
		h.t.Fatalf("failed to get HEAD: %v", err)
	}

	_, err = h.repo.CreateTag(name, head.Hash(), nil)
	if err != nil {
		h.t.Fatalf("failed to create lightweight tag: %v", err)
	}
}

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of GitVCS: repository detection,
// commit identification, and version control operations.
// =============================================================================

// TestName_ReturnsGit validates that the VCS identifies itself as "git".
//
// Why: Consumers may need to determine the VCS type for conditional logic or
// logging. This ensures the VCS correctly self-identifies.
//
// What: Call Name() on a GitVCS instance and verify it returns "git".
func TestName_ReturnsGit(t *testing.T) {
	// Precondition: A GitVCS instance is created
	vcs := NewGitVCSDefault()

	// Action: Get the VCS name
	name := vcs.Name()

	// Expected: The name should be "git"
	if name != "git" {
		t.Errorf("expected name 'git', got '%s'", name)
	}
}

// TestIsRepository_InGitRepo_ReturnsTrue validates repository detection inside
// a valid git repository.
//
// Why: The VCS must correctly detect when operating inside a git repository to
// enable version control operations. False negatives would prevent versioning.
//
// What: Initialize a git repo with at least one commit, then verify
// IsRepository() returns true.
func TestIsRepository_InGitRepo_ReturnsTrue(t *testing.T) {
	// Precondition: A valid git repository with at least one commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Check if we're in a repository
	vcs := NewGitVCSDefault()
	isRepo := vcs.IsRepository()

	// Expected: Should return true since we're in a git repo
	if !isRepo {
		t.Error("expected IsRepository() to return true in git repo")
	}
}

// TestGetRepositoryRoot_ReturnsCorrectPath validates that the repository root
// is correctly identified.
//
// Why: Many operations need to locate the repository root to find config files
// or construct relative paths. Incorrect root detection breaks these workflows.
//
// What: Create a git repo, then verify GetRepositoryRoot() returns the
// directory containing the .git folder.
func TestGetRepositoryRoot_ReturnsCorrectPath(t *testing.T) {
	// Precondition: A git repository initialized in a known directory
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the repository root
	vcs := NewGitVCSDefault()
	root, err := vcs.GetRepositoryRoot()

	// Expected: Should return the temp directory where the repo was initialized
	if err != nil {
		t.Fatalf("GetRepositoryRoot() error: %v", err)
	}

	// Resolve symlinks for comparison (macOS /tmp -> /private/tmp)
	expectedRoot, _ := filepath.EvalSymlinks(h.dir)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualRoot != expectedRoot {
		t.Errorf("expected root '%s', got '%s'", expectedRoot, actualRoot)
	}
}

// TestGetVCSIdentifier_ReturnsCorrectLength validates commit hash retrieval
// with various length specifications.
//
// Why: Version strings often embed commit hashes for traceability. The hash
// length must be configurable to balance uniqueness vs. readability.
//
// What: Request hashes of various lengths and verify the returned strings
// match the requested lengths.
func TestGetVCSIdentifier_ReturnsCorrectLength(t *testing.T) {
	// Precondition: A git repository with at least one commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	vcs := NewGitVCSDefault()

	tests := []struct {
		length   int
		wantErr  bool
		wantLen  int
		testName string
	}{
		{7, false, 7, "default length"},
		{1, false, 1, "minimum length"},
		{40, false, 40, "maximum length"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Action: Get the VCS identifier with specified length
			hash, err := vcs.GetVCSIdentifier(tt.length)

			// Expected: Should return hash of exact length requested
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(hash) != tt.wantLen {
				t.Errorf("expected hash length %d, got %d", tt.wantLen, len(hash))
			}
		})
	}
}

// TestGetBranchName_OnBranch_ReturnsBranchName validates branch name retrieval.
//
// Why: CI/CD systems often need the current branch name for build metadata,
// conditional deployments, or versioning schemes based on branch patterns.
//
// What: Create a commit on a branch, then verify GetBranchName() returns the
// correct branch name.
func TestGetBranchName_OnBranch_ReturnsBranchName(t *testing.T) {
	// Precondition: A git repository with a commit on a branch
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the current branch name
	vcs := NewGitVCSDefault()
	branch, err := vcs.GetBranchName()

	// Expected: Should return "master" or "main" (depends on git config)
	if err != nil {
		t.Fatalf("GetBranchName() error: %v", err)
	}

	if branch != "master" && branch != "main" {
		t.Errorf("expected branch 'master' or 'main', got '%s'", branch)
	}
}

// TestGetCommitDate_ReturnsValidDate validates commit timestamp retrieval.
//
// Why: Version metadata often includes timestamps for tracking when builds
// were created. This must accurately reflect the actual commit time.
//
// What: Create a commit and verify GetCommitDate() returns a timestamp within
// an acceptable range around when the commit was made.
func TestGetCommitDate_ReturnsValidDate(t *testing.T) {
	// Precondition: A git repository with a commit made at a known time
	h := NewTestHelper(t)
	defer h.Cleanup()

	beforeCommit := time.Now().Add(-time.Second).UTC()
	h.CreateCommit("initial commit")
	afterCommit := time.Now().Add(time.Second).UTC()

	// Action: Get the commit date
	vcs := NewGitVCSDefault()
	commitDate, err := vcs.GetCommitDate()

	// Expected: Date should be between before and after timestamps
	if err != nil {
		t.Fatalf("GetCommitDate() error: %v", err)
	}

	if commitDate.Before(beforeCommit) || commitDate.After(afterCommit) {
		t.Errorf("commit date %v not within expected range [%v, %v]",
			commitDate, beforeCommit, afterCommit)
	}
}

// TestGetCommitAuthor_ReturnsAuthorName validates author name retrieval.
//
// Why: Build metadata may include author information for audit trails or
// attribution in changelogs.
//
// What: Create a commit with a known author, then verify GetCommitAuthor()
// returns the correct name.
func TestGetCommitAuthor_ReturnsAuthorName(t *testing.T) {
	// Precondition: A commit created with a known author name
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the commit author
	vcs := NewGitVCSDefault()
	author, err := vcs.GetCommitAuthor()

	// Expected: Should return "Test Author" as configured in TestHelper
	if err != nil {
		t.Fatalf("GetCommitAuthor() error: %v", err)
	}

	if author != "Test Author" {
		t.Errorf("expected 'Test Author', got '%s'", author)
	}
}

// TestGetCommitAuthorEmail_ReturnsAuthorEmail validates author email retrieval.
//
// Why: Author email can be used for notifications, attribution, or linking to
// user profiles in CI/CD systems.
//
// What: Create a commit with a known author email, then verify
// GetCommitAuthorEmail() returns the correct email.
func TestGetCommitAuthorEmail_ReturnsAuthorEmail(t *testing.T) {
	// Precondition: A commit created with a known author email
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the commit author email
	vcs := NewGitVCSDefault()
	email, err := vcs.GetCommitAuthorEmail()

	// Expected: Should return "test@example.com" as configured in TestHelper
	if err != nil {
		t.Fatalf("GetCommitAuthorEmail() error: %v", err)
	}

	if email != "test@example.com" {
		t.Errorf("expected 'test@example.com', got '%s'", email)
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests demonstrating important alternate flows and common use patterns.
// =============================================================================

// TestIsWorkingDirectoryClean_CleanRepo_ReturnsTrue validates clean state
// detection when there are no uncommitted changes.
//
// Why: Release tooling often requires a clean working directory to ensure
// deterministic builds. This test confirms clean detection works correctly.
//
// What: Create a commit with no pending changes, then verify
// IsWorkingDirectoryClean() returns true.
func TestIsWorkingDirectoryClean_CleanRepo_ReturnsTrue(t *testing.T) {
	// Precondition: A git repository with all changes committed
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Check if the working directory is clean
	vcs := NewGitVCSDefault()
	clean, err := vcs.IsWorkingDirectoryClean()

	// Expected: Should return true since everything is committed
	if err != nil {
		t.Fatalf("IsWorkingDirectoryClean() error: %v", err)
	}

	if !clean {
		t.Error("expected clean working directory")
	}
}

// TestIsWorkingDirectoryClean_DirtyRepo_ReturnsFalse validates dirty state
// detection when uncommitted changes exist.
//
// Why: Detecting uncommitted changes prevents releasing untested code and
// ensures version strings accurately reflect repository state.
//
// What: Create an uncommitted file after a commit, then verify
// IsWorkingDirectoryClean() returns false.
func TestIsWorkingDirectoryClean_DirtyRepo_ReturnsFalse(t *testing.T) {
	// Precondition: A git repository with uncommitted changes
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Create an uncommitted file
	filename := filepath.Join(h.dir, "dirty.txt")
	if err := os.WriteFile(filename, []byte("dirty"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Action: Check if the working directory is clean
	vcs := NewGitVCSDefault()
	clean, err := vcs.IsWorkingDirectoryClean()

	// Expected: Should return false due to uncommitted file
	if err != nil {
		t.Fatalf("IsWorkingDirectoryClean() error: %v", err)
	}

	if clean {
		t.Error("expected dirty working directory")
	}
}

// TestCreateTag_CreatesAnnotatedTag validates tag creation functionality.
//
// Why: Version tagging is fundamental to release management. Tags mark specific
// commits as releases for downstream consumption.
//
// What: Create an annotated tag, then verify it exists via TagExists().
func TestCreateTag_CreatesAnnotatedTag(t *testing.T) {
	// Precondition: A git repository with a commit to tag
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Create an annotated tag
	vcs := NewGitVCSDefault()
	err := vcs.CreateTag("v1.0.0", "Release 1.0.0")

	// Expected: Tag should be created successfully
	if err != nil {
		t.Fatalf("CreateTag() error: %v", err)
	}

	// Verify tag exists
	exists, err := vcs.TagExists("v1.0.0")
	if err != nil {
		t.Fatalf("TagExists() error: %v", err)
	}
	if !exists {
		t.Error("expected tag to exist")
	}
}

// TestTagExists_ExistingTag_ReturnsTrue validates tag existence detection for
// existing tags.
//
// Why: Before creating a tag, tools need to check if the version already
// exists to prevent duplicate releases.
//
// What: Create a tag, then verify TagExists() returns true.
func TestTagExists_ExistingTag_ReturnsTrue(t *testing.T) {
	// Precondition: A git repository with an existing tag
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	h.CreateTag("v1.0.0", "Test tag")

	// Action: Check if the tag exists
	vcs := NewGitVCSDefault()
	exists, err := vcs.TagExists("v1.0.0")

	// Expected: Should return true for existing tag
	if err != nil {
		t.Fatalf("TagExists() error: %v", err)
	}
	if !exists {
		t.Error("expected tag to exist")
	}
}

// TestTagExists_NonExistingTag_ReturnsFalse validates tag existence detection
// for non-existent tags.
//
// Why: Tools must differentiate between existing and non-existent tags to
// allow creating new versions while preventing overwrites.
//
// What: Query for a tag that was never created, verify TagExists() returns false.
func TestTagExists_NonExistingTag_ReturnsFalse(t *testing.T) {
	// Precondition: A git repository without the queried tag
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Check if a non-existent tag exists
	vcs := NewGitVCSDefault()
	exists, err := vcs.TagExists("v1.0.0")

	// Expected: Should return false for non-existent tag
	if err != nil {
		t.Fatalf("TagExists() error: %v", err)
	}
	if exists {
		t.Error("expected tag to not exist")
	}
}

// TestGetCommitsSinceTag_OnTaggedCommit_ReturnsZero validates commit counting
// when HEAD is exactly at a tagged commit.
//
// Why: Build numbers derived from commit counts must correctly show zero when
// on a tagged release to generate clean version strings.
//
// What: Create a tag on HEAD, then verify GetCommitsSinceTag() returns 0.
func TestGetCommitsSinceTag_OnTaggedCommit_ReturnsZero(t *testing.T) {
	// Precondition: A tag exists on the current HEAD commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	h.CreateTag("v1.0.0", "Release 1.0.0")

	// Action: Get commits since the tag
	vcs := NewGitVCSDefault()
	count, err := vcs.GetCommitsSinceTag()

	// Expected: Should return 0 since HEAD is at the tagged commit
	if err != nil {
		t.Fatalf("GetCommitsSinceTag() error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 commits since tag, got %d", count)
	}
}

// TestGetCommitsSinceTag_CommitsAfterTag_ReturnsCorrectCount validates commit
// counting when commits exist after the last tag.
//
// Why: Pre-release version strings often include commit distance (e.g.,
// 1.0.0-beta.2) to show development progress since last release.
//
// What: Create commits after a tag, verify GetCommitsSinceTag() returns the
// correct count.
func TestGetCommitsSinceTag_CommitsAfterTag_ReturnsCorrectCount(t *testing.T) {
	// Precondition: Two commits exist after the last tag
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	h.CreateTag("v1.0.0", "Release 1.0.0")
	h.CreateCommit("second commit")
	h.CreateCommit("third commit")

	// Action: Get commits since the tag
	vcs := NewGitVCSDefault()
	count, err := vcs.GetCommitsSinceTag()

	// Expected: Should return 2 for the two commits after the tag
	if err != nil {
		t.Fatalf("GetCommitsSinceTag() error: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 commits since tag, got %d", count)
	}
}

// TestGetLastTag_WithTags_ReturnsMostRecent validates that GetLastTag returns
// the most recently created tag when multiple tags exist.
//
// Why: Version determination must identify the latest release to calculate
// next version or commit distance correctly.
//
// What: Create multiple tags on different commits, verify GetLastTag() returns
// the most recent one.
func TestGetLastTag_WithTags_ReturnsMostRecent(t *testing.T) {
	// Precondition: Multiple tags exist on different commits
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	h.CreateTag("v1.0.0", "Release 1.0.0")
	h.CreateCommit("second commit")
	h.CreateTag("v1.1.0", "Release 1.1.0")

	// Action: Get the last tag
	vcs := NewGitVCSDefault()
	tag, err := vcs.GetLastTag()

	// Expected: Should return the most recent tag "v1.1.0"
	if err != nil {
		t.Fatalf("GetLastTag() error: %v", err)
	}

	if tag != "v1.1.0" {
		t.Errorf("expected 'v1.1.0', got '%s'", tag)
	}
}

// TestGetLastTagCommit_WithTag_ReturnsCommitHash validates that the correct
// commit hash is returned for the last tag.
//
// Why: Some workflows need the exact commit that was tagged to compare changes
// or determine what code is in a specific release.
//
// What: Create a tag, add more commits, verify GetLastTagCommit() returns the
// hash of the tagged commit (not HEAD).
func TestGetLastTagCommit_WithTag_ReturnsCommitHash(t *testing.T) {
	// Precondition: A tag exists followed by additional commits
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	h.CreateTag("v1.0.0", "Release 1.0.0")

	// Get the commit hash we tagged
	head, _ := h.repo.Head()
	expectedHash := head.Hash().String()

	h.CreateCommit("second commit")

	// Action: Get the last tag's commit hash
	vcs := NewGitVCSDefault()
	commit, err := vcs.GetLastTagCommit()

	// Expected: Should return the hash of the tagged commit, not HEAD
	if err != nil {
		t.Fatalf("GetLastTagCommit() error: %v", err)
	}

	if commit != expectedHash {
		t.Errorf("expected '%s', got '%s'", expectedHash, commit)
	}
}

// TestGetUncommittedChanges_WithChanges_ReturnsCount validates counting of
// uncommitted changes.
//
// Why: Dirty state indicators in version strings (e.g., "1.0.0+dirty") require
// knowing how many files have uncommitted changes.
//
// What: Create multiple uncommitted files, verify GetUncommittedChanges()
// returns the correct count.
func TestGetUncommittedChanges_WithChanges_ReturnsCount(t *testing.T) {
	// Precondition: Multiple uncommitted files exist
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Create uncommitted files
	for i := 0; i < 3; i++ {
		filename := filepath.Join(h.dir, "file"+string(rune('a'+i))+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	// Action: Get the count of uncommitted changes
	vcs := NewGitVCSDefault()
	count, err := vcs.GetUncommittedChanges()

	// Expected: Should return 3 for the three uncommitted files
	if err != nil {
		t.Fatalf("GetUncommittedChanges() error: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 uncommitted changes, got %d", count)
	}
}

// TestCreateBranch_Success validates branch creation functionality.
//
// Why: Release workflows may create release branches for maintenance or
// hotfixes. Proper branch creation is essential for these workflows.
//
// What: Create a new branch, then verify it exists via BranchExists().
func TestCreateBranch_Success(t *testing.T) {
	// Precondition: A git repository with a commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Create a new branch
	vcs := NewGitVCSDefault()
	err := vcs.CreateBranch("release/v1.0.0")

	// Expected: Branch should be created successfully
	if err != nil {
		t.Fatalf("CreateBranch() error: %v", err)
	}

	// Verify branch was created
	exists, err := vcs.BranchExists("release/v1.0.0")
	if err != nil {
		t.Fatalf("BranchExists() error: %v", err)
	}
	if !exists {
		t.Error("expected branch to exist after creation")
	}
}

// TestBranchExists_ExistingBranch_ReturnsTrue validates branch existence
// detection for existing branches.
//
// Why: Before creating a branch, tools must check if it already exists to
// prevent errors or accidental overwrites.
//
// What: Query for the default branch after a commit, verify BranchExists()
// returns true.
func TestBranchExists_ExistingBranch_ReturnsTrue(t *testing.T) {
	// Precondition: A git repository with a commit (creates default branch)
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	vcs := NewGitVCSDefault()

	// Master/main branch should exist after commit
	// First check what branches exist
	branches, err := h.repo.Branches()
	if err != nil {
		t.Fatalf("failed to get branches: %v", err)
	}

	var branchName string
	_ = branches.ForEach(func(ref *plumbing.Reference) error {
		branchName = ref.Name().Short()
		return errStopIteration
	})

	// Action: Check if the branch exists
	exists, err := vcs.BranchExists(branchName)

	// Expected: Should return true for existing branch
	if err != nil {
		t.Fatalf("BranchExists() error: %v", err)
	}
	if !exists {
		t.Errorf("expected branch '%s' to exist", branchName)
	}
}

// TestBranchExists_NonExistingBranch_ReturnsFalse validates branch existence
// detection for non-existent branches.
//
// Why: Branch existence checks must correctly identify missing branches to
// allow creation without false collision errors.
//
// What: Query for a branch that was never created, verify BranchExists()
// returns false.
func TestBranchExists_NonExistingBranch_ReturnsFalse(t *testing.T) {
	// Precondition: A git repository without the queried branch
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Check if a non-existent branch exists
	vcs := NewGitVCSDefault()
	exists, err := vcs.BranchExists("non-existent-branch-xyz")

	// Expected: Should return false for non-existent branch
	if err != nil {
		t.Fatalf("BranchExists() error: %v", err)
	}
	if exists {
		t.Error("expected non-existent branch to return false")
	}
}

// =============================================================================
// ERROR HANDLING
// Tests demonstrating expected failure modes and error conditions.
// =============================================================================

// TestIsRepository_NotInGitRepo_ReturnsFalse validates repository detection
// outside a git repository.
//
// Why: Graceful handling of non-repository directories prevents crashes and
// allows clear error messaging to users.
//
// What: Create a plain directory without git init, verify IsRepository()
// returns false without error.
func TestIsRepository_NotInGitRepo_ReturnsFalse(t *testing.T) {
	// Precondition: A directory without git initialization
	dir, err := os.MkdirTemp("", "no-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Action: Check if we're in a repository
	vcs := NewGitVCSDefault()
	isRepo := vcs.IsRepository()

	// Expected: Should return false outside a git repo
	if isRepo {
		t.Error("expected IsRepository() to return false outside git repo")
	}
}

// TestGetRepositoryRoot_NotInRepo_ReturnsError validates error handling when
// trying to get repository root outside a git repository.
//
// Why: Operations that depend on repository context must fail clearly when
// invoked outside a repository.
//
// What: Call GetRepositoryRoot() outside a git repo, verify an error is returned.
func TestGetRepositoryRoot_NotInRepo_ReturnsError(t *testing.T) {
	// Precondition: A directory without git initialization
	dir, err := os.MkdirTemp("", "no-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Action: Attempt to get repository root outside a repo
	vcs := NewGitVCSDefault()
	_, err = vcs.GetRepositoryRoot()

	// Expected: Should return an error
	if err == nil {
		t.Error("expected error when not in git repo")
	}
}

// TestGetVCSIdentifier_InvalidLength_ReturnsError validates error handling for
// invalid hash length requests.
//
// Why: Invalid length parameters must be rejected to prevent undefined behavior
// or confusing output in version strings.
//
// What: Request hashes with invalid lengths (zero, negative, too large), verify
// errors are returned.
func TestGetVCSIdentifier_InvalidLength_ReturnsError(t *testing.T) {
	// Precondition: A git repository with at least one commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	vcs := NewGitVCSDefault()

	tests := []struct {
		length   int
		testName string
	}{
		{0, "zero length invalid"},
		{-1, "negative length invalid"},
		{41, "exceeds max length"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Action: Request hash with invalid length
			_, err := vcs.GetVCSIdentifier(tt.length)

			// Expected: Should return an error
			if err == nil {
				t.Errorf("expected error for length %d", tt.length)
			}
		})
	}
}

// TestGetCommitsSinceTag_NoTags_ReturnsNegativeOne validates behavior when no
// tags exist in the repository.
//
// Why: Repositories without any tags are common during initial development.
// The VCS must handle this gracefully rather than erroring.
//
// What: Create commits without any tags, verify GetCommitsSinceTag() returns -1.
func TestGetCommitsSinceTag_NoTags_ReturnsNegativeOne(t *testing.T) {
	// Precondition: A repository with commits but no tags
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get commits since tag when no tags exist
	vcs := NewGitVCSDefault()
	count, err := vcs.GetCommitsSinceTag()

	// Expected: Should return -1 to indicate no tags exist
	if err != nil {
		t.Fatalf("GetCommitsSinceTag() error: %v", err)
	}

	if count != -1 {
		t.Errorf("expected -1 when no tags exist, got %d", count)
	}
}

// TestGetLastTag_NoTags_ReturnsEmptyString validates behavior when querying
// the last tag in a repository without tags.
//
// Why: New repositories have no tags; tools must handle this gracefully and
// allow determining that versioning should start from scratch.
//
// What: Create commits without tags, verify GetLastTag() returns empty string.
func TestGetLastTag_NoTags_ReturnsEmptyString(t *testing.T) {
	// Precondition: A repository with commits but no tags
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the last tag when no tags exist
	vcs := NewGitVCSDefault()
	tag, err := vcs.GetLastTag()

	// Expected: Should return empty string when no tags exist
	if err != nil {
		t.Fatalf("GetLastTag() error: %v", err)
	}

	if tag != "" {
		t.Errorf("expected empty string when no tags, got '%s'", tag)
	}
}

// TestGetLastTagCommit_NoTags_ReturnsEmptyString validates behavior when
// querying the last tag's commit in a repository without tags.
//
// Why: Consistent behavior with GetLastTag - both should indicate absence of
// tags gracefully without errors.
//
// What: Create commits without tags, verify GetLastTagCommit() returns empty
// string.
func TestGetLastTagCommit_NoTags_ReturnsEmptyString(t *testing.T) {
	// Precondition: A repository with commits but no tags
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the last tag's commit when no tags exist
	vcs := NewGitVCSDefault()
	commit, err := vcs.GetLastTagCommit()

	// Expected: Should return empty string when no tags exist
	if err != nil {
		t.Fatalf("GetLastTagCommit() error: %v", err)
	}

	if commit != "" {
		t.Errorf("expected empty string when no tags, got '%s'", commit)
	}
}

// =============================================================================
// EDGE CASES
// Tests demonstrating boundary conditions and less common scenarios.
// =============================================================================

// TestGetUncommittedChanges_CleanRepo_ReturnsZero validates zero uncommitted
// changes in a clean repository.
//
// Why: Clean state must report zero changes; false positives would incorrectly
// trigger dirty indicators in version strings.
//
// What: Create a commit with no pending changes, verify GetUncommittedChanges()
// returns 0.
func TestGetUncommittedChanges_CleanRepo_ReturnsZero(t *testing.T) {
	// Precondition: A clean repository with all changes committed
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get the count of uncommitted changes
	vcs := NewGitVCSDefault()
	count, err := vcs.GetUncommittedChanges()

	// Expected: Should return 0 for a clean repository
	if err != nil {
		t.Fatalf("GetUncommittedChanges() error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 uncommitted changes, got %d", count)
	}
}

// TestGetCommitsSinceTag_LightweightTag_ReturnsCorrectCount validates commit
// counting works correctly with lightweight (non-annotated) tags.
//
// Why: Git supports both annotated and lightweight tags. Both types must be
// handled correctly for commit distance calculations.
//
// What: Create a lightweight tag, add commits, verify GetCommitsSinceTag()
// returns the correct count.
func TestGetCommitsSinceTag_LightweightTag_ReturnsCorrectCount(t *testing.T) {
	// Precondition: A lightweight tag exists followed by a commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")
	h.CreateLightweightTag("v1.0.0")
	h.CreateCommit("second commit")

	// Action: Get commits since the lightweight tag
	vcs := NewGitVCSDefault()
	count, err := vcs.GetCommitsSinceTag()

	// Expected: Should return 1 for the commit after the lightweight tag
	if err != nil {
		t.Fatalf("GetCommitsSinceTag() error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 commit since lightweight tag, got %d", count)
	}
}

// TestFindGitDir_InGitRepo_ReturnsRoot validates that findGitDir correctly
// locates the repository root from within the repo.
//
// Why: Internal helper function used for repository detection; must correctly
// identify the .git directory location.
//
// What: Call findGitDir from the repository root, verify it returns the root
// directory.
func TestFindGitDir_InGitRepo_ReturnsRoot(t *testing.T) {
	// Precondition: Inside a git repository
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Find the git directory from the repo root
	vcs := NewGitVCSDefault()
	root := vcs.findGitDir(h.dir)

	// Expected: Should return the repository root directory
	expectedRoot, _ := filepath.EvalSymlinks(h.dir)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualRoot != expectedRoot {
		t.Errorf("expected '%s', got '%s'", expectedRoot, actualRoot)
	}
}

// TestFindGitDir_InSubdirectory_ReturnsRoot validates that findGitDir correctly
// locates the repository root from a nested subdirectory.
//
// Why: Users may run versionator from any subdirectory within their project.
// Repository detection must traverse upward to find .git.
//
// What: Create nested subdirectories, call findGitDir from the deepest one,
// verify it returns the repository root.
func TestFindGitDir_InSubdirectory_ReturnsRoot(t *testing.T) {
	// Precondition: Inside a nested subdirectory of a git repository
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Create subdirectory
	subdir := filepath.Join(h.dir, "subdir", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Action: Find the git directory from a nested subdirectory
	vcs := NewGitVCSDefault()
	root := vcs.findGitDir(subdir)

	// Expected: Should return the repository root, not the subdirectory
	expectedRoot, _ := filepath.EvalSymlinks(h.dir)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualRoot != expectedRoot {
		t.Errorf("expected '%s', got '%s'", expectedRoot, actualRoot)
	}
}

// TestFindGitDir_NotInGitRepo_ReturnsEmptyString validates findGitDir behavior
// outside any git repository.
//
// Why: findGitDir must not crash or return incorrect paths when operating
// outside a git repository.
//
// What: Call findGitDir from a non-git directory, verify it returns empty
// string.
func TestFindGitDir_NotInGitRepo_ReturnsEmptyString(t *testing.T) {
	// Precondition: A directory outside any git repository
	dir, err := os.MkdirTemp("", "no-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Action: Find the git directory outside a repo
	vcs := NewGitVCSDefault()
	root := vcs.findGitDir(dir)

	// Expected: Should return empty string when not in a git repo
	if root != "" {
		t.Errorf("expected empty string, got '%s'", root)
	}
}

// TestCreateBranch_MultipleBranches validates creating multiple branches
// successively.
//
// Why: Release management may create multiple branches (e.g., release/v1.0.0,
// release/v1.1.0). Each creation must succeed independently.
//
// What: Create multiple branches, verify all exist via BranchExists().
func TestCreateBranch_MultipleBranches(t *testing.T) {
	// Precondition: A git repository with a commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	vcs := NewGitVCSDefault()

	// Action: Create multiple release branches
	branches := []string{"release/v1.0.0", "release/v1.1.0", "release/v2.0.0"}
	for _, branch := range branches {
		err := vcs.CreateBranch(branch)
		if err != nil {
			t.Fatalf("CreateBranch(%s) error: %v", branch, err)
		}
	}

	// Expected: All branches should exist
	for _, branch := range branches {
		exists, err := vcs.BranchExists(branch)
		if err != nil {
			t.Fatalf("BranchExists(%s) error: %v", branch, err)
		}
		if !exists {
			t.Errorf("expected branch '%s' to exist", branch)
		}
	}
}

// =============================================================================
// MINUTIAE
// Tests covering configuration, environment variables, and utility functions.
// =============================================================================

// TestGetHashLength_Default_ReturnsSeven validates the default hash length
// when no environment variable is set.
//
// Why: Default configuration must be predictable and well-documented. The
// default of 7 matches git's own default for short hashes.
//
// What: Unset the environment variable, verify GetHashLength() returns 7.
func TestGetHashLength_Default_ReturnsSeven(t *testing.T) {
	// Precondition: No environment variable set
	os.Unsetenv("VERSIONATOR_HASH_LENGTH")

	// Action: Get the hash length
	length := GetHashLength()

	// Expected: Should return default of 7
	if length != 7 {
		t.Errorf("expected default length 7, got %d", length)
	}
}

// TestGetHashLength_FromEnv_ReturnsEnvValue validates that hash length can be
// configured via environment variable.
//
// Why: Environment-based configuration allows customization without code
// changes, useful for CI/CD environments.
//
// What: Set the environment variable to 12, verify GetHashLength() returns 12.
func TestGetHashLength_FromEnv_ReturnsEnvValue(t *testing.T) {
	// Precondition: Environment variable set to a valid value
	os.Setenv("VERSIONATOR_HASH_LENGTH", "12")
	defer os.Unsetenv("VERSIONATOR_HASH_LENGTH")

	// Action: Get the hash length
	length := GetHashLength()

	// Expected: Should return the environment variable value
	if length != 12 {
		t.Errorf("expected length 12 from env, got %d", length)
	}
}

// TestGetHashLength_InvalidEnv_ReturnsDefault validates fallback to default
// when environment variable contains invalid values.
//
// Why: Invalid configuration must not break the tool; graceful fallback to
// defaults ensures stability.
//
// What: Set various invalid environment values, verify GetHashLength() returns
// the default 7 for each.
func TestGetHashLength_InvalidEnv_ReturnsDefault(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
	}{
		{"non-numeric", "abc"},
		{"too low", "0"},
		{"too high", "41"},
		{"negative", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Invalid environment variable value
			os.Setenv("VERSIONATOR_HASH_LENGTH", tt.envValue)
			defer os.Unsetenv("VERSIONATOR_HASH_LENGTH")

			// Action: Get the hash length
			length := GetHashLength()

			// Expected: Should fallback to default of 7
			if length != 7 {
				t.Errorf("expected default 7 for invalid env '%s', got %d", tt.envValue, length)
			}
		})
	}
}

// TestGetHashLengthFromConfig_ValidConfig_ReturnsConfigValue validates that
// hash length from config takes precedence over environment.
//
// Why: Explicit config file settings should override environment defaults for
// project-specific configuration.
//
// What: Pass a valid config value, verify GetHashLengthFromConfig() returns it.
func TestGetHashLengthFromConfig_ValidConfig_ReturnsConfigValue(t *testing.T) {
	// Precondition: No conflicting environment variable
	os.Unsetenv("VERSIONATOR_HASH_LENGTH")

	// Action: Get hash length from config
	length := GetHashLengthFromConfig(10)

	// Expected: Should return the config value
	if length != 10 {
		t.Errorf("expected 10 from config, got %d", length)
	}
}

// TestGetHashLengthFromConfig_InvalidConfig_FallsBackToEnvOrDefault validates
// fallback chain: config -> env -> default.
//
// Why: The configuration hierarchy must be predictable: explicit config wins,
// then environment, then hardcoded default.
//
// What: Pass invalid config values with various env settings, verify correct
// fallback behavior.
func TestGetHashLengthFromConfig_InvalidConfig_FallsBackToEnvOrDefault(t *testing.T) {
	tests := []struct {
		name        string
		configValue int
		envValue    string
		expected    int
	}{
		{"zero config, no env", 0, "", 7},
		{"negative config, no env", -1, "", 7},
		{"too high config, no env", 50, "", 7},
		{"zero config, valid env", 0, "15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Precondition: Set up config and env as specified
			if tt.envValue != "" {
				os.Setenv("VERSIONATOR_HASH_LENGTH", tt.envValue)
			} else {
				os.Unsetenv("VERSIONATOR_HASH_LENGTH")
			}
			defer os.Unsetenv("VERSIONATOR_HASH_LENGTH")

			// Action: Get hash length from config
			length := GetHashLengthFromConfig(tt.configValue)

			// Expected: Should follow fallback chain
			if length != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, length)
			}
		})
	}
}

// TestPackageLevel_IsGitRepository_InRepo_ReturnsTrue validates the package-
// level convenience function for repository detection.
//
// Why: Package-level functions provide a simplified API; they must correctly
// delegate to the underlying VCS implementation.
//
// What: Inside a git repo, verify IsGitRepository() returns true.
func TestPackageLevel_IsGitRepository_InRepo_ReturnsTrue(t *testing.T) {
	// Precondition: Inside a git repository with a commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Force re-registration of VCS
	vcs := NewGitVCSDefault()
	if !vcs.IsRepository() {
		t.Skip("VCS not detecting repository - likely registration issue")
	}

	// Action: Use package-level function
	isRepo := IsGitRepository()

	// Expected: Should return true
	if !isRepo {
		t.Error("expected IsGitRepository() to return true")
	}
}

// TestPackageLevel_GetGitShortHash_ReturnsHash validates the package-level
// convenience function for hash retrieval.
//
// Why: Package-level functions simplify common operations; they must work
// correctly without requiring VCS instance management.
//
// What: Inside a git repo, verify GetGitShortHash() returns a hash of the
// requested length.
func TestPackageLevel_GetGitShortHash_ReturnsHash(t *testing.T) {
	// Precondition: Inside a git repository with a commit
	h := NewTestHelper(t)
	defer h.Cleanup()
	h.CreateCommit("initial commit")

	// Action: Get short hash using package-level function
	hash, err := GetGitShortHash(7)

	// Expected: Should return a 7-character hash
	if err != nil {
		t.Fatalf("GetGitShortHash() error: %v", err)
	}

	if len(hash) != 7 {
		t.Errorf("expected hash length 7, got %d", len(hash))
	}
}
