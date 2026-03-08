package commitparser

import (
	"regexp"
	"strings"
)

// Parser handles commit message parsing
type Parser struct {
	mode ParseMode
}

// NewParser creates a parser with the specified mode
func NewParser(mode ParseMode) *Parser {
	return &Parser{mode: mode}
}

// Regular expressions for parsing
var (
	// +semver:major, +semver:minor, +semver:patch, +semver:skip
	// Case insensitive, can appear anywhere in the commit message
	semverMarkerRegex = regexp.MustCompile(`(?i)\+semver:\s*(major|minor|patch|skip)`)

	// Conventional commits: type(scope)!: description
	// Breaking change indicated by ! before colon
	// type is required, scope is optional
	conventionalRegex = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?\s*:\s*(.+)`)

	// BREAKING CHANGE: or BREAKING-CHANGE: in commit body/footer
	breakingChangeFooter = regexp.MustCompile(`(?im)^BREAKING[ -]CHANGE:\s*(.+)`)
)

// ParseCommit parses a single commit message and detects bump level
func (p *Parser) ParseCommit(message string) ParsedCommit {
	result := ParsedCommit{
		Message:   message,
		BumpLevel: BumpNone,
		Format:    "unknown",
	}

	// Check +semver: markers first (they take precedence)
	if p.mode&ModeSemverMarkers != 0 {
		if match := semverMarkerRegex.FindStringSubmatch(message); match != nil {
			result.Format = "semver-marker"
			switch strings.ToLower(match[1]) {
			case "major":
				result.BumpLevel = BumpMajor
			case "minor":
				result.BumpLevel = BumpMinor
			case "patch":
				result.BumpLevel = BumpPatch
			case "skip":
				result.BumpLevel = BumpSkip
			}
			return result
		}
	}

	// Check conventional commits
	if p.mode&ModeConventionalCommits != 0 {
		lines := strings.Split(message, "\n")
		if match := conventionalRegex.FindStringSubmatch(lines[0]); match != nil {
			result.Format = "conventional"
			result.Type = strings.ToLower(match[1])
			result.Scope = match[2] // May be empty

			// Check for breaking change indicator
			hasBangBreaking := match[3] == "!"
			hasFooterBreaking := breakingChangeFooter.MatchString(message)
			result.IsBreaking = hasBangBreaking || hasFooterBreaking

			if result.IsBreaking {
				result.BumpLevel = BumpMajor
			} else {
				// Map conventional commit types to bump levels
				switch result.Type {
				case "feat":
					result.BumpLevel = BumpMinor
				case "fix":
					result.BumpLevel = BumpPatch
				// Other types (chore, docs, style, refactor, test, ci, perf, build) don't bump
				default:
					result.BumpLevel = BumpNone
				}
			}
			return result
		}
	}

	return result
}

// AnalyzeCommits analyzes multiple commits and determines the bump level
// Conflict resolution:
//   - Highest level wins (major > minor > patch > none)
//   - Exception: if ANY commit has +semver:skip, return skip
func (p *Parser) AnalyzeCommits(messages []string) CommitAnalysis {
	analysis := CommitAnalysis{
		BumpLevel:   BumpNone,
		Commits:     make([]ParsedCommit, 0, len(messages)),
		CommitCount: len(messages),
	}

	for _, msg := range messages {
		parsed := p.ParseCommit(msg)
		analysis.Commits = append(analysis.Commits, parsed)

		// Skip takes precedence over everything
		if parsed.BumpLevel == BumpSkip {
			analysis.BumpLevel = BumpSkip
			analysis.SkipReason = "Commit contains +semver:skip"
			analysis.TriggeringCommit = msg
			return analysis
		}

		// Highest bump level wins
		if parsed.BumpLevel > analysis.BumpLevel {
			analysis.BumpLevel = parsed.BumpLevel
			analysis.TriggeringCommit = msg
		}
	}

	return analysis
}
