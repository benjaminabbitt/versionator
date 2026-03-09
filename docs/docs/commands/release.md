---
title: release
description: Create git tag and release branch for current version
---

# release

Create git tag and release branch for current version

Create a git tag and release branch for the current version.

This command will:
1. Check that you're in a git repository
2. If only the VERSION file is dirty, commit it automatically
3. Verify there are no other uncommitted changes
4. Get the current version
5. Create a git tag with the version (prefixed with 'v')
6. Create a release branch (e.g., 'release/v1.2.3') if enabled

This is the recommended workflow after bumping a version:

```bash
versionator bump patch increment
versionator release push
```

Release branch creation is enabled by default. Configure in .versionator.yaml:

```yaml
release:
  createBranch: true    # set to false to disable
  branchPrefix: "release/"
```

Use --no-branch to skip branch creation for a single invocation.

The command will fail if there are uncommitted changes (other than VERSION)
or if the tag already exists.

## Usage

```bash
versionator release [flags]
versionator release push [flags]
```

## Subcommands

### push

Create a release and push both the tag and branch to the remote repository.

```bash
versionator release push
```

This combines the release command with git push operations:
1. Perform the standard release (create tag and branch)
2. Push the tag to origin
3. Push the release branch to origin (if created)

Example workflow:
```bash
# Commit with semver marker
git add -A
git commit -m "Fix bug +semver:patch"

# Auto-bump, release, and push in two commands
versionator bump
versionator release push
```

## Flags

These flags apply to both `release` and `release push`:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --force` | bool | false | Force creation even if tag exists |
| `-m, --message` | string | - | Tag message (default: 'Release \<version\>') |
| `--no-branch` | bool | false | Skip creating release branch |
| `-p, --prefix` | string | v | Tag prefix (default: 'v') |
| `-v, --verbose` | bool | false | Show additional information |

