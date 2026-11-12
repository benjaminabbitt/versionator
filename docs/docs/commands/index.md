---
title: Commands Reference
description: Complete reference of all versionator commands
sidebar_position: 0
---

# Commands Reference

Versionator provides commands for managing semantic versions.

## Available Commands

| Command | Description |
|---------|-------------|
| [`bump`](./bump) | Auto-bump version based on commit messages |
| [`config`](./config) | Manage versionator configuration |
| [`init`](./init) | Initialize versionator in this directory |
| [`output`](./output) | Output version in various formats |
| [`release`](./release) | Create git tag and release branch for current version |
| [`set`](./set) | Set version to an arbitrary value |
| [`support`](./support) | Shell completion and tooling support |

## Global Flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `--ignore-untracked` | Untracked files do not make the tree dirty, for the dirty template variables and release's clean-tree check (overrides dirty.ignoreUntracked) |
| `--log-format` | Log output format (quiet, console, json, development) |
| `-h, --help` | Help for any command |
