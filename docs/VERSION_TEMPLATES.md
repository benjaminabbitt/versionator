# Versionator: Version Templates Guide

This guide covers how to use pre-release identifiers, build metadata, and prefixes in versionator following [SemVer 2.0.0](https://semver.org/).

## SemVer 2.0.0 Format

```
{prefix}{major}.{minor}.{patch}-{prerelease}+{metadata}
         \_____core version____/
```

| Component | Separator | Example | Purpose |
|-----------|-----------|---------|---------|
| Prefix | (none) | `v`, `release-` | Optional version prefix |
| Core Version | `.` (dots) | `1.2.3` | Major.Minor.Patch |
| Pre-release | `-` (dash) then `-` between items | `-alpha-1` | Unstable version identifier |
| Metadata | `+` (plus) then `.` between items | `+20241211.abc1234` | Build information |

**Full Example**: `v1.2.3-alpha-5+20241211103045.abc1234`

---

## CLI Flags

All three flags (`--prefix`, `--prerelease`, `--metadata`) accept **optional template values**.

### Flag Syntax

**Important**: Use `=` syntax when providing values:

```bash
# Correct - use = for values
--prefix=release-
--prerelease="alpha-{{CommitsSinceTag}}"
--metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"

# Flag without value - uses defaults
--prefix      # Uses "v"
--prerelease  # Uses config defaults
--metadata    # Uses config defaults
```

### Flag Behavior Summary

| Flag | Without Value | With Value |
|------|---------------|------------|
| `--prefix` | Enable with default `"v"` | Use provided prefix |
| `--prerelease` | Enable with config template | Render provided template |
| `--metadata` | Enable with config template | Render provided template |

---

## Separator Conventions

### Pre-release: Use DASHES

Pre-release components are separated by **dashes** (`-`).

```bash
# YOU provide the dash separators in your template
--prerelease="alpha-1"
--prerelease="beta-{{CommitsSinceTag}}"
--prerelease="rc-1-{{EscapedBranchName}}"
```

The **leading dash** is auto-prepended when you use `{{PreReleaseWithDash}}` in your output template.

### Metadata: Use DOTS

Metadata components are separated by **dots** (`.`).

```bash
# YOU provide the dot separators in your template
--metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"
--metadata="{{CommitsSinceTag}}.{{ShortSha}}.{{BuildDateUTC}}"
```

The **leading plus** is auto-prepended when you use `{{MetadataWithPlus}}` in your output template.

---

## Template Variables

### Version Components

| Variable | Example | Description |
|----------|---------|-------------|
| `{{Major}}` | `1` | Major version number |
| `{{Minor}}` | `2` | Minor version number |
| `{{Patch}}` | `3` | Patch version number |
| `{{MajorMinorPatch}}` | `1.2.3` | Core version string |
| `{{MajorMinor}}` | `1.2` | Major.Minor only |
| `{{Prefix}}` | `v` | Version prefix |

### Pre-release (Rendered)

| Variable | Example | Description |
|----------|---------|-------------|
| `{{PreRelease}}` | `alpha-5` | Rendered pre-release (no leading dash) |
| `{{PreReleaseWithDash}}` | `-alpha-5` | With auto-prepended dash (empty if none) |

### Metadata (Rendered)

| Variable | Example | Description |
|----------|---------|-------------|
| `{{Metadata}}` | `20241211.abc1234` | Rendered metadata (no leading plus) |
| `{{MetadataWithPlus}}` | `+20241211.abc1234` | With auto-prepended plus (empty if none) |

### VCS/Git Information

| Variable | Example | Description |
|----------|---------|-------------|
| `{{Hash}}` | `abc1234def567...` | Full commit hash (40 chars for git) |
| `{{ShortHash}}` | `abc1234` | Short hash (7 chars) |
| `{{MediumHash}}` | `abc1234def01` | Medium hash (12 chars, configurable) |
| `{{BranchName}}` | `feature/foo` | Current branch name |
| `{{EscapedBranchName}}` | `feature-foo` | Branch with `/` replaced by `-` |
| `{{CommitsSinceTag}}` | `42` | Commits since last tag |
| `{{BuildNumber}}` | `42` | Alias for CommitsSinceTag (GitVersion compat) |
| `{{BuildNumberPadded}}` | `0042` | Padded to 4 digits |
| `{{UncommittedChanges}}` | `3` | Count of uncommitted changes |
| `{{Dirty}}` | `dirty` or `` | Non-empty if uncommitted changes exist |
| `{{VersionSourceHash}}` | `abc1234...` | Hash of commit the last tag points to |

### Commit Author

| Variable | Example | Description |
|----------|---------|-------------|
| `{{CommitAuthor}}` | `John Doe` | Name of the commit author |
| `{{CommitAuthorEmail}}` | `john@example.com` | Email of the commit author |

### Commit Timestamps (UTC)

| Variable | Example | Description |
|----------|---------|-------------|
| `{{CommitDate}}` | `2024-12-11T10:30:45Z` | ISO 8601 format |
| `{{CommitDateCompact}}` | `20241211103045` | Compact (YYYYMMDDHHmmss) |
| `{{CommitDateShort}}` | `2024-12-11` | Date only |
| `{{CommitYear}}` | `2024` | Year |
| `{{CommitMonth}}` | `12` | Month (zero-padded) |
| `{{CommitDay}}` | `11` | Day (zero-padded) |

### Build Timestamps (UTC)

| Variable | Example | Description |
|----------|---------|-------------|
| `{{BuildDateTimeUTC}}` | `2024-12-11T10:30:45Z` | ISO 8601 format |
| `{{BuildDateTimeCompact}}` | `20241211103045` | Compact (YYYYMMDDHHmmss) |
| `{{BuildDateUTC}}` | `2024-12-11` | Date only |
| `{{BuildYear}}` | `2024` | Year |
| `{{BuildMonth}}` | `12` | Month (zero-padded) |
| `{{BuildDay}}` | `11` | Day (zero-padded) |

---

## Usage Examples

### Basic Version Output

```bash
# Core version only (default)
versionator version
# Output: 1.2.3

# With prefix
versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
# Output: v1.2.3

# Custom prefix
versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix=release-
# Output: release-1.2.3
```

### Pre-release Versions

```bash
# Alpha release
versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" \
  --prerelease="alpha-1"
# Output: 1.2.3-alpha-1

# Beta with commit count
versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" \
  --prerelease="beta-{{CommitsSinceTag}}"
# Output: 1.2.3-beta-42

# Release candidate
versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" \
  --prerelease="rc-1"
# Output: 1.2.3-rc-1
```

### Build Metadata

```bash
# Timestamp and SHA
versionator version -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"
# Output: 1.2.3+20241211103045.abc1234

# Just the SHA
versionator version -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
  --metadata="{{MediumSha}}"
# Output: 1.2.3+abc1234def01
```

### Full SemVer 2.0.0

```bash
# Complete version string
versionator version \
  -t "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prefix \
  --prerelease="alpha-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"
# Output: v1.2.3-alpha-42+20241211103045.abc1234
```

### Emit to Files

```bash
# Python with pre-release
versionator emit python \
  --prerelease="beta-1" \
  --metadata="{{ShortSha}}" \
  --output mypackage/_version.py

# JSON with full version info
versionator emit json \
  --prerelease="rc-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}" \
  --output version.json
```

---

## Configuration File

Configure defaults in `.versionator.yaml`:

```yaml
# .versionator.yaml
prefix: "v"

prerelease:
  enabled: true
  template:
    - "alpha"                    # Static text
    - "{{CommitsSinceTag}}"      # Dynamic value
  # Result: "alpha-5" (items joined with dashes)

metadata:
  enabled: true
  type: "git"
  template:
    - "{{BuildDateTimeCompact}}" # 20241211103045
    - "{{ShortSha}}"             # abc1234
  # Result: "20241211103045.abc1234" (items joined with dots)
  git:
    hashLength: 12               # Length for {{MediumSha}}

logging:
  output: "console"              # console, json, development
```

### Config Template Arrays

The config uses **arrays** where items are automatically joined:
- Pre-release items joined with **dashes** (`-`)
- Metadata items joined with **dots** (`.`)

| Config | Result |
|--------|--------|
| `template: ["alpha", "1"]` | `alpha-1` |
| `template: ["{{BuildDateTimeCompact}}", "{{ShortSha}}"]` | `20241211103045.abc1234` |

### Using Config Defaults

When you use flags without values, the config templates are used:

```bash
# Uses prerelease.template and metadata.template from config
versionator version \
  -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prerelease \
  --metadata
```

---

## Common Patterns

### Development Builds

```yaml
# .versionator.yaml
prerelease:
  enabled: true
  template: ["dev", "{{CommitsSinceTag}}"]
metadata:
  enabled: true
  template: ["{{BuildDateTimeCompact}}", "{{ShortSha}}"]
```
Result: `1.2.3-dev-42+20241211103045.abc1234`

### CI/CD Builds

```bash
# In CI pipeline - dynamic pre-release based on branch
versionator version \
  -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prerelease="{{EscapedBranchName}}-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"
# Output: 1.2.3-feature-login-5+20241211103045.abc1234
```

### Release Candidates

```bash
# RC with number
versionator version \
  -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" \
  --prerelease="rc-1"
# Output: 1.2.3-rc-1
```

### Production (Metadata Only)

```bash
# No pre-release, just build metadata
versionator version \
  -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
  --metadata="{{MediumSha}}"
# Output: 1.2.3+abc1234def01
```

---

## Viewing Current Values

```bash
# Show all template variables and their current values
versionator vars

# Show just the core version
versionator version

# Show help with all template variables
versionator version --help
versionator emit --help
```
