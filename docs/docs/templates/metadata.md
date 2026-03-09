---
title: Metadata Templates
description: Configuring build metadata with templates
sidebar_position: 3
---

# Metadata Templates

Build metadata provides additional information about a build without affecting version precedence. Per SemVer, versions differing only in metadata are considered equal for sorting.

## Static vs Dynamic Metadata

### Static (Stored in VERSION file)

Set a fixed metadata value:

```bash
versionator config metadata set build.123
# VERSION: 1.0.0+build.123

versionator config metadata set 20241212
# VERSION: 1.0.0+20241212
```

Clear metadata:

```bash
versionator config metadata clear
# VERSION: 1.0.0
```

### Dynamic (Template-based)

Use templates for values computed at build time:

```bash
versionator output version \
  -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortHash}}"
# Output: 1.0.0+20241211103045.abc1234
```

Dynamic values are computed at runtime and don't modify the VERSION file.

## Template Configuration

Set a default template in `.versionator.yaml`:

```yaml
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"
  git:
    hashLength: 12    # For {{MediumHash}}
```

Then use with the flag:

```bash
versionator output version -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" --metadata
# Output: 1.0.0+20241211103045.abc1234
```

## Separator Convention

Metadata components use **dots** (`.`) as separators:

```yaml
# Correct - use dots
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"       # 20241211103045.abc1234
  template: "{{CommitsSinceTag}}.{{ShortHash}}.{{BuildDateUTC}}"  # 42.abc1234.2024-01-15

# Incorrect - don't use dashes for metadata components
metadata:
  template: "{{BuildDateTimeCompact}}-{{ShortHash}}"       # Avoid
```

The leading plus before metadata is automatically added when using `{{MetadataWithPlus}}`.

## Common Patterns

### Timestamp + Hash

```yaml
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"
```

Result: `1.0.0+20241211103045.abc1234`

### Build Number + Hash

```yaml
metadata:
  template: "{{CommitsSinceTag}}.{{ShortHash}}"
```

Result: `1.0.0+42.abc1234`

### Date Only

```yaml
metadata:
  template: "{{BuildDateUTC}}"
```

Result: `1.0.0+2024-01-15`

### CI Build Info

```yaml
metadata:
  template: "ci.{{BuildNumberPadded}}.{{ShortHash}}"
```

Result: `1.0.0+ci.0042.abc1234`

### Long Hash

```yaml
metadata:
  template: "{{MediumHash}}"
  git:
    hashLength: 12
```

Result: `1.0.0+abc1234def01`

## Enable/Disable Commands

### metadata set

Sets a static value in the VERSION file:

```bash
versionator config metadata set build.123
# VERSION: 1.0.0+build.123
```

### metadata template

Sets a template in config and renders it to VERSION:

```bash
versionator config metadata template "{{BuildDateTimeCompact}}.{{ShortHash}}"
# Config: template = "{{BuildDateTimeCompact}}.{{ShortHash}}"
# VERSION: 1.0.0+20241211103045.abc1234
```

### metadata enable

Renders the config template to VERSION:

```bash
versionator config metadata enable
# Reads template from config
# VERSION: 1.0.0+20241211103045.abc1234
```

### metadata disable

Clears metadata from VERSION (preserves config template):

```bash
versionator config metadata disable
# VERSION: 1.0.0
# Config template still saved
```

### metadata clear

Clears metadata from VERSION:

```bash
versionator config metadata clear
# VERSION: 1.0.0
```

### metadata status

Shows current state:

```bash
versionator config metadata status
# Metadata: ENABLED
# Value: 20241211103045.abc1234
```

## Variables for Metadata

Common variables useful in metadata templates:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{BuildDateTimeCompact}}` | Build timestamp | `20241211103045` |
| `{{BuildDateUTC}}` | Build date | `2024-12-11` |
| `{{ShortHash}}` | 7-char commit hash | `abc1234` |
| `{{MediumHash}}` | Configurable length hash | `abc1234def01` |
| `{{Hash}}` | Full 40-char hash | `abc1234def...` |
| `{{CommitsSinceTag}}` | Commits since tag | `42` |
| `{{BranchName}}` | Current branch | `main` |

## Git Hash Length

Configure the `{{MediumHash}}` length:

```yaml
metadata:
  git:
    hashLength: 12    # Default is 12
```

| Setting | Output |
|---------|--------|
| `hashLength: 7` | `abc1234` (same as ShortHash) |
| `hashLength: 12` | `abc1234def01` |
| `hashLength: 20` | `abc1234def0123456789` |

## Version Precedence

:::note
Build metadata does **NOT** affect version precedence. Two versions that differ only in metadata are equal for sorting purposes.
:::

```
1.0.0+build.1 == 1.0.0+build.2    # Equal precedence
1.0.0+abc == 1.0.0+xyz            # Equal precedence
1.0.0 < 1.0.1+any.metadata        # Different core versions
```

## Use Cases

### Reproducible Builds

Include enough information to reproduce the build:

```yaml
metadata:
  template: "{{Hash}}"    # Full commit hash
```

### CI/CD Tracking

Include CI build information:

```bash
versionator output version --metadata="ci.$CI_BUILD_NUMBER.$CI_COMMIT_SHA"
```

### Nightly Builds

Timestamp-based builds:

```yaml
metadata:
  template: "nightly.{{BuildDateTimeCompact}}"
```

## See Also

- [Semantic Versioning](../concepts/semver) - SemVer spec details
- [Template Variables](./variables) - All available variables
- [CI/CD Integration](../integration/cicd) - Build automation
