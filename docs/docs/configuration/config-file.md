---
title: Configuration File
description: Configure versionator with .versionator.yaml
sidebar_position: 1
---

# Configuration File

Versionator can be configured using a `.versionator.yaml` file in your project directory.

## Creating Config File

```bash
# Create VERSION and .versionator.yaml with defaults
versionator init --config

# Create with custom initial version
versionator init --config --version 1.0.0 --prefix v
```

## Configuration Options

### Full Example

```yaml
# .versionator.yaml

# Version prefix (e.g., "v" for v1.0.0)
prefix: "v"

# Pre-release template configuration
prerelease:
  template: "build-{{CommitsSinceTag}}"  # Template for generating pre-release
  stable: false                           # false = generate from template at output

# Build metadata template configuration
metadata:
  template: "{{ShortHash}}"               # Template for generating metadata
  stable: false                           # false = generate from template at output
  git:
    hashLength: 12    # Length for {{MediumHash}}

# Logging configuration
logging:
  output: "console"   # console, json, or development

# Release branch configuration
release:
  createBranch: true
  branchPrefix: "release/"

# Custom template variables
custom:
  AppName: "MyApp"
  Environment: "production"
```

## Options Reference

### prefix

The version prefix string.

```yaml
prefix: "v"           # Results in v1.0.0 (recommended)
prefix: "V"           # Results in V1.0.0
prefix: ""            # No prefix: 1.0.0
```

Only `v` or `V` prefixes are allowed per SemVer convention.

### prerelease

Pre-release template configuration.

```yaml
prerelease:
  # Template string (Mustache syntax)
  template: "build-{{CommitsSinceTag}}"

  # Stability setting (default: false)
  # false = generate from template at output time (CD workflow)
  # true = store in VERSION file (release workflow)
  stable: false
```

**Stability**: When `stable: false` (default), the pre-release is generated from the template each time you run `emit`, `ci`, or `output` commands. This is ideal for continuous delivery workflows where you want dynamic values like commit counts.

When `stable: true`, you must explicitly set the pre-release value and it will be stored in the VERSION file.

**Separator Convention**: Use dashes (`-`) between pre-release components:

```yaml
prerelease:
  template: "alpha-{{CommitsSinceTag}}"      # alpha-5
  template: "beta-1-{{EscapedBranchName}}"   # beta-1-feature-foo
```

### metadata

Build metadata template configuration.

```yaml
metadata:
  # Template string (Mustache syntax)
  template: "{{ShortHash}}"

  # Stability setting (default: false)
  # false = generate from template at output time (CD workflow)
  # true = store in VERSION file (release workflow)
  stable: false

  # Git-specific settings
  git:
    hashLength: 12    # Length for {{MediumHash}}
```

**Stability**: When `stable: false` (default), the metadata is generated from the template each time you run `emit`, `ci`, or `output` commands. This is ideal for continuous delivery workflows where you want dynamic values like commit hashes.

When `stable: true`, you must explicitly set the metadata value and it will be stored in the VERSION file.

**Separator Convention**: Use dots (`.`) between metadata components:

```yaml
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"  # 20241211103045.abc1234
```

### logging

Logging output format.

```yaml
logging:
  output: "console"     # Human-readable (default)
  output: "json"        # JSON format for parsing
  output: "development" # Verbose development output
```

### release

Git release configuration.

```yaml
release:
  createBranch: true        # Create release branch when tagging
  branchPrefix: "release/"  # Branch name prefix
```

When enabled, `versionator release` creates both:
- A git tag (e.g., `v1.0.0`)
- A release branch (e.g., `release/v1.0.0`)

### custom

Custom template variables for use in templates.

```yaml
custom:
  AppName: "MyApp"
  Environment: "production"
  DeployTarget: "aws"
```

Use in templates:

```bash
versionator output version -t "{{AppName}} v{{MajorMinorPatch}}"
# Output: MyApp v1.0.0
```

Manage via CLI:

```bash
versionator config custom set AppName "MyApp"
versionator config custom get AppName
versionator config custom list
versionator config custom delete AppName
```

## Config File Discovery

Versionator looks for `.versionator.yaml` in the same directory as the VERSION file. Config files are not inherited from parent directories.

```
myproject/
├── .versionator.yaml    # Config for myproject/VERSION
├── VERSION
└── packages/
    ├── .versionator.yaml # Config for packages/VERSION
    └── VERSION
```

## Environment Variables

Some settings can be overridden via environment variables:

| Variable | Description |
|----------|-------------|
| `VERSIONATOR_LOG_FORMAT` | Logging output format |

## Relationship with VERSION File

:::important
The VERSION file stores stable version components. Non-stable components are generated at output time.
:::

The VERSION file stores:
- **Major, minor, patch** - always stored in VERSION
- **Prefix** - always stored in VERSION
- **Pre-release** - stored in VERSION only when `stable: true`
- **Metadata** - stored in VERSION only when `stable: true`

The config file stores:
- **Templates** for pre-release and metadata (used when `stable: false`)
- **Stability settings** for pre-release and metadata
- **Custom variables** for templating

### Stability Model

| Component | Default | Behavior |
|-----------|---------|----------|
| Major/Minor/Patch | Always stable | Always stored in VERSION |
| Prefix | Always stable | Always stored in VERSION |
| Pre-release | `stable: false` | Generated from template at output |
| Metadata | `stable: false` | Generated from template at output |

When `stable: false` (default for pre-release and metadata):
- Values are generated from templates each time you output the version
- `set` commands will error (use `--force` to override and set template instead)
- Ideal for continuous delivery workflows

When `stable: true`:
- Values are stored directly in the VERSION file
- `set` commands modify the VERSION file
- Ideal for traditional release workflows

## See Also

- [Template Variables](../templates/variables) - Available template variables
- [Pre-release Templates](../templates/prerelease) - Pre-release configuration
- [Metadata Templates](../templates/metadata) - Build metadata configuration
