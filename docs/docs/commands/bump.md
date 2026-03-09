---
title: bump
description: Auto-bump version based on commit messages
---

# bump

Auto-bump version based on commit messages

Analyze commits since the last tag and bump the version accordingly.

Supported commit message formats:

  +semver: markers (can appear anywhere in the commit message):
    +semver:major - Bump major version (1.0.0 -\> 2.0.0)
    +semver:minor - Bump minor version (1.0.0 -\> 1.1.0)
    +semver:patch - Bump patch version (1.0.0 -\> 1.0.1)
    +semver:skip  - Skip version bump entirely

  Conventional Commits (https://conventionalcommits.org):
    feat: ...        - Bump minor version
    fix: ...         - Bump patch version
    feat!: ...       - Bump major version (breaking change)
    BREAKING CHANGE: - Bump major version (in commit footer)

Conflict resolution:
  - Highest bump level wins (major \> minor \> patch)
  - +semver:skip takes precedence and prevents any bump

**Examples:**

```bash
versionator bump                   # Auto-bump and amend last commit
versionator bump --dry-run         # Show what would happen
versionator bump --no-amend        # Bump without amending the commit
versionator bump --mode=semver     # Only use +semver: markers
versionator bump --mode=conventional  # Only use conventional commits
```

## Usage

```bash
versionator bump [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `major` | Manage major version |
| `minor` | Manage minor version |
| `patch` | Manage patch version |

### major

Manage major version

Commands to increment or decrement the major version component

```bash
versionator bump major
```

### minor

Manage minor version

Commands to increment or decrement the minor version component

```bash
versionator bump minor
```

### patch

Manage patch version

Commands to increment or decrement the patch version component

```bash
versionator bump patch
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Show what would happen without making changes |
| `--mode` | string | all | Parse mode: semver, conventional, or all |
| `--no-amend` | bool | false | Update VERSION file but do not amend the last commit |

