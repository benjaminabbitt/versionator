package ci

import "os"

// Environment represents a detected CI/CD platform
type Environment string

const (
	EnvNone          Environment = "none"
	EnvGitHubActions Environment = "github"
	EnvGitLabCI      Environment = "gitlab"
	EnvAzureDevOps   Environment = "azure"
	EnvCircleCI      Environment = "circleci"
	EnvJenkins       Environment = "jenkins"
	EnvGeneric       Environment = "shell"
)

// String returns the string representation of the environment
func (e Environment) String() string {
	return string(e)
}

// Detect returns the current CI environment based on environment variables
func Detect() Environment {
	// GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return EnvGitHubActions
	}

	// GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		return EnvGitLabCI
	}

	// Azure DevOps
	if os.Getenv("TF_BUILD") == "True" {
		return EnvAzureDevOps
	}

	// CircleCI
	if os.Getenv("CIRCLECI") == "true" {
		return EnvCircleCI
	}

	// Jenkins
	if os.Getenv("JENKINS_URL") != "" {
		return EnvJenkins
	}

	// No CI detected
	return EnvNone
}

// Variables defines the standard set of version variables to export
type Variables struct {
	Version           string // Full version string (e.g., "v1.2.3" or "1.2.3-alpha+build")
	VersionSemver     string // SemVer without prefix (e.g., "1.2.3-alpha")
	VersionCore       string // Core version only (e.g., "1.2.3")
	Major             string // Major version
	Minor             string // Minor version
	Patch             string // Patch version
	Revision          string // Revision (4th component, may be empty)
	PreRelease        string // Pre-release identifier (may be empty)
	Metadata          string // Build metadata (may be empty)
	GitSHA            string // Full commit SHA
	GitSHAShort       string // Short commit SHA (7 chars)
	GitBranch         string // Current branch name
	BuildNumber       string // Commits since last tag
	Dirty             string // "true" if uncommitted changes, "false" otherwise
}

// VariableNames returns the standard variable names with optional prefix
func VariableNames(prefix string) map[string]string {
	return map[string]string{
		"Version":       prefix + "VERSION",
		"VersionSemver": prefix + "VERSION_SEMVER",
		"VersionCore":   prefix + "VERSION_CORE",
		"Major":         prefix + "VERSION_MAJOR",
		"Minor":         prefix + "VERSION_MINOR",
		"Patch":         prefix + "VERSION_PATCH",
		"Revision":      prefix + "VERSION_REVISION",
		"PreRelease":    prefix + "VERSION_PRERELEASE",
		"Metadata":      prefix + "VERSION_METADATA",
		"GitSHA":        prefix + "GIT_SHA",
		"GitSHAShort":   prefix + "GIT_SHA_SHORT",
		"GitBranch":     prefix + "GIT_BRANCH",
		"BuildNumber":   prefix + "BUILD_NUMBER",
		"Dirty":         prefix + "DIRTY",
	}
}

// ToMap converts Variables to a map with the given prefix
func (v *Variables) ToMap(prefix string) map[string]string {
	names := VariableNames(prefix)
	return map[string]string{
		names["Version"]:       v.Version,
		names["VersionSemver"]: v.VersionSemver,
		names["VersionCore"]:   v.VersionCore,
		names["Major"]:         v.Major,
		names["Minor"]:         v.Minor,
		names["Patch"]:         v.Patch,
		names["Revision"]:      v.Revision,
		names["PreRelease"]:    v.PreRelease,
		names["Metadata"]:      v.Metadata,
		names["GitSHA"]:        v.GitSHA,
		names["GitSHAShort"]:   v.GitSHAShort,
		names["GitBranch"]:     v.GitBranch,
		names["BuildNumber"]:   v.BuildNumber,
		names["Dirty"]:         v.Dirty,
	}
}
