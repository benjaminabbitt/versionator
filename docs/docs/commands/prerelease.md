---
title: prerelease
description: Manage pre-release identifier
---

# prerelease

Manage pre-release identifier

Commands to enable or disable pre-release identifiers.

Pre-release follows SemVer 2.0.0 specification:
- Appended with a dash (-) - this is added automatically
- Use DASHES (-) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3-alpha-5

## Usage

```bash
versionator prerelease [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `clear` | Clear pre-release value from VERSION file |
| `disable` | Disable pre-release identifier |
| `enable` | Enable pre-release identifier |
| `set` | Set pre-release value |
| `status` | Show pre-release status |
| `template` | Get or set the pre-release template |

### clear

Clear pre-release value from VERSION file

Remove the pre-release identifier from VERSION file

```bash
versionator prerelease clear
```

### disable

Disable pre-release identifier

Disable pre-release identifier by clearing it from the VERSION file.

The VERSION file is the source of truth - this command removes the pre-release from it directly.

```bash
versionator prerelease disable
```

### enable

Enable pre-release identifier

Enable pre-release identifier by rendering the config template and setting it in VERSION file.

If a template is configured in .versionator.yaml, it will be rendered and set as a static value.
If no template is configured, defaults to "alpha".

The VERSION file is the source of truth - this command writes to it directly.

```bash
versionator prerelease enable
```

### set

Set pre-release value

```
Set a static pre-release value in both config and VERSION file.

This updates:
1. The config file (.versionator.yaml) - so 'prerelease enable' can restore it
2. The VERSION file - the source of truth for the current version

Use 'prerelease template' for dynamic values with variables like {{CommitsSinceTag}}.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dots (e.g., "alpha.1")

Examples:
  versionator prerelease set alpha
  versionator prerelease set beta.1
  versionator prerelease set rc.2
```

```bash
versionator prerelease set
```

### status

Show pre-release status

Show current pre-release status from VERSION file (source of truth).

Also shows the configured template from .versionator.yaml if set.

```bash
versionator prerelease status
```

### template

Get or set the pre-release template

```
Get or set the pre-release template.

When setting a template, it is saved to .versionator.yaml config AND rendered
immediately to set the pre-release value in VERSION file.

IMPORTANT: Use DASHES (-) to separate pre-release identifiers per SemVer 2.0.0.
The leading dash (-) is added automatically - do NOT include it in your template.

The template uses Mustache syntax. Available variables:
  {{ShortHash}}            - Short git commit hash, 7 chars (e.g., "abc1234")
  {{MediumHash}}           - Medium git commit hash, 12 chars (e.g., "abc1234def01")
  {{Hash}}                 - Full git commit hash (40 chars)
  {{BranchName}}           - Current branch name
  {{EscapedBranchName}}    - Branch name with / replaced by -
  {{CommitsSinceTag}}      - Commits since last tag
  {{BuildDateTimeCompact}} - Compact timestamp (20241211103045)
  {{BuildDateUTC}}         - Date only (2024-12-11)
  {{CommitDate}}           - Commit date ISO 8601
  {{CommitDateCompact}}    - Commit date compact (20241211103045)

Examples:
  versionator prerelease template                              # Show current template
  versionator prerelease template "alpha"                      # Static "alpha"
  versionator prerelease template "alpha-{{CommitsSinceTag}}"  # "alpha-5"
  versionator prerelease template "rc-{{CommitsSinceTag}}"     # "rc-5"
  versionator prerelease template "beta-{{EscapedBranchName}}" # "beta-feature-foo"
```

```bash
versionator prerelease template
```

