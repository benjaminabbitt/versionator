---
title: Pre-release Templates
description: Configuring pre-release identifiers with templates
sidebar_position: 2
---

# Pre-release Templates

Pre-release identifiers mark versions as unstable or in-progress (e.g., `1.0.0-alpha.1`, `2.0.0-rc.1`).

## Static vs Dynamic Pre-release

### Static (Stored in VERSION file)

Set a fixed pre-release value:

```bash
versionator prerelease set alpha
# VERSION: 1.0.0-alpha

versionator prerelease set beta.1
# VERSION: 1.0.0-beta.1

versionator prerelease set rc.2
# VERSION: 1.0.0-rc.2
```

Clear when ready to release:

```bash
versionator prerelease clear
# VERSION: 1.0.0
```

### Dynamic (Template-based)

Use templates for values that change (like commit count):

```bash
versionator version \
  -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" \
  --prerelease="alpha-{{CommitsSinceTag}}"
# Output: 1.0.0-alpha-5
```

Dynamic values are computed at runtime and don't modify the VERSION file.

## Template Configuration

Set a default template in `.versionator.yaml`:

```yaml
prerelease:
  template: "alpha-{{CommitsSinceTag}}"
```

Then use with the flag:

```bash
versionator version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" --prerelease
# Output: 1.0.0-alpha-5
```

## Separator Convention

Pre-release components use **dashes** (`-`) as separators:

```yaml
# Correct - use dashes
prerelease:
  template: "alpha-{{CommitsSinceTag}}"      # alpha-5
  template: "beta-1-{{EscapedBranchName}}"   # beta-1-feature-foo

# Incorrect - don't use dots for pre-release components
prerelease:
  template: "alpha.{{CommitsSinceTag}}"      # Avoid
```

The leading dash before the pre-release is automatically added when using `{{PreReleaseWithDash}}`.

## Common Patterns

### Development Builds

```yaml
prerelease:
  template: "dev-{{CommitsSinceTag}}"
```

Result: `1.0.0-dev-42`

### Branch-based

```yaml
prerelease:
  template: "{{EscapedBranchName}}-{{CommitsSinceTag}}"
```

Result: `1.0.0-feature-login-5`

### CI Build Numbers

```yaml
prerelease:
  template: "build-{{BuildNumberPadded}}"
```

Result: `1.0.0-build-0042`

### Alpha/Beta/RC Workflow

```bash
# Alpha phase
versionator prerelease set alpha

# Alpha iterations
versionator prerelease set alpha.2
versionator prerelease set alpha.3

# Move to beta
versionator prerelease set beta

# Release candidate
versionator prerelease set rc.1
versionator prerelease set rc.2

# Final release
versionator prerelease clear
```

## Enable/Disable Commands

### prerelease set

Sets a static value in the VERSION file:

```bash
versionator prerelease set alpha
# VERSION: 1.0.0-alpha
```

### prerelease template

Sets a template in config and renders it to VERSION:

```bash
versionator prerelease template "alpha-{{CommitsSinceTag}}"
# Config: template = "alpha-{{CommitsSinceTag}}"
# VERSION: 1.0.0-alpha-5
```

### prerelease enable

Renders the config template to VERSION:

```bash
versionator prerelease enable
# Reads template from config
# VERSION: 1.0.0-alpha-5
```

### prerelease disable

Clears pre-release from VERSION (preserves config template):

```bash
versionator prerelease disable
# VERSION: 1.0.0
# Config template still saved
```

### prerelease clear

Clears pre-release from VERSION:

```bash
versionator prerelease clear
# VERSION: 1.0.0
```

### prerelease status

Shows current state:

```bash
versionator prerelease status
# Pre-release: ENABLED
# Value: alpha-5
```

## Variables for Pre-release

Common variables useful in pre-release templates:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{CommitsSinceTag}}` | Commits since last tag | `42` |
| `{{BuildNumber}}` | Alias for CommitsSinceTag | `42` |
| `{{BuildNumberPadded}}` | Padded to 4 digits | `0042` |
| `{{EscapedBranchName}}` | Branch name (safe chars) | `feature-login` |
| `{{ShortHash}}` | Short commit hash | `abc1234` |

## Version Increment Behavior

:::note
Pre-release is **automatically cleared** when incrementing major, minor, or patch version. This follows SemVer 2.0.0 specification.
:::

```bash
# Starting: 1.0.0-alpha
versionator patch increment
# Result: 1.0.1 (not 1.0.1-alpha)
```

## Sorting Pre-releases

SemVer pre-release precedence (lowest to highest):

```
1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta
1.0.0-alpha < 1.0.0-beta < 1.0.0-rc < 1.0.0
```

## See Also

- [Semantic Versioning](../concepts/semver) - SemVer spec details
- [Template Variables](./variables) - All available variables
