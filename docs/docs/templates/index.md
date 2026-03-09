---
title: Templates Overview
description: Using Mustache templates in versionator
sidebar_position: 0
---

# Templates Overview

Versionator uses [Mustache](https://mustache.github.io/) templating for flexible version output formatting. Templates can be used with the `--template` flag and in configuration files.

## Basic Usage

Use `{{VariableName}}` syntax to insert template variables:

```bash
# Simple template
versionator output version -t "{{MajorMinorPatch}}"
# Output: 1.2.3

# With prefix
versionator output version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix
# Output: v1.2.3

# Custom format
versionator output version -t "Version: {{Major}}.{{Minor}}.{{Patch}}"
# Output: Version: 1.2.3
```

## Available Variables

See [Template Variables](./variables) for the complete list. Common variables include:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{Major}}` | Major version | `1` |
| `{{Minor}}` | Minor version | `2` |
| `{{Patch}}` | Patch version | `3` |
| `{{MajorMinorPatch}}` | Core version | `1.2.3` |
| `{{Prefix}}` | Version prefix | `v` |
| `{{ShortHash}}` | Git short hash | `abc1234` |
| `{{BranchName}}` | Current branch | `main` |

## Pre-release and Metadata

Pre-release and metadata are rendered from their own templates:

```bash
versionator output version \
  -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prerelease="alpha-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortHash}}"
# Output: 1.2.3-alpha-5+20241211103045.abc1234
```

### WithDash and WithPlus Variants

- `{{PreReleaseWithDash}}` includes the leading `-` (or empty if no pre-release)
- `{{MetadataWithPlus}}` includes the leading `+` (or empty if no metadata)

This makes it easy to build valid SemVer strings:

```bash
# Pre-release only
versionator output version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}" --prerelease="beta"
# Output: 1.2.3-beta

# No pre-release (variable is empty)
versionator output version -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}"
# Output: 1.2.3
```

## Configuration Templates

Set default templates in `.versionator.yaml`:

```yaml
prerelease:
  template: "alpha-{{CommitsSinceTag}}"

metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"
```

Then use without specifying the template:

```bash
versionator output version \
  -t "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prerelease \
  --metadata
```

## Custom Variables

Define custom variables in config:

```yaml
custom:
  AppName: "MyApp"
  Environment: "prod"
```

Use them in templates:

```bash
versionator output version -t "{{AppName}}-{{MajorMinorPatch}}"
# Output: MyApp-1.2.3
```

Or set inline:

```bash
versionator output version -t "{{AppName}}-{{MajorMinorPatch}}" --set AppName="MyApp"
```

## Code Generation

Templates are also used by the `emit` command:

```bash
# Use built-in Python template
versionator output emit python --output _version.py

# Dump template for customization
versionator output emit dump python > custom_python.tmpl

# Use custom template
versionator output emit --template-file custom_python.tmpl --output _version.py
```

## View Current Values

See all variables with their current values:

```bash
versionator config vars
```

## Template Syntax Reference

Versionator uses Mustache syntax:

| Syntax | Description |
|--------|-------------|
| `{{var}}` | Insert variable value |
| `{{#var}}...{{/var}}` | Section (if var is truthy) |
| `{{^var}}...{{/var}}` | Inverted section (if var is falsy) |
| `{{! comment }}` | Comment (not rendered) |

### Conditional Example

```bash
# Include "dirty" suffix only if there are uncommitted changes
versionator output version -t "{{MajorMinorPatch}}{{#Dirty}}-dirty{{/Dirty}}"
# Output with changes: 1.2.3-dirty
# Output without: 1.2.3
```

## See Also

- [Template Variables](./variables) - Complete variable reference
- [Pre-release Templates](./prerelease) - Pre-release configuration
- [Metadata Templates](./metadata) - Build metadata configuration
