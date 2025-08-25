# Versionator Development Context

This file contains comprehensive instructions for executing tests and other operations on the Versionator codebase.

## ⚠️ IMPORTANT NOTICE: NO BACKWARDS COMPATIBILITY

**This project does NOT preserve backwards compatibility.** Breaking changes may be introduced at any time without deprecation warnings or migration paths. Always test thoroughly when upgrading and be prepared to update your integrations accordingly.

## Quick Start

```bash
# Initialize the project (first time setup)
just init

# Set up development environment
just dev-setup
```

## Testing

### Running Tests

```bash
# Run all tests
just test
# OR
go test ./...

# Run tests with coverage
just test-coverage
# OR 
go test -cover ./...

# Run tests for specific package
go test ./internal/config/
go test ./internal/application/
go test ./internal/vcs/
go test ./cmd/
```

## Build Operations

### Development Builds

```bash
# Standard build
just build
# OR
go build -o bin/application .

# Verbose build (for debugging)
just build-verbose

# Run application with arguments
just run [args]
# OR
./bin/application [args]
```

### Cross-Platform Builds

```bash
# Build for all supported platforms
just build-all
```

Supported platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64) 
- Windows (amd64, arm64)
- FreeBSD (amd64)

### Installation

```bash
# Install to /usr/local/bin (Unix-like systems)
just install
```

## Code Quality

### Formatting

```bash
# Format all Go code
just fmt
# OR
go fmt ./...
```

### Linting

```bash
# Run linter (golangci-lint preferred, falls back to go vet)
just lint
```

### Static Analysis

```bash
# Run go vet
go vet ./...

# Check for race conditions
go test -race ./...
```

## Dependency Management

```bash
# Download and tidy dependencies
just deps
# OR
go mod download
go mod tidy

# Clean module cache
just clean-cache
# OR
go clean -modcache
```

## Cleanup

```bash
# Clean build artifacts
just clean
```

## Project Status

```bash
# Show comprehensive project status
just status
```


## Development Workflow

1. **Setup**: Run `just dev-setup` for initial setup
2. **Development**: Make changes and run `just test` frequently
3. **Code Quality**: Run `just fmt` and `just lint` before commits
4. **Testing**: Ensure all tests pass with `just test-coverage`
5. **Build**: Test cross-platform builds with `just build-all`

## Debugging

### Verbose Builds
Use `just build-verbose` to see detailed build information.

### Test Debugging
```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestSpecificFunction ./internal/package/

# Run tests with race detection
go test -race ./...
```

### Environment Variables
- `GO111MODULE=on` - Enforced throughout the project
- `CGO_ENABLED=0` - Used for static builds

## Architecture Notes

- **No Backwards Compatibility**: This project explicitly does not maintain backwards compatibility
- **Filesystem Abstraction**: Uses afero.Fs for testable file operations
- **VCS Abstraction**: Pluggable version control system support
- **Static Linking**: Cross-platform builds use static linking for portability
- **Semantic Versioning**: Core functionality built around semver principles

## Libraries and approaches
### File System
File system access will be via afero, ensuring testing is feasible.

### Testing
Testing will occur via testify.

#### Mocks
Mocks will be produced with Gomock.  Gomock mocks will be generated via commands in `just`

#### Cobra
Within cobra commands, use cmd.Print, not stdio operations to allow testing to work reliably.

## Troubleshooting

### Common Issues

1. **Permission Issues**: Run `just clean-cache` to fix module cache permissions
2. **Build Failures**: Ensure Go 1.23+ is installed and `GO111MODULE=on`
3. **Test Failures**: Check filesystem permissions and Git repository state
4. **Missing Dependencies**: Run `just deps` to refresh modules

### Useful Commands

```bash
# Check Go version
go version

# List all modules
go list -m all

# Check for updates
go list -u -m all

# Verify module integrity
go mod verify
```

---

**Remember: This project does NOT maintain backwards compatibility. Breaking changes can and will be introduced without notice or migration paths.**