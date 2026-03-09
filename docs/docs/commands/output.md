---
title: output
description: Output version in various formats
---

# output

Output version in various formats

Output the current version in various formats for different use cases.

Use subcommands to output version information:
  output version  - Show current version (with optional template)
  output emit     - Generate version files for programming languages
  output ci       - Output version variables for CI/CD systems

## Usage

```bash
versionator output [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `ci` | Output version variables for CI/CD systems |
| `emit` | Emit version in various formats |
| `version` | Show current version |

### ci

Output version variables for CI/CD systems

Output version variables in CI/CD-specific formats.

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
  versionator ci --prefix=MYAPP_    # Variable prefix (MYAPP_VERSION, etc.)

```bash
versionator output ci [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --format` | string | - | Output format (github, gitlab, azure, circleci, jenkins, shell) |
| `-o, --output` | string | - | Output file (default: stdout or CI-specific location) |
| `--prefix` | string | - | Variable name prefix (e.g., 'MYAPP_' -\> MYAPP_VERSION) |

### emit

Emit version in various formats

```
Emit the current version in various programming language formats.

Supported formats: python, json, yaml, go, c, c-header, cpp, cpp-header, js, ts, java, kotlin, csharp, php, swift, ruby, rust

FLAGS WITH OPTIONAL VALUES (use = syntax for values, e.g., --prefix=value):
  --prefix, -p            Enable prefix (default "v" if no value given)
  --prefix="V"            Use uppercase V prefix (only 'v' or 'V' allowed)
  --prerelease            Enable pre-release with config defaults
  --prerelease="..."      Use custom template (YOU provide dash separators)
  --metadata              Enable metadata with config defaults
  --metadata="..."        Use custom template (YOU provide dot separators)

IMPORTANT - SEPARATOR CONVENTIONS (per SemVer 2.0.0):
  Pre-release: Components separated by DASHES (e.g., "alpha-1", "beta-{{CommitsSinceTag}}")
               The leading dash (-) is auto-prepended via {{PreReleaseWithDash}}
  Metadata:    Components separated by DOTS (e.g., "{{BuildDateTimeCompact}}.{{ShortSha}}")
               The leading plus (+) is auto-prepended via {{MetadataWithPlus}}

TEMPLATE VARIABLES (Mustache syntax):

  Version Components:
    {{Major}}                - Major version (e.g., "1")
    {{Minor}}                - Minor version (e.g., "2")
    {{Patch}}                - Patch version (e.g., "3")
    {{MajorMinorPatch}}      - Core version: Major.Minor.Patch (e.g., "1.2.3")
    {{MajorMinor}}           - Major.Minor (e.g., "1.2")
    {{Prefix}}               - Version prefix (e.g., "v")

  Pre-release (rendered from --prerelease template):
    {{PreRelease}}           - Rendered pre-release (e.g., "alpha-5")
    {{PreReleaseWithDash}}   - With dash prefix (e.g., "-alpha-5") or empty

  Metadata (rendered from --metadata template):
    {{Metadata}}             - Rendered metadata (e.g., "20241211.abc1234")
    {{MetadataWithPlus}}     - With plus prefix (e.g., "+20241211.abc1234")

  VCS/Git Information:
    {{Hash}}                 - Full commit hash (40 chars for git)
    {{ShortHash}}            - Short commit hash (7 chars)
    {{MediumHash}}           - Medium commit hash (12 chars)
    {{BranchName}}           - Current branch (e.g., "feature/foo")
    {{EscapedBranchName}}    - Branch with slashes replaced (e.g., "feature-foo")
    {{CommitsSinceTag}}      - Commits since last tag (e.g., "42")
    {{BuildNumber}}          - Alias for CommitsSinceTag (GitVersion compatibility)
    {{BuildNumberPadded}}    - Padded to 4 digits (e.g., "0042")
    {{UncommittedChanges}}   - Count of dirty files (e.g., "3")
    {{Dirty}}                - "dirty" if uncommitted changes > 0, empty otherwise
    {{VersionSourceHash}}    - Hash of commit the last tag points to

  Commit Author:
    {{CommitAuthor}}         - Name of the commit author
    {{CommitAuthorEmail}}    - Email of the commit author

  Commit Timestamp (UTC):
    {{CommitDate}}           - ISO 8601: 2024-01-15T10:30:00Z
    {{CommitDateCompact}}    - Compact: 20240115103045 (YYYYMMDDHHmmss)
    {{CommitDateShort}}      - Date only: 2024-01-15

  Build Timestamp (UTC):
    {{BuildDateTimeUTC}}     - ISO 8601: 2024-01-15T10:30:00Z
    {{BuildDateTimeCompact}} - Compact: 20240115103045 (YYYYMMDDHHmmss)
    {{BuildDateUTC}}         - Date only: 2024-01-15
    {{BuildYear}}            - Year: 2024
    {{BuildMonth}}           - Month: 01 (zero-padded)
    {{BuildDay}}             - Day: 15 (zero-padded)

Use 'versionator vars' to see all template variables and their current values.

EXAMPLES:
  # Print Python version to stdout
  versionator emit python

  # With pre-release and metadata
  versionator emit python --prerelease "alpha" --metadata "{{ShortSha}}"

  # Use config defaults for prerelease/metadata
  versionator emit python --prerelease --metadata

  # Use custom template string
  versionator emit --template '{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}' \
    --prerelease "rc-1" --metadata "{{BuildDateTimeCompact}}"

  # Write to file
  versionator emit python --output mypackage/_version.py

  # Use template file
  versionator emit --template-file _version.tmpl.py --output _version.py

  # Dump a template for customization
  versionator emit dump python --output _version.tmpl.py
```

```bash
versionator output emit [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metadata` | string | - | Metadata template (uses config default if flag provided without value) |
| `-o, --output` | string | - | Output file path (default: stdout) |
| `-p, --prefix` | string | - | Version prefix (default 'v' if flag provided without value) |
| `--prerelease` | string | - | Pre-release template (uses config default if flag provided without value) |
| `-t, --template` | string | - | Custom Mustache template string |
| `-f, --template-file` | string | - | Path to template file |

### version

Show current version

```
Show the current version from VERSION file.

By default, outputs the full SemVer version (Major.Minor.Patch[-PreRelease][+Metadata]).

Use --template to customize the output format with Mustache syntax.

FLAGS WITH OPTIONAL VALUES (use = syntax for values, e.g., --prefix=value):
  --prefix, -p            Enable prefix (default "v" if no value given)
  --prefix="V"            Use uppercase V prefix (only 'v' or 'V' allowed)
  --prerelease            Enable pre-release with config defaults
  --prerelease="..."      Use custom template (YOU provide dash separators)
  --metadata              Enable metadata with config defaults
  --metadata="..."        Use custom template (YOU provide dot separators)

IMPORTANT - SEPARATOR CONVENTIONS (per SemVer 2.0.0):
  Pre-release: Components separated by DASHES (e.g., "alpha-1", "beta-{{CommitsSinceTag}}")
               The leading dash (-) is auto-prepended via {{PreReleaseWithDash}}
  Metadata:    Components separated by DOTS (e.g., "{{BuildDateTimeCompact}}.{{ShortSha}}")
               The leading plus (+) is auto-prepended via {{MetadataWithPlus}}

TEMPLATE VARIABLES:
  Version Components:
    {{Major}}            - Major version number
    {{Minor}}            - Minor version number
    {{Patch}}            - Patch version number
    {{MajorMinorPatch}}  - Major.Minor.Patch
    {{Prefix}}           - Version prefix (e.g., "v")

  Pre-release (rendered from --prerelease template):
    {{PreRelease}}         - Rendered pre-release (e.g., "alpha-5")
    {{PreReleaseWithDash}} - With dash prefix (e.g., "-alpha-5")

  Metadata (rendered from --metadata template):
    {{Metadata}}           - Rendered metadata (e.g., "20241211.abc1234")
    {{MetadataWithPlus}}   - With plus prefix (e.g., "+20241211.abc1234")

  VCS/Git (available in all templates):
    {{ShortHash}}        - Short commit hash (7 chars)
    {{MediumHash}}       - Medium commit hash (12 chars)
    {{Hash}}             - Full commit hash
    {{BranchName}}       - Current branch name
    {{CommitsSinceTag}}  - Commits since last tag
    {{BuildNumber}}      - Alias for CommitsSinceTag
    {{BuildNumberPadded}} - Padded to 4 digits (e.g., "0042")

  Commit Info:
    {{CommitDate}}       - Last commit datetime (ISO 8601)
    {{CommitDateCompact}} - Compact: 20241211103045
    {{CommitAuthor}}     - Commit author name
    {{CommitAuthorEmail}} - Commit author email

  Build Timestamps:
    {{BuildDateTimeCompact}} - Compact: 20241211103045
    {{BuildDateUTC}}         - Date only: 2024-12-11

  Custom Variables:
    Use --set key=value to inject custom variables
    Custom vars from .versionator.yaml config are also available

EXAMPLES:
  # Basic version (includes prerelease/metadata from VERSION file)
  versionator version                              # Output: 1.2.3-alpha+build.1

  # With prefix
  versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
                                                   # Output: v1.2.3

  # Full SemVer with prerelease and metadata
  versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
    --prerelease "alpha-{{CommitsSinceTag}}" \
    --metadata "{{BuildDateTimeCompact}}.{{ShortSha}}"
                                                   # Output: 1.2.3-alpha-5+20241211103045.abc1234

  # Use config defaults for prerelease/metadata
  versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
    --prerelease --metadata

  # With custom variables
  versionator version -t "{{AppName}} v{{MajorMinorPatch}}" --set AppName="My App"
```

```bash
versionator output version [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metadata` | string | - | Metadata template (uses config default if flag provided without value) |
| `-p, --prefix` | string | - | Version prefix (default 'v' if flag provided without value) |
| `--prerelease` | string | - | Pre-release template (uses config default if flag provided without value) |
| `--set` | stringArray | [] | Set custom variable (key=value), can be repeated |
| `-t, --template` | string | - | Template string for version output (Mustache syntax) |

