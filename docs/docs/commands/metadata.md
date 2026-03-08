---
title: metadata
description: Manage build metadata
---

# metadata

Manage build metadata

Commands to enable or disable appending build metadata to version numbers.

Build metadata follows SemVer 2.0.0 specification:
- Appended with a plus sign (+) - this is added automatically
- Use DOTS (.) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3+20241211103045.abc1234def5

## Usage

```bash
versionator metadata [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `clear` | Clear metadata value from VERSION file |
| `configure` | Show metadata configuration |
| `disable` | Disable build metadata |
| `enable` | Enable build metadata |
| `set` | Set metadata value |
| `status` | Show metadata status |
| `template` | Get or set the metadata template |

### clear

Clear metadata value from VERSION file

Remove the build metadata from VERSION file

```bash
versionator metadata clear
```

### configure

Show metadata configuration

Show metadata configuration from .versionator.yaml

```bash
versionator metadata configure
```

### disable

Disable build metadata

Disable build metadata by clearing it from the VERSION file.

The VERSION file is the source of truth - this command removes the metadata from it directly.

```bash
versionator metadata disable
```

### enable

Enable build metadata

Enable build metadata by rendering the config template and setting it in VERSION file.

If a template is configured in .versionator.yaml, it will be rendered and set as a static value.
If no template is configured, defaults to the git short hash.

The VERSION file is the source of truth - this command writes to it directly.

```bash
versionator metadata enable
```

### set

Set metadata value

```
Set a static metadata value in both config and VERSION file.

This updates:
1. The config file (.versionator.yaml) - so 'metadata enable' can restore it
2. The VERSION file - the source of truth for the current version

Use 'metadata template' for dynamic values with variables like {{ShortHash}}.

The value must follow SemVer 2.0.0:
- Only alphanumerics and hyphens [0-9A-Za-z-]
- Separate identifiers with dots (e.g., "build.123")

Examples:
  versionator metadata set build.123
  versionator metadata set 20241211103045
  versionator metadata set ci.456.linux
```

```bash
versionator metadata set
```

### status

Show metadata status

Show current metadata status from VERSION file (source of truth).

Also shows the configured template from .versionator.yaml if set.

```bash
versionator metadata status
```

### template

Get or set the metadata template

```
Get or set the metadata template used for build metadata.

When setting a template, it is saved to .versionator.yaml config AND rendered
immediately to set the metadata value in VERSION file.

IMPORTANT: Use DOTS (.) to separate metadata identifiers per SemVer 2.0.0.
The leading plus (+) is added automatically - do NOT include it in your template.

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
  versionator metadata template                                              # Show current
  versionator metadata template "{{BuildDateTimeCompact}}.{{MediumHash}}"    # Timestamp.hash
  versionator metadata template "{{ShortHash}}"                              # Just git hash
  versionator metadata template "{{CommitsSinceTag}}.{{ShortHash}}"          # Build number.hash
```

```bash
versionator metadata template
```

