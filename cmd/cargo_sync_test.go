package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
)

// =============================================================================
// CARGO.TOML SYNC ON EVERY VERSION CHANGE
// An `updates:` entry for Cargo.toml is applied by each command that writes the
// VERSION file, not only by release.
// =============================================================================

const cargoSyncConfig = `prefix: ""
prerelease:
  stable: true
metadata:
  stable: true
updates:
  - file: Cargo.toml
    path: workspace.package.version
    template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}"
`

const cargoSyncManifest = `[workspace]
members = ["crates/*"]

[workspace.package]
version = "0.6.2" # kept in sync by versionator
`

func cargoSyncDir(t *testing.T, version string) {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	_ = os.WriteFile("VERSION", []byte(version+"\n"), 0644)
	_ = os.WriteFile(".versionator.yaml", []byte(cargoSyncConfig), 0644)
	_ = os.WriteFile("Cargo.toml", []byte(cargoSyncManifest), 0644)
}

func runVersionator(t *testing.T, args ...string) error {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})
	return rootCmd.Execute()
}

func cargoVersionLine(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("read Cargo.toml: %v", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "version") {
			return line
		}
	}
	return ""
}

func TestCargoSync_Set_UpdatesWorkspaceVersion(t *testing.T) {
	cargoSyncDir(t, "0.6.2")

	if err := runVersionator(t, "set", "0.7.0-pre1"); err != nil {
		t.Fatalf("set: %v", err)
	}

	if got := cargoVersionLine(t); got != `version = "0.7.0-pre1" # kept in sync by versionator` {
		t.Errorf("got %q", got)
	}
}

func TestCargoSync_BumpPatch_UpdatesWorkspaceVersion(t *testing.T) {
	cargoSyncDir(t, "0.6.2")

	if err := runVersionator(t, "bump", "patch", "increment"); err != nil {
		t.Fatalf("bump: %v", err)
	}

	if got := cargoVersionLine(t); got != `version = "0.6.3" # kept in sync by versionator` {
		t.Errorf("got %q", got)
	}
}

func TestCargoSync_PrereleaseSet_UpdatesWorkspaceVersion(t *testing.T) {
	cargoSyncDir(t, "0.7.0")

	if err := runVersionator(t, "config", "prerelease", "set", "rc.1"); err != nil {
		t.Fatalf("config prerelease set: %v", err)
	}

	if got := cargoVersionLine(t); got != `version = "0.7.0-rc.1" # kept in sync by versionator` {
		t.Errorf("got %q", got)
	}
}

func TestCargoSync_PrereleaseClear_UpdatesWorkspaceVersion(t *testing.T) {
	cargoSyncDir(t, "0.7.0-rc.1")

	if err := runVersionator(t, "config", "prerelease", "clear"); err != nil {
		t.Fatalf("config prerelease clear: %v", err)
	}

	if got := cargoVersionLine(t); got != `version = "0.7.0" # kept in sync by versionator` {
		t.Errorf("got %q", got)
	}
}

func TestCargoSync_SetFourPart_IsRefused(t *testing.T) {
	cargoSyncDir(t, "0.6.2")

	err := runVersionator(t, "set", "0.7.0.1")

	if err == nil || !strings.Contains(err.Error(), "not a valid Cargo version") {
		t.Fatalf("expected the Cargo refusal, got %v", err)
	}
	if got := cargoVersionLine(t); got != `version = "0.6.2" # kept in sync by versionator` {
		t.Errorf("Cargo.toml must be left as it was, got %q", got)
	}
	assertVersionFile(t, "0.6.2\n")
}

func assertVersionFile(t *testing.T, want string) {
	t.Helper()
	b, _ := os.ReadFile("VERSION")
	if string(b) != want {
		t.Errorf("VERSION must be left as it was (%q), got %q", want, b)
	}
}

// Every command that changes VERSION checks the updates first: a refusal
// changes neither VERSION nor any configured file.
func TestCargoSync_RefusedUpdate_LeavesVersionUnchanged(t *testing.T) {
	for name, args := range map[string][]string{
		"bump":               {"bump", "patch", "increment"},
		"bump decrement":     {"bump", "patch", "decrement"},
		"prerelease set":     {"config", "prerelease", "set", "rc.1"},
		"prerelease enable":  {"config", "prerelease", "enable"},
		"prerelease clear":   {"config", "prerelease", "clear"},
		"metadata set":       {"config", "metadata", "set", "b.1"},
		"prerelease templ.":  {"config", "prerelease", "template", "rc"},
		"prefix set":         {"config", "prefix", "set", "v"},
		"prefix enable":      {"config", "prefix", "enable"},
		"set":                {"set", "0.9.0"},
		"set four-part":      {"set", "0.9.0.1"},
		"prerelease stable0": {"config", "prerelease", "stable", "false"},
	} {
		t.Run(name, func(t *testing.T) {
			cargoSyncDir(t, "0.6.2-rc.0")
			broken := "[workspace]\nmembers = []\n"
			_ = os.WriteFile("Cargo.toml", []byte(broken), 0644)

			if err := runVersionator(t, args...); err == nil {
				t.Fatal("expected the missing workspace.package.version to refuse the command")
			}

			assertVersionFile(t, "0.6.2-rc.0\n")
			if b, _ := os.ReadFile("Cargo.toml"); string(b) != broken {
				t.Errorf("Cargo.toml changed: %q", b)
			}
			if b, _ := os.ReadFile(".versionator.yaml"); string(b) != cargoSyncConfig {
				t.Errorf(".versionator.yaml changed:\n%s", b)
			}
		})
	}
}

