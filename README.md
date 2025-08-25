# versionator

A semantic version management tool that manages versions in a `VERSION` file.

## Quick Start

Download the pre-compiled binary for your platform from [GitHub Releases](https://github.com/your-username/versionator/releases) - no installation required.

```bash
# Show current version
./application version

# Increment versions
./application major inc     # 1.0.0 → 2.0.0
./application minor inc     # 1.0.0 → 1.1.0  
./application patch inc     # 1.0.0 → 1.0.1

# Create git tag
./application commit
```

## Integration Examples

See the `examples/` directory for complete integration examples in multiple languages (Go, C++, C, Rust, Java).

### Using Make

```makefile
# Makefile example - inject version into Go binary
VERSION := $(shell ./versionator version)
build:
	go build -ldflags "-X main.VERSION=$(VERSION)" -o app
```

```makefile
# Makefile example - inject version into C++ binary  
VERSION := $(shell ./versionator version)
build:
	g++ -DVERSION="\"$(VERSION)\"" -o app main.cpp
```

### Using Just

[Just](https://github.com/casey/just) is a modern command runner alternative to Make.

```just
# justfile example - inject version into Go binary
build:
    #!/bin/zsh
    VERSION=$(./versionator version)
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

```just
# justfile example - inject version into C++ binary
build:
    #!/bin/zsh
    VERSION=$(./versionator version)
    g++ -DVERSION="\"$VERSION\"" -o app main.cpp
```

## Available Binaries

- `versionator-linux-amd64`
- `versionator-linux-arm64` 
- `versionator-darwin-amd64`
- `versionator-darwin-arm64`
- `versionator-windows-amd64.exe`
- `versionator-windows-arm64.exe`
- `versionator-freebsd-amd64`

All binaries are statically compiled - just download and run.

## Development

