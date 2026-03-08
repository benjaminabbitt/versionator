---
title: tag
description: Create git tag and release branch for current version
---

# tag

Create git tag and release branch for current version.

This command will:
1. Check that you're in a git repository
2. Verify there are no uncommitted changes
3. Get the current version
4. Create a git tag with the version
5. Create a release branch (e.g., `release/v1.2.3`) if enabled

## Usage

```bash
versionator tag [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --force` | bool | false | Force creation even if tag/branch exists |
| `-m, --message` | string | - | Tag message (default: 'Release \<version\>') |
| `-p, --prefix` | string | v | Tag prefix (default: 'v') |
| `--no-branch` | bool | false | Skip creating release branch |
| `-v, --verbose` | bool | false | Show additional information |

## Examples

### Basic Tag Creation

```bash
# Create tag for current version
versionator tag
# Output:
# Successfully created tag 'v1.2.3' for version 1.2.3 using Git
# Successfully created branch 'release/v1.2.3'
```

### Custom Tag Message

```bash
versionator tag -m "Release 1.2.3 with performance improvements"
```

### Tag Only (No Branch)

```bash
versionator tag --no-branch
# Only creates the tag, skips branch creation
```

### Force Overwrite

```bash
versionator tag --force
# Overwrites existing tag if it exists
```

### Verbose Output

```bash
versionator tag -v
# Output:
# Successfully created tag 'v1.2.3' for version 1.2.3 using Git
# Successfully created branch 'release/v1.2.3'
#   Message: Release 1.2.3
#   Git ID: abc1234
```

## Configuration

Release branch creation is controlled by `.versionator.yaml`:

```yaml
release:
  createBranch: true        # Enable/disable branch creation (default: true)
  branchPrefix: "release/"  # Prefix for branch names (default: "release/")
```

With the default configuration, `versionator tag` on version `v1.2.3` creates:
- Tag: `v1.2.3`
- Branch: `release/v1.2.3`

### Disable Branch Creation Globally

```yaml
release:
  createBranch: false
```

### Custom Branch Prefix

```yaml
release:
  branchPrefix: "releases/"   # Creates releases/v1.2.3
```

## What Gets Created

| VERSION File | Tag | Branch |
|--------------|-----|--------|
| `1.0.0` | `v1.0.0` | `release/v1.0.0` |
| `v1.0.0` | `v1.0.0` | `release/v1.0.0` |
| `1.0.0-beta.1` | `v1.0.0-beta.1` | `release/v1.0.0-beta.1` |
| `2.0.0-rc.1+build.123` | `v2.0.0-rc.1+build.123` | `release/v2.0.0-rc.1+build.123` |

## Requirements

- Working directory must be clean (no uncommitted changes)
- Tag must not already exist (unless `--force` is used)
- Must be in a git repository

## Pushing to Remote

Tags and branches are created locally. Push them to your remote:

```bash
# Push tag
git push origin v1.2.3

# Push release branch
git push origin release/v1.2.3

# Or push all tags
git push --tags
```

## See Also

- [Git Integration](../integration/git) - Complete git workflow
- [Configuration File](../configuration/config-file) - Release configuration options
