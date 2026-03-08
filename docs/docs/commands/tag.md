---
title: tag
description: Create git tag for current version
---

# tag

Create git tag for current version

Create a git tag for the current version after ensuring the working directory is clean.

This command will:
1. Check that you're in a git repository
2. Verify there are no uncommitted changes
3. Get the current version
4. Create a git tag with the version (prefixed with 'v')

The command will fail if there are uncommitted changes or if the tag already exists.

## Usage

```bash
versionator tag [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --force` | bool | false | Force creation even if tag exists |
| `-m, --message` | string | - | Tag message (default: 'Release \<version\>') |
| `-p, --prefix` | string | v | Tag prefix (default: 'v') |
| `-v, --verbose` | bool | false | Show additional information |

