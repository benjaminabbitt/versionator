---
title: custom
description: Manage custom key-value pairs in config
---

# custom

Manage custom key-value pairs in config

```
Manage custom key-value pairs that can be used in templates.

Custom variables are stored in .versionator.yaml and available as {{KeyName}} in templates.

Examples:
  versionator custom set AppName "My Application"
  versionator custom set BuildEnv production
  versionator custom get AppName
  versionator custom list
  versionator custom delete AppName

Then use in templates:
  versionator version -t "{{AppName}} v{{MajorMinorPatch}}"
```

## Usage

```bash
versionator custom [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `delete` | Delete a custom key-value pair |
| `get` | Get a custom value by key |
| `list` | List all custom key-value pairs |
| `set` | Set a custom key-value pair |

### delete

Delete a custom key-value pair

```bash
versionator custom delete
```

### get

Get a custom value by key

```bash
versionator custom get
```

### list

List all custom key-value pairs

```bash
versionator custom list
```

### set

Set a custom key-value pair

```
Set a custom key-value pair in .versionator.yaml.

The key becomes a template variable accessible as {{Key}}.

Examples:
  versionator custom set AppName "My Application"
  versionator custom set Environment production
  versionator custom set Copyright "2024 Acme Inc"
```

```bash
versionator custom set
```

