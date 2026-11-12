package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CARGO.TOML VERSION SYNC
// An `updates:` entry on a Cargo.toml edits the one string at its path in
// place. Every other byte of the file stays as it was, a lookalike value in
// another table is never touched, and a value Cargo would refuse is refused
// before any file is written.
// =============================================================================

func writeTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0755))
	require.NoError(t, os.WriteFile(p, []byte(content), 0644))
	return p
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return string(b)
}

func cargoUpdate(file, path string) config.UpdateConfig {
	return config.UpdateConfig{File: file, Path: path, Template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}"}
}

func runUpdates(t *testing.T, data emit.TemplateData, cfgs ...config.UpdateConfig) error {
	t.Helper()
	return NewUpdater(cfgs, NewDaselFileParser(), newTestLogger(t)).UpdateFiles(data)
}

const workspaceManifest = `# The workspace root.
[workspace]
resolver = "2"
members = ["crates/*"]

# Internal crates are pinned to the workspace version, and listed FIRST.
[workspace.dependencies]
core-lib = { path = "crates/core-lib", version = "0.6.2" }
other = "0.6.2"

[workspace.package]
edition   = "2021"
version   = "0.6.2"   # stamped by versionator
license = "AGPL-3.0-only"
`

func TestCargoSync_WorkspacePackageVersion_OnlyThatValueChanges(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.7.0", PreReleaseWithDash: "-pre1"},
		cargoUpdate(p, "workspace.package.version"))
	require.NoError(t, err)

	want := strings.Replace(workspaceManifest, `version   = "0.6.2"`, `version   = "0.7.0-pre1"`, 1)
	assert.Equal(t, want, readFile(t, p), "only [workspace.package] version may change, byte for byte")
}

const packageManifest = `[package]
name = "ctxloom"
# the crate's own version
version = '1.0.0'
edition = "2021"

[dependencies]
serde = { version = "1.0.0", features = ["derive"] }

[package.metadata.docs]
version = "1.0.0"
`

func TestCargoSync_PackageVersion_LiteralStringKeepsItsQuotes(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", packageManifest)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.1.0"}, cargoUpdate(p, "package.version")))

	want := strings.Replace(packageManifest, `version = '1.0.0'`, `version = '1.1.0'`, 1)
	assert.Equal(t, want, readFile(t, p))
}

func TestCargoSync_PackageVersionAfterLookalike_TargetsThePath(t *testing.T) {
	dir := t.TempDir()
	content := `[dependencies]
serde = { version = "1.0.0" }

[package]
name = "late"
version = "1.0.0"
`
	p := writeTemp(t, dir, "Cargo.toml", content)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "package.version")))

	want := `[dependencies]
serde = { version = "1.0.0" }

[package]
name = "late"
version = "2.0.0"
`
	assert.Equal(t, want, readFile(t, p))
}

func TestCargoSync_DottedKey_IsFound(t *testing.T) {
	dir := t.TempDir()
	content := "[workspace]\nmembers = []\npackage.version = \"0.1.0\" # dotted\n"
	p := writeTemp(t, dir, "Cargo.toml", content)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.2.0"}, cargoUpdate(p, "workspace.package.version")))

	assert.Equal(t, "[workspace]\nmembers = []\npackage.version = \"0.2.0\" # dotted\n", readFile(t, p))
}

func TestCargoSync_SameValue_LeavesFileAlone(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.6.2"}, cargoUpdate(p, "workspace.package.version")))

	assert.Equal(t, workspaceManifest, readFile(t, p))
}

func TestCargoSync_CargoLock_IsNotTouched(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)
	lock := "version = 4\n\n[[package]]\nname = \"core-lib\"\nversion = \"0.6.2\"\n"
	lockPath := writeTemp(t, dir, "Cargo.lock", lock)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.7.0"}, cargoUpdate(p, "workspace.package.version")))

	assert.Equal(t, lock, readFile(t, lockPath))
}

