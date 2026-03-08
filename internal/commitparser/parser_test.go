package commitparser

import (
	"testing"
)

func TestParseCommit_SemverMarkerMajor(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("fix: something +semver:major")

	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
	if result.Format != "semver-marker" {
		t.Errorf("expected format 'semver-marker', got %q", result.Format)
	}
}

func TestParseCommit_SemverMarkerMinor(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("Add feature +semver:minor")

	if result.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor, got %v", result.BumpLevel)
	}
}

func TestParseCommit_SemverMarkerPatch(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("fix bug +semver:patch")

	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch, got %v", result.BumpLevel)
	}
}

func TestParseCommit_SemverMarkerSkip(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("chore: update deps +semver:skip")

	if result.BumpLevel != BumpSkip {
		t.Errorf("expected BumpSkip, got %v", result.BumpLevel)
	}
}

func TestParseCommit_SemverMarkerCaseInsensitive(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("+SEMVER:MAJOR")

	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
}

func TestParseCommit_SemverMarkerInBody(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("Title\n\nSome description\n+semver:major")

	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
}

func TestParseCommit_ConventionalFeat(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("feat: add new feature")

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

func TestParseCommit_ConventionalFix(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("fix: resolve bug")

	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch, got %v", result.BumpLevel)
	}
	if result.Type != "fix" {
		t.Errorf("expected type 'fix', got %q", result.Type)
	}
}

func TestParseCommit_ConventionalBreakingWithBang(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("feat!: breaking change")

	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
	if !result.IsBreaking {
		t.Error("expected IsBreaking to be true")
	}
}

func TestParseCommit_ConventionalBreakingInFooter(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("feat: change API\n\nBREAKING CHANGE: removed old endpoint")

	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
	if !result.IsBreaking {
		t.Error("expected IsBreaking to be true")
	}
}

func TestParseCommit_ConventionalBreakingChangeHyphen(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("feat: change API\n\nBREAKING-CHANGE: removed old endpoint")

	if result.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", result.BumpLevel)
	}
}

func TestParseCommit_ConventionalWithScope(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("feat(api): add endpoint")

	if result.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor, got %v", result.BumpLevel)
	}
	if result.Scope != "api" {
		t.Errorf("expected scope 'api', got %q", result.Scope)
	}
}

func TestParseCommit_ConventionalChoreNoBump(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("chore: update deps")

	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", result.BumpLevel)
	}
}

func TestParseCommit_ConventionalDocsNoBump(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("docs: update readme")

	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", result.BumpLevel)
	}
}

func TestParseCommit_SemverMarkerTakesPrecedence(t *testing.T) {
	parser := NewParser(ModeAll)
	// feat: would be minor, but +semver:patch overrides
	result := parser.ParseCommit("feat: add feature +semver:patch")

	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch (semver marker takes precedence), got %v", result.BumpLevel)
	}
	if result.Format != "semver-marker" {
		t.Errorf("expected format 'semver-marker', got %q", result.Format)
	}
}

func TestParseCommit_ModeOnlySemver(t *testing.T) {
	parser := NewParser(ModeSemverMarkers)
	result := parser.ParseCommit("feat: add feature")

	// Should not detect conventional commit in semver-only mode
	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone in semver-only mode, got %v", result.BumpLevel)
	}
}

func TestParseCommit_ModeOnlyConventional(t *testing.T) {
	parser := NewParser(ModeConventionalCommits)
	result := parser.ParseCommit("fix: bug +semver:major")

	// Should ignore semver marker, detect fix as patch
	if result.BumpLevel != BumpPatch {
		t.Errorf("expected BumpPatch (ignore semver marker in conventional-only mode), got %v", result.BumpLevel)
	}
}

func TestParseCommit_UnknownFormat(t *testing.T) {
	parser := NewParser(ModeAll)
	result := parser.ParseCommit("Updated something")

	if result.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", result.BumpLevel)
	}
	if result.Format != "unknown" {
		t.Errorf("expected format 'unknown', got %q", result.Format)
	}
}

func TestAnalyzeCommits_HighestBumpWins(t *testing.T) {
	parser := NewParser(ModeAll)

	messages := []string{
		"fix: minor fix",
		"feat: new feature",
		"docs: update readme",
	}

	analysis := parser.AnalyzeCommits(messages)

	if analysis.BumpLevel != BumpMinor {
		t.Errorf("expected BumpMinor (highest), got %v", analysis.BumpLevel)
	}
	if analysis.CommitCount != 3 {
		t.Errorf("expected CommitCount 3, got %d", analysis.CommitCount)
	}
}

func TestAnalyzeCommits_MajorWins(t *testing.T) {
	parser := NewParser(ModeAll)

	messages := []string{
		"fix: patch fix",
		"feat: minor feature",
		"feat!: breaking change",
	}

	analysis := parser.AnalyzeCommits(messages)

	if analysis.BumpLevel != BumpMajor {
		t.Errorf("expected BumpMajor (highest), got %v", analysis.BumpLevel)
	}
}

func TestAnalyzeCommits_SkipTakesPrecedence(t *testing.T) {
	parser := NewParser(ModeAll)

	messages := []string{
		"feat!: breaking change +semver:major",
		"chore: maintenance +semver:skip",
	}

	analysis := parser.AnalyzeCommits(messages)

	if analysis.BumpLevel != BumpSkip {
		t.Errorf("expected BumpSkip (takes precedence), got %v", analysis.BumpLevel)
	}
	if analysis.SkipReason == "" {
		t.Error("expected SkipReason to be set")
	}
}

func TestAnalyzeCommits_EmptyList(t *testing.T) {
	parser := NewParser(ModeAll)
	analysis := parser.AnalyzeCommits([]string{})

	if analysis.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone for empty list, got %v", analysis.BumpLevel)
	}
	if analysis.CommitCount != 0 {
		t.Errorf("expected CommitCount 0, got %d", analysis.CommitCount)
	}
}

func TestAnalyzeCommits_AllNoBump(t *testing.T) {
	parser := NewParser(ModeAll)

	messages := []string{
		"chore: update deps",
		"docs: readme",
		"style: formatting",
	}

	analysis := parser.AnalyzeCommits(messages)

	if analysis.BumpLevel != BumpNone {
		t.Errorf("expected BumpNone, got %v", analysis.BumpLevel)
	}
}

func TestAnalyzeCommits_TriggeringCommitIsSet(t *testing.T) {
	parser := NewParser(ModeAll)

	messages := []string{
		"chore: update deps",
		"feat: new feature",
		"docs: readme",
	}

	analysis := parser.AnalyzeCommits(messages)

	if analysis.TriggeringCommit != "feat: new feature" {
		t.Errorf("expected TriggeringCommit 'feat: new feature', got %q", analysis.TriggeringCommit)
	}
}

func TestBumpLevel_String(t *testing.T) {
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
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("BumpLevel(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}
