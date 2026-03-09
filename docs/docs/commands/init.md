---
title: init
description: Initialize versionator in this directory
---

# init

Initialize versionator in this directory

Initialize versionator by creating a VERSION file.

Creates a VERSION file with the specified initial version and prefix.
Optionally creates a .versionator.yaml configuration file.

Only 'v' or 'V' prefixes are allowed per SemVer convention.

**Examples:**

```bash
versionator init                        # Create VERSION with 0.0.1
versionator init --version 1.0.0        # Create VERSION with 1.0.0
versionator init --prefix v             # Create VERSION with v0.0.1
versionator init --config               # Also create .versionator.yaml
versionator init --force                # Overwrite existing files
```

## Usage

```bash
versionator init [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `hook` | Install post-commit hook for automatic version bumping |

### hook

Install post-commit hook for automatic version bumping

Install a git post-commit hook that runs 'versionator bump'.

This automatically bumps the VERSION file based on +semver: tags in commit
messages and amends the commit to include the VERSION change.

The hook only triggers when the commit message contains:
  +semver:major - Bump major version
  +semver:minor - Bump minor version
  +semver:patch - Bump patch version

**Examples:**

```bash
versionator init hook              # Install the post-commit hook
versionator init hook --uninstall  # Remove the post-commit hook
```

```bash
versionator init hook [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --force` | bool | false | Overwrite existing hook |
| `--uninstall` | bool | false | Remove the post-commit hook |

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | bool | false | Also create .versionator.yaml |
| `-f, --force` | bool | false | Overwrite existing files |
| `-p, --prefix` | string | - | Version prefix ('v' or 'V' only) |
| `-v, --version` | string | 0.0.1 | Initial version |

