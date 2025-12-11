# Versionator

A semantic version management CLI tool that manages versions in a `VERSION` file.

## Installation

### Go Install (Recommended for Go developers)

```bash
go install github.com/benjaminabbitt/versionator@latest
```

### Homebrew (macOS/Linux)

```bash
brew tap benjaminabbitt/tap
brew install versionator
```

### Chocolatey (Windows)

```powershell
choco install versionator
```

### Debian/Ubuntu (.deb)

```bash
# Download the latest .deb package (amd64)
VERSION="1.0.0"  # Replace with desired version
wget https://github.com/benjaminabbitt/versionator/releases/download/v${VERSION}/versionator_${VERSION}_amd64.deb
sudo dpkg -i versionator_${VERSION}_amd64.deb

# Or for arm64
wget https://github.com/benjaminabbitt/versionator/releases/download/v${VERSION}/versionator_${VERSION}_arm64.deb
sudo dpkg -i versionator_${VERSION}_arm64.deb
```

### Manual Installation

Download the pre-compiled binary for your platform from [GitHub Releases](https://github.com/benjaminabbitt/versionator/releases).

#### Linux/macOS

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

#### Windows

1. Download `versionator-windows-amd64.exe` from [Releases](https://github.com/benjaminabbitt/versionator/releases)
2. Rename to `versionator.exe`
3. Move to a directory in your PATH (e.g., `C:\Users\<username>\bin`)
4. Or add the download location to your PATH environment variable

### Available Binaries

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | x64 | `versionator-linux-amd64` |
| Linux | arm64 | `versionator-linux-arm64` |
| macOS | Intel | `versionator-darwin-amd64` |
| macOS | Apple Silicon | `versionator-darwin-arm64` |
| Windows | x64 | `versionator-windows-amd64.exe` |
| Windows | arm64 | `versionator-windows-arm64.exe` |
| FreeBSD | x64 | `versionator-freebsd-amd64` |

All binaries are statically compiled - no dependencies required.

## Quick Start

```bash
# Create a VERSION file
echo "1.0.0" > VERSION

# Show current version
versionator version

# Increment versions
versionator major increment   # 1.0.0 -> 2.0.0
versionator minor increment   # 1.0.0 -> 1.1.0
versionator patch increment   # 1.0.0 -> 1.0.1

# Decrement versions
versionator patch decrement   # 1.0.1 -> 1.0.0

# Create git tag for current version
versionator commit
```

## Usage

```
versionator [command]

Available Commands:
  version     Show current version
  major       Increment or decrement major version
  minor       Increment or decrement minor version
  patch       Increment or decrement patch version
  prefix      Manage version prefix (e.g., "v")
  suffix      Manage version suffix (e.g., "-beta")
  commit      Create git tag for current version
  help        Help about any command

Flags:
  --log-format string   Log output format (console, json, development) (default "console")
  -h, --help            Help for versionator
```

## Configuration

Versionator can be configured via a `.versionator.yaml` file in your project root:

```yaml
# .versionator.yaml
prefix: "v"           # Version prefix (e.g., v1.0.0)
suffix: ""            # Version suffix (e.g., -beta)
logging:
  output: "console"   # console, json, or development
```

## Integration Examples

See the `examples/` directory for complete integration examples in multiple languages (Go, C++, C, Rust, Java).

### Using Make

```makefile
# Makefile example - inject version into Go binary
VERSION := $(shell versionator version)
build:
	go build -ldflags "-X main.VERSION=$(VERSION)" -o app
```

```makefile
# Makefile example - inject version into C++ binary
VERSION := $(shell versionator version)
build:
	g++ -DVERSION="\"$(VERSION)\"" -o app main.cpp
```

### Using Just

[Just](https://github.com/casey/just) is a modern command runner alternative to Make.

```just
# justfile example - inject version into Go binary
build:
    #!/bin/bash
    VERSION=$(versionator version)
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Get version
  id: version
  run: echo "version=$(versionator version)" >> $GITHUB_OUTPUT

- name: Build with version
  run: go build -ldflags "-X main.VERSION=${{ steps.version.outputs.version }}" -o app
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/benjaminabbitt/versionator.git
cd versionator

# Build
go build -o versionator .

# Or use just
just build
```

### Running Tests

```bash
go test ./...

# Or use just
just test
```

## License

BSD 3-Clause License - see [LICENSE](LICENSE) for details.
