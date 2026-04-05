package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/benjaminabbitt/versionator/internal/plugin"
	"github.com/benjaminabbitt/versionator/internal/vcs"
)

// errStopIteration is a sentinel error used to break out of ForEach iteration
// This is a standard Go pattern (similar to io.EOF) for signaling iteration completion
var errStopIteration = errors.New("stop iteration")

// DefaultMaxCommitDepth is the maximum number of commits to walk when searching for tags
const DefaultMaxCommitDepth = 10000

// GitVersionControlSystem implements VersionControlSystem for Git
type GitVersionControlSystem struct {
	repoRoot   string
	repo       Repository       // cached repository (interface)
	repoOpener RepositoryOpener // injected repository opener
	tagInfo    *TagInfo         // cached tag information
	tagInfoErr error            // cached error from tag info fetch
}

// TagInfo holds pre-computed tag-related information from a single walk
type TagInfo struct {
	CommitsSinceTag   int    // -1 if no tags exist
	LastTagName       string // empty if no tags
	LastTagCommitHash string // empty if no tags
}

// NewGitVCS creates a new GitVersionControlSystem with a custom repository opener.
// Use this constructor for testing with mock repositories.
func NewGitVCS(opener RepositoryOpener) *GitVersionControlSystem {
	return &GitVersionControlSystem{
		repoOpener: opener,
	}
}

// NewGitVCSDefault creates a new GitVersionControlSystem with the default go-git opener.
func NewGitVCSDefault() *GitVersionControlSystem {
	return NewGitVCS(DefaultRepositoryOpener)
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
// Files matching gitignore patterns are not considered when checking for cleanliness
func (g *GitVersionControlSystem) IsWorkingDirectoryClean() (bool, error) {
	// Reuse GetDirtyFiles which already handles gitignore filtering
	files, err := g.GetDirtyFiles()
	if err != nil {
		return false, err
	}
	return len(files) == 0, nil
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
			return errStopIteration // Early exit once found
		}
		return nil
	})

	if errors.Is(err, errStopIteration) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return exists, nil
}

// CreateBranch creates a branch with the specified name from the current HEAD
func (g *GitVersionControlSystem) CreateBranch(branchName string) error {
	repo, err := g.openRepository()
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Create the branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())

	err = repo.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// BranchExists checks if a branch with the specified name exists
func (g *GitVersionControlSystem) BranchExists(branchName string) (bool, error) {
	repo, err := g.openRepository()
	if err != nil {
		return false, err
	}

	branches, err := repo.Branches()
	if err != nil {
		return false, fmt.Errorf("failed to get branches: %w", err)
	}

	exists := false
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		if branch.Name().Short() == branchName {
			exists = true
			return errStopIteration // Early exit once found
		}
		return nil
	})

	if errors.Is(err, errStopIteration) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to iterate branches: %w", err)
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
	info, err := g.getTagInfo()
	if err != nil {
		return 0, err
	}
	return info.CommitsSinceTag, nil
}

// GetUncommittedChanges returns the count of uncommitted changes (staged + unstaged + untracked)
// Files matching gitignore patterns are excluded from the count
func (g *GitVersionControlSystem) GetUncommittedChanges() (int, error) {
	// Reuse GetDirtyFiles which already handles gitignore filtering
	files, err := g.GetDirtyFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// GetLastTag returns the most recent semver tag
// Returns empty string if no tags exist
func (g *GitVersionControlSystem) GetLastTag() (string, error) {
	info, err := g.getTagInfo()
	if err != nil {
		return "", err
	}
	return info.LastTagName, nil
}

// GetLastTagCommit returns the SHA of the commit the last tag points to
func (g *GitVersionControlSystem) GetLastTagCommit() (string, error) {
	info, err := g.getTagInfo()
	if err != nil {
		return "", err
	}
	return info.LastTagCommitHash, nil
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

// GetCommitMessagesSinceTag returns all commit messages since the most recent tag
// Returns commit messages newest first, empty slice if on tagged commit or no tags
func (g *GitVersionControlSystem) GetCommitMessagesSinceTag() ([]string, error) {
	info, err := g.getTagInfo()
	if err != nil {
		return nil, err
	}

	// No commits since tag (on tagged commit) or no tags exist with 0 commits
	if info.CommitsSinceTag <= 0 {
		return []string{}, nil
	}

	repo, err := g.openRepository()
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	messages := make([]string, 0, info.CommitsSinceTag)
	count := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		if count >= info.CommitsSinceTag {
			return errStopIteration
		}
		messages = append(messages, c.Message)
		count++
		return nil
	})

	if err != nil && !errors.Is(err, errStopIteration) {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return messages, nil
}

// GetDirtyFiles returns the list of files with uncommitted changes.
// Files matching gitignore patterns are excluded from the result.
// Uses go-git's worktree.Status() with gitignore filtering.
// Falls back to git CLI if go-git fails (e.g., permission errors on unreadable gitignored files).
func (g *GitVersionControlSystem) GetDirtyFiles() ([]string, error) {
	files, err := g.getDirtyFilesGoGit()
	if err != nil {
		// Fall back to git CLI — handles permission errors from unreadable gitignored files
		return g.getDirtyFilesGitCLI()
	}
	return files, nil
}

// getDirtyFilesGoGit uses go-git's worktree.Status() with gitignore filtering.
func (g *GitVersionControlSystem) getDirtyFilesGoGit() ([]string, error) {
	repo, err := g.openRepository()
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get working tree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	matcher, _ := g.getGitignoreMatcher()

	var files []string
	for file := range status {
		if g.isFileIgnored(matcher, file) {
			continue
		}
		files = append(files, file)
	}
	return files, nil
}

// getDirtyFilesGitCLI falls back to git CLI for status.
// git status --porcelain natively respects .gitignore and handles unreadable files.
func (g *GitVersionControlSystem) getDirtyFilesGitCLI() ([]string, error) {
	root, err := g.GetRepositoryRoot()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// Porcelain format: XY filename (first 3 chars are status + space)
		if len(line) > 3 {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	return files, nil
}

// CommitFiles stages and commits the specified files with the given message
func (g *GitVersionControlSystem) CommitFiles(files []string, message string) error {
	repo, err := g.openRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %w", err)
	}

	// Stage each file
	for _, file := range files {
		_, err := worktree.Add(file)
		if err != nil {
			return fmt.Errorf("failed to stage file %s: %w", file, err)
		}
	}

	// Get author info from last commit
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	// Create commit
	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// AmendCommit stages the specified files and amends the last commit
func (g *GitVersionControlSystem) AmendCommit(files []string) error {
	repo, err := g.openRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %w", err)
	}

	// Stage each file
	for _, file := range files {
		_, err := worktree.Add(file)
		if err != nil {
			return fmt.Errorf("failed to stage file %s: %w", file, err)
		}
	}

	// Get the last commit to preserve its message and author
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	lastCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	// Amend the commit (reuse message, update author timestamp)
	_, err = worktree.Commit(lastCommit.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  lastCommit.Author.Name,
			Email: lastCommit.Author.Email,
			When:  time.Now(),
		},
		Amend: true,
	})
	if err != nil {
		return fmt.Errorf("failed to amend commit: %w", err)
	}

	return nil
}

