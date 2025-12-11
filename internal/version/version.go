package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/benjaminabbitt/versionator/internal/vcs"

	"github.com/Masterminds/semver/v3"
)

const versionFile = "VERSION"

// getVersionFilePath returns the path to the VERSION file
// If we're in a VCS repository, use the repository root, otherwise use current directory
func getVersionFilePath() (string, error) {
	activeVCS := vcs.GetActiveVCS()
	if activeVCS != nil {
		root, err := activeVCS.GetRepositoryRoot()
		if err == nil {
			return filepath.Join(root, versionFile), nil
		}
	}

	// Fallback to current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, versionFile), nil
}

// GetCurrentVersion reads the current version from the VERSION file
func GetCurrentVersion() (string, error) {
	filePath, err := getVersionFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If VERSION file doesn't exist, create it with 0.0.0
			err := writeVersion("0.0.0")
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

// VersionLevel represents the semantic version component to modify
type VersionLevel int

const (
	MajorLevel VersionLevel = iota
	MinorLevel
	PatchLevel
)

// writeVersion writes the version to the VERSION file
func writeVersion(version string) error {
	filePath, err := getVersionFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(version+"\n"), 0644)
}

// Increment increments the specified version level
func Increment(level VersionLevel) error {
	current, err := GetCurrentVersion()
	if err != nil {
		return err
	}

	v, err := semver.NewVersion(current)
	if err != nil {
		return err
	}

	var newVersion semver.Version
	switch level {
	case MajorLevel:
		newVersion = v.IncMajor()
	case MinorLevel:
		newVersion = v.IncMinor()
	case PatchLevel:
		newVersion = v.IncPatch()
	default:
		return fmt.Errorf("invalid version level: %d", level)
	}

	return writeVersion(newVersion.String())
}

// Decrement decrements the specified version level
func Decrement(level VersionLevel) error {
	current, err := GetCurrentVersion()
	if err != nil {
		return err
	}

	v, err := semver.NewVersion(current)
	if err != nil {
		return err
	}

	var newVersion semver.Version
	switch level {
	case MajorLevel:
		if v.Major() == 0 {
			return fmt.Errorf("cannot decrement major version below 0")
		}
		newVersion = *semver.MustParse(fmt.Sprintf("%d.0.0", v.Major()-1))
	case MinorLevel:
		if v.Minor() == 0 {
			return fmt.Errorf("cannot decrement minor version below 0")
		}
		newVersion = *semver.MustParse(fmt.Sprintf("%d.%d.0", v.Major(), v.Minor()-1))
	case PatchLevel:
		if v.Patch() == 0 {
			return fmt.Errorf("cannot decrement patch version below 0")
		}
		newVersion = *semver.MustParse(fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch()-1))
	default:
		return fmt.Errorf("invalid version level: %d", level)
	}

	return writeVersion(newVersion.String())
}
