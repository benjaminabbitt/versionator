package commitparser

import (
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the commit parser: parsing
// conventional commit messages and analyzing commit histories to determine
// the appropriate semantic version bump level.
// =============================================================================

// TestParseCommit_ConventionalFeat_ReturnsBumpMinor validates that the parser
// correctly identifies "feat:" conventional commits as requiring a minor version bump.
//
// Why: Features represent new functionality which, per semver, requires a minor bump.
// This is the most common case for adding functionality to a project.
//
// What: Given a conventional commit with "feat:" prefix, the parser should return
// BumpMinor with format "conventional" and type "feat".
func TestParseCommit_ConventionalFeat_ReturnsBumpMinor(t *testing.T) {
	// Precondition: Parser is initialized with ModeAll to accept all commit formats
	parser := NewParser(ModeAll)

	// Action: Parse a conventional commit with feat: prefix
	result := parser.ParseCommit("feat: add new feature")

	// Expected: BumpMinor level, conventional format, feat type
	if result.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor, got %v", result.BumpLevel)
	}
	if result.Format != "conventional" {
		t.Errorf("expected format 'conventional', got %q", result.Format)
	}
	if result.Type != "feat" {
		t.Errorf("expected type 'feat', got %q", result.Type)
	}
}

// TestParseCommit_ConventionalFix_ReturnsBumpPatch validates that the parser
// correctly identifies "fix:" conventional commits as requiring a patch version bump.
//
// Why: Bug fixes represent corrections to existing functionality which, per semver,
// requires a patch bump. This is the standard way to indicate backward-compatible fixes.
//
// What: Given a conventional commit with "fix:" prefix, the parser should return
// BumpPatch with type "fix".
func TestParseCommit_ConventionalFix_ReturnsBumpPatch(t *testing.T) {
	// Precondition: Parser is initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse a conventional commit with fix: prefix
	result := parser.ParseCommit("fix: resolve bug")

	// Expected: BumpPatch level and fix type
	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch, got %v", result.BumpLevel)
	}
	if result.Type != "fix" {
		t.Errorf("expected type 'fix', got %q", result.Type)
	}
}

// TestAnalyzeCommits_MultipleCommits_ReturnsHighestBumpLevel validates that when
// analyzing multiple commits, the parser returns the highest bump level found.
//
// Why: When preparing a release, all commits since the last release are analyzed.
// The release version must accommodate the most significant change (major > minor > patch).
//
// What: Given commits containing fix (patch), feat (minor), and docs (none), the
// analysis should return BumpMinor as the highest applicable level.
func TestAnalyzeCommits_MultipleCommits_ReturnsHighestBumpLevel(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Analyze a mix of commits with different bump levels
	messages := []string{
		"fix: minor fix",      // BumpPatch
		"feat: new feature",   // BumpMinor (highest)
		"docs: update readme", // BumpNone
	}
	analysis := parser.AnalyzeCommits(messages)

	// Expected: BumpMinor (highest level) with correct commit count
	if analysis.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor (highest), got %v", analysis.BumpLevel)
	}
	if analysis.CommitCount != 3 {
		t.Errorf("expected CommitCount 3, got %d", analysis.CommitCount)
	}
}

// TestAnalyzeCommits_WithBreakingChange_ReturnsBumpMajor validates that breaking
// changes result in a major version bump, even when mixed with lesser changes.
//
// Why: Breaking changes require consumers to modify their code and must be clearly
// signaled with a major version increment per semver.
//
// What: Given commits including a breaking change (feat!), the analysis should
// return BumpMajor regardless of other commits present.
func TestAnalyzeCommits_WithBreakingChange_ReturnsBumpMajor(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Analyze commits including a breaking change
	messages := []string{
		"fix: patch fix",
		"feat: minor feature",
		"feat!: breaking change",
	}
	analysis := parser.AnalyzeCommits(messages)

	// Expected: BumpMajor takes precedence
	if analysis.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor (highest), got %v", analysis.BumpLevel)
	}
}

