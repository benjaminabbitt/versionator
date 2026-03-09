---
title: version
description: Show current version
---

# version

Show current version

```
Show the current version from VERSION file.

By default, outputs the full SemVer version (Major.Minor.Patch[-PreRelease][+Metadata]).

Use --template to customize the output format with Mustache syntax.

FLAGS WITH OPTIONAL VALUES (use = syntax for values, e.g., --prefix=value):
  --prefix, -p            Enable prefix (default "v" if no value given)
  --prefix="V"            Use uppercase V prefix (only "v" or "V" allowed)
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

## Usage

```bash
versionator version [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metadata` | string | - | Metadata template (uses config default if flag provided without value) |
| `-p, --prefix` | string | - | Version prefix (default 'v' if flag provided without value) |
| `--prerelease` | string | - | Pre-release template (uses config default if flag provided without value) |
| `--set` | stringArray | [] | Set custom variable (key=value), can be repeated |
| `-t, --template` | string | - | Template string for version output (Mustache syntax) |

