package version

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/vcs"
)

// ModulePath returns the directory holding the active VERSION file, relative to
// the repository root, slash-separated and terminated with a slash. It is empty
// when the VERSION file sits at the repository root, when no VCS is active, or
// when the repository root cannot be determined.
//
// The trailing slash is deliberate. Go requires a submodule's tag to be the
// module's directory path followed by the version ("internal/tools/v1.2.3"), so
// a single template — "{{ModulePath}}{{Prefix}}{{MajorMinorPatch}}" — renders
// correctly both inside a module subdirectory and at the repository root, where
// ModulePath contributes nothing.
func ModulePath() (string, error) {
	dir, err := moduleDir()
	if err != nil || dir == "" {
		return "", err
	}
	return dir + "/", nil
}

// ModuleName returns the bare name of the directory holding the active VERSION
// file ("tools" for internal/tools), empty at the repository root. It suits
// component-style tag conventions outside the Go ecosystem, such as
// "tools-v1.2.3", where the separator is a matter of convention rather than a
// requirement of the module system.
func ModuleName() (string, error) {
	dir, err := moduleDir()
	if err != nil || dir == "" {
		return "", err
	}
	return path.Base(dir), nil
}

// moduleDir returns the VERSION file's directory relative to the repository
// root, slash-separated and without a trailing slash. It returns an empty
// string — not an error — whenever the path cannot be placed inside a
// repository, so that a tag template still renders outside a repo.
func moduleDir() (string, error) {
	versionPath, err := getVersionPath()
	if err != nil {
		return "", err
	}

	activeVCS := vcs.GetActiveVCS()
	if activeVCS == nil {
		return "", nil
	}
	root, err := activeVCS.GetRepositoryRoot()
	if err != nil || root == "" {
		return "", nil
	}

	rel, err := filepath.Rel(resolvePath(root), resolvePath(filepath.Dir(versionPath)))
	if err != nil {
		return "", nil
	}
	// "." is the root itself; ".." means the VERSION file was found outside the
	// repository, which no tag should encode.
	if rel == "." || rel == string(filepath.Separator) || strings.HasPrefix(rel, "..") {
		return "", nil
	}

	return filepath.ToSlash(rel), nil
}

// resolvePath expands symlinks so that two paths describing the same directory
// compare equal. Temporary directories are symlinks on some platforms, which
// would otherwise make a VERSION file at the repository root look external to
// it.
func resolvePath(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	// The VERSION file may not exist yet; fall back to the cleaned path.
	if abs, err := filepath.Abs(p); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(p)
}
