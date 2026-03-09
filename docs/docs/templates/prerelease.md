---
title: Pre-release Templates
description: Configuring pre-release identifiers with templates
sidebar_position: 2
---

# Pre-release Templates

Pre-release identifiers mark versions as unstable or in-progress (e.g., `1.0.0-alpha.1`, `2.0.0-rc.1`).

## Stability Model

Pre-release supports two modes controlled by the `stable` setting:

### Dynamic (Default: stable: false)

When `stable: false` (the default), pre-release is **generated from template at output time**:

```yaml
# .versionator.yaml
prerelease:
  template: "build-{{CommitsSinceTag}}"
  stable: false  # Default
```

Every time you run `emit`, `ci`, or `output` commands, the template is evaluated:

```bash
versionator output version
# Output: 1.0.0-build-42

# After more commits:
versionator output version
# Output: 1.0.0-build-45
```

This is ideal for **continuous delivery** workflows where every build should have a unique version.

### Static (stable: true)

When `stable: true`, pre-release is **stored in the VERSION file**:

```bash
# Enable stable mode
versionator config prerelease stable true

# Now set a fixed value
versionator config prerelease set alpha
# VERSION: 1.0.0-alpha

versionator config prerelease set beta.1
# VERSION: 1.0.0-beta.1
```

This is ideal for **traditional release** workflows (alpha → beta → rc → release).

Clear when ready to release:

```bash
versionator config prerelease clear
# VERSION: 1.0.0
```

### Checking Current Mode

```bash
versionator config prerelease stable
# Output: false (or true)
```

## Template Configuration

Set a default template in `.versionator.yaml`:

```yaml
prerelease:
  template: "alpha-{{CommitsSinceTag}}"
```

Then use with the flag:

```bash
versionator output version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" --prerelease
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
versionator config prerelease set alpha

# Alpha iterations
versionator config prerelease set alpha.2
versionator config prerelease set alpha.3

# Move to beta
versionator config prerelease set beta

# Release candidate
versionator config prerelease set rc.1
versionator config prerelease set rc.2

# Final release
versionator config prerelease clear
```

## Commands

### prerelease set

Sets a pre-release value. Behavior depends on stability:

**When stable: true:**
```bash
versionator config prerelease set alpha
# VERSION: 1.0.0-alpha
```

**When stable: false (default):**
```bash
versionator config prerelease set alpha
# Error: cannot set literal pre-release when stable is false
# Use --force to set as template, or set stable to true first

# Force it (sets as template):
versionator config prerelease set alpha --force
# Config: template = "alpha"
```

### prerelease template

Sets a template in config:

```bash
versionator config prerelease template "build-{{CommitsSinceTag}}"
# Config: template = "build-{{CommitsSinceTag}}"
```

When `stable: false`, this template is rendered at output time.
When `stable: true`, the template is rendered and stored in VERSION immediately.

### prerelease stable

Get or set the stability mode:

```bash
# Get current setting
versionator config prerelease stable
# Output: false

# Enable stable mode (store in VERSION)
versionator config prerelease stable true

# Disable stable mode (generate from template)
versionator config prerelease stable false
```

### prerelease clear

Clears pre-release from VERSION file:

```bash
versionator config prerelease clear
# VERSION: 1.0.0
```

Only works when `stable: true`. When `stable: false`, use an empty template instead.

### prerelease status

Shows current state:

```bash
versionator config prerelease status
# Stable: false
# Template: build-{{CommitsSinceTag}}
# Rendered: build-42
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
