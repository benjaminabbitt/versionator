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
  versionator bump                   # Auto-bump and amend last commit
  versionator bump --dry-run         # Show what would happen
  versionator bump --no-amend        # Bump without amending the commit
  versionator bump --mode=semver     # Only use +semver: markers
  versionator bump --mode=conventional  # Only use conventional commits`,
	RunE: runBump,
}

// runLevelIncrement handles incrementing a version level
func runLevelIncrement(cmd *cobra.Command, level version.VersionLevel, titleName string) error {
	if err := version.Increment(level); err != nil {
		return err
	}
	ver, err := version.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("error reading updated version: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s version incremented to: %s\n", titleName, ver)
	return runConfiguredUpdates(cmd)
}

// runLevelDecrement handles decrementing a version level
func runLevelDecrement(cmd *cobra.Command, level version.VersionLevel, titleName string) error {
	if err := version.Decrement(level); err != nil {
		return err
	}
	ver, err := version.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("error reading updated version: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s version decremented to: %s\n", titleName, ver)
	return runConfiguredUpdates(cmd)
}

// makeLevelCmd creates a parent command for a version level with increment/decrement subcommands
func makeLevelCmd(level version.VersionLevel, name string) *cobra.Command {
	titleName := strings.ToUpper(name[:1]) + name[1:]

	cmd := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Increment %s version (default), or use subcommands", name),
		Long:  fmt.Sprintf("Increment the %s version. Use 'decrement' subcommand to decrement instead.", name),
		RunE: func(c *cobra.Command, args []string) error {
			return runLevelIncrement(c, level, titleName)
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "increment",
		Aliases: []string{"inc", "+", "up"},
		Short:   fmt.Sprintf("Increment %s version", name),
		Long:    fmt.Sprintf("Increment the %s version", name),
		RunE: func(c *cobra.Command, args []string) error {
			return runLevelIncrement(c, level, titleName)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "decrement",
		Aliases: []string{"dec", "-", "down"},
		Short:   fmt.Sprintf("Decrement %s version", name),
		Long:    fmt.Sprintf("Decrement the %s version", name),
		RunE: func(c *cobra.Command, args []string) error {
			return runLevelDecrement(c, level, titleName)
		},
	})

	return cmd
}

func init() {
	rootCmd.AddCommand(bumpCmd)

	bumpCmd.Flags().Bool("dry-run", false, "Show what would happen without making changes")
	bumpCmd.Flags().Bool("no-amend", false, "Update VERSION file but do not amend the last commit")
	bumpCmd.Flags().String("mode", "all", "Parse mode: semver, conventional, or all")

	// Add level commands to bump
	bumpCmd.AddCommand(makeLevelCmd(version.MajorLevel, "major"))
	bumpCmd.AddCommand(makeLevelCmd(version.MinorLevel, "minor"))
	bumpCmd.AddCommand(makeLevelCmd(version.PatchLevel, "patch"))
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

	if err := runConfiguredUpdates(cmd); err != nil {
		return err
	}

	// Amend the last commit by default (unless --no-amend is specified)
	noAmend, _ := cmd.Flags().GetBool("no-amend")
	if !noAmend {
		if err := activeVCS.AmendCommit([]string{"VERSION"}); err != nil {
			return fmt.Errorf("failed to amend commit: %w", err)
		}
		cmd.Println("Amended last commit to include VERSION change")
	}

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
