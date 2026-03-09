---
title: Installation
description: How to install versionator on your system
sidebar_position: 2
---

# Installation

Versionator is distributed as a single static binary with no dependencies, packaged in compressed archives.

## Linux

```bash
# Download and extract (x64)
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64.tar.gz
tar xzf versionator-linux-amd64.tar.gz

# Move to PATH
sudo mv versionator-linux-amd64 /usr/local/bin/versionator

# Verify installation
versionator version
```

For ARM64 (e.g., Raspberry Pi, AWS Graviton), use `versionator-linux-arm64.tar.gz`.

## macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-darwin-arm64.tar.gz
tar xzf versionator-darwin-arm64.tar.gz
sudo mv versionator-darwin-arm64 /usr/local/bin/versionator

# Intel
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-darwin-amd64.tar.gz
tar xzf versionator-darwin-amd64.tar.gz
sudo mv versionator-darwin-amd64 /usr/local/bin/versionator
```

## Windows

```powershell
# Download and extract (x64)
Invoke-WebRequest -Uri https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-windows-amd64.zip -OutFile versionator.zip
Expand-Archive versionator.zip -DestinationPath .

# Move to a directory in your PATH
Move-Item versionator-windows-amd64.exe C:\Users\$env:USERNAME\bin\versionator.exe
```

For ARM64 Windows, use `versionator-windows-arm64.zip`.

## FreeBSD

```bash
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-freebsd-amd64.tar.gz
tar xzf versionator-freebsd-amd64.tar.gz
sudo mv versionator-freebsd-amd64 /usr/local/bin/versionator
```

## Available Archives

| Platform | Architecture | Archive |
|----------|--------------|---------|
| Linux | x64 | `versionator-linux-amd64.tar.gz` |
| Linux | arm64 | `versionator-linux-arm64.tar.gz` |
| macOS | Intel | `versionator-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `versionator-darwin-arm64.tar.gz` |
| Windows | x64 | `versionator-windows-amd64.zip` |
| Windows | arm64 | `versionator-windows-arm64.zip` |
| FreeBSD | x64 | `versionator-freebsd-amd64.tar.gz` |

All binaries are statically compiled with no dependencies.

## Security Verification

Each release includes:

1. **SHA256 checksums** (`checksums.txt`) - Verify archive integrity
2. **VirusTotal scans** - Independent malware analysis (links in release notes)

```bash
# Verify checksum
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

:::warning Disclaimer
VirusTotal scans are provided for transparency. A clean scan does not guarantee safety, and false positives can occur. For maximum security, review the source code and build from source.
:::

### Build from Source

```bash
git clone https://github.com/benjaminabbitt/versionator.git
cd versionator
go build -o versionator .
```

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
