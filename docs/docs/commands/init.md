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
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | bool | false | Also create .versionator.yaml |
| `-f, --force` | bool | false | Overwrite existing files |
| `-p, --prefix` | string | - | Version prefix (e.g., 'v') |
| `-v, --version` | string | 0.0.1 | Initial version |