// TestAnalyzeCommits_TriggeringCommitIsSet validates that the analysis correctly
// identifies which commit triggered the bump level determination.
//
// Why: For release notes and debugging, it's valuable to know which specific commit
// was responsible for the version bump decision.
//
// What: Given multiple commits where only one causes a bump, the TriggeringCommit
// field should contain that commit's message.
func TestAnalyzeCommits_TriggeringCommitIsSet(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Analyze commits where feat: is the triggering commit
	messages := []string{
		"chore: update deps",
		"feat: new feature",
		"docs: readme",
	}
	analysis := parser.AnalyzeCommits(messages)

	// Expected: TriggeringCommit points to the feat commit
	if analysis.TriggeringCommit != "feat: new feature" {
		t.Errorf("expected TriggeringCommit 'feat: new feature', got %q", analysis.TriggeringCommit)
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows: different commit formats, semver markers,
// breaking change indicators, and scoped commits.
// =============================================================================

// TestParseCommit_SemverMarkerMajor_ReturnsBumpMajor validates that explicit
// semver markers override conventional commit semantics.
//
// Why: Teams may need to force a specific version bump regardless of the commit
// type. The +semver:major marker provides this explicit control.
//
// What: Given a commit with +semver:major marker, parser returns BumpMajor
// with format "semver-marker".
func TestParseCommit_SemverMarkerMajor_ReturnsBumpMajor(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse a commit with semver marker
	result := parser.ParseCommit("fix: something +semver:major")

	// Expected: BumpMajor from marker, not BumpPatch from fix:
	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
	if result.Format != "semver-marker" {
		t.Errorf("expected format 'semver-marker', got %q", result.Format)
	}
}

// TestParseCommit_SemverMarkerMinor_ReturnsBumpMinor validates minor semver markers.
//
// Why: Allows forcing a minor bump on non-feature commits when appropriate.
//
// What: Given +semver:minor marker, parser returns BumpMinor.
func TestParseCommit_SemverMarkerMinor_ReturnsBumpMinor(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with +semver:minor
	result := parser.ParseCommit("Add feature +semver:minor")

	// Expected: BumpMinor
	if result.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor, got %v", result.BumpLevel)
	}
}

// TestParseCommit_SemverMarkerPatch_ReturnsBumpPatch validates patch semver markers.
//
// Why: Allows forcing a patch bump when commit message doesn't follow conventions.
//
// What: Given +semver:patch marker, parser returns BumpPatch.
func TestParseCommit_SemverMarkerPatch_ReturnsBumpPatch(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with +semver:patch
	result := parser.ParseCommit("fix bug +semver:patch")

	// Expected: BumpPatch
	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch, got %v", result.BumpLevel)
	}
}

// TestParseCommit_SemverMarkerSkip_ReturnsBumpSkip validates the skip marker.
//
// Why: Some commits (CI changes, typo fixes) should not trigger any release.
// The +semver:skip marker explicitly prevents version bumps.
//
// What: Given +semver:skip marker, parser returns BumpSkip.
func TestParseCommit_SemverMarkerSkip_ReturnsBumpSkip(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with +semver:skip
	result := parser.ParseCommit("chore: update deps +semver:skip")

	// Expected: BumpSkip
	if result.BumpLevel != BumpSkip {
		t.Errorf("expected BumpSkip, got %v", result.BumpLevel)
	}
}

// TestParseCommit_ConventionalBreakingWithBang_ReturnsBumpMajor validates that
// the conventional commit "!" breaking change indicator is recognized.
//
// Why: The "!" suffix on commit types is a standard way to indicate breaking
// changes per the conventional commits specification.
//
// What: Given "feat!:" prefix, parser returns BumpMajor with IsBreaking=true.
func TestParseCommit_ConventionalBreakingWithBang_ReturnsBumpMajor(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with ! breaking indicator
	result := parser.ParseCommit("feat!: breaking change")

	// Expected: BumpMajor and IsBreaking flag set
	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
	if !result.IsBreaking {
		t.Error("expected IsBreaking to be true")
	}
}

// TestParseCommit_ConventionalBreakingInFooter_ReturnsBumpMajor validates that
// BREAKING CHANGE in the commit footer is recognized.
//
// Why: Per conventional commits spec, "BREAKING CHANGE:" in the footer is an
// alternative way to indicate breaking changes with more detail.
//
// What: Given BREAKING CHANGE in footer, parser returns BumpMajor with IsBreaking=true.
func TestParseCommit_ConventionalBreakingInFooter_ReturnsBumpMajor(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with BREAKING CHANGE in footer
	result := parser.ParseCommit("feat: change API\n\nBREAKING CHANGE: removed old endpoint")

	// Expected: BumpMajor and IsBreaking flag set
	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
	if !result.IsBreaking {
		t.Error("expected IsBreaking to be true")
	}
}

