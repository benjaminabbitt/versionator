package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/benjaminabbitt/versionator/internal/plugin"
	"github.com/benjaminabbitt/versionator/internal/vcs"
)

// errStopIteration is a sentinel error used to break out of ForEach iteration
// This is a standard Go pattern (similar to io.EOF) for signaling iteration completion
var errStopIteration = errors.New("stop iteration")

// GitVersionControlSystem implements VersionControlSystem for Git
type GitVersionControlSystem struct {
	repoRoot string
}

// NewGitVCS creates a new GitVersionControlSystem
func NewGitVCS() *GitVersionControlSystem {
	return &GitVersionControlSystem{}
}

// Name returns "git"
func (g *GitVersionControlSystem) Name() string {
	return "git"
}

// Types returns the set of plugin types this VCS implements
func (g *GitVersionControlSystem) Types() plugin.PluginTypeSet {
	return plugin.NewPluginTypeSet(plugin.TypeVCS, plugin.TypeTemplateProvider)
}

// GetTemplateVariables returns git-specific template variables
func (g *GitVersionControlSystem) GetTemplateVariables(context map[string]string) map[string]string {
	shortHash := context["ShortHash"]
	if shortHash == "" {
		return nil
	}

	return map[string]string{
		"GitShortHash": "git." + shortHash,
		"ShaShortHash": "sha." + shortHash,
	}
}

// IsRepository checks if we're in a git repository
func (g *GitVersionControlSystem) IsRepository() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	root := g.findGitDir(cwd)
	if root != "" {
		g.repoRoot = root
		return true
	}
	return false
}

