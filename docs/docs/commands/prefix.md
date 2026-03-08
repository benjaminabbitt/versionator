---
title: prefix
description: Manage version prefix
---

# prefix

Manage version prefix

Commands to enable, disable, or set version prefix in VERSION file

## Usage

```bash
versionator prefix [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `disable` | Disable version prefix |
| `enable` | Enable version prefix |
| `set` | Set version prefix |
| `status` | Show prefix status |

### disable

Disable version prefix

Disable version prefix by setting it to empty string

```bash
versionator prefix disable
```

### enable

Enable version prefix

Enable version prefix using config value if set, otherwise 'v'

```bash
versionator prefix enable
```

### set

Set version prefix

Set a custom version prefix in both config and VERSION file.

This updates:
1. The config file (.versionator.yaml) - so 'prefix enable' can restore it
2. The VERSION file - the source of truth for the current version

```bash
versionator prefix set
```

### status

Show prefix status

Show current prefix status from VERSION file (source of truth).

Also shows the configured prefix from .versionator.yaml that will be used on 'prefix enable'.

```bash
versionator prefix status
```

