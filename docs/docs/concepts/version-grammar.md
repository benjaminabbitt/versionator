---
title: Version Grammar
description: Understanding version string format and syntax
sidebar_position: 4
---

# Version Grammar

This guide explains how version strings work in plain English. Whether you're tagging releases, configuring CI/CD, or just trying to understand what `v1.2.3-beta.1+build.456` means, this page has you covered.

For the formal specification, see the [EBNF grammar file](https://github.com/benjaminabbitt/versionator/blob/master/docs/grammar/version.ebnf).

## What is a Version String?

A version string is a structured way to communicate information about a software release. Instead of saying "the new version" or "the February update," version strings provide precise, comparable identifiers.

```
v1.2.3-beta.1+build.456
```

This single string tells us:
- It's version 1.2.3
- It's a beta pre-release (the first one)
- It was built with identifier "build.456"

## Anatomy of a Version

Every version string follows this pattern:

```
[prefix]MAJOR.MINOR.PATCH[-prerelease][+metadata]
```

Let's break down each part:

### The Prefix (Optional)

```
v1.0.0
V2.3.4
1.0.0
```

The prefix is a single letter that appears before the numbers. Versionator only allows:

| Prefix | Meaning |
|--------|---------|
| `v` | Lowercase "v" (most common) |
| `V` | Uppercase "V" |
| *(none)* | No prefix at all |

**Why only v/V?** This follows conventions established by Go modules, npm, and git tags. Other prefixes like "ver" or "version-" are not supported because they create inconsistency across tools.

### The Core Version (Required)

The core version has three numbers separated by dots:

```
MAJOR.MINOR.PATCH
```

| Component | What It Means | When to Change |
|-----------|---------------|----------------|
| **MAJOR** | Breaking changes | You changed something that breaks existing code |
| **MINOR** | New features | You added something new (backwards compatible) |
| **PATCH** | Bug fixes | You fixed a bug (backwards compatible) |

**Examples:**

| Version | Interpretation |
|---------|----------------|
| `1.0.0` | First stable release |
| `2.0.0` | Major rewrite or breaking changes from 1.x |
| `1.5.0` | Added new features to version 1 |
| `1.5.3` | Third bug fix release for version 1.5 |

**The rules:**
- All three numbers are required
- Numbers can be any non-negative integer (0, 1, 2, ... 999, etc.)
- No leading zeros (use `1.2.3`, not `01.02.03`)

### Pre-release Identifier (Optional)

Pre-release versions come *before* the final release. They're for testing and early access.

```
1.0.0-alpha
1.0.0-beta.1
1.0.0-rc.2
```

The pre-release starts with a hyphen (`-`) followed by identifiers separated by dots.

**Common pre-release labels:**

| Label | Meaning | Typical Use |
|-------|---------|-------------|
| `alpha` | Very early, unstable | Internal testing |
| `beta` | Feature complete, may have bugs | External testing |
| `rc` | Release candidate | Final testing before release |

**Adding numbers:**

You can add numbers to track iterations:

```
1.0.0-alpha      # First alpha
1.0.0-alpha.1    # Alpha iteration 1
1.0.0-alpha.2    # Alpha iteration 2
1.0.0-beta       # Move to beta
1.0.0-beta.1     # Beta iteration 1
1.0.0-rc.1       # First release candidate
1.0.0-rc.2       # Second release candidate
1.0.0            # Final release!
```

**Important:** A pre-release version is always *less than* the normal version:

```
1.0.0-alpha < 1.0.0-beta < 1.0.0-rc.1 < 1.0.0
```

This means `1.0.0-rc.1` comes before `1.0.0` in sort order, which is what you want.

### Build Metadata (Optional)

Build metadata provides additional information about a specific build. It starts with a plus sign (`+`).

```
1.0.0+20241215
1.0.0+build.123
1.0.0-beta.1+sha.abc1234
```

**Common metadata:**

| Example | What It Contains |
|---------|------------------|
| `+20241215` | Build date |
| `+build.123` | CI build number |
| `+sha.abc1234` | Git commit hash |
| `+20241215.abc1234` | Date and commit |

**Important:** Build metadata is ignored when comparing versions:

```
1.0.0+build.1 == 1.0.0+build.999  # Same version!
```

This makes sense because both builds represent the same *release*, just built at different times or on different machines.

## Putting It All Together

Here are complete examples showing all parts:

| Version | Prefix | Core | Pre-release | Metadata |
|---------|--------|------|-------------|----------|
| `1.0.0` | - | 1.0.0 | - | - |
| `v1.0.0` | v | 1.0.0 | - | - |
| `v2.1.3-alpha` | v | 2.1.3 | alpha | - |
| `v1.0.0-rc.1` | v | 1.0.0 | rc.1 | - |
| `1.0.0+build.456` | - | 1.0.0 | - | build.456 |
| `v3.2.1-beta.2+20241215.abc1234` | v | 3.2.1 | beta.2 | 20241215.abc1234 |

## Character Rules

Version strings have specific rules about what characters are allowed:

### In Core Version Numbers

Only digits `0-9`. No letters, no symbols.

```
1.2.3      ✓ Valid
1.2.3a     ✗ Invalid (letter in patch)
1.2.3.4    ✗ Invalid for SemVer (too many parts)
```

### In Pre-release and Metadata

Allowed characters:
- Digits: `0-9`
- Letters: `a-z`, `A-Z`
- Hyphen: `-`

```
1.0.0-alpha       ✓ Valid
1.0.0-alpha.1     ✓ Valid
1.0.0-my-feature  ✓ Valid (hyphens OK)
1.0.0-alpha_1     ✗ Invalid (underscore not allowed)
1.0.0-alpha 1     ✗ Invalid (space not allowed)
```

### No Leading Zeros in Numbers

Numeric identifiers cannot have leading zeros:

```
1.0.0-alpha.1    ✓ Valid
1.0.0-alpha.01   ✗ Invalid (leading zero)
1.0.0-alpha.0    ✓ Valid (just zero is fine)
```

## Version Comparison (Sorting)

When versions are sorted, they follow these rules:

### 1. Compare Core Numbers Left to Right

```
1.0.0 < 2.0.0      # Major differs
1.1.0 < 1.2.0      # Minor differs
1.1.1 < 1.1.2      # Patch differs
```

### 2. Pre-release Comes Before Release

```
1.0.0-alpha < 1.0.0
1.0.0-rc.99 < 1.0.0
```

### 3. Pre-release Identifiers Compare Left to Right

```
1.0.0-alpha < 1.0.0-beta           # "alpha" < "beta" alphabetically
1.0.0-alpha.1 < 1.0.0-alpha.2      # 1 < 2 numerically
1.0.0-alpha < 1.0.0-alpha.1        # Shorter < longer
```

### 4. Numbers Compare as Numbers, Not Text

```
1.0.0-alpha.2 < 1.0.0-alpha.10     # 2 < 10 (numeric comparison)
```

This is important! Text sorting would put `10` before `2`, but version sorting correctly recognizes `10 > 2`.

### 5. Build Metadata is Ignored

```
1.0.0+build.1 == 1.0.0+build.999   # Same precedence
```

## Complete Example: A Release Cycle

Here's how versions might progress through a typical release:

```
0.1.0              # Initial development
0.2.0              # More features
0.9.0              # Getting close to 1.0
1.0.0-alpha        # First 1.0 preview
1.0.0-alpha.1      # Alpha bugfix
1.0.0-alpha.2      # More alpha fixes
1.0.0-beta         # Feature complete, testing
1.0.0-beta.1       # Beta bugfix
1.0.0-rc.1         # Release candidate
1.0.0-rc.2         # Fix last-minute issue
1.0.0              # Official release!
1.0.1              # Patch release (bugfix)
1.1.0              # Minor release (new feature)
2.0.0              # Major release (breaking changes)
```

## Special Formats

### Go Module Versions

Go requires a `v` prefix and has a special format for untagged commits called "pseudo-versions":

```
v1.2.3                                    # Normal version
v0.0.0-20241215103045-abc123def456        # Pseudo-version
```

The pseudo-version contains:
- A base version (`v0.0.0`)
- A timestamp (`20241215103045` = 2024-12-15 10:30:45 UTC)
- A commit hash prefix (`abc123def456`)

### Dynamic Content in Versionator

Versionator can generate version components dynamically using templates:

```yaml
# .versionator.yaml
prerelease:
  template: "alpha-{{CommitsSinceTag}}"
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"
```

This might produce: `1.0.0-alpha-5+20241215103045.abc1234`

The `{{...}}` parts are replaced with actual values when you run versionator.

## Quick Reference

| Part | Required | Starts With | Allowed Characters | Example |
|------|----------|-------------|-------------------|---------|
| Prefix | No | - | `v` or `V` only | `v` |
| Major | Yes | - | digits | `1` |
| Minor | Yes | `.` | digits | `.2` |
| Patch | Yes | `.` | digits | `.3` |
| Pre-release | No | `-` | digits, letters, `-`, `.` | `-beta.1` |
| Metadata | No | `+` | digits, letters, `-`, `.` | `+build.456` |

## See Also

- [VERSION File](./version-file) - How versionator stores versions
- [Semantic Versioning](./semver) - SemVer 2.0.0 explained
- [Grammar-Based Parser](./grammar) - How versionator parses versions (technical)
- [EBNF Grammar](https://github.com/benjaminabbitt/versionator/blob/master/docs/grammar/version.ebnf) - Formal grammar specification
