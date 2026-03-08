---
title: minor
description: Manage minor version
---

# minor

Manage minor version

Commands to increment or decrement the minor version component

## Usage

```bash
versionator minor [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `decrement` | Decrement minor version |
| `increment` | Increment minor version |

### decrement

Decrement minor version

Decrement the minor version and reset patch to 0

```bash
versionator minor decrement
```

**Aliases:** `dec`, `-`

### increment

Increment minor version

Increment the minor version and reset patch to 0

```bash
versionator minor increment
```

**Aliases:** `inc`, `+`

