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
| `V1.0.0` | `V1.0.0` |
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
versionator config prerelease set alpha.1
versionator release

# Iterate on alpha
versionator config prerelease set alpha.2
versionator release

# Move to beta
versionator config prerelease set beta.1
versionator release

# Final release
versionator config prerelease clear
versionator release
git push --tags
```

## Semantic Commit Automation

Versionator can automatically determine version bumps by analyzing commit messages since the last tag, similar to [semantic-release](https://semantic-release.gitbook.io/).

### The `bump` Command

The `bump` command parses commits and determines the appropriate version bump:

```bash
# Analyze commits and bump version automatically
versionator bump

# Preview what would happen without making changes
versionator bump --dry-run

# Only use +semver: markers
versionator bump --mode=semver

# Only use conventional commits
versionator bump --mode=conventional
```

### Supported Commit Formats

**+semver: markers** (can appear anywhere in commit message):

| Marker | Effect |
|--------|--------|
| `+semver:major` | Bump major (1.0.0 → 2.0.0) |
| `+semver:minor` | Bump minor (1.0.0 → 1.1.0) |
| `+semver:patch` | Bump patch (1.0.0 → 1.0.1) |
| `+semver:skip` | Skip version bump entirely |

**Conventional Commits** ([conventionalcommits.org](https://conventionalcommits.org)):

| Commit Type | Effect |
|-------------|--------|
| `feat: ...` | Bump minor |
| `fix: ...` | Bump patch |
| `feat!: ...` | Bump major (breaking) |
| `BREAKING CHANGE:` in footer | Bump major |

### Conflict Resolution

- Highest bump level wins (major > minor > patch)
- `+semver:skip` takes precedence and prevents any bump

### Automated Release Workflow

Combine `bump` and `release` for a fully automated workflow:

```bash
# After merging feature branches to main
versionator bump && versionator release && git push && git push --tags
```

### Lefthook Post-Merge Automation

Automate releases after merging to main with [Lefthook](https://github.com/evilmartians/lefthook):

```yaml
# lefthook.yml
post-merge:
  commands:
    auto-release:
      run: |
        # Only run on main branch
        if [ "$(git branch --show-current)" = "main" ]; then
          versionator bump --dry-run
          if [ $? -eq 0 ]; then
            versionator bump && versionator release && git push && git push --tags
          fi
        fi
```

### CI Workflow (GitHub Actions)

For teams that prefer CI-based releases:

```yaml
name: Auto Release

on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Need full history for commit analysis

      - name: Install versionator
        run: go install github.com/benjaminabbitt/versionator@latest

      - name: Check for version bump
        id: bump
        run: |
          if versionator bump --dry-run 2>&1 | grep -q "Would bump"; then
            echo "should_release=true" >> $GITHUB_OUTPUT
          else
            echo "should_release=false" >> $GITHUB_OUTPUT
          fi

      - name: Bump and Release
        if: steps.bump.outputs.should_release == 'true'
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          versionator bump
          versionator release
          git push
          git push --tags
```

### Example Workflow

```bash
# Developer workflow with conventional commits
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login timeout issue"
git commit -m "feat!: redesign API endpoints"  # Breaking change

# Merge to main, then release
git checkout main
git merge feature-branch

# Versionator analyzes commits and bumps appropriately
versionator bump --dry-run
# Output: Would bump from 1.2.3 to 2.0.0 (major)
# Triggering commit: feat!: redesign API endpoints

versionator bump
versionator release
git push && git push --tags
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
if ! versionator output version > /dev/null 2>&1; then
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
versionator output emit python --output src/_version.py
```

## Lefthook Integration

With [Lefthook](https://github.com/evilmartians/lefthook):

```yaml
# lefthook.yml
pre-commit:
  commands:
    version-check:
      run: versionator output version > /dev/null
      fail_text: "Invalid VERSION file"

post-merge:
  commands:
    version-emit:
      run: versionator output emit python --output src/_version.py
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
