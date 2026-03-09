---
title: init
description: Initialize versionator in this directory
---

# init

Initialize versionator in this directory

Initialize versionator by creating a VERSION file.

Creates a VERSION file with the specified initial version and prefix.
Optionally creates a .versionator.yaml configuration file.

Examples:
  versionator init                        # Create VERSION with 0.0.1
  versionator init --version 1.0.0        # Create VERSION with 1.0.0
  versionator init --prefix v             # Create VERSION with v0.0.1
  versionator init --config               # Also create .versionator.yaml
  versionator init --force                # Overwrite existing files

## Usage

```bash
versionator init [flags]
versionator init [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `hook` | Install post-commit hook for automatic version bumping |

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | bool | false | Also create .versionator.yaml |
| `-f, --force` | bool | false | Overwrite existing files |
| `-p, --prefix` | string | - | Version prefix ('v' or 'V' only) |
| `-v, --version` | string | 0.0.1 | Initial version |

## hook

Install a git post-commit hook that automatically bumps the VERSION file.

The hook runs `versionator bump` after each commit that contains a `+semver:` tag in the commit message. Since bump amends by default, the VERSION change is automatically included in the commit.

### Usage

```bash
versionator init hook              # Install the post-commit hook
versionator init hook --uninstall  # Remove the post-commit hook
versionator init hook --force      # Overwrite existing hook
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --force` | bool | false | Overwrite existing hook |
| `--uninstall` | bool | false | Remove the post-commit hook |

### How It Works

When installed, the hook:
1. Checks if the commit message contains `+semver:major`, `+semver:minor`, or `+semver:patch`
2. If found, runs `versionator bump --amend` to bump VERSION and amend the commit

This automates the version bumping workflow - just include the appropriate `+semver:` tag in your commit message.

### Example Workflow

```bash
# Install the hook once
versionator init hook

# Make commits with +semver: tags
git commit -m "feat: add new feature +semver:minor"
# VERSION is automatically bumped and included in the commit

git commit -m "fix: bug fix +semver:patch"
# VERSION is automatically bumped and included in the commit
```

### Safety

- The hook only triggers on `+semver:major`, `+semver:minor`, or `+semver:patch`
- `+semver:skip` does NOT trigger the hook (handled by versionator bump)
- The hook will not overwrite a non-versionator hook (use `--force` to override)

