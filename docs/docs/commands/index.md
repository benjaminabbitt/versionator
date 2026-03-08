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
| [`ci`](./ci) | Output version variables for CI/CD systems |
| [`custom`](./custom) | Manage custom key-value pairs in config |
| [`emit`](./emit) | Emit version in various formats |
| [`init`](./init) | Initialize versionator in this directory |
| [`major`](./major) | Manage major version |
| [`metadata`](./metadata) | Manage build metadata |
| [`minor`](./minor) | Manage minor version |
| [`mode`](./mode) | Manage versioning mode (release or continuous-delivery) |
| [`patch`](./patch) | Manage patch version |
| [`prefix`](./prefix) | Manage version prefix |
| [`prerelease`](./prerelease) | Manage pre-release identifier |
| [`release`](./release) | Create git tag and release branch for current version |
| [`vars`](./vars) | Show all template variables and their current values |
| [`version`](./version) | Show current version |

## Global Flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `--log-format` | Log output format (console, json, development) |
| `-h, --help` | Help for any command |
