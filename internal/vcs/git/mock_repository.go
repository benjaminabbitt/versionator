package git

import (
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// MockRepository is a test double for the Repository interface.
// It allows configuring return values for each method.
type MockRepository struct {
	// Head configuration
	HeadRef *plumbing.Reference
	HeadErr error

	// CommitObject configuration
	Commits    map[plumbing.Hash]*object.Commit
	CommitErr  error

	// CreateTag configuration
	CreateTagRef *plumbing.Reference
	CreateTagErr error

	// Tags configuration
	TagRefs []plumbing.Reference
	TagsErr error

	// Branches configuration
	BranchRefs []plumbing.Reference
	BranchesErr error

	// TagObject configuration
	TagObjects map[plumbing.Hash]*object.Tag
	TagObjErr  error

	// Log configuration
	LogCommits []*object.Commit
	LogErr     error

	// SetReference configuration
	SetRefErr error
	SetRefCalled bool
	SetRefArg *plumbing.Reference

	// Worktree configuration
	MockWorktree *MockWorktree
	WorktreeErr  error
}

// NewMockRepository creates a new MockRepository with sensible defaults.
func NewMockRepository() *MockRepository {
	return &MockRepository{
		Commits:    make(map[plumbing.Hash]*object.Commit),
		TagObjects: make(map[plumbing.Hash]*object.Tag),
	}
}

func (m *MockRepository) Head() (*plumbing.Reference, error) {
	return m.HeadRef, m.HeadErr
}

func (m *MockRepository) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	if m.CommitErr != nil {
		return nil, m.CommitErr
	}
	if commit, ok := m.Commits[h]; ok {
		return commit, nil
	}
	return nil, plumbing.ErrObjectNotFound
}

func (m *MockRepository) CreateTag(name string, hash plumbing.Hash, opts *git.CreateTagOptions) (*plumbing.Reference, error) {
	return m.CreateTagRef, m.CreateTagErr
}

func (m *MockRepository) Tags() (storer.ReferenceIter, error) {
	if m.TagsErr != nil {
		return nil, m.TagsErr
	}
	return &mockRefIter{refs: m.TagRefs}, nil
}

func (m *MockRepository) Branches() (storer.ReferenceIter, error) {
	if m.BranchesErr != nil {
		return nil, m.BranchesErr
	}
	return &mockRefIter{refs: m.BranchRefs}, nil
}

func (m *MockRepository) TagObject(h plumbing.Hash) (*object.Tag, error) {
	if m.TagObjErr != nil {
		return nil, m.TagObjErr
	}
	if tag, ok := m.TagObjects[h]; ok {
		return tag, nil
	}
	return nil, plumbing.ErrObjectNotFound
}

func (m *MockRepository) Log(opts *git.LogOptions) (object.CommitIter, error) {
	if m.LogErr != nil {
		return nil, m.LogErr
	}
	return &mockCommitIter{commits: m.LogCommits}, nil
}

func (m *MockRepository) SetReference(ref *plumbing.Reference) error {
	m.SetRefCalled = true
	m.SetRefArg = ref
	return m.SetRefErr
}

func (m *MockRepository) Worktree() (Worktree, error) {
	if m.WorktreeErr != nil {
		return nil, m.WorktreeErr
	}
	return m.MockWorktree, nil
}

// MockWorktree is a test double for the Worktree interface.
type MockWorktree struct {
	StatusResult git.Status
	StatusErr    error

	AddHash plumbing.Hash
	AddErr  error
	AddCalls []string

	CommitHash plumbing.Hash
	CommitErr  error
	CommitCalls []struct{
		Msg  string
		Opts *git.CommitOptions
	}
}

// NewMockWorktree creates a new MockWorktree with sensible defaults.
func NewMockWorktree() *MockWorktree {
	return &MockWorktree{
		StatusResult: make(git.Status),
	}
}

func (m *MockWorktree) Status() (git.Status, error) {
	return m.StatusResult, m.StatusErr
}

func (m *MockWorktree) Add(path string) (plumbing.Hash, error) {
	m.AddCalls = append(m.AddCalls, path)
	return m.AddHash, m.AddErr
}

func (m *MockWorktree) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	m.CommitCalls = append(m.CommitCalls, struct{Msg string; Opts *git.CommitOptions}{msg, opts})
	return m.CommitHash, m.CommitErr
}

// mockRefIter implements storer.ReferenceIter for testing.
type mockRefIter struct {
	refs  []plumbing.Reference
	index int
}

func (i *mockRefIter) Next() (*plumbing.Reference, error) {
	if i.index >= len(i.refs) {
		return nil, storer.ErrStop
	}
	ref := &i.refs[i.index]
	i.index++
	return ref, nil
}

func (i *mockRefIter) ForEach(fn func(*plumbing.Reference) error) error {
	for idx := range i.refs {
		if err := fn(&i.refs[idx]); err != nil {
			if err == storer.ErrStop {
				return nil
			}
			return err
		}
	}
	return nil
}

func (i *mockRefIter) Close() {}

// mockCommitIter implements object.CommitIter for testing.
type mockCommitIter struct {
	commits []*object.Commit
	index   int
}

func (i *mockCommitIter) Next() (*object.Commit, error) {
	if i.index >= len(i.commits) {
		return nil, storer.ErrStop
	}
	commit := i.commits[i.index]
	i.index++
	return commit, nil
}

func (i *mockCommitIter) ForEach(fn func(*object.Commit) error) error {
	for _, c := range i.commits {
		if err := fn(c); err != nil {
			if err == storer.ErrStop {
				return nil
			}
			return err
		}
	}
	return nil
}

func (i *mockCommitIter) Close() {}

// Helper functions for creating test data

// MakeTestCommit creates a commit object for testing.
func MakeTestCommit(hash plumbing.Hash, message, authorName, authorEmail string, when time.Time) *object.Commit {
	return &object.Commit{
		Hash:    hash,
		Message: message,
		Author: object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  when,
		},
	}
}

// MakeTestReference creates a reference for testing.
func MakeTestReference(name plumbing.ReferenceName, hash plumbing.Hash) plumbing.Reference {
	return *plumbing.NewHashReference(name, hash)
}

// MakeTestTag creates a tag reference for testing.
func MakeTestTag(name string, hash plumbing.Hash) plumbing.Reference {
	return MakeTestReference(plumbing.NewTagReferenceName(name), hash)
}

// MakeTestBranch creates a branch reference for testing.
func MakeTestBranch(name string, hash plumbing.Hash) plumbing.Reference {
	return MakeTestReference(plumbing.NewBranchReferenceName(name), hash)
}

// MockRepositoryOpener returns a RepositoryOpener that returns the given mock.
func MockRepositoryOpener(mock Repository) RepositoryOpener {
	return func(path string) (Repository, error) {
		return mock, nil
	}
}

// ErrorRepositoryOpener returns a RepositoryOpener that always returns an error.
func ErrorRepositoryOpener(err error) RepositoryOpener {
	return func(path string) (Repository, error) {
		return nil, err
	}
}