// GetRepositoryRoot returns the root directory of the git repository
func (g *GitVersionControlSystem) GetRepositoryRoot() (string, error) {
	if g.repoRoot != "" {
		return g.repoRoot, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	root := g.findGitDir(cwd)
	if root == "" {
		return "", fmt.Errorf("not a git repository")
	}

	g.repoRoot = root
	return root, nil
}

// IsWorkingDirectoryClean checks if there are no uncommitted changes
func (g *GitVersionControlSystem) IsWorkingDirectoryClean() (bool, error) {
	repo, err := g.openRepository()
	if err != nil {
		return false, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get working tree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get git status: %w", err)
	}

	return status.IsClean(), nil
}

// GetVCSIdentifier returns a short hash of the current commit
func (g *GitVersionControlSystem) GetVCSIdentifier(length int) (string, error) {
	if length < 1 || length > 40 {
		return "", fmt.Errorf("invalid hash length: %d (must be between 1 and 40)", length)
	}

	repo, err := g.openRepository()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	fullHash := commit.Hash.String()
	if length > len(fullHash) {
		length = len(fullHash)
	}

	return fullHash[:length], nil
}

// CreateTag creates a git tag
func (g *GitVersionControlSystem) CreateTag(tagName, message string) error {
	repo, err := g.openRepository()
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Get commit to use author info for tagger
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	_, err = repo.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Message: message,
		Tagger: &object.Signature{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// TagExists checks if a tag exists
func (g *GitVersionControlSystem) TagExists(tagName string) (bool, error) {
	repo, err := g.openRepository()
	if err != nil {
		return false, err
	}

	tags, err := repo.Tags()
	if err != nil {
		return false, fmt.Errorf("failed to get tags: %w", err)
	}

	exists := false
	err = tags.ForEach(func(tag *plumbing.Reference) error {
		if tag.Name().Short() == tagName {
			exists = true
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return exists, nil
}

// GetBranchName returns the current branch name
func (g *GitVersionControlSystem) GetBranchName() (string, error) {
	repo, err := g.openRepository()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Check if HEAD is a branch reference
	if ref.Name().IsBranch() {
		return ref.Name().Short(), nil
	}

	// Detached HEAD state - return empty string or commit hash
	return "", nil
}

// GetCommitDate returns the date of the current commit
func (g *GitVersionControlSystem) GetCommitDate() (time.Time, error) {
	repo, err := g.openRepository()
	if err != nil {
		return time.Time{}, err
	}

	ref, err := repo.Head()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get commit object: %w", err)
	}

	return commit.Author.When.UTC(), nil
}

// GetCommitsSinceTag returns the number of commits since the most recent tag
// Returns 0 if on a tagged commit, -1 if no tags exist
func (g *GitVersionControlSystem) GetCommitsSinceTag() (int, error) {
	repo, err := g.openRepository()
	if err != nil {
		return 0, err
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return 0, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return 0, fmt.Errorf("failed to get commit object: %w", err)
	}

	// Build a map of tag hashes for quick lookup
	tagHashes := make(map[plumbing.Hash]bool)
	tags, err := repo.Tags()
	if err != nil {
		return 0, fmt.Errorf("failed to get tags: %w", err)
	}

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		// Handle both lightweight and annotated tags
		tagHashes[ref.Hash()] = true

		// For annotated tags, also add the target commit hash
		tagObj, err := repo.TagObject(ref.Hash())
		if err == nil {
			tagHashes[tagObj.Target] = true
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to iterate tags: %w", err)
	}

	// No tags exist
	if len(tagHashes) == 0 {
		return -1, nil
	}

	// Walk commits from HEAD until we find a tagged commit
	count := 0
	commitIter, err := repo.Log(&git.LogOptions{From: headCommit.Hash})
	if err != nil {
		return 0, fmt.Errorf("failed to get commit log: %w", err)
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		if tagHashes[c.Hash] {
			return errStopIteration // Use sentinel error to break iteration
		}
		count++
		return nil
	})

	// If we broke out because we found a tag, return the count
	if errors.Is(err, errStopIteration) {
		return count, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to iterate commits: %w", err)
	}

	// No tagged ancestor found (shouldn't happen if tags exist)
	return count, nil
}

// GetUncommittedChanges returns the count of uncommitted changes (staged + unstaged + untracked)
func (g *GitVersionControlSystem) GetUncommittedChanges() (int, error) {
	repo, err := g.openRepository()
	if err != nil {
		return 0, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return 0, fmt.Errorf("failed to get working tree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return 0, fmt.Errorf("failed to get git status: %w", err)
	}

	return len(status), nil
}

// GetLastTag returns the most recent semver tag
// Returns empty string if no tags exist
func (g *GitVersionControlSystem) GetLastTag() (string, error) {
	repo, err := g.openRepository()
	if err != nil {
		return "", err
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	// Build a map of commit hash -> tag name
	tagMap := make(map[plumbing.Hash]string)
	tags, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()

		// For annotated tags, get the target commit
		tagObj, err := repo.TagObject(ref.Hash())
		if err == nil {
			tagMap[tagObj.Target] = tagName
		} else {
			// Lightweight tag - ref.Hash() is the commit
			tagMap[ref.Hash()] = tagName
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate tags: %w", err)
	}

	if len(tagMap) == 0 {
		return "", nil
	}

	// Walk commits from HEAD until we find a tagged commit
	commitIter, err := repo.Log(&git.LogOptions{From: headCommit.Hash})
	if err != nil {
		return "", fmt.Errorf("failed to get commit log: %w", err)
	}

	var lastTag string
	err = commitIter.ForEach(func(c *object.Commit) error {
		if tag, ok := tagMap[c.Hash]; ok {
			lastTag = tag
			return errStopIteration // Use sentinel error to break iteration
		}
		return nil
	})

	if errors.Is(err, errStopIteration) {
		return lastTag, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to iterate commits: %w", err)
	}

	return lastTag, nil
}

// GetLastTagCommit returns the SHA of the commit the last tag points to
func (g *GitVersionControlSystem) GetLastTagCommit() (string, error) {
	repo, err := g.openRepository()
	if err != nil {
		return "", err
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	// Build a map of commit hash -> tag exists
	tagCommits := make(map[plumbing.Hash]bool)
	tags, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		// For annotated tags, get the target commit
		tagObj, err := repo.TagObject(ref.Hash())
		if err == nil {
			tagCommits[tagObj.Target] = true
		} else {
			// Lightweight tag
			tagCommits[ref.Hash()] = true
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate tags: %w", err)
	}

	if len(tagCommits) == 0 {
		return "", nil
	}

	// Walk commits from HEAD until we find a tagged commit
	commitIter, err := repo.Log(&git.LogOptions{From: headCommit.Hash})
	if err != nil {
		return "", fmt.Errorf("failed to get commit log: %w", err)
	}

	var lastTagCommit string
	err = commitIter.ForEach(func(c *object.Commit) error {
		if tagCommits[c.Hash] {
			lastTagCommit = c.Hash.String()
			return errStopIteration // Use sentinel error to break iteration
		}
		return nil
	})

	if errors.Is(err, errStopIteration) {
		return lastTagCommit, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to iterate commits: %w", err)
	}

	return lastTagCommit, nil
}

// GetCommitAuthor returns the name of the current commit's author
func (g *GitVersionControlSystem) GetCommitAuthor() (string, error) {
	repo, err := g.openRepository()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	return commit.Author.Name, nil
}

// GetCommitAuthorEmail returns the email of the current commit's author
func (g *GitVersionControlSystem) GetCommitAuthorEmail() (string, error) {
	repo, err := g.openRepository()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	return commit.Author.Email, nil
}

// Helper methods
func (g *GitVersionControlSystem) findGitDir(startPath string) string {
	currentPath := startPath

	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return currentPath
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break
		}
		currentPath = parentPath
	}

	return ""
}

func (g *GitVersionControlSystem) openRepository() (*git.Repository, error) {
	root, err := g.GetRepositoryRoot()
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(root)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return repo, nil
}

// GetHashLength returns the configured hash length from config file or environment variable
// Priority: 1) Config file, 2) VERSIONATOR_HASH_LENGTH env var, 3) Default (7)
func GetHashLength() int {
	const defaultLength = 7
	const envVar = "VERSIONATOR_HASH_LENGTH"

	// First check environment variable for backward compatibility
	if lengthStr := os.Getenv(envVar); lengthStr != "" {
		if length, err := strconv.Atoi(lengthStr); err == nil && length >= 1 && length <= 40 {
			return length
		}
	}

	return defaultLength
}

// GetHashLengthFromConfig returns hash length from config, with fallback to environment/default
func GetHashLengthFromConfig(configHashLength int) int {
	// If config has a valid hash length, use it
	if configHashLength >= 1 && configHashLength <= 40 {
		return configHashLength
	}

	// Fall back to environment variable or default
	return GetHashLength()
}

// GetGitShortHash returns the short hash of the current HEAD commit with specified length
func GetGitShortHash(hashLength int) (string, error) {
	gitVCS := vcs.GetVCS("git")
	if gitVCS == nil || !gitVCS.IsRepository() {
		return "", fmt.Errorf("not in a git repository")
	}

	return gitVCS.GetVCSIdentifier(hashLength)
}

// IsGitRepository checks if we're in a git repository
func IsGitRepository() bool {
	gitVCS := vcs.GetVCS("git")
	if gitVCS == nil {
		return false
	}
	return gitVCS.IsRepository()
}

// IsWorkingDirectoryClean checks if the git working directory is clean (no uncommitted changes)
func IsWorkingDirectoryClean() (bool, error) {
	gitVCS := vcs.GetVCS("git")
	if gitVCS == nil || !gitVCS.IsRepository() {
		return false, fmt.Errorf("not a git repository")
	}

	return gitVCS.IsWorkingDirectoryClean()
}

// CreateTag creates a git tag with the specified name and message
func CreateTag(tagName, message string) error {
	gitVCS := vcs.GetVCS("git")
	if gitVCS == nil || !gitVCS.IsRepository() {
		return fmt.Errorf("not a git repository")
	}

	return gitVCS.CreateTag(tagName, message)
}

// TagExists checks if a git tag with the specified name already exists
func TagExists(tagName string) (bool, error) {
	gitVCS := vcs.GetVCS("git")
	if gitVCS == nil || !gitVCS.IsRepository() {
		return false, fmt.Errorf("not a git repository")
	}

	return gitVCS.TagExists(tagName)
}

// Auto-registration as both VCS and plugin
func init() {
	gitVCS := NewGitVCS()
	vcs.RegisterVCS(gitVCS)
	plugin.RegisterTemplateProvider(gitVCS)
}
