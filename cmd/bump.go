package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/commitparser"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Auto-bump version based on commit messages",
	Long: `Analyze commits since the last tag and bump the version accordingly.

Supported commit message formats:

  +semver: markers (can appear anywhere in the commit message):
    +semver:major - Bump major version (1.0.0 -> 2.0.0)
    +semver:minor - Bump minor version (1.0.0 -> 1.1.0)
    +semver:patch - Bump patch version (1.0.0 -> 1.0.1)
    +semver:skip  - Skip version bump entirely

  Conventional Commits (https://conventionalcommits.org):
    feat: ...        - Bump minor version
    fix: ...         - Bump patch version
    feat!: ...       - Bump major version (breaking change)
    BREAKING CHANGE: - Bump major version (in commit footer)

Conflict resolution:
  - Highest bump level wins (major > minor > patch)
  - +semver:skip takes precedence and prevents any bump

Examples:
  versionator bump                   # Auto-bump based on commits
  versionator bump --dry-run         # Show what would happen
  versionator bump --mode=semver     # Only use +semver: markers
  versionator bump --mode=conventional  # Only use conventional commits`,
	RunE: runBump,
}

func init() {
	rootCmd.AddCommand(bumpCmd)

	bumpCmd.Flags().Bool("dry-run", false, "Show what would happen without making changes")
	bumpCmd.Flags().String("mode", "all", "Parse mode: semver, conventional, or all")
}

func runBump(cmd *cobra.Command, args []string) error {
	// Get active VCS
	activeVCS := vcs.GetActiveVCS()
	if activeVCS == nil {
		return fmt.Errorf(commitparser.ErrNoVCSDetected)
	}

	// Get commits since last tag
	messages, err := activeVCS.GetCommitMessagesSinceTag()
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(messages) == 0 {
		cmd.Println("No commits since last tag. Nothing to bump.")
		return nil
	}

	// Determine parse mode
	modeFlag, _ := cmd.Flags().GetString("mode")
	mode := getParseMode(modeFlag)

	// Parse and analyze commits
	parser := commitparser.NewParser(mode)
	analysis := parser.AnalyzeCommits(messages)

	// Handle skip
	if analysis.BumpLevel == commitparser.BumpSkip {
		cmd.Printf("Version bump skipped: %s\n", analysis.SkipReason)
		return nil
	}

	// Handle no bump detected
	if analysis.BumpLevel == commitparser.BumpNone {
		cmd.Println("No version bump detected in commits.")
		cmd.Println("Use +semver:patch/minor/major markers or conventional commits (feat:, fix:) to trigger bumps.")
		return nil
	}

	// Get current version
	v, err := version.Load()
	if err != nil {
		return fmt.Errorf("failed to load version: %w", err)
	}

	oldVersion := v.FullString()

	// Calculate new version
	newVersion := calculateNewVersion(v, analysis.BumpLevel)

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if dryRun {
		cmd.Printf("Analyzed %d commit(s)\n", analysis.CommitCount)
		cmd.Printf("Detected bump level: %s\n", analysis.BumpLevel.String())
		cmd.Printf("Triggering commit: %s\n", truncateCommit(analysis.TriggeringCommit))
		cmd.Printf("Would bump from %s to %s\n", oldVersion, newVersion)
		return nil
	}

	// Actually perform the bump
	versionLevel := analysis.BumpLevel.ToVersionLevel()
	if versionLevel < 0 {
		return fmt.Errorf("invalid bump level: %s", analysis.BumpLevel.String())
	}

	if err := version.Increment(versionLevel); err != nil {
		return fmt.Errorf("failed to bump version: %w", err)
	}

	cmd.Printf("Version bumped from %s to %s (%s)\n",
		oldVersion, newVersion, analysis.BumpLevel.String())
	cmd.Printf("Triggering commit: %s\n", truncateCommit(analysis.TriggeringCommit))

	return nil
}

func getParseMode(mode string) commitparser.ParseMode {
	switch strings.ToLower(mode) {
	case "semver":
		return commitparser.ModeSemverMarkers
	case "conventional":
		return commitparser.ModeConventionalCommits
	default:
		return commitparser.ModeAll
	}
}

func calculateNewVersion(v *version.Version, level commitparser.BumpLevel) string {
	// Create a copy to calculate new version without modifying original
	newMajor := v.Major
	newMinor := v.Minor
	newPatch := v.Patch

	switch level {
	case commitparser.BumpMajor:
		newMajor++
		newMinor = 0
		newPatch = 0
	case commitparser.BumpMinor:
		newMinor++
		newPatch = 0
	case commitparser.BumpPatch:
		newPatch++
	}

	return fmt.Sprintf("%s%d.%d.%d", v.Prefix, newMajor, newMinor, newPatch)
}

func truncateCommit(msg string) string {
	// Get first line and truncate
	lines := strings.Split(msg, "\n")
	first := strings.TrimSpace(lines[0])
	if len(first) > 60 {
		return first[:57] + "..."
	}
	return first
}
