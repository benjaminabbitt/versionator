---
title: Template Variables
description: Complete reference of all template variables available in versionator
sidebar_position: 1
---

# Template Variables

Versionator uses [Mustache](https://mustache.github.io/) templating. Use `{{VariableName}}` syntax in templates.

:::tip
Run `versionator vars` to see all variables with their current values.
:::

## Version Components

Core version numbers and formatting.

| Variable | Description | Example |
|----------|-------------|--------|
| `{{Major}}` | Major version number | `1` |
| `{{Minor}}` | Minor version number | `2` |
| `{{Patch}}` | Patch version number | `3` |
| `{{MajorMinorPatch}}` | Core version: Major.Minor.Patch | `1.2.3` |
| `{{MajorMinor}}` | Major.Minor | `1.2` |
| `{{Prefix}}` | Version prefix | `v` |

## Pre-release

Pre-release identifier variables.

| Variable | Description | Example |
|----------|-------------|--------|
| `{{PreRelease}}` | Rendered pre-release identifier | `alpha-5` |
| `{{PreReleaseWithDash}}` | Pre-release with leading dash | `-alpha-5` |
| `{{PreReleaseLabel}}` | Label part of pre-release | `alpha` |
| `{{PreReleaseNumber}}` | Number part of pre-release | `5` |

## Build Metadata

Build metadata variables.

| Variable | Description | Example |
|----------|-------------|--------|
| `{{Metadata}}` | Rendered build metadata | `20241211.abc1234` |
| `{{MetadataWithPlus}}` | Metadata with leading plus | `+20241211.abc1234` |

## VCS / Git Information

Version control information.

| Variable | Description | Example |
|----------|-------------|--------|
| `{{Hash}}` | Full commit hash (40 chars) | `abc1234def5678...` |
| `{{ShortHash}}` | Short commit hash (7 chars) | `abc1234` |
| `{{MediumHash}}` | Medium commit hash (12 chars) | `abc1234def01` |
| `{{BranchName}}` | Current branch name | `feature/foo` |
| `{{EscapedBranchName}}` | Branch with slashes replaced | `feature-foo` |
| `{{CommitsSinceTag}}` | Commits since last tag | `42` |
| `{{BuildNumber}}` | Alias for CommitsSinceTag | `42` |
| `{{BuildNumberPadded}}` | Padded to 4 digits | `0042` |
| `{{UncommittedChanges}}` | Count of uncommitted files | `3` |
| `{{Dirty}}` | 'dirty' if uncommitted changes exist | `dirty` |
| `{{VersionSourceHash}}` | Hash of commit that last tag points to | `def5678` |

## Commit Information

Details about the current commit.

| Variable | Description | Example |
|----------|-------------|--------|
| `{{CommitAuthor}}` | Commit author name | `John Doe` |
| `{{CommitAuthorEmail}}` | Commit author email | `john@example.com` |
| `{{CommitDate}}` | ISO 8601 commit date | `2024-01-15T10:30:00Z` |
| `{{CommitDateCompact}}` | Compact commit date | `20240115103045` |
| `{{CommitDateShort}}` | Date only | `2024-01-15` |
| `{{CommitYear}}` | Commit year | `2024` |
| `{{CommitMonth}}` | Commit month (zero-padded) | `01` |
| `{{CommitDay}}` | Commit day (zero-padded) | `15` |

## Build Timestamps

Timestamps at build time.

| Variable | Description | Example |
|----------|-------------|--------|
| `{{BuildDateTimeUTC}}` | ISO 8601 build time | `2024-01-15T10:30:00Z` |
| `{{BuildDateTimeCompact}}` | Compact build time | `20240115103045` |
| `{{BuildDateUTC}}` | Build date only | `2024-01-15` |
| `{{BuildYear}}` | Build year | `2024` |
| `{{BuildMonth}}` | Build month (zero-padded) | `01` |
| `{{BuildDay}}` | Build day (zero-padded) | `15` |