func TestCargoSync_MissingKey_IsAnError(t *testing.T) {
	dir := t.TempDir()
	content := "[package]\nname = \"noversion\"\n"
	p := writeTemp(t, dir, "Cargo.toml", content)

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0"}, cargoUpdate(p, "package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrPathNotFound)
	assert.Contains(t, err.Error(), "package.version")
	assert.Equal(t, content, readFile(t, p))
}

func TestCargoSync_MissingTable_IsAnError(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", "[package]\nname = \"x\"\nversion = \"1.0.0\"\n")

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0"}, cargoUpdate(p, "workspace.package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrPathNotFound)
}

// A member crate that inherits its version (`version.workspace = true`) has no
// string to write: the version lives in the workspace root.
func TestCargoSync_InheritedVersion_IsAnError(t *testing.T) {
	dir := t.TempDir()
	content := "[package]\nname = \"member\"\nversion.workspace = true\n"
	p := writeTemp(t, dir, "Cargo.toml", content)

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0"}, cargoUpdate(p, "package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValueNotString)
	assert.Equal(t, content, readFile(t, p))
}

func TestCargoSync_FourPartVersion_IsRefused(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.7.0.4"}, cargoUpdate(p, "workspace.package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrCargoVersionInvalid)
	assert.Contains(t, err.Error(), "0.7.0.4")
	assert.Equal(t, workspaceManifest, readFile(t, p))
}

func TestCargoSync_PrefixedVersion_IsRefused(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)

	cfg := config.UpdateConfig{File: p, Path: "workspace.package.version", Template: "{{Prefix}}{{MajorMinorPatch}}"}
	err := runUpdates(t, emit.TemplateData{Prefix: "v", MajorMinorPatch: "0.7.0"}, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrCargoVersionInvalid)
	assert.Equal(t, workspaceManifest, readFile(t, p))
}

func TestCargoSync_BuildMetadata_IsAccepted(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)

	cfg := config.UpdateConfig{File: p, Path: "workspace.package.version", Template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"}
	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.7.0", PreReleaseWithDash: "-rc.1", MetadataWithPlus: "+build.5"}, cfg))

	assert.Contains(t, readFile(t, p), `version   = "0.7.0-rc.1+build.5"`)
}

// The SemVer check is Cargo's rule, so it applies to a Cargo manifest's
// version keys only; another TOML file takes what its template renders.
func TestTOMLSync_NonCargoFile_IsNotHeldToCargoRules(t *testing.T) {
	dir := t.TempDir()
	content := "[tool.poetry]\nversion = \"1.0.0\"\n"
	p := writeTemp(t, dir, "pyproject.toml", content)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0.4"}, cargoUpdate(p, "tool.poetry.version")))

	assert.Equal(t, "[tool.poetry]\nversion = \"1.0.0.4\"\n", readFile(t, p))
}

func TestCargoSync_NestedCargoToml_IsHeldToCargoRules(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "crates/app/Cargo.toml", "[package]\nname = \"app\"\nversion = \"1.0.0\"\n")

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0.4"}, cargoUpdate(p, "package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrCargoVersionInvalid)
}

// A refused value leaves every configured file as it was, including the ones
// listed before it: the updates are all rendered and checked first.
func TestUpdateFiles_OneRefused_WritesNothing(t *testing.T) {
	dir := t.TempDir()
	pkgJSON := writeTemp(t, dir, "package.json", `{"version": "1.0.0"}`)
	cargo := writeTemp(t, dir, "Cargo.toml", workspaceManifest)

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0.4"},
		config.UpdateConfig{File: pkgJSON, Path: "version", Template: "{{MajorMinorPatch}}"},
		cargoUpdate(cargo, "workspace.package.version"))

	require.Error(t, err)
	assert.Equal(t, `{"version": "1.0.0"}`, readFile(t, pkgJSON))
	assert.Equal(t, workspaceManifest, readFile(t, cargo))
}

