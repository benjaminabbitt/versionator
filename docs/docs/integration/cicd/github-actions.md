---
title: GitHub Actions
description: Using versionator with GitHub Actions
sidebar_position: 1
---

# GitHub Actions

**Platform:** [GitHub Actions](https://github.com/features/actions)

## Get Version

Capture the current version as an output:

```yaml
- name: Get version
  id: version
  run: echo "version=$(versionator version)" >> $GITHUB_OUTPUT

- name: Use version
  run: echo "Building version ${{ steps.version.outputs.version }}"
```

## Build with Version

Inject version at build time:

```yaml
- name: Build with version
  run: |
    VERSION=$(versionator version)
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

## Full Version Info

Get extended version information:

```yaml
- name: Get full version info
  id: version
  run: |
    echo "version=$(versionator version)" >> $GITHUB_OUTPUT
    echo "full=$(versionator version -t '{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}' --prefix --metadata='{{ShortHash}}')" >> $GITHUB_OUTPUT
```

## Release Workflow

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Get version from tag
        id: version
        run: echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Build
        run: |
          VERSION=${{ steps.version.outputs.version }}
          go build -ldflags "-X main.VERSION=$VERSION" -o app

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: app
          body: "Release ${{ steps.version.outputs.version }}"
```

## Install Versionator

```yaml
- name: Install versionator
  run: go install github.com/benjaminabbitt/versionator@latest
```

Or cache it for faster builds:

```yaml
- name: Cache versionator
  uses: actions/cache@v4
  with:
    path: ~/go/bin/versionator
    key: versionator-${{ runner.os }}

- name: Install versionator
  run: |
    if [ ! -f ~/go/bin/versionator ]; then
      go install github.com/benjaminabbitt/versionator@latest
    fi
```
