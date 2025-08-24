package version

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"versionator/internal/vcs"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
)

const versionFile = "VERSION"

// getVersionFilePath returns the path to the VERSION file
// If we're in a VCS repository, use the repository root, otherwise use working directory
func (v *Version) getVersionFilePath() (string, error) {
	if v.vcs != nil && v.vcs.IsRepository() {
		root, err := v.vcs.GetRepositoryRoot()
		if err == nil {
			return filepath.Join(root, versionFile), nil
		}
	}

	// Fallback to working directory
	return filepath.Join(v.workingDir, versionFile), nil
}

// GetCurrentVersion reads the current version from the VERSION file
func (v *Version) GetCurrentVersion() (string, error) {
	filePath, err := v.getVersionFilePath()
	if err != nil {
		return "", err
	}

	data, err := afero.ReadFile(v.fs, filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If VERSION file doesn't exist, create it with 0.0.0
			err := v.WriteVersion("0.0.0")
			if err != nil {
				return "", fmt.Errorf("failed to create VERSION file: %w", err)
			}
			return "0.0.0", nil
		}
		return "", fmt.Errorf("failed to read VERSION file: %w", err)
	}

	version := strings.TrimSpace(string(data))
	if version == "" {
		return "0.0.0", nil
	}

	// Validate the version format
	_, err = semver.NewVersion(version)
	if err != nil {
		return "", fmt.Errorf("invalid version format in VERSION file: %s", version)
	}

	return version, nil
}

// Version provides version management operations with a filesystem
type Version struct {
	fs         afero.Fs
	workingDir string
	vcs        vcs.VersionControlSystem
}

// NewVersion creates a new Version instance with the provided filesystem, working directory, and VCS
func NewVersion(fs afero.Fs, workingDir string, vcsInstance vcs.VersionControlSystem) *Version {
	return &Version{fs: fs, workingDir: workingDir, vcs: vcsInstance}
}


// VersionLevel represents the semantic version component to modify
type VersionLevel int

const (
	MajorLevel VersionLevel = iota
	MinorLevel
	PatchLevel
)

// WriteVersion writes the version to the VERSION file
func (v *Version) WriteVersion(version string) error {
	filePath, err := v.getVersionFilePath()
	if err != nil {
		return err
	}
	return afero.WriteFile(v.fs, filePath, []byte(version+"\n"), 0644)
}

// Increment increments the specified version level
func (v *Version) Increment(level VersionLevel) error {
	current, err := v.GetCurrentVersion()
	if err != nil {
		return err
	}

	ver, err := semver.NewVersion(current)
	if err != nil {
		return err
	}

	var newVersion semver.Version
	switch level {
	case MajorLevel:
		newVersion = ver.IncMajor()
	case MinorLevel:
		newVersion = ver.IncMinor()
	case PatchLevel:
		newVersion = ver.IncPatch()
	default:
		return fmt.Errorf("invalid version level: %d", level)
	}

	return v.WriteVersion(newVersion.String())
}

// Decrement decrements the specified version level
func (v *Version) Decrement(level VersionLevel) error {
	current, err := v.GetCurrentVersion()
	if err != nil {
		return err
	}

	ver, err := semver.NewVersion(current)
	if err != nil {
		return err
	}

	var newVersion semver.Version
	switch level {
	case MajorLevel:
		if ver.Major() == 0 {
			return fmt.Errorf("cannot decrement major version below 0")
		}
		newVersion = *semver.MustParse(fmt.Sprintf("%d.0.0", ver.Major()-1))
	case MinorLevel:
		if ver.Minor() == 0 {
			return fmt.Errorf("cannot decrement minor version below 0")
		}
		newVersion = *semver.MustParse(fmt.Sprintf("%d.%d.0", ver.Major(), ver.Minor()-1))
	case PatchLevel:
		if ver.Patch() == 0 {
			return fmt.Errorf("cannot decrement patch version below 0")
		}
		newVersion = *semver.MustParse(fmt.Sprintf("%d.%d.%d", ver.Major(), ver.Minor(), ver.Patch()-1))
	default:
		return fmt.Errorf("invalid version level: %d", level)
	}

	return v.WriteVersion(newVersion.String())
}

