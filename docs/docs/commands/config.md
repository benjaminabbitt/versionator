---
title: config
description: Manage configuration
---

# config

Manage configuration

Commands for managing versionator configuration files.

## Usage

```bash
versionator config [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `dump` | Dump default configuration to stdout or file |

### dump

Dump default configuration to stdout or file

Dump a default .versionator.yaml configuration file.

By default, outputs to stdout. Use --output to write to a file.

Examples:
  versionator config dump                              # Print default config to stdout
  versionator config dump --output .versionator.yaml   # Write to file

```bash
versionator config dump [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-o, --output` | string | - | Output file path (default: stdout) |

