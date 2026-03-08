package branch

// Error messages
const (
	ErrEmptyBranchName = "branch name is empty"
	ErrDetachedHead    = "detached HEAD state"
	ErrInvalidPattern  = "invalid branch pattern"
)

// Log messages
const (
	LogBranchDetected     = "branch_detected"
	LogMainBranchMatched  = "main_branch_matched"
	LogBranchPrerelease   = "branch_prerelease_generated"
	LogBranchNotEnabled   = "branch_versioning_not_enabled"
)
