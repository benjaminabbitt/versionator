---
title: emit
description: Emit version in various formats
---

# emit

Emit version in various formats

```
Emit the current version in various programming language formats.

Supported formats: python, json, yaml, go, c, c-header, cpp, cpp-header, js, ts, java, kotlin, csharp, php, swift, ruby, rust

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

## Installation in CI/Build Systems

The `emit` command renders dynamic content (git hashes, timestamps, commit counts) at **build time**. This requires versionator to be installed where you generate code embeddings.

Versionator is a **static binary** with no dependencies:

```bash
# Add to your CI pipeline or build container
curl -sSL https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64 -o /usr/local/bin/versionator
chmod +x /usr/local/bin/versionator
```

:::tip
If you're not using dynamic pre-release or metadata templates, you can simply read the VERSION file directly without installing versionator. See [VERSION File - Static vs Dynamic Content](../concepts/version-file#static-vs-dynamic-content).
:::

## Usage

```bash
versionator emit [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `dump` | Dump embedded template to filesystem for customization |

### dump

Dump embedded template to filesystem for customization

Dump the embedded template for a format to the filesystem.

This allows you to customize the template and use it with --template-file.

Supported formats: python, json, yaml, go, c, c-header, cpp, cpp-header, js, ts, java, kotlin, csharp, php, swift, ruby, rust

See 'versionator emit --help' for the full list of template variables.

Examples:
  # Print Python template to stdout
  versionator emit dump python

  # Save Python template to file for editing
  versionator emit dump python --output _version.tmpl.py

  # Then use your customized template
  versionator emit --template-file _version.tmpl.py --output _version.py

```bash
versionator emit dump [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-o, --output` | string | - | Output file path (default: stdout) |

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metadata` | string | - | Metadata template (uses config default if flag provided without value) |
| `-o, --output` | string | - | Output file path (default: stdout) |
| `-p, --prefix` | string | - | Version prefix (default 'v' if flag provided without value) |
| `--prerelease` | string | - | Pre-release template (uses config default if flag provided without value) |
| `-t, --template` | string | - | Custom Mustache template string |
| `-f, --template-file` | string | - | Path to template file |

