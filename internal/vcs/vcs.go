package vcs

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
}