// TestParseCommit_ConventionalWithScope_ExtractsScope validates scope extraction.
//
// Why: Scopes help categorize changes by component (e.g., api, cli, core) for
// better change tracking and release notes generation.
//
// What: Given "feat(api):" format, parser extracts "api" as the scope.
func TestParseCommit_ConventionalWithScope_ExtractsScope(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with scope
	result := parser.ParseCommit("feat(api): add endpoint")

	// Expected: BumpMinor and scope extracted
	if result.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor, got %v", result.BumpLevel)
	}
	if result.Scope != "api" {
		t.Errorf("expected scope 'api', got %q", result.Scope)
	}
}

// TestParseCommit_ConventionalChore_ReturnsBumpNone validates that non-bump
// commit types like "chore:" don't trigger version changes.
//
// Why: Maintenance tasks, dependency updates, and other housekeeping changes
// typically don't affect the public API and shouldn't trigger releases.
//
// What: Given "chore:" prefix, parser returns BumpNone.
func TestParseCommit_ConventionalChore_ReturnsBumpNone(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse chore commit
	result := parser.ParseCommit("chore: update deps")

	// Expected: BumpNone (no version change)
	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", result.BumpLevel)
	}
}

// TestParseCommit_ConventionalDocs_ReturnsBumpNone validates that documentation
// changes don't trigger version bumps.
//
// Why: Documentation changes don't affect runtime behavior and shouldn't
// require consumers to update their dependency versions.
//
// What: Given "docs:" prefix, parser returns BumpNone.
func TestParseCommit_ConventionalDocs_ReturnsBumpNone(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse docs commit
	result := parser.ParseCommit("docs: update readme")

	// Expected: BumpNone
	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", result.BumpLevel)
	}
}

// TestParseCommit_SemverMarkerTakesPrecedence_OverridesConventional validates
// that semver markers override conventional commit type semantics.
//
// Why: Explicit version control via markers should always win over inferred
// semantics, allowing teams to override when the commit type doesn't match
// the actual impact.
//
// What: Given "feat:" (minor) with "+semver:patch", parser returns BumpPatch.
func TestParseCommit_SemverMarkerTakesPrecedence_OverridesConventional(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse feat commit with +semver:patch override
	result := parser.ParseCommit("feat: add feature +semver:patch")

	// Expected: BumpPatch from marker, not BumpMinor from feat:
	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch (semver marker takes precedence), got %v", result.BumpLevel)
	}
	if result.Format != "semver-marker" {
		t.Errorf("expected format 'semver-marker', got %q", result.Format)
	}
}

