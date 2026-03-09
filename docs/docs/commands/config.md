---
title: config
description: Manage versionator configuration
---

# config

Manage versionator configuration

Manage versionator configuration including version prefix, pre-release,
metadata, custom variables, and versioning mode.

Use subcommands to configure specific aspects:
  config prefix      - Manage version prefix (v, V)
  config prerelease  - Manage pre-release identifiers
  config metadata    - Manage build metadata
  config custom      - Manage custom key-value pairs
  config mode        - Switch between release and continuous-delivery modes
  config vars        - Show all available template variables

## Usage

```bash
versionator config [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `custom` | Manage custom key-value pairs in config |
| `metadata` | Manage build metadata |
| `mode` | Manage versioning mode (release or continuous-delivery) |
| `prefix` | Manage version prefix |
| `prerelease` | Manage pre-release identifier |
| `vars` | Show all template variables and their current values |

### custom

Manage custom key-value pairs in config

```
Manage custom key-value pairs that can be used in templates.

Custom variables are stored in .versionator.yaml and available as {{KeyName}} in templates.
```

**Examples:**

```bash
versionator custom set AppName "My Application"
versionator custom set BuildEnv production
versionator custom get AppName
versionator custom list
versionator custom delete AppName

Then use in templates:
versionator version -t "{{AppName}} v{{MajorMinorPatch}}"
```

```bash
versionator config custom
```

### metadata

Manage build metadata

Commands to enable or disable appending build metadata to version numbers.

Build metadata follows SemVer 2.0.0 specification:
- Appended with a plus sign (+) - this is added automatically
- Use DOTS (.) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3+20241211103045.abc1234def5

```bash
versionator config metadata
```

### mode

Manage versioning mode (release or continuous-delivery)

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

**Examples:**

```bash
versionator mode                           # Show current mode
versionator mode release                   # Set to release mode
versionator mode cd                        # Set to continuous-delivery mode
versionator mode cd --prerelease "build-{{CommitsSinceTag}}"
versionator mode cd --metadata "{{ShortHash}}"
```

```bash
versionator config mode [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metadata` | string | - | Metadata template for CD mode (Mustache) |
| `--prerelease` | string | - | Pre-release template for CD mode (Mustache) |

### prefix

Manage version prefix

Commands to enable, disable, or set version prefix in VERSION file.

Only 'v' or 'V' prefixes are allowed per SemVer convention.

```bash
versionator config prefix
```

### prerelease

Manage pre-release identifier

Commands to enable or disable pre-release identifiers.

Pre-release follows SemVer 2.0.0 specification:
- Appended with a dash (-) - this is added automatically
- Use DASHES (-) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3-alpha-5

```bash
versionator config prerelease
```

### vars

Show all template variables and their current values

Display all available template variables and their current values.

This is useful for understanding what variables are available when
creating custom templates for version, prerelease, or metadata output.

```bash
versionator config vars
```

