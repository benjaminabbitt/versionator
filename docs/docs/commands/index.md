---
title: Commands Reference
description: Complete reference of all versionator commands
sidebar_position: 0
---

# Commands Reference

Versionator provides commands organized into logical groups for managing semantic versions.

## Command Structure

```
versionator
├── init                    # Initialize versionator
│   └── hook               # Install post-commit git hook
├── bump                    # Auto-bump from commits, or manual bump
│   ├── major              # Major version operations
│   │   ├── increment      # Increment major (aliases: inc, +, up)
│   │   └── decrement      # Decrement major (aliases: dec, -, down)
│   ├── minor              # Minor version operations
│   │   ├── increment      # Increment minor (aliases: inc, +, up)
│   │   └── decrement      # Decrement minor (aliases: dec, -, down)
│   └── patch              # Patch version operations
│       ├── increment      # Increment patch (aliases: inc, +, up)
│       └── decrement      # Decrement patch (aliases: dec, -, down)
├── release                 # Create git tag and release branch
│   └── push               # Release and push to remote
├── config                  # Configuration management
│   ├── prefix             # Manage version prefix (v/V)
│   ├── prerelease         # Manage pre-release identifiers
│   ├── metadata           # Manage build metadata
│   ├── custom             # Manage custom key-value pairs
│   ├── mode               # Switch release/continuous-delivery modes
│   └── vars               # Show all template variables
├── output                  # Output version in various formats
│   ├── version            # Show current version
│   ├── emit               # Generate version files for languages
│   └── ci                 # Output for CI/CD systems
└── support                 # Shell completion and tooling
    ├── completion         # Generate shell completions
    └── schema             # Generate CLI schema for tooling
```

## Top-Level Commands

| Command | Description |
|---------|-------------|
| [`init`](./init) | Initialize versionator in this directory |
| [`bump`](./bump) | Auto-bump version based on commit messages |
| [`release`](./release) | Create git tag and release branch for current version |
| `config` | Manage versionator configuration |
| `output` | Output version in various formats |
| `support` | Shell completion and tooling support |

## Config Subcommands

| Command | Description |
|---------|-------------|
| `config prefix` | Manage version prefix (v, V) |
| `config prerelease` | Manage pre-release identifier |
| `config metadata` | Manage build metadata |
| `config custom` | Manage custom key-value pairs |
| `config mode` | Manage versioning mode |
| `config vars` | Show all template variables |

## Output Subcommands

| Command | Description |
|---------|-------------|
| `output version` | Show current version |
| `output emit` | Emit version in various formats |
| `output ci` | Output version variables for CI/CD systems |

## Bump Subcommands

| Command | Aliases | Description |
|---------|---------|-------------|
| `bump` | | Auto-bump based on commit messages |
| `bump major increment` | `inc`, `+`, `up` | Increment major version |
| `bump major decrement` | `dec`, `-`, `down` | Decrement major version |
| `bump minor increment` | `inc`, `+`, `up` | Increment minor version |
| `bump minor decrement` | `dec`, `-`, `down` | Decrement minor version |
| `bump patch increment` | `inc`, `+`, `up` | Increment patch version |
| `bump patch decrement` | `dec`, `-`, `down` | Decrement patch version |

## Global Flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `--log-format` | Log output format (quiet, console, json, development) |
| `-h, --help` | Help for any command |