func TestCargoSync_PrefixSet_RunsUpdates(t *testing.T) {
	cargoSyncDir(t, "1.0.0")
	_ = os.WriteFile(".versionator.yaml", []byte(`prefix: ""
updates:
  - file: tag.json
    path: tag
    template: "{{Prefix}}{{MajorMinorPatch}}"
`), 0644)
	_ = os.WriteFile("tag.json", []byte(`{"tag": "1.0.0"}`), 0644)

	if err := runVersionator(t, "config", "prefix", "set", "v"); err != nil {
		t.Fatalf("config prefix set: %v", err)
	}

	b, _ := os.ReadFile("tag.json")
	if string(b) != `{"tag": "v1.0.0"}` {
		t.Errorf("got %s", b)
	}
}

func TestUpdateTemplateData_PreReleaseAndMetadataSources(t *testing.T) {
	cases := []struct {
		name               string
		stored             string // VERSION file pre-release
		stable             bool
		template, wantDash string
	}{
		{"stored value wins over a dynamic template", "rc.1", false, "alpha", "-rc.1"},
		{"dynamic template renders when nothing is stored", "", false, "alpha", "-alpha"},
		{"stable with nothing stored has no pre-release", "", true, "alpha", ""},
		{"stable stored value", "beta", true, "alpha", "-beta"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := &version.Version{Major: 1, PreRelease: c.stored, BuildMetadata: c.stored}
			cfg := &config.Config{
				PreRelease: config.PreReleaseConfig{Template: c.template, Stable: c.stable},
				Metadata:   config.MetadataConfig{Template: c.template, Stable: c.stable},
			}
			data := updateTemplateData(v, cfg)
			if data.PreReleaseWithDash != c.wantDash {
				t.Errorf("PreReleaseWithDash = %q, want %q", data.PreReleaseWithDash, c.wantDash)
			}
			wantPlus := strings.Replace(c.wantDash, "-", "+", 1)
			if data.MetadataWithPlus != wantPlus {
				t.Errorf("MetadataWithPlus = %q, want %q", data.MetadataWithPlus, wantPlus)
			}
		})
	}
}

// A refused update fails the command that changed VERSION, for each writer.
func TestCargoSync_PrereleaseSet_RefusedUpdateFailsCommand(t *testing.T) {
	cargoSyncDir(t, "0.7.0")
	_ = os.WriteFile("Cargo.toml", []byte("[workspace]\nmembers = []\n"), 0644)

	if err := runVersionator(t, "config", "prerelease", "set", "rc.1"); err == nil {
		t.Fatal("expected the missing workspace.package.version to fail the command")
	}
}

func TestCargoSync_PrereleaseTemplateStable_RefusedUpdateFailsCommand(t *testing.T) {
	cargoSyncDir(t, "0.7.0")
	_ = os.WriteFile("Cargo.toml", []byte("[workspace]\nmembers = []\n"), 0644)

	if err := runVersionator(t, "config", "prerelease", "template", "rc"); err == nil {
		t.Fatal("expected the missing workspace.package.version to fail the command")
	}
}

func TestCargoSync_PrereleaseTemplateStable_UpdatesWorkspaceVersion(t *testing.T) {
	cargoSyncDir(t, "0.7.0")

	if err := runVersionator(t, "config", "prerelease", "template", "rc"); err != nil {
		t.Fatalf("config prerelease template: %v", err)
	}

	if got := cargoVersionLine(t); got != `version = "0.7.0-rc" # kept in sync by versionator` {
		t.Errorf("got %q", got)
	}
}

func TestCargoSync_PrereleaseStableFalse_UpdatesWorkspaceVersion(t *testing.T) {
	cargoSyncDir(t, "0.7.0-rc.1")

	if err := runVersionator(t, "config", "prerelease", "stable", "false"); err != nil {
		t.Fatalf("config prerelease stable false: %v", err)
	}

	if got := cargoVersionLine(t); got != `version = "0.7.0" # kept in sync by versionator` {
		t.Errorf("got %q", got)
	}
}
