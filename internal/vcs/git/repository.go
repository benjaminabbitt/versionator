package git

import (
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Repository abstracts git repository operations for testability.
// This interface wraps go-git's *git.Repository to enable mocking in tests.
type Repository interface {
	// Head returns the HEAD reference
	Head() (*plumbing.Reference, error)

	// CommitObject returns a commit object by hash
	CommitObject(h plumbing.Hash) (*object.Commit, error)

	// CreateTag creates an annotated tag
	CreateTag(name string, hash plumbing.Hash, opts *git.CreateTagOptions) (*plumbing.Reference, error)

	// Tags returns an iterator over all tags
	Tags() (storer.ReferenceIter, error)

	// Branches returns an iterator over all branches
	Branches() (storer.ReferenceIter, error)

	// TagObject returns a tag object by hash (for annotated tags)
	TagObject(h plumbing.Hash) (*object.Tag, error)

	// Log returns a commit log iterator
	Log(opts *git.LogOptions) (object.CommitIter, error)

	// SetReference stores a reference (used for branch creation)
	SetReference(ref *plumbing.Reference) error

	// Worktree returns the repository worktree
	Worktree() (Worktree, error)
}

// Worktree abstracts git worktree operations for testability.
type Worktree interface {
	// Status returns the working tree status
	Status() (git.Status, error)

	// Add stages a file
	Add(path string) (plumbing.Hash, error)

	// Commit creates a commit with the staged changes
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
}

// RepositoryOpener is a function that opens a git repository at the given path.
type RepositoryOpener func(path string) (Repository, error)

// DefaultRepositoryOpener opens a git repository using go-git's PlainOpen.
func DefaultRepositoryOpener(path string) (Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &GoGitRepository{repo: repo}, nil
}

// GoGitRepository wraps go-git's *git.Repository to implement the Repository interface.
type GoGitRepository struct {
	repo *git.Repository
}

func (r *GoGitRepository) Head() (*plumbing.Reference, error) {
	return r.repo.Head()
}

func (r *GoGitRepository) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	return r.repo.CommitObject(h)
}

func (r *GoGitRepository) CreateTag(name string, hash plumbing.Hash, opts *git.CreateTagOptions) (*plumbing.Reference, error) {
	return r.repo.CreateTag(name, hash, opts)
}

func (r *GoGitRepository) Tags() (storer.ReferenceIter, error) {
	return r.repo.Tags()
}

func (r *GoGitRepository) Branches() (storer.ReferenceIter, error) {
	return r.repo.Branches()
}

func (r *GoGitRepository) TagObject(h plumbing.Hash) (*object.Tag, error) {
	return r.repo.TagObject(h)
}

func (r *GoGitRepository) Log(opts *git.LogOptions) (object.CommitIter, error) {
	return r.repo.Log(opts)
}

func (r *GoGitRepository) SetReference(ref *plumbing.Reference) error {
	return r.repo.Storer.SetReference(ref)
}

func (r *GoGitRepository) Worktree() (Worktree, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, err
	}
	return &GoGitWorktree{wt: wt}, nil
}

// GoGitWorktree wraps go-git's *git.Worktree to implement the Worktree interface.
type GoGitWorktree struct {
	wt *git.Worktree
}

func (w *GoGitWorktree) Status() (git.Status, error) {
	return w.wt.Status()
}

func (w *GoGitWorktree) Add(path string) (plumbing.Hash, error) {
	return w.wt.Add(path)
}

func (w *GoGitWorktree) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	return w.wt.Commit(msg, opts)
}

// CommitInfo holds extracted information about a commit for testing purposes.
type CommitInfo struct {
	Hash        plumbing.Hash
	AuthorName  string
	AuthorEmail string
	AuthorTime  time.Time
	Message     string
}

// ExtractCommitInfo extracts relevant fields from a commit object.
func ExtractCommitInfo(c *object.Commit) CommitInfo {
	return CommitInfo{
		Hash:        c.Hash,
		AuthorName:  c.Author.Name,
		AuthorEmail: c.Author.Email,
		AuthorTime:  c.Author.When,
		Message:     c.Message,
	}
}