// GetHooksPath returns the path to the git hooks directory
func (g *GitVersionControlSystem) GetHooksPath() (string, error) {
	root, err := g.GetRepositoryRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".git", "hooks"), nil
}

// Helper methods

// getGitignoreMatcher creates a gitignore matcher from repository patterns
// It reads patterns from .gitignore files and .git/info/exclude
func (g *GitVersionControlSystem) getGitignoreMatcher() (gitignore.Matcher, error) {
	root, err := g.GetRepositoryRoot()
	if err != nil {
		return nil, err
	}

	// Create a filesystem interface from the repository root
	fs := osfs.New(root)

	// Read gitignore patterns from repository
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		// If we can't read patterns, return nil matcher (no filtering)
		return nil, nil
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	return gitignore.NewMatcher(patterns), nil
}

// isFileIgnored checks if a file path should be ignored according to gitignore rules
func (g *GitVersionControlSystem) isFileIgnored(matcher gitignore.Matcher, filePath string) bool {
	if matcher == nil {
		return false
	}

	// Split path into components for the matcher
	pathComponents := strings.Split(filePath, string(filepath.Separator))

	// Check if the file matches gitignore patterns
	// We assume it's not a directory since we're checking file status
	return matcher.Match(pathComponents, false)
}

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

func (g *GitVersionControlSystem) openRepository() (Repository, error) {
	// Return cached repo if available
	if g.repo != nil {
		return g.repo, nil
	}

	root, err := g.GetRepositoryRoot()
	if err != nil {
		return nil, err
	}

	repo, err := g.repoOpener(root)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Cache for future calls
	g.repo = repo
	return repo, nil
}

// getTagInfo performs a single walk to gather all tag-related information
// Results are cached for subsequent calls
func (g *GitVersionControlSystem) getTagInfo() (*TagInfo, error) {
	// Return cached result if available
	if g.tagInfo != nil || g.tagInfoErr != nil {
		return g.tagInfo, g.tagInfoErr
	}

	info, err := g.computeTagInfo()
	g.tagInfo = info
	g.tagInfoErr = err
	return info, err
}

// computeTagInfo does the actual work of walking commits to find tag info
func (g *GitVersionControlSystem) computeTagInfo() (*TagInfo, error) {
	repo, err := g.openRepository()
	if err != nil {
		return nil, err
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	// Build a map of commit hash -> tag name (for both lightweight and annotated tags)
	tagMap := make(map[plumbing.Hash]string)
	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
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
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	// No tags exist
	if len(tagMap) == 0 {
		return &TagInfo{CommitsSinceTag: -1}, nil
	}

	// Walk commits from HEAD until we find a tagged commit (with depth limit)
	commitIter, err := repo.Log(&git.LogOptions{From: headCommit.Hash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	count := 0
	var result *TagInfo

	err = commitIter.ForEach(func(c *object.Commit) error {
		if tagName, ok := tagMap[c.Hash]; ok {
			result = &TagInfo{
				CommitsSinceTag:   count,
				LastTagName:       tagName,
				LastTagCommitHash: c.Hash.String(),
			}
			return errStopIteration
		}
		count++

		// Depth limit to avoid walking huge histories
		if count >= DefaultMaxCommitDepth {
			return errStopIteration
		}
		return nil
	})

	if result != nil {
		return result, nil
	}

	// Hit depth limit or no tagged ancestor found
	if errors.Is(err, errStopIteration) || err == nil {
		return &TagInfo{CommitsSinceTag: count}, nil
	}

	return nil, fmt.Errorf("failed to iterate commits: %w", err)
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

// PushTag pushes a tag to the remote repository
func (g *GitVersionControlSystem) PushTag(tagName string) error {
	root, err := g.GetRepositoryRoot()
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "push", "origin", tagName)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push tag: %w: %s", err, string(output))
	}
	return nil
}

// PushBranch pushes a branch to the remote repository
func (g *GitVersionControlSystem) PushBranch(branchName string) error {
	root, err := g.GetRepositoryRoot()
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push branch: %w: %s", err, string(output))
	}
	return nil
}

// Auto-registration as both VCS and plugin
func init() {
	gitVCS := NewGitVCSDefault()
	vcs.RegisterVCS(gitVCS)
	plugin.RegisterTemplateProvider(gitVCS)
}
