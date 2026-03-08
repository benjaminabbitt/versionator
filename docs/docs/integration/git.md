---
title: Git Integration
description: Using versionator with Git for tagging and releases
sidebar_position: 1
---

# Git Integration

Versionator integrates with Git to create annotated tags and release branches for your versions.

## Creating Tags

The `tag` command creates an annotated Git tag for the current version:

```bash
# Create tag for current version
versionator tag
# Creates: v1.0.0 (if prefix is enabled)

# With custom message
versionator tag -m "Release 1.0.0 with new features"

# Force overwrite existing tag
versionator tag --force
```

### Tag Naming

The tag name follows your version format:

| VERSION File | Tag Created |
|--------------|-------------|
| `1.0.0` | `1.0.0` |
| `v1.0.0` | `v1.0.0` |
| `release-1.0.0` | `release-1.0.0` |
| `v1.0.0-beta.1` | `v1.0.0-beta.1` |

### Requirements

Before creating a tag:
- Working directory must be clean (no uncommitted changes)
- The current commit will be tagged

```bash
# Check status first
git status

# Stage and commit any changes
git add .
git commit -m "Prepare release v1.0.0"

# Now tag
versionator tag
```

## Release Branches

Configure automatic release branch creation in `.versionator.yaml`:

```yaml
release:
  createBranch: true
  branchPrefix: "release/"
```

With this configuration, `versionator tag` creates:
- A tag (e.g., `v1.0.0`)
- A branch (e.g., `release/v1.0.0`)

### Skip Branch Creation

```bash
versionator tag --no-branch
```

## Pushing to Remote

Tags are local by default. Push them to your remote:

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

A typical release workflow:

```bash
# 1. Ensure working directory is clean
git status

# 2. Bump version
versionator minor increment

# 3. Stage the VERSION file change
git add VERSION

# 4. Commit
git commit -m "Bump version to 1.1.0"

# 5. Create tag
versionator tag

# 6. Push commit and tag
git push
git push --tags
```

## Pre-release Workflow

For pre-release versions:

```bash
# Set pre-release
versionator prerelease set alpha.1
git add VERSION
git commit -m "Start alpha release cycle"

# Iterate on alpha
versionator prerelease set alpha.2
git add VERSION
git commit -m "Alpha 2"

# Move to beta
versionator prerelease set beta.1
git add VERSION
git commit -m "Start beta release cycle"

# Release
versionator prerelease clear
git add VERSION
git commit -m "Release 1.1.0"
versionator tag
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

1. **Commit VERSION changes**: Always commit the VERSION file change before tagging
2. **Clean working directory**: Ensure no uncommitted changes before tagging
3. **Semantic commits**: Use descriptive commit messages for version bumps
4. **Push tags explicitly**: Tags don't push automatically with `git push`
5. **Use annotated tags**: Versionator creates annotated tags by default

## See Also

- [CI/CD Integration](./cicd) - Automation workflows
- [Template Variables](../templates/variables) - Git-related variables
