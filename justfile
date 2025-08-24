# Default recipe to display help
default:
    @just --list

# Clean Go module cache and fix permissions
clean-cache:
    #!/bin/zsh
    echo "Cleaning Go module cache..."
    GO111MODULE=on go clean -modcache || true

# Download dependencies with permission fix
deps:
    #!/bin/zsh
    set -e
    echo "Fixing permissions before downloading dependencies..."
    echo "Downloading Go modules..."
    GO111MODULE=on go mod download
    echo "Tidying go.mod..."
    GO111MODULE=on go mod tidy
    echo "Dependencies downloaded successfully!"

# Build the application
build:
    #!/bin/zsh
    set -e
    mkdir -p bin/
    echo "Building versionator..."
    GO111MODULE=on go build -o bin/versionator .
    echo "Build completed: bin/versionator"

# Build with verbose output for debugging
build-verbose:
    #!/bin/zsh
    set -e
    mkdir -p bin/
    echo "Building versionator (verbose)..."
    GO111MODULE=on go build -v -o bin/versionator .

# Run the application with arguments
run *args: build
    ./bin/versionator {{args}}

# Install the binary to /usr/local/bin
install: build
    sudo cp bin/versionator /usr/local/bin/

# Clean build artifacts
clean:
    rm -rf bin/
    GO111MODULE=on go clean

# Run tests
test:
    GO111MODULE=on go test ./...

# Run tests with coverage
test-coverage:
    GO111MODULE=on go test -cover ./...

# Format code
fmt:
    GO111MODULE=on go fmt ./...

# Run linter (if golangci-lint is available)
lint:
    #!/bin/zsh
    if command -v golangci-lint >/dev/null 2>&1; then
        GO111MODULE=on golangci-lint run
    else
        echo "golangci-lint not found, running go vet instead..."
        GO111MODULE=on go vet ./...
    fi

# Initialize project (run once after cloning)
init:
    #!/bin/zsh
    set -e
    echo "Initializing versionator project..."
    echo "Step 1: Cleaning cache and fixing permissions..."
    just clean-cache
    echo "Step 2: Downloading dependencies..."
    just deps
    echo "Step 3: Creating bin directory..."
    mkdir -p bin/
    echo "Initialization complete!"

# Development setup
dev-setup:
    #!/bin/zsh
    set -e
    echo "Setting up development environment..."
    just init
    echo "Building application..."
    just build
    echo "Development environment ready!"

# Build for all platforms with static linking
build-all:
    #!/bin/zsh
    set -e
    echo "Building for all platforms with static linking..."
    mkdir -p bin/

    # Linux amd64
    echo "Building for Linux amd64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-linux-amd64 .

    # Linux arm64
    echo "Building for Linux arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-linux-arm64 .

    # macOS amd64 (Intel)
    echo "Building for macOS amd64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-darwin-amd64 .

    # macOS arm64 (Apple Silicon)
    echo "Building for macOS arm64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-darwin-arm64 .

    # Windows amd64
    echo "Building for Windows amd64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-windows-amd64.exe .

    # Windows arm64
    echo "Building for Windows arm64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=arm64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-windows-arm64.exe .

    # FreeBSD amd64
    echo "Building for FreeBSD amd64..."
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 GO111MODULE=on go build -ldflags='-s -w' -trimpath -o bin/versionator-freebsd-amd64 .

    echo "All builds completed successfully!"
    echo "Build artifacts:"
    ls -la bin/

# Show project status
status:
    @echo "=== Versionator Project Status ==="
    @echo "Go version:"
    @go version
    @echo ""
    @echo "Current version:"
    @just version 2>/dev/null || echo "No VERSION file found"
    @echo ""
    @echo "Git Commands:"
    @echo "  commit          - Create git tag for current version"
    @echo "  commit-with-message MSG - Create git tag with custom message"
    @echo ""
    @echo "Module status:"
    @GO111MODULE=on go list -m all
    @echo ""
    @echo "Build status:"
    @ls -la bin/ 2>/dev/null || echo "No build artifacts found"


