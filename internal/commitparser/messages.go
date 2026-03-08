package commitparser

// Error messages
const (
	ErrNoCommitsSinceTag = "no commits since last tag"
	ErrNoVCSDetected     = "not in a version control repository"
	ErrNoBumpDetected    = "no version bump detected in commits"
)

// Log messages
const (
	LogCommitsAnalyzed   = "commits_analyzed"
	LogBumpLevelDetected = "bump_level_detected"
	LogSkipDetected      = "skip_detected"
)
