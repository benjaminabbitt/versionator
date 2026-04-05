package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/ci"
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/vcs"
	"github.com/benjaminabbitt/versionator/internal/version"

	"github.com/spf13/cobra"
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Output version variables for CI/CD systems",
	Long: `Output version variables in CI/CD-specific formats.

Auto-detects CI environment or use --format to specify:
  github   - GitHub Actions ($GITHUB_OUTPUT, $GITHUB_ENV)
  gitlab   - GitLab CI (dotenv artifact format)
  azure    - Azure DevOps (##vso[task.setvariable])
  circleci - CircleCI ($BASH_ENV)
  jenkins  - Jenkins (properties file format)
  shell    - Generic shell exports

Variables exported:
  VERSION, VERSION_SEMVER, VERSION_CORE,
  VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH,
  VERSION_PRERELEASE, VERSION_METADATA,
  GIT_SHA, GIT_SHA_SHORT, GIT_BRANCH, BUILD_NUMBER, DIRTY

Examples:
  versionator ci                    # Auto-detect CI and set vars
  versionator ci --format=github    # Force GitHub Actions format
  versionator ci --format=shell     # Print shell exports to stdout
  versionator ci --output=vars.env  # Write to file
  versionator ci --prefix=MYAPP_    # Variable prefix (MYAPP_VERSION, etc.)`,
	RunE: runCI,
}

func init() {
	outputCmd.AddCommand(ciCmd)

	ciCmd.Flags().StringP("format", "f", "", "Output format (github, gitlab, azure, circleci, jenkins, shell)")
	ciCmd.Flags().StringP("output", "o", "", "Output file (default: stdout or CI-specific location)")
	ciCmd.Flags().String("prefix", "", "Variable name prefix (e.g., 'MYAPP_' -> MYAPP_VERSION)")
}

func runCI(cmd *cobra.Command, args []string) error {
	// Get current version
	v, err := version.Load()
	if err != nil {
		return fmt.Errorf("failed to load version: %w", err)
	}

	// Build CI variables
	vars := buildCIVariables(v)

	// Get prefix
	prefix, _ := cmd.Flags().GetString("prefix")

	// Convert to map with prefix
	varMap := vars.ToMap(prefix)

	// Determine format
	formatFlag, _ := cmd.Flags().GetString("format")
	var formatter ci.Formatter

	if formatFlag != "" {
		var err error
		formatter, err = ci.GetFormatterByName(formatFlag)
		if err != nil {
			return fmt.Errorf("invalid format: %s (available: %s)",
				formatFlag, strings.Join(ci.AvailableFormatters(), ", "))
		}
	} else {
		// Auto-detect CI environment
		env := ci.Detect()
		formatter = ci.GetFormatter(env)

		if env == ci.EnvNone {
			cmd.Println("No CI environment detected, using shell format")
		} else {
			cmd.Printf("Detected CI environment: %s\n", env)
		}
	}

	// Determine output destination
	outputFile, _ := cmd.Flags().GetString("output")
	var writer = cmd.OutOrStdout()

	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// Write variables
	if err := formatter.Write(varMap, writer); err != nil {
		return fmt.Errorf("failed to write variables: %w", err)
	}

	return nil
}

func buildCIVariables(v *version.Version) *ci.Variables {
	// Load config to check stability settings
	cfg, _ := config.ReadConfig()

	// Build template data for rendering
	templateData := emit.BuildTemplateDataFromVersion(v)

	// Determine pre-release based on stability
	preRelease := v.PreRelease
	if cfg != nil && !cfg.PreRelease.Stable && cfg.PreRelease.Template != "" {
		// Non-stable: render from template
		if rendered, err := emit.RenderTemplateWithData(cfg.PreRelease.Template, templateData); err == nil {
			preRelease = strings.TrimSpace(rendered)
		}
	}

	// Determine metadata based on stability
	metadata := v.BuildMetadata
	if cfg != nil && !cfg.Metadata.Stable && cfg.Metadata.Template != "" {
		// Non-stable: render from template
		if rendered, err := emit.RenderTemplateWithData(cfg.Metadata.Template, templateData); err == nil {
			metadata = strings.TrimSpace(rendered)
		}
	}

	// Build version strings with potentially rendered pre-release/metadata
	versionFull := v.CoreVersion()
	if preRelease != "" {
		versionFull += "-" + preRelease
	}
	if metadata != "" {
		versionFull += "+" + metadata
	}

	vars := &ci.Variables{
		Version:       v.Prefix + versionFull,
		VersionSemver: versionFull,
		VersionCore:   v.CoreVersion(),
		Major:         strconv.Itoa(v.Major),
		Minor:         strconv.Itoa(v.Minor),
		Patch:         strconv.Itoa(v.Patch),
		Revision:      v.RevisionString(),
		PreRelease:    preRelease,
		Metadata:      metadata,
	}

	// Get VCS information
	activeVCS := vcs.GetActiveVCS()
	if activeVCS != nil {
		if sha, err := activeVCS.GetVCSIdentifier(40); err == nil {
			vars.GitSHA = sha
		}
		if shortSHA, err := activeVCS.GetVCSIdentifier(7); err == nil {
			vars.GitSHAShort = shortSHA
		}
		if branch, err := activeVCS.GetBranchName(); err == nil {
			vars.GitBranch = branch
		}
		if commits, err := activeVCS.GetCommitsSinceTag(); err == nil {
			vars.BuildNumber = strconv.Itoa(commits)
		}
		if changes, err := activeVCS.GetUncommittedChanges(); err == nil {
			if changes > 0 {
				vars.Dirty = "true"
			} else {
				vars.Dirty = "false"
			}
		}
	}

	return vars
}
