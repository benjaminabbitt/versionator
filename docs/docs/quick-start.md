---
title: Quick Start
description: Get up and running with versionator in minutes
sidebar_position: 3
---

# Quick Start

This guide will get you up and running with versionator in just a few minutes.

## Initialize Your Project

Navigate to your project directory and run:

```bash
cd my-project
versionator init
```

This creates a `VERSION` file with the initial version `0.0.0`.

```bash
cat VERSION
# Output: 0.0.0

# Or initialize with a specific version and prefix
versionator init --version 1.0.0 --prefix v
cat VERSION
# Output: v1.0.0
```

## Increment Versions

Use the `major`, `minor`, and `patch` commands to increment versions following SemVer:

```bash
# Bump to first release
versionator major increment
# VERSION: 1.0.0

# Add a feature
versionator minor increment
# VERSION: 1.1.0

# Fix a bug
versionator patch increment
# VERSION: 1.1.1
```

### Shorthand Aliases

You can use `inc` and `dec` as shortcuts:

```bash
versionator patch inc    # Same as patch increment
versionator minor dec    # Same as minor decrement
```

## Add a Prefix

Many projects use `v` prefix for versions (e.g., `v1.0.0`):

```bash
# Enable prefix
versionator prefix set v

cat VERSION
# Output: v1.1.1
```

## Create Releases

Create git tags and release branches:

```bash
# Bump version and release
versionator patch increment
versionator release

# The release command will:
# 1. Auto-commit VERSION file (if it's the only dirty file)
# 2. Create an annotated tag (e.g., v1.1.1)
# 3. Create a release branch (e.g., release/v1.1.1)

# Push to remote
git push --tags
git push origin release/v1.1.1
```

## Use with Templates

Output version in custom formats using Mustache templates:

```bash
# Simple version
versionator version
# Output: 1.1.1

# With prefix
versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
# Output: v1.1.1

# Full SemVer with pre-release and metadata
versionator version \
  -t "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prefix \
  --prerelease="alpha-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortHash}}"
# Output: v1.1.1-alpha-5+20241211103045.abc1234
```

## Generate Version Files

Generate version information in your preferred programming language:

```bash
# Python
versionator emit python --output mypackage/_version.py

# JSON
versionator emit json --output version.json

# Go (compile-time injection recommended instead)
versionator emit go
```

See [Binary Embedding](./integration/binary-embedding) for language-specific examples and best practices.

## View All Variables

See all available template variables and their current values:

```bash
versionator vars
```

## Common Workflows

### Feature Development

```bash
# Start work on a feature
git checkout -b feature/new-feature

# Make changes...

# Commit your changes
git add .
git commit -m "Add new feature"
git checkout main
git merge feature/new-feature

# Bump version and release
versionator minor increment
versionator release
git push --tags
```

### Bug Fix

```bash
# Fix the bug and commit
git add .
git commit -m "Fix critical bug"

# Bump version and release
versionator patch increment
versionator release
git push --tags
```

### Pre-release

```bash
# Set up pre-release
versionator prerelease set alpha

cat VERSION
# Output: v1.2.0-alpha

# Increment pre-release number
versionator prerelease set alpha.2

# When ready to release
versionator prerelease clear
```

## Next Steps

- Learn about the [VERSION File Format](./concepts/version-file)
- Explore [Template Variables](./templates/variables)
- Set up [CI/CD Integration](./integration/cicd)
- See [Command Reference](./commands/) for all commands
