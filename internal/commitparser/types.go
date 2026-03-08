package commitparser

import "github.com/benjaminabbitt/versionator/internal/version"

// BumpLevel represents the detected version bump level
type BumpLevel int

const (
	BumpNone  BumpLevel = iota // No bump detected
	BumpPatch                  // Patch bump (fix:, +semver:patch)
	BumpMinor                  // Minor bump (feat:, +semver:minor)
	BumpMajor                  // Major bump (BREAKING CHANGE, !, +semver:major)
	BumpSkip                   // Explicitly skip versioning (+semver:skip)
)

// String returns the string representation of a BumpLevel
func (b BumpLevel) String() string {
	switch b {
	case BumpNone:
		return "none"
	case BumpPatch:
		return "patch"
	case BumpMinor:
		return "minor"
	case BumpMajor:
		return "major"
	case BumpSkip:
		return "skip"
	default:
		return "unknown"
	}
}

// ToVersionLevel converts BumpLevel to version.VersionLevel
// Returns -1 for BumpNone and BumpSkip (no version change)
func (b BumpLevel) ToVersionLevel() version.VersionLevel {
	switch b {
	case BumpMajor:
		return version.MajorLevel
	case BumpMinor:
		return version.MinorLevel
	case BumpPatch:
		return version.PatchLevel
	default:
		return -1
	}
}

// ParseMode controls which commit formats to recognize
type ParseMode int

const (
	ModeSemverMarkers       ParseMode = 1 << iota // +semver: markers only
	ModeConventionalCommits                       // Conventional commits only
	ModeAll                 = ModeSemverMarkers | ModeConventionalCommits
)

// ParsedCommit represents a single parsed commit
type ParsedCommit struct {
	Message    string    // Original commit message
	BumpLevel  BumpLevel // Detected bump level
	Format     string    // "semver-marker", "conventional", or "unknown"
	Type       string    // For conventional: "feat", "fix", etc.
	Scope      string    // For conventional: optional scope in parentheses
	IsBreaking bool      // Whether this is a breaking change
}

// CommitAnalysis contains the result of analyzing multiple commits
type CommitAnalysis struct {
	BumpLevel        BumpLevel      // Final determined bump level
	Commits          []ParsedCommit // All parsed commits
	SkipReason       string         // If skip, why
	TriggeringCommit string         // The commit that determined the bump level
	CommitCount      int            // Total commits analyzed
}