// TestAnalyzeCommits_SkipTakesPrecedence_OverridesAllOther validates that
// BumpSkip takes precedence over all other bump levels in analysis.
//
// Why: When a commit explicitly signals skip, the entire changeset should be
// skipped regardless of other changes. This is useful for release management.
//
// What: Given commits with major and skip markers, analysis returns BumpSkip.
func TestAnalyzeCommits_SkipTakesPrecedence_OverridesAllOther(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Analyze commits including one with +semver:skip
	messages := []string{
		"feat!: breaking change +semver:major",
		"chore: maintenance +semver:skip",
	}
	analysis := parser.AnalyzeCommits(messages)

	// Expected: BumpSkip takes precedence over major
	if analysis.BumpLevel != BumpSkip {
		t.Errorf("expected BumpSkip (takes precedence), got %v", analysis.BumpLevel)
	}
	if analysis.SkipReason == "" {
		t.Error("expected SkipReason to be set")
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes: parser mode restrictions that cause
// certain commit formats to be ignored.
// =============================================================================

// TestParseCommit_ModeOnlySemver_IgnoresConventional validates that in
// semver-markers-only mode, conventional commits are not recognized.
//
// Why: Teams may want to use only explicit semver markers and ignore
// conventional commit inference entirely for more control.
//
// What: Given ModeSemverMarkers and a "feat:" commit without marker,
// parser returns BumpNone.
func TestParseCommit_ModeOnlySemver_IgnoresConventional(t *testing.T) {
	// Precondition: Parser initialized with semver-markers-only mode
	parser := NewParser(ModeSemverMarkers)

	// Action: Parse conventional commit without semver marker
	result := parser.ParseCommit("feat: add feature")

	// Expected: BumpNone because conventional commits are ignored
	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone in semver-only mode, got %v", result.BumpLevel)
	}
}

// TestParseCommit_ModeOnlyConventional_IgnoresSemverMarker validates that in
// conventional-commits-only mode, semver markers are ignored.
//
// Why: Teams using strict conventional commits may want to ensure markers
// don't accidentally override their commit type semantics.
//
// What: Given ModeConventionalCommits and "fix: ... +semver:major",
// parser returns BumpPatch (from fix:), ignoring the marker.
func TestParseCommit_ModeOnlyConventional_IgnoresSemverMarker(t *testing.T) {
	// Precondition: Parser initialized with conventional-commits-only mode
	parser := NewParser(ModeConventionalCommits)

	// Action: Parse fix commit with +semver:major that should be ignored
	result := parser.ParseCommit("fix: bug +semver:major")

	// Expected: BumpPatch from fix:, marker ignored
	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch (ignore semver marker in conventional-only mode), got %v", result.BumpLevel)
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions: empty input, unknown formats, commits that
// don't match any pattern.
// =============================================================================

// TestParseCommit_UnknownFormat_ReturnsBumpNone validates handling of commits
// that don't match any recognized format.
//
// Why: Not all commits follow conventional or semver-marker formats. The parser
// must gracefully handle freeform commit messages without crashing.
//
// What: Given a plain commit message without any markers, parser returns
// BumpNone with format "unknown".
func TestParseCommit_UnknownFormat_ReturnsBumpNone(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse a freeform commit message
	result := parser.ParseCommit("Updated something")

	// Expected: BumpNone with unknown format
	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", result.BumpLevel)
	}
	if result.Format != "unknown" {
		t.Errorf("expected format 'unknown', got %q", result.Format)
	}
}

// TestAnalyzeCommits_EmptyList_ReturnsBumpNone validates handling of empty
// commit lists.
//
// Why: Edge case where there are no commits to analyze (e.g., no changes
// since last release). Parser must handle gracefully without panic.
//
// What: Given an empty slice, analysis returns BumpNone with zero count.
func TestAnalyzeCommits_EmptyList_ReturnsBumpNone(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Analyze empty commit list
	analysis := parser.AnalyzeCommits([]string{})

	// Expected: BumpNone with zero commits
	if analysis.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone for empty list, got %v", analysis.BumpLevel)
	}
	if analysis.CommitCount != 0 {
		t.Errorf("expected CommitCount 0, got %d", analysis.CommitCount)
	}
}

// TestAnalyzeCommits_AllNoBumpTypes_ReturnsBumpNone validates that when all
// commits are non-bumping types, the result is BumpNone.
//
// Why: A release cycle with only chore/docs/style commits should not trigger
// a version bump - there's no user-facing change to release.
//
// What: Given only chore, docs, and style commits, analysis returns BumpNone.
func TestAnalyzeCommits_AllNoBumpTypes_ReturnsBumpNone(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Analyze commits that all have BumpNone
	messages := []string{
		"chore: update deps",
		"docs: readme",
		"style: formatting",
	}
	analysis := parser.AnalyzeCommits(messages)

	// Expected: BumpNone because no commit triggers a bump
	if analysis.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", analysis.BumpLevel)
	}
}

// TestBumpLevel_String_UnknownValue_ReturnsUnknown validates that undefined
// BumpLevel values return "unknown" string representation.
//
// Why: Defensive programming - if a BumpLevel value is somehow invalid,
// String() should return a safe fallback rather than panic.
//
// What: Given BumpLevel(999), String() returns "unknown".
func TestBumpLevel_String_UnknownValue_ReturnsUnknown(t *testing.T) {
	// Precondition: Create an invalid BumpLevel value
	unknown := BumpLevel(999)

	// Action: Get string representation
	got := unknown.String()

	// Expected: "unknown" fallback
	if got != "unknown" {
		t.Errorf("BumpLevel(999).String() = %q, want %q", got, "unknown")
	}
}

// =============================================================================
// MINUTIAE
// Tests for obscure but important scenarios: case insensitivity, marker
// placement variations, hyphenated keywords.
// =============================================================================

