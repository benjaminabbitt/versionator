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

Examples:
  versionator bump                   # Auto-bump and amend last commit
  versionator bump --dry-run         # Show what would happen
  versionator bump --no-amend        # Bump without amending the commit
  versionator bump --mode=semver     # Only use +semver: markers
  versionator bump --mode=conventional  # Only use conventional commits

## Usage

```bash
versionator bump [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Show what would happen without making changes |
| `--mode` | string | all | Parse mode: semver, conventional, or all |
| `--no-amend` | bool | false | Update VERSION file but do not amend the last commit |

## Manual Version Bumping

For manual version control, use the nested subcommands:

```bash
# Increment versions
versionator bump major increment   # or: bump major inc, bump major +, bump major up
versionator bump minor increment   # or: bump minor inc, bump minor +, bump minor up
versionator bump patch increment   # or: bump patch inc, bump patch +, bump patch up

# Decrement versions
versionator bump major decrement   # or: bump major dec, bump major -, bump major down
versionator bump minor decrement   # or: bump minor dec, bump minor -, bump minor down
versionator bump patch decrement   # or: bump patch dec, bump patch -, bump patch down
```

### Subcommands

| Command | Aliases | Description |
|---------|---------|-------------|
| `bump major increment` | `inc`, `+`, `up` | Increment major version (resets minor and patch to 0) |
| `bump major decrement` | `dec`, `-`, `down` | Decrement major version |
| `bump minor increment` | `inc`, `+`, `up` | Increment minor version (resets patch to 0) |
| `bump minor decrement` | `dec`, `-`, `down` | Decrement minor version |
| `bump patch increment` | `inc`, `+`, `up` | Increment patch version |
| `bump patch decrement` | `dec`, `-`, `down` | Decrement patch version |

## Automatic Bumping with Git Hook

You can install a post-commit hook to automatically bump the VERSION file when commits contain `+semver:` tags:

```bash
versionator init hook
```

This installs a hook that runs `versionator bump` after each commit containing `+semver:major`, `+semver:minor`, or `+semver:patch`. The commit is automatically amended to include the VERSION change.

See [init hook](./init#hook) for details.

