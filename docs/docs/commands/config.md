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
  config prerelease  - Manage pre-release identifiers and stability
  config metadata    - Manage build metadata and stability
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
| `metadata` | Manage build metadata and stability |
| `prefix` | Manage version prefix |
| `prerelease` | Manage pre-release identifier and stability |
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

Manage build metadata and stability

Commands to manage build metadata configuration including stability settings.

Build metadata follows SemVer 2.0.0 specification:
- Appended with a plus sign (+) - added automatically
- Multiple identifiers separated by DOTS (.)
- Each identifier: alphanumerics and hyphens only [0-9A-Za-z-]

Example: 1.2.3+20241211103045.abc1234
         └─────────────────┘ └──────┘
          identifier 1       identifier 2

**Stability**: Controls whether metadata is stored in VERSION file or generated from template.

| Stability | Default | Behavior |
|-----------|---------|----------|
| `false` | Yes | Generated from template at output time |
| `true` | No | Stored directly in VERSION file |

**Examples:**

```bash
versionator config metadata                    # Show current status
versionator config metadata template           # Get current template
versionator config metadata template "{{ShortHash}}"  # Set template
versionator config metadata stable             # Get stability setting
versionator config metadata stable true        # Enable stable mode
versionator config metadata stable false       # Disable stable mode (use template)
versionator config metadata set "build123"     # Set literal (requires stable:true)
versionator config metadata set "build123" --force  # Force set template
```

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

Manage pre-release identifier and stability

Commands to manage pre-release configuration including stability settings.

Pre-release follows SemVer 2.0.0 specification:
- Appended with a dash (-) - this is added automatically
- Use DASHES (-) to separate identifiers in your template
- Must contain only alphanumerics and hyphens [0-9A-Za-z-]

Example output: 1.2.3-alpha-5

**Stability**: Controls whether pre-release is stored in VERSION file or generated from template.

| Stability | Default | Behavior |
|-----------|---------|----------|
| `false` | Yes | Generated from template at output time |
| `true` | No | Stored directly in VERSION file |

**Examples:**

```bash
versionator config prerelease                    # Show current status
versionator config prerelease template           # Get current template
versionator config prerelease template "build-{{CommitsSinceTag}}"  # Set template
versionator config prerelease stable             # Get stability setting
versionator config prerelease stable true        # Enable stable mode
versionator config prerelease stable false       # Disable stable mode (use template)
versionator config prerelease set "alpha"        # Set literal (requires stable:true)
versionator config prerelease set "alpha" --force  # Force set template
```

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

