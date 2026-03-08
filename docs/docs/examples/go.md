---
title: Go
description: Complete Go integration example
sidebar_position: 2
---

# Go Integration

This guide shows how to integrate versionator with a Go project using build-time version injection.

## Project Structure

```
myproject/
├── VERSION
├── .versionator.yaml
├── Makefile
├── go.mod
├── main.go
└── internal/
    └── version/
        └── version.go
```

## Setup

### Initialize Version

```bash
cd myproject
versionator version
# Creates VERSION with 0.0.0

# Enable prefix
versionator prefix set v
# VERSION: v0.0.0
```

## Version Variable

Define a version variable to inject at build time:

### Simple (main package)

```go
// main.go
package main

import "fmt"

var VERSION = "dev"

func main() {
    fmt.Printf("MyApp version %s\n", VERSION)
}
```

### Package-based

```go
// internal/version/version.go
package version

var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)
```

```go
// main.go
package main

import (
    "fmt"
    "myproject/internal/version"
)

func main() {
    fmt.Printf("MyApp %s (commit: %s, built: %s)\n",
        version.Version,
        version.GitCommit,
        version.BuildDate)
}
```

## Build Commands

### Basic Build

```bash
VERSION=$(versionator version)
go build -ldflags "-X main.VERSION=$VERSION" -o myapp
```

### Full Build Info

```bash
VERSION=$(versionator version)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build -ldflags "\
    -X myproject/internal/version.Version=$VERSION \
    -X myproject/internal/version.GitCommit=$COMMIT \
    -X myproject/internal/version.BuildDate=$DATE" \
    -o myapp
```

### Using Versionator Variables

```bash
VERSION=$(versionator version -t "{{Prefix}}{{MajorMinorPatch}}" --prefix)
COMMIT=$(versionator version -t "{{ShortHash}}")
DATE=$(versionator version -t "{{BuildDateTimeUTC}}")

go build -ldflags "\
    -X myproject/internal/version.Version=$VERSION \
    -X myproject/internal/version.GitCommit=$COMMIT \
    -X myproject/internal/version.BuildDate=$DATE" \
    -o myapp
```

## Build Automation

### Makefile

```makefile
BINARY := myapp
VERSION := $(shell versionator version)
COMMIT := $(shell versionator version -t "{{ShortHash}}")
DATE := $(shell versionator version -t "{{BuildDateTimeUTC}}")
LDFLAGS := -X myproject/internal/version.Version=$(VERSION) \
           -X myproject/internal/version.GitCommit=$(COMMIT) \
           -X myproject/internal/version.BuildDate=$(DATE)

.PHONY: build test clean release

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)

# Cross-compile
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-linux-amd64

build-darwin:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-arm64

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-windows-amd64.exe

build-all: build-linux build-darwin build-windows

# Version management
bump-patch:
	versionator patch increment

bump-minor:
	versionator minor increment

bump-major:
	versionator major increment

release: bump-patch
	git add VERSION
	git commit -m "Release $(VERSION)"
	versionator tag
	git push --tags
```

### Just

```just
binary := "myapp"
version := `versionator version`
commit := `versionator version -t "{{ShortHash}}"`
date := `versionator version -t "{{BuildDateTimeUTC}}"`
ldflags := "-X myproject/internal/version.Version=" + version + " -X myproject/internal/version.GitCommit=" + commit + " -X myproject/internal/version.BuildDate=" + date

# Build binary
build:
    go build -ldflags "{{ldflags}}" -o {{binary}}

# Run tests
test:
    go test ./...

# Build for all platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -ldflags "{{ldflags}}" -o {{binary}}-linux-amd64
    GOOS=darwin GOARCH=arm64 go build -ldflags "{{ldflags}}" -o {{binary}}-darwin-arm64
    GOOS=windows GOARCH=amd64 go build -ldflags "{{ldflags}}" -o {{binary}}-windows-amd64.exe

# Release
release bump="patch":
    versionator {{bump}} increment
    git add VERSION
    git commit -m "Release $(versionator version)"
    versionator tag
    git push --tags
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install versionator
        run: go install github.com/benjaminabbitt/versionator@latest

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          VERSION=$(versionator version)
          COMMIT=$(versionator version -t "{{ShortHash}}")
          DATE=$(versionator version -t "{{BuildDateTimeUTC}}")
          EXT=""
          if [ "$GOOS" = "windows" ]; then EXT=".exe"; fi

          go build -ldflags "\
            -X myproject/internal/version.Version=$VERSION \
            -X myproject/internal/version.GitCommit=$COMMIT \
            -X myproject/internal/version.BuildDate=$DATE" \
            -o myapp-$GOOS-$GOARCH$EXT

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: myapp-${{ matrix.goos }}-${{ matrix.goarch }}
          path: myapp-*

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download artifacts
        uses: actions/download-artifact@v3

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: myapp-*/myapp-*
```

## Version Command

Add a version subcommand using [Cobra](https://github.com/spf13/cobra):

```go
// cmd/version.go
package cmd

import (
    "fmt"
    "myproject/internal/version"
    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Version:    %s\n", version.Version)
        fmt.Printf("Git Commit: %s\n", version.GitCommit)
        fmt.Printf("Built:      %s\n", version.BuildDate)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

## See Also

- [Language Integration](../integration/languages) - Other languages
- [CI/CD Integration](../integration/cicd) - Full CI/CD examples