func TestCargoVersionValid(t *testing.T) {
	for v, ok := range map[string]bool{
		"0.7.0": true, "0.7.0-pre1": true, "1.0.0-alpha.1+build.42": true, "10.20.30": true,
		"0.7.0.1": false, "v0.7.0": false, "V1.0.0": false, "1.0": false, "01.0.0": false,
		"1.0.0-": false, "1.0.0-01": false, "": false, " 1.0.0": false,
	} {
		assert.Equal(t, ok, cargoVersionValid(v), v)
	}
}

func TestIsCargoVersionKey(t *testing.T) {
	assert.True(t, isCargoVersionKey("Cargo.toml", "package.version"))
	assert.True(t, isCargoVersionKey("a/b/Cargo.toml", "workspace.package.version"))
	assert.False(t, isCargoVersionKey("Cargo.toml", "package.metadata.version"))
	assert.False(t, isCargoVersionKey("pyproject.toml", "package.version"))
}

func TestCargoSync_InlineTablePath_IsFound(t *testing.T) {
	dir := t.TempDir()
	content := "[workspace]\npackage = { edition = \"2021\", version = \"0.1.0\" }\n"
	p := writeTemp(t, dir, "Cargo.toml", content)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.2.0"}, cargoUpdate(p, "workspace.package.version")))

	assert.Equal(t, "[workspace]\npackage = { edition = \"2021\", version = \"0.2.0\" }\n", readFile(t, p))
}

func TestCargoSync_MultiLineString_IsAnError(t *testing.T) {
	dir := t.TempDir()
	content := "[package]\nversion = \"\"\"1.0.0\"\"\"\n"
	p := writeTemp(t, dir, "Cargo.toml", content)

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValueNotString)
}

func TestCargoSync_NumberAtPath_IsAnError(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", "[package]\nversion = 1\n")

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValueNotString)
}

func TestCargoSync_ArrayTableEntries_AreNotMatched(t *testing.T) {
	dir := t.TempDir()
	content := "[[bin]]\nname = \"x\"\n\n[package]\nname = \"y\"\nversion = \"1.0.0\"\n"
	p := writeTemp(t, dir, "Cargo.toml", content)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "package.version")))

	assert.Equal(t, "[[bin]]\nname = \"x\"\n\n[package]\nname = \"y\"\nversion = \"2.0.0\"\n", readFile(t, p))
}

func TestCargoSync_TableAtPath_IsAnError(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", "[package.version]\nworkspace = true\n")

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "package.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValueNotString)
}

func TestCargoSync_Unchanged_IsNotRewritten(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)
	require.NoError(t, os.Chmod(p, 0444)) // a write would fail
	t.Cleanup(func() { _ = os.Chmod(p, 0644) })

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "0.6.2"}, cargoUpdate(p, "workspace.package.version")))
}

func TestQuoteTOMLLike(t *testing.T) {
	assert.Equal(t, `'1.0.0'`, string(quoteTOMLLike([]byte(`'0.1.0'`), "1.0.0")))
	assert.Equal(t, `"it's"`, string(quoteTOMLLike([]byte(`'0.1.0'`), "it's")))
	assert.Equal(t, `"a\"b\\c"`, string(quoteTOMLLike([]byte(`"x"`), `a"b\c`)))
}

// [[array]] entries are not addressable by a dotted path; dasel's index
// syntax reaches them through the earlier text replacement.
func TestTOMLSync_ArrayTableByDottedPath_IsNotFound(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "x.toml", "[[bin]]\nversion = \"1.0.0\"\n")

	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "bin.version"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrPathNotFound)
}

func TestTOMLSync_DaselIndexSelector_UsesTextReplacement(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "x.toml", "[[bin]]\nname = \"a\"\nversion = '1.0.0'\n")

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "bin[0].version")))

	assert.Equal(t, "[[bin]]\nname = \"a\"\nversion = \"2.0.0\"\n", readFile(t, p))
}

func TestTOMLSync_DaselIndexSelector_SameValue_Unchanged(t *testing.T) {
	dir := t.TempDir()
	content := "[[bin]]\nversion = \"2.0.0\"\n"
	p := writeTemp(t, dir, "x.toml", content)

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"}, cargoUpdate(p, "bin[0].version")))

	assert.Equal(t, content, readFile(t, p))
}

