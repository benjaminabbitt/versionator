---
title: Semantic Versioning
description: Understanding SemVer 2.0.0 and how versionator implements it
sidebar_position: 2
---

# Semantic Versioning

Versionator follows [Semantic Versioning 2.0.0](https://semver.org/) (SemVer) for version management. This page explains the specification and how versionator implements it.

## Version Format

```
MAJOR.MINOR.PATCH[-PRERELEASE][+METADATA]
```

### Core Version

The core version consists of three non-negative integers:

| Component | When to Increment | Resets |
|-----------|-------------------|--------|
| **MAJOR** | Incompatible API changes | MINOR, PATCH to 0 |
| **MINOR** | New backwards-compatible functionality | PATCH to 0 |
| **PATCH** | Backwards-compatible bug fixes | Nothing |

### Pre-release Identifier

Pre-release versions are denoted by a hyphen followed by identifiers:

```
1.0.0-alpha
1.0.0-alpha.1
1.0.0-beta.2
1.0.0-rc.1
```

**Rules:**
- Identifiers consist of alphanumerics and hyphens `[0-9A-Za-z-]`
- Numeric identifiers must not have leading zeros
- Pre-release versions have lower precedence than normal versions
- Identifiers are separated by dots (`.`) in SemVer, but versionator uses dashes (`-`) for template input for clarity

### Build Metadata

Build metadata is denoted by a plus sign followed by identifiers:

```
1.0.0+20241212
1.0.0+build.123
1.0.0-alpha.1+001
```

**Rules:**
- Identifiers consist of alphanumerics and hyphens `[0-9A-Za-z-]`
- Identifiers are separated by dots (`.`)
- Build metadata is **ignored** when determining version precedence
- Two versions differing only in metadata are considered equal for ordering

## Versionator Implementation

### Incrementing Behavior

When you increment a version component, versionator follows SemVer rules:

```bash
# Starting version: 1.2.3-alpha

versionator major increment
# Result: 2.0.0 (minor, patch reset; pre-release cleared)

versionator minor increment
# Result: 1.3.0 (patch reset; pre-release cleared)

versionator patch increment
# Result: 1.2.4 (pre-release cleared)
```

:::note
Pre-release is always cleared when incrementing any version component, per SemVer spec.
:::

### Decrementing Behavior

Decrementing works similarly but doesn't reset other components:

```bash
# Starting version: 2.1.3

versionator major decrement
# Result: 1.1.3

versionator minor decrement
# Result: 2.0.3

versionator patch decrement
# Result: 2.1.2
```

### Pre-release Separators

In the VERSION file and output, pre-release components use **dashes** (`-`):

```
1.2.3-alpha-1
1.2.3-beta-2
1.2.3-rc-1
```

When specifying pre-release templates, you provide the dashes:

```bash
versionator version --prerelease="alpha-{{CommitsSinceTag}}"
# Output: 1.2.3-alpha-5
```

### Metadata Separators

Build metadata components use **dots** (`.`):

```
1.2.3+20241212.abc1234
1.2.3+build.123.sha.abc1234
```

When specifying metadata templates:

```bash
versionator version --metadata="{{BuildDateTimeCompact}}.{{ShortHash}}"
# Output: 1.2.3+20241211103045.abc1234
```

## Version Precedence

Versions are compared for precedence according to SemVer:

1. Compare MAJOR, MINOR, PATCH numerically
2. A version with pre-release has lower precedence than the normal version
3. Pre-release identifiers are compared left-to-right
4. Numeric identifiers compared numerically, alphanumeric lexically
5. Numeric identifiers have lower precedence than alphanumeric
6. Build metadata is ignored in precedence

### Examples

```
1.0.0 < 2.0.0 < 2.1.0 < 2.1.1
1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta
1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0
```

## Common Patterns

### Alpha/Beta/RC Workflow

```bash
# Development starts
versionator minor increment
versionator prerelease set alpha
# Result: 1.1.0-alpha

# Alpha iterations
versionator prerelease set alpha.2
versionator prerelease set alpha.3

# Move to beta
versionator prerelease set beta

# Release candidate
versionator prerelease set rc.1

# Final release
versionator prerelease clear
# Result: 1.1.0
```

### Nightly Builds

Using metadata for build identification:

```bash
versionator version \
  -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortHash}}"
# Output: 1.1.0+20241211103045.abc1234
```

## See Also

- [Pre-release Templates](../templates/prerelease) - Dynamic pre-release identifiers
- [Metadata Templates](../templates/metadata) - Build metadata configuration
- [SemVer.org](https://semver.org/) - Official specification
