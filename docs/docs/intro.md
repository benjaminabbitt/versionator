---
title: Introduction
description: What is versionator and why use it
sidebar_position: 1
slug: /
---

# Introduction

Versionator is a CLI tool for managing semantic versions following [SemVer 2.0.0](https://semver.org/). It stores the current version in a plain text `VERSION` file, making version management explicit and deterministic.

## Why Versionator?

### The Problem with Auto-Versioning

Many versioning tools automatically calculate versions from git history. While convenient, this approach has drawbacks:

- **Non-deterministic**: The same commit can produce different versions depending on branch state
- **Complex configuration**: Branching strategies require extensive configuration
- **Debugging difficulty**: Hard to understand why a particular version was generated

See [Competitors](./competitors) for detailed comparisons with GitVersion, semantic-release, and others.

### The Versionator Approach

Versionator takes a different approach: **explicit version management**.

- The `VERSION` file is the **single source of truth**
- Version changes are deliberate actions (`versionator patch increment`)
- Versions are predictable and reproducible
- Works seamlessly in monorepos with independent package versions

## Key Features

- **Version in your binary**: Embed version directly into compiled binaries—know exactly what's running in production
- **Simple VERSION file**: Human-readable plain text file as single source of truth
- **Full SemVer 2.0.0 support**: Major.Minor.Patch with pre-release and metadata
- **10+ language support**: Go, Rust, C, C++, Java, Python, JavaScript, TypeScript, Ruby, and more
- **Container-ready**: Embed version in Docker images via OCI labels and compiled binaries
- **Git integration**: Create annotated tags and release branches with `versionator release`
- **Mustache templating**: Flexible output formatting with template variables
- **Monorepo support**: Independent versions for nested packages
- **Single binary**: No runtime dependencies, works everywhere

## The Real Benefit: Version in Your Binary

The VERSION file is just the start. The real power is getting that version **into your compiled binary or container image**:

```bash
# Build with version embedded
$ VERSION=$(versionator version)
$ go build -ldflags "-X main.Version=$VERSION" -o myapp

# Now your binary knows its version
$ ./myapp --version
myapp v1.1.1 (commit: abc1234, built: 2024-01-15T10:30:00Z)
```

When you're debugging at 2 AM, you'll know exactly what's running. See [Binary Embedding](./integration/binary-embedding) for examples in Go, Rust, C, C++, Java, Python, JavaScript, Docker, and more.

## Quick Example

```bash
# Initialize version (creates VERSION file with 0.0.0)
versionator version

# Increment versions
versionator major increment   # 0.0.0 -> 1.0.0
versionator minor increment   # 1.0.0 -> 1.1.0
versionator patch increment   # 1.1.0 -> 1.1.1

# Create git tag and release branch
versionator release           # Creates tag v1.1.1 and branch release/v1.1.1

# Generate version file for Python
versionator emit python --output mypackage/_version.py
```

## When to Use Versionator

Versionator is ideal for:

- **Monorepos** with multiple packages needing independent versions
- **CI/CD pipelines** where version bumps are explicit steps
- **Projects** that want predictable, reproducible versioning
- **Teams** that prefer deliberate version management over auto-calculation
- **Multi-language projects** needing version info in multiple formats

## Getting Started

Ready to get started? Head to the [Installation](./installation) guide, then follow the [Quick Start](./quick-start) tutorial.
