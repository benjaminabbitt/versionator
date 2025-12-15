package vcs

import "time"

// VersionControlSystem defines the interface for version control operations
type VersionControlSystem interface {
	// Name returns the name of the VCS (e.g., "git", "svn")
	Name() string

	// IsRepository checks if we're in a repository of this type
	IsRepository() bool

	// GetRepositoryRoot returns the root directory of the repository
	GetRepositoryRoot() (string, error)

	// IsWorkingDirectoryClean checks if there are no uncommitted changes
	IsWorkingDirectoryClean() (bool, error)

	// GetVCSIdentifier returns a VCS-specific identifier for the current state
	GetVCSIdentifier(length int) (string, error)

	// CreateTag creates a tag with the specified name and message
	CreateTag(tagName, message string) error

	// TagExists checks if a tag with the specified name exists
	TagExists(tagName string) (bool, error)

	// GetBranchName returns the current branch name
	GetBranchName() (string, error)

	// GetCommitDate returns the date of the current commit
	GetCommitDate() (time.Time, error)

	// GetCommitsSinceTag returns the number of commits since the most recent semver tag
	// Returns 0 if on a tagged commit, -1 if no tags exist
	GetCommitsSinceTag() (int, error)

	// GetUncommittedChanges returns the count of uncommitted changes (staged + unstaged + untracked)
	GetUncommittedChanges() (int, error)

	// GetLastTag returns the most recent semver tag
	// Returns empty string if no matching tags exist
	GetLastTag() (string, error)

	// GetLastTagCommit returns the SHA of the commit the last tag points to
	GetLastTagCommit() (string, error)

	// GetCommitAuthor returns the name of the current commit's author
	GetCommitAuthor() (string, error)

	// GetCommitAuthorEmail returns the email of the current commit's author
	GetCommitAuthorEmail() (string, error)
}
