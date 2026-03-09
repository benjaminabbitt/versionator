---
title: Installation
description: How to install versionator on your system
sidebar_position: 2
---

# Installation

Versionator is distributed as a single static binary with no dependencies.

## Linux/macOS

```bash
# Download (example for Linux amd64)
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64

# Make executable
chmod +x versionator-linux-amd64

# Move to PATH
sudo mv versionator-linux-amd64 /usr/local/bin/versionator

# Verify installation
versionator version
```

## Windows

1. Download `versionator-windows-amd64.exe` from [Releases](https://github.com/benjaminabbitt/versionator/releases)
2. Rename to `versionator.exe`
3. Move to a directory in your PATH (e.g., `C:\Users\<username>\bin`)
4. Or add the download location to your PATH environment variable

## Available Binaries

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | x64 | `versionator-linux-amd64` |
| Linux | arm64 | `versionator-linux-arm64` |
| macOS | Intel | `versionator-darwin-amd64` |
| macOS | Apple Silicon | `versionator-darwin-arm64` |
| Windows | x64 | `versionator-windows-amd64.exe` |
| Windows | arm64 | `versionator-windows-arm64.exe` |
| FreeBSD | x64 | `versionator-freebsd-amd64` |

All binaries are statically compiled with no dependencies.

## Verify Installation

After installation, verify versionator is working:

```bash
versionator version
```

This should output `0.0.1` (or create a VERSION file if one doesn't exist).

## Shell Completion

For tab-completion support, see the [Shell Completion](./configuration/shell-completion) guide.

## Next Steps

Now that versionator is installed, continue to the [Quick Start](./quick-start) tutorial.