func TestTOMLSync_DaselIndexSelector_Errors(t *testing.T) {
	dir := t.TempDir()
	parser := NewDaselFileParser()

	_, err := parser.setTOMLString([]byte("not = [valid"), "bin[0].version", "2.0.0")
	assert.ErrorContains(t, err, ErrFileParseFailed)

	_, err = parser.setTOMLString([]byte("[[bin]]\nname = \"a\"\n"), "bin[0].version", "2.0.0")
	assert.ErrorContains(t, err, ErrPathNotFound)

	// A non-string value has no quoted text to replace.
	_, err = parser.setTOMLString([]byte("[[bin]]\nversion = 1\n"), "bin[0].version", "2.0.0")
	assert.ErrorContains(t, err, "could not find value")
	_ = dir
}

func TestTOMLSync_UnparseableFile_IsAnError(t *testing.T) {
	_, err := NewDaselFileParser().setTOMLString([]byte("[package\nversion = \"1\"\n"), "package.version", "2.0.0")
	assert.ErrorContains(t, err, ErrFileParseFailed)
}

func TestCargoSync_EmptyTableHeaderAtPath_IsAnError(t *testing.T) {
	_, err := NewDaselFileParser().setTOMLString([]byte("[package.version]\n"), "package.version", "2.0.0")
	assert.ErrorContains(t, err, ErrValueNotString)
}

func TestTOMLSync_ArrayTableName_IsATable(t *testing.T) {
	_, err := NewDaselFileParser().setTOMLString([]byte("[[bin]]\nversion = \"1\"\n"), "bin", "2.0.0")
	assert.ErrorContains(t, err, ErrValueNotString)
}

func TestUpdateFiles_MissingFile_IsAnError(t *testing.T) {
	err := runUpdates(t, emit.TemplateData{MajorMinorPatch: "1.0.0"},
		cargoUpdate(filepath.Join(t.TempDir(), "Cargo.toml"), "package.version"))
	assert.ErrorContains(t, err, ErrFileNotFound)
}

// A file already at the new version is passed over, and the files after it
// are still written.
func TestUpdateFiles_UnchangedFileFirst_LaterFileStillWritten(t *testing.T) {
	dir := t.TempDir()
	first := writeTemp(t, dir, "a/Cargo.toml", "[package]\nversion = \"2.0.0\"\n")
	second := writeTemp(t, dir, "b/Cargo.toml", "[package]\nversion = \"1.0.0\"\n")

	require.NoError(t, runUpdates(t, emit.TemplateData{MajorMinorPatch: "2.0.0"},
		cargoUpdate(first, "package.version"), cargoUpdate(second, "package.version")))

	assert.Equal(t, "[package]\nversion = \"2.0.0\"\n", readFile(t, second))
}

func TestCargoSync_NonStringInsideInlineTable_IsAnError(t *testing.T) {
	_, err := NewDaselFileParser().setTOMLString([]byte("[workspace]\npackage = { version = 1 }\n"), "workspace.package.version", "2.0.0")
	assert.ErrorContains(t, err, ErrValueNotString)
}

// Check runs every update in memory and writes nothing, success or refusal.
func TestUpdater_Check_WritesNothing(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "Cargo.toml", workspaceManifest)
	u := NewUpdater([]config.UpdateConfig{cargoUpdate(p, "workspace.package.version")}, NewDaselFileParser(), newTestLogger(t))

	require.NoError(t, u.Check(emit.TemplateData{MajorMinorPatch: "0.7.0"}))
	assert.Equal(t, workspaceManifest, readFile(t, p))

	err := u.Check(emit.TemplateData{MajorMinorPatch: "0.7.0.1"})
	assert.ErrorContains(t, err, ErrCargoVersionInvalid)
	assert.Equal(t, workspaceManifest, readFile(t, p))
	assert.Empty(t, u.GetFilesToCommit())
}
