---
title: Installation
description: How to install versionator on your system
sidebar_position: 2
---

# Installation

Versionator is distributed as a single static binary with no dependencies. Choose the installation method that works best for your system.

## Package Managers

### Homebrew (macOS/Linux)

```bash
brew tap benjaminabbitt/tap
brew install versionator
```

### Chocolatey (Windows)

```powershell
choco install versionator
```

### Go Install

If you have Go installed:

```bash
go install github.com/benjaminabbitt/versionator@latest
```

## Manual Installation

### Debian/Ubuntu (.deb)

```bash
# Download the latest .deb package
VERSION="1.0.0"  # Replace with desired version

# For amd64
wget https://github.com/benjaminabbitt/versionator/releases/download/v${VERSION}/versionator_${VERSION}_amd64.deb
sudo dpkg -i versionator_${VERSION}_amd64.deb

# For arm64
wget https://github.com/benjaminabbitt/versionator/releases/download/v${VERSION}/versionator_${VERSION}_arm64.deb
sudo dpkg -i versionator_${VERSION}_arm64.deb
```

### Linux/macOS Binary

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

### Windows

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

This should output `0.0.0` (or create a VERSION file if one doesn't exist).

## Shell Completion

For tab-completion support, see the [Shell Completion](./configuration/shell-completion) guide.

## Next Steps

Now that versionator is installed, continue to the [Quick Start](./quick-start) tutorial.
