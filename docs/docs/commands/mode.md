---
title: mode
description: Manage versioning mode (release or continuous-delivery)
---

# mode

Manage versioning mode (release or continuous-delivery)

```
Manage versioning mode configuration.

Versioning modes control how pre-release and metadata are generated:

  release (default):
    - Pre-release and metadata come from VERSION file
    - Used for standard release workflows
    - Developer controls version components

  continuous-delivery:
    - Pre-release and metadata are auto-generated from templates
    - Every build gets a unique version (e.g., 1.2.3-build-42+abc1234)
    - Templates use Mustache syntax with VCS variables

Examples:
  versionator mode                           # Show current mode
  versionator mode release                   # Set to release mode
  versionator mode cd                        # Set to continuous-delivery mode
  versionator mode cd --prerelease "build-{{CommitsSinceTag}}"
  versionator mode cd --metadata "{{ShortHash}}"
```

## Usage

```bash
versionator mode [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metadata` | string | - | Metadata template for CD mode (Mustache) |
| `--prerelease` | string | - | Pre-release template for CD mode (Mustache) |

