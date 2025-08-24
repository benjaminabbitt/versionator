package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
)

// GitVersionControlSystem implements VersionControlSystem for Git
type GitVersionControlSystem struct {
	repoRoot string
	fs       afero.Fs
}

// NewGitVCS creates a new GitVersionControlSystem
func NewGitVCS(fs afero.Fs) *GitVersionControlSystem {
	return &GitVersionControlSystem{fs: fs}
}

// NewGitVCSDefault creates a new GitVersionControlSystem using the OS filesystem
func NewGitVCSDefault() *GitVersionControlSystem {
	return NewGitVCS(afero.NewOsFs())
}

// Name returns "git"
func (g *GitVersionControlSystem) Name() string {
	return "git"
}

// IsRepository checks if we're in a git repository
func (g *GitVersionControlSystem) IsRepository() bool {
	cwd, err := g.getWorkingDir()
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

// getWorkingDir gets the current working directory using the filesystem interface
func (g *GitVersionControlSystem) getWorkingDir() (string, error) {
	// For testing scenarios with memory filesystem, we start from root "/"
	// For OS filesystem, we get the actual working directory using afero's capabilities
	if _, ok := g.fs.(*afero.MemMapFs); ok {
		return "/", nil
	}
	
	// For OS filesystem, we still need to use os.Getwd() since afero doesn't provide
	// a direct equivalent for getting current working directory
	return os.Getwd()
}

// GetRepositoryRoot returns the root directory of the git repository
func (g *GitVersionControlSystem) GetRepositoryRoot() (string, error) {
	if g.repoRoot != "" {
		return g.repoRoot, nil
	}

	cwd, err := g.getWorkingDir()
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

	_, err = repo.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Message: message,
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

// Helper methods
func (g *GitVersionControlSystem) findGitDir(startPath string) string {
	currentPath := startPath

	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := g.fs.Stat(gitPath); err == nil && info.IsDir() {
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
func (g *GitVersionControlSystem) GetHashLength() int {
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


