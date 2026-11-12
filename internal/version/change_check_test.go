package version

import (
	"errors"
	"os"
	"testing"
)

// =============================================================================
// CHANGE CHECK
// A check set with SetChangeCheck judges the version a command is about to
// write, before the VERSION file is touched. A refusal leaves VERSION as it was.
// =============================================================================

func inVersionDir(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	_ = os.Chdir(dir)
	if content != "" {
		_ = os.WriteFile(versionFile, []byte(content), 0644)
	}
	t.Cleanup(func() { SetChangeCheck(nil) })
}

func versionFileContent(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	return string(b)
}

var errRefused = errors.New("refused by check")

func TestChangeCheck_Refusal_LeavesVersionUnchanged(t *testing.T) {
	writers := map[string]func() error{
		"SetVersion":    func() error { return SetVersion("2.0.0") },
		"Increment":     func() error { return Increment(PatchLevel) },
		"Decrement":     func() error { return Decrement(PatchLevel) },
		"SetPrefix":     func() error { return SetPrefix("v") },
		"SetPreRelease": func() error { return SetPreRelease("rc.1") },
		"SetMetadata":   func() error { return SetMetadata("b.1") },
		"Save":          func() error { return Save(&Version{Major: 3}) },
	}
	for name, write := range writers {
		t.Run(name, func(t *testing.T) {
			inVersionDir(t, "1.0.1\n")
			SetChangeCheck(func(*Version) error { return errRefused })

			err := write()

			if !errors.Is(err, errRefused) {
				t.Fatalf("expected the check's refusal, got %v", err)
			}
			if got := versionFileContent(t); got != "1.0.1\n" {
				t.Errorf("VERSION changed to %q", got)
			}
		})
	}
}

func TestChangeCheck_SeesTheNextVersion(t *testing.T) {
	inVersionDir(t, "1.0.1\n")
	var seen string
	SetChangeCheck(func(next *Version) error { seen = next.FullString(); return nil })

	if err := Increment(MinorLevel); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	if seen != "1.1.0" {
		t.Errorf("check saw %q, want 1.1.0", seen)
	}
	if got := versionFileContent(t); got != "1.1.0\n" {
		t.Errorf("VERSION = %q", got)
	}
}

// Creating the default VERSION on first read is not a change a command asked
// for, so it is not checked.
func TestChangeCheck_DefaultCreationOnLoad_IsNotChecked(t *testing.T) {
	inVersionDir(t, "")
	SetChangeCheck(func(*Version) error { return errRefused })

	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := versionFileContent(t); got != "v0.0.1\n" { // the config default prefix
		t.Errorf("VERSION = %q", got)
	}
}

// An invalid version is refused by Save's own validation before the check.
func TestChangeCheck_InvalidVersion_NotChecked(t *testing.T) {
	inVersionDir(t, "1.0.1\n")
	called := false
	SetChangeCheck(func(*Version) error { called = true; return nil })

	if err := Save(&Version{Major: 1, PreRelease: "bad..pre"}); err == nil {
		t.Fatal("expected an invalid version to be refused")
	}
	if called {
		t.Error("the check ran on a version Save refuses anyway")
	}
}
