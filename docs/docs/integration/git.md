---
title: Git Integration
description: Using versionator with Git for tagging and releases
sidebar_position: 1
---

# Git Integration

Versionator integrates with Git to create annotated tags and release branches for your versions.

## Creating Releases

The `release` command creates an annotated Git tag and release branch for the current version:

```bash
# Bump version and release
versionator patch increment
versionator release
# Creates: tag v1.0.0 and branch release/v1.0.0

# With custom message
versionator release -m "Release 1.0.0 with new features"

# Force overwrite existing tag
versionator release --force
```

### Auto-Commit Behavior

The `release` command will **automatically commit the VERSION file** if:
- The working directory is dirty
- The VERSION file is the **only** dirty file

This enables a clean workflow:

```bash
versionator patch increment    # VERSION file is now dirty
versionator release            # Auto-commits VERSION, creates tag and branch
```

If other files are dirty, the command will fail with a list of dirty files.

### Tag Naming

The tag name follows your version format:

| VERSION File | Tag Created |
|--------------|-------------|
| `1.0.0` | `1.0.0` |
| `v1.0.0` | `v1.0.0` |
| `release-1.0.0` | `release-1.0.0` |
| `v1.0.0-beta.1` | `v1.0.0-beta.1` |

## Release Branches

By default, `versionator release` creates **both a tag and a release branch**:

```bash
versionator release
# Output:
# Committed VERSION file: Release 1.0.0
# Successfully created tag 'v1.0.0' for version 1.0.0 using git
# Successfully created branch 'release/v1.0.0'
```

### Why Release Branches?

Release branches serve important purposes:

- **Hotfix isolation**: Apply patches to a specific release without pulling in main branch changes
- **Support multiple versions**: Maintain `release/v1.x` and `release/v2.x` simultaneously
- **CI/CD triggers**: Many pipelines trigger on `release/*` branch patterns
- **Clear release history**: Each release has a named branch for easy navigation

### Configuration

Configure release branches in `.versionator.yaml`:

```yaml
release:
  createBranch: true        # Enable/disable (default: true)
  branchPrefix: "release/"  # Prefix for branch names
```

### What Gets Created

| VERSION | Tag | Branch |
|---------|-----|--------|
| `1.0.0` | `v1.0.0` | `release/v1.0.0` |
| `1.0.0-beta.1` | `v1.0.0-beta.1` | `release/v1.0.0-beta.1` |
| `2.0.0-rc.1` | `v2.0.0-rc.1` | `release/v2.0.0-rc.1` |

### Skip Branch Creation

For a single invocation:

```bash
versionator release --no-branch
```

To disable globally:

```yaml
# .versionator.yaml
release:
  createBranch: false
```

## Pushing to Remote

Tags and branches are local by default. Push them to your remote:

```bash
# Push specific tag
git push origin v1.0.0

# Push all tags
git push --tags

# Push tag and release branch
git push origin v1.0.0
git push origin release/v1.0.0
```

## Version Bump Workflow

A typical release workflow with versionator:

```bash
# 1. Make sure other changes are committed
git status

# 2. Bump version and release
versionator minor increment
versionator release

# 3. Push everything
git push
git push --tags
git push origin release/v1.1.0
```

The `release` command handles the VERSION file commit automatically.

## Pre-release Workflow

For pre-release versions:

```bash
# Set pre-release
versionator prerelease set alpha.1
versionator release

# Iterate on alpha
versionator prerelease set alpha.2
versionator release

# Move to beta
versionator prerelease set beta.1
versionator release

# Final release
versionator prerelease clear
versionator release
git push --tags
```

## Template Variables from Git

Versionator extracts information from Git for templates:

| Variable | Description |
|----------|-------------|
| `{{Hash}}` | Full commit hash |
| `{{ShortHash}}` | 7-character hash |
| `{{MediumHash}}` | 12-character hash |
| `{{BranchName}}` | Current branch name |
| `{{EscapedBranchName}}` | Branch with `/` replaced by `-` |
| `{{CommitsSinceTag}}` | Commits since last tag |
| `{{CommitAuthor}}` | Commit author name |
| `{{CommitAuthorEmail}}` | Commit author email |
| `{{CommitDate}}` | Commit timestamp (ISO 8601) |
| `{{Dirty}}` | Non-empty if uncommitted changes |
| `{{UncommittedChanges}}` | Count of uncommitted files |

## Git Hooks Integration

Integrate versionator with Git hooks:

### Pre-commit Hook

Ensure VERSION file is valid:

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Validate VERSION file format
if ! versionator version > /dev/null 2>&1; then
    echo "ERROR: Invalid VERSION file"
    exit 1
fi
```

### Post-merge Hook

Regenerate version files after merge:

```bash
#!/bin/bash
# .git/hooks/post-merge

# Regenerate version file for interpreted languages
versionator emit python --output src/_version.py
```

## Lefthook Integration

With [Lefthook](https://github.com/evilmartians/lefthook):

```yaml
# lefthook.yml
pre-commit:
  commands:
    version-check:
      run: versionator version > /dev/null
      fail_text: "Invalid VERSION file"

post-merge:
  commands:
    version-emit:
      run: versionator emit python --output src/_version.py
```

## GitHub Actions Integration

See [CI/CD Integration](./cicd) for detailed GitHub Actions workflows.

## Best Practices

1. **Use release command**: Let `versionator release` handle VERSION commits
2. **Clean working directory**: Commit other changes before bumping version
3. **Semantic commits**: Use descriptive commit messages for version bumps
4. **Push tags explicitly**: Tags don't push automatically with `git push`
5. **Use annotated tags**: Versionator creates annotated tags by default

## See Also

- [CI/CD Integration](./cicd) - Automation workflows
- [Template Variables](../templates/variables) - Git-related variables
- [Release Command](../commands/release) - Command reference
