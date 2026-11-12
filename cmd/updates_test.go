package cmd

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/version"
)

func inTempDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	_ = os.Chdir(dir)
}

func TestWithConfigRollback_SnapshotFails_ChangeNotRun(t *testing.T) {
	inTempDir(t)
	_ = os.Mkdir(".versionator.yaml", 0755)
	ran := false

	err := withConfigRollback(func() error { ran = true; return nil })

	if err == nil || ran {
		t.Errorf("expected the snapshot error and no change, got err=%v ran=%v", err, ran)
	}
}

func TestWithConfigRollback_RestoreFails_SaysBoth(t *testing.T) {
	inTempDir(t)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: v\n"), 0644)
	t.Cleanup(func() { _ = os.Chmod(".versionator.yaml", 0644) })
	refused := errors.New("refused")

	err := withConfigRollback(func() error {
		_ = os.Chmod(".versionator.yaml", 0444) // the restore cannot write it back
		return refused
	})

	if !errors.Is(err, refused) || !strings.Contains(err.Error(), "restoring .versionator.yaml also failed") {
		t.Errorf("got %v", err)
	}
}

// A config that does not read checks nothing (the command reports the config).
func TestCheckConfiguredUpdates_UnreadableConfig_ChecksNothing(t *testing.T) {
	inTempDir(t)
	_ = os.WriteFile(".versionator.yaml", []byte("prefix: [unclosed\n"), 0644)

	if err := checkConfiguredUpdates(&version.Version{Major: 1}); err != nil {
		t.Errorf("got %v", err)
	}
}
