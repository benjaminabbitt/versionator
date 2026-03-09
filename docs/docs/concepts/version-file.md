---
title: VERSION File
description: Understanding the VERSION file format and discovery
sidebar_position: 1
---

# VERSION File

The `VERSION` file is the single source of truth for your project's version. It's a plain text file containing the current semantic version.

## File Format

The VERSION file contains a SemVer 2.0.0 version string:

```
[prefix]major.minor.patch[-prerelease][+metadata]
```

### Examples

```
0.0.0
1.2.3
v1.0.0
v2.5.3-alpha.1
v3.0.0+20241212.abc1234
v1.2.3-rc.1+build.456
```

### Components

| Component | Description | Example |
|-----------|-------------|---------|
| Prefix | Optional version prefix (`v` or `V` only) | `v` |
| Major | Major version number | `1` |
| Minor | Minor version number | `2` |
| Patch | Patch version number | `3` |
| Pre-release | Optional pre-release identifier | `-alpha.1`, `-rc.2` |
| Metadata | Optional build metadata | `+20241212.abc1234` |

## Source of Truth

:::important
The VERSION file is always the source of truth. Its content takes priority over any configuration in `.versionator.yaml`.
:::

- The prefix is parsed directly from the VERSION file (everything before the first digit)
- Config settings only apply as defaults when creating a new VERSION file
- Pre-release and metadata in the VERSION file are static values

## Static vs Dynamic Content

The VERSION file contains **static** values only. Dynamic content like git hashes, timestamps, or commit counts are rendered at **emit time** by versionator.

### When You Need Versionator

If your workflow uses **templates** or **dynamic content** in pre-release or metadata:

```yaml
# .versionator.yaml
prerelease:
  template: "alpha-{{CommitsSinceTag}}"
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"
```

You'll need versionator installed in your CI container to:
- Generate code embeddings (`versionator emit`)
- Set CI/CD environment variables (`versionator ci`)
- Render templates with current git/build context

Versionator is a **static binary** with no dependencies, making installation straightforward:

```bash
# Download and install (single binary, no dependencies)
curl -sSL https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64 -o /usr/local/bin/versionator
chmod +x /usr/local/bin/versionator
```

### When You Can Just Read VERSION

If you're **not** using dynamic pre-release or metadata content, the VERSION file is plain text that any tool can read:

```bash
# Shell
VERSION=$(cat VERSION)

# Make
VERSION := $(shell cat VERSION)

# Python
version = open('VERSION').read().strip()
```

No versionator installation required for these simple use cases.

## File Discovery

Versionator walks up the directory tree from the current working directory looking for a VERSION file. This enables nested projects with independent versions.

### Discovery Order

1. Check current directory for `VERSION`
2. Walk up to parent directory
3. Repeat until found or filesystem root is reached
4. If not found, create `VERSION` in current directory with `0.0.1`

### Example Directory Structure

```
myproject/
├── VERSION              # 1.0.0
├── packages/
│   ├── VERSION          # 2.0.0
│   └── core/
│       ├── VERSION      # 3.0.0
│       └── src/
└── apps/
    └── web/             # No VERSION file
```

Running `versionator version` from different directories:

```bash
# From myproject/
versionator version          # 1.0.0

# From myproject/packages/core/
versionator version          # 3.0.0

# From myproject/packages/core/src/
versionator version          # 3.0.0 (walks up to packages/core/)

# From myproject/apps/web/
versionator version          # Creates VERSION with 0.0.1
```

## Creating the VERSION File

The VERSION file is created automatically on first use:

```bash
# In a directory without VERSION
versionator version
# Creates VERSION with: 0.0.1

# With prefix enabled in config
versionator version
# Creates VERSION with: v0.0.1
```

You can also create it manually:

```bash
echo "1.0.0" > VERSION
```

## Editing the VERSION File

While you can edit the VERSION file manually, it's recommended to use versionator commands:

```bash
# Increment versions
versionator major increment
versionator minor increment
versionator patch increment

# Set prefix
versionator prefix set v

# Set pre-release
versionator prerelease set alpha.1

# Set metadata
versionator metadata set build.123
```

This ensures the version remains valid SemVer format.

## Version Control

The VERSION file should be committed to version control:

```bash
git add VERSION
git commit -m "Bump version to 1.2.3"
```

This provides a clear history of version changes in your repository.

## See Also

- [SemVer 2.0.0](./semver) - Semantic Versioning specification
- [Version Grammar](./version-grammar) - Understanding version string format
- [Monorepo Support](./monorepo) - Managing versions in monorepos
- [VERSION Grammar (EBNF)](https://github.com/benjaminabbitt/versionator/blob/master/docs/grammar/version.ebnf) - Formal grammar specification