// TestParseCommit_SemverMarkerCaseInsensitive_RecognizesUppercase validates
// that semver markers are case-insensitive.
//
// Why: Users may type markers in various cases (SEMVER, Semver, semver).
// The parser should accept all variations for usability.
//
// What: Given "+SEMVER:MAJOR" (all caps), parser returns BumpMajor.
func TestParseCommit_SemverMarkerCaseInsensitive_RecognizesUppercase(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with uppercase semver marker
	result := parser.ParseCommit("+SEMVER:MAJOR")

	// Expected: BumpMajor regardless of case
	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
}

// TestParseCommit_SemverMarkerInBody_RecognizesMarkerInMultilineCommit validates
// that semver markers are recognized when placed in the commit body.
//
// Why: Developers may place markers in the body rather than the subject line
// for readability. The parser should scan the entire commit message.
//
// What: Given a multiline commit with +semver:major in body, parser returns BumpMajor.
func TestParseCommit_SemverMarkerInBody_RecognizesMarkerInMultilineCommit(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse multiline commit with marker in body
	result := parser.ParseCommit("Title\n\nSome description\n+semver:major")

	// Expected: BumpMajor from marker in body
	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
}

// TestParseCommit_ConventionalBreakingChangeHyphen_RecognizesHyphenatedForm validates
// that BREAKING-CHANGE (with hyphen) is recognized as equivalent to BREAKING CHANGE.
//
// Why: The conventional commits spec allows both "BREAKING CHANGE:" and
// "BREAKING-CHANGE:" as valid breaking change footers.
//
// What: Given BREAKING-CHANGE in footer, parser returns BumpMajor.
func TestParseCommit_ConventionalBreakingChangeHyphen_RecognizesHyphenatedForm(t *testing.T) {
	// Precondition: Parser initialized with ModeAll
	parser := NewParser(ModeAll)

	// Action: Parse commit with hyphenated BREAKING-CHANGE footer
	result := parser.ParseCommit("feat: change API\n\nBREAKING-CHANGE: removed old endpoint")

	// Expected: BumpMajor from hyphenated breaking change
	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
}

// TestBumpLevel_String_AllLevels_ReturnsCorrectStrings validates string
// representation for all defined BumpLevel values.
//
// Why: String() is used for logging, output, and debugging. Each level must
// have a correct, human-readable representation.
//
// What: Each BumpLevel constant returns its expected string name.
func TestBumpLevel_String_AllLevels_ReturnsCorrectStrings(t *testing.T) {
	// Precondition: Define all expected mappings
	tests := []struct {
		level    BumpLevel
		expected string
	}{
		{BumpNone, "none"},
		{BumpPatch, "patch"},
		{BumpMinor, "minor"},
		{BumpMajor, "major"},
		{BumpSkip, "skip"},
	}

	for _, tt := range tests {
		// Action: Get string representation
		got := tt.level.String()

		// Expected: Matches predefined string
		if got != tt.expected {
			t.Errorf("BumpLevel(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

// TestBumpLevel_ToVersionLevel_MapsCorrectly validates that BumpLevel correctly
// converts to version.VersionLevel for version incrementing operations.
//
// Why: The commit parser's BumpLevel must integrate with the version package's
// VersionLevel type for actual version calculations.
//
// What: Each BumpLevel maps to its corresponding VersionLevel integer value,
// with non-bump levels (None, Skip) mapping to -1.
func TestBumpLevel_ToVersionLevel_MapsCorrectly(t *testing.T) {
	// Precondition: Define expected mappings
	// version.MajorLevel = 0, MinorLevel = 1, PatchLevel = 2
	tests := []struct {
		name     string
		level    BumpLevel
		expected int
	}{
		{name: "BumpMajor -> MajorLevel", level: BumpMajor, expected: 0},
		{name: "BumpMinor -> MinorLevel", level: BumpMinor, expected: 1},
		{name: "BumpPatch -> PatchLevel", level: BumpPatch, expected: 2},
		{name: "BumpNone -> -1", level: BumpNone, expected: -1},
		{name: "BumpSkip -> -1", level: BumpSkip, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Convert to VersionLevel
			got := int(tt.level.ToVersionLevel())

			// Expected: Matches predefined mapping
			if got != tt.expected {
				t.Errorf("BumpLevel(%d).ToVersionLevel() = %d, want %d", tt.level, got, tt.expected)
			}
		})
	}
}
