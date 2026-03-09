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

This creates a `VERSION` file with the initial version `0.0.1`.

```bash
cat VERSION
# Output: 0.0.1

# Or initialize with a specific version and prefix
versionator init --version 1.0.0 --prefix v
cat VERSION
# Output: v1.0.0
```

## Increment Versions

Use `bump` with `major`, `minor`, or `patch` subcommands to increment versions following SemVer:

```bash
# Bump to first release
versionator bump major increment
# VERSION: 1.0.0

# Add a feature
versionator bump minor increment
# VERSION: 1.1.0

# Fix a bug
versionator bump patch increment
# VERSION: 1.1.1
```

### Auto-bump from Commit Messages

Use `versionator bump` without arguments to automatically detect version changes from commit messages:

```bash
# Commit with +semver: marker
git commit -m "Add new feature +semver:minor"

# Auto-bump based on commit message
versionator bump
# VERSION: 1.2.0
```

## Add a Prefix

Many projects use `v` prefix for versions (e.g., `v1.0.0`):

```bash
# Enable prefix
versionator config prefix set v

cat VERSION
# Output: v1.1.1
```

## Create Releases

Create git tags and release branches:

```bash
# Bump version and release
versionator bump patch increment
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
versionator output version
# Output: 1.1.1

# With prefix
versionator output version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
# Output: v1.1.1

# Full SemVer with pre-release and metadata
versionator output version \
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
versionator output emit python --output mypackage/_version.py

# JSON
versionator output emit json --output version.json

# Go (compile-time injection recommended instead)
versionator output emit go
```

See [Binary Embedding](./integration/binary-embedding) for language-specific examples and best practices.

## View All Variables

See all available template variables and their current values:

```bash
versionator config vars
```

## Common Workflows

### Auto-bump with Release Push (Recommended)

The fastest workflow uses `+semver:` markers in commit messages for automatic version detection:

```bash
# Stage and commit with semver marker
git add -A
git commit -m "Add new feature +semver:minor"

# Auto-detect bump level from commit, then release and push
versionator bump
versionator release push
```

The `+semver:` marker tells versionator what type of change this is:
- `+semver:major` - Breaking changes (1.0.0 → 2.0.0)
- `+semver:minor` - New features (1.0.0 → 1.1.0)
- `+semver:patch` - Bug fixes (1.0.0 → 1.0.1)
- `+semver:skip` - No version bump needed

The `release push` command creates the tag and release branch, then pushes both to the remote in one step.

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
versionator bump minor increment
versionator release push
```

### Bug Fix

```bash
# Fix the bug and commit
git add .
git commit -m "Fix critical bug"

# Bump version and release
versionator bump patch increment
versionator release push
```

### Pre-release

```bash
# Set up pre-release
versionator config prerelease set alpha

cat VERSION
# Output: v1.2.0-alpha

# Increment pre-release number
versionator config prerelease set alpha.2

# When ready to release
versionator config prerelease clear
```

## Next Steps

- Learn about the [VERSION File Format](./concepts/version-file)
- Explore [Template Variables](./templates/variables)
- Set up [CI/CD Integration](./integration/cicd)
- See [Command Reference](./commands/) for all commands
