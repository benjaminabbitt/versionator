---
title: major
description: Manage major version
---

# major

Manage major version

Commands to increment or decrement the major version component

## Usage

```bash
versionator major [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `decrement` | Decrement major version |
| `increment` | Increment major version |

### decrement

Decrement major version

Decrement the major version and reset minor and patch to 0

```bash
versionator major decrement
```

**Aliases:** `dec`, `-`

### increment

Increment major version

Increment the major version and reset minor and patch to 0

```bash
versionator major increment
```

**Aliases:** `inc`, `+`

