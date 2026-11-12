---
title: config
description: Manage versionator configuration
---

# config

Manage versionator configuration

Manage versionator configuration including version prefix, pre-release,
metadata, and custom variables.

Use subcommands to configure specific aspects:
  config prefix      - Manage version prefix (v, V)
  config prerelease  - Manage pre-release identifiers (includes stability setting)
  config metadata    - Manage build metadata (includes stability setting)
  config custom      - Manage custom key-value pairs
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

Commands to manage build metadata.

Build metadata follows SemVer 2.0.0 specification:
- Appended with a plus sign (+) - added automatically
- Multiple identifiers separated by DOTS (.)
- Each identifier: alphanumerics and hyphens only [0-9A-Za-z-]

Stability controls where the metadata value lives:
  stable: true  - Value is written to VERSION file
  stable: false - Value is generated from template at output time (default)

Example: 1.2.3+abc1234
         └─────┘
          metadata

```bash
versionator config metadata
```

### prefix

Manage version prefix

Commands to enable, disable, or set version prefix in VERSION file.

Only 'v' or 'V' prefixes are allowed per SemVer convention.

```bash
versionator config prefix
```

### prerelease

Manage pre-release identifier

Commands to manage pre-release identifiers.

Pre-release follows SemVer 2.0.0 specification:
- Appended with a dash (-) - this is added automatically
- Use DASHES (-) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Stability controls where the pre-release value lives:
  stable: true  - Value is written to VERSION file (traditional release workflow)
  stable: false - Value is generated from template at output time (default, CD workflow)

Example output: 1.2.3-build-5

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

