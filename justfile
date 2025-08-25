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
    echo "Getting version from versionator..."
    VERSION=$(./bin/versionator version 2>/dev/null || cat VERSION 2>/dev/null || echo "dev")
    echo "Building versionator with version: $VERSION"
    GO111MODULE=on go build -ldflags "-X main.VERSION=$VERSION" -o bin/versionator .
    echo "Build completed: bin/versionator"

# Build with verbose output for debugging
build-verbose:
    #!/bin/zsh
    set -e
    mkdir -p bin/
    echo "Getting version from versionator..."
    VERSION=$(./bin/versionator version 2>/dev/null || cat VERSION 2>/dev/null || echo "dev")
    echo "Building versionator (verbose) with version: $VERSION"
    GO111MODULE=on go build -v -ldflags "-X main.VERSION=$VERSION" -o bin/versionator .

# Run the application with arguments
run *args: build
    ./bin/versionator {{args}}

# Install the binary to /usr/local/bin
install: build
    sudo cp bin/application /usr/local/bin/

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
    echo "Building for all platforms with static linking and UPX compression..."
    mkdir -p bin/
    
    # Get version information
    VERSION=$(./bin/versionator version 2>/dev/null || cat VERSION 2>/dev/null || echo "dev")
    echo "Building with version: $VERSION"

    # Linux amd64
    echo "Building for Linux amd64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-linux-amd64 .
    echo "Compressing Linux amd64 binary with UPX..."
    upx --best --lzma bin/versionator-linux-amd64

    # Linux arm64
    echo "Building for Linux arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-linux-arm64 .
    echo "Compressing Linux arm64 binary with UPX..."
    upx --best --lzma bin/versionator-linux-arm64

    # macOS amd64 (Intel)
    echo "Building for macOS amd64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-darwin-amd64 .
    # echo "Compressing macOS amd64 binary with UPX..."
    # MacOS is not supported
    # upx --best --lzma bin/versionator-darwin-amd64

    # macOS arm64 (Apple Silicon)
    echo "Building for macOS arm64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-darwin-arm64 .
    # echo "Compressing macOS arm64 binary with UPX..."
    # MacOS is not supported
    # upx --best --lzma bin/versionator-darwin-arm64

    # Windows amd64
    echo "Building for Windows amd64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-windows-amd64.exe .
    echo "Compressing Windows amd64 binary with UPX..."
    upx --best --lzma bin/versionator-windows-amd64.exe

    # Windows arm64
    echo "Building for Windows arm64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=arm64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-windows-arm64.exe .
    # echo "Compressing Windows arm64 binary with UPX..."
    # Windows ARM64 is not supported
    # upx --best --lzma bin/versionator-windows-arm64.exe

    # FreeBSD amd64
    echo "Building for FreeBSD amd64..."
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 GO111MODULE=on go build -ldflags="-s -w -X main.VERSION=$VERSION" -trimpath -o bin/versionator-freebsd-amd64 .
    # echo "Compressing FreeBSD amd64 binary with UPX..."
    # FreeBSD amd64 is not supported
    # upx --best --lzma bin/versionator-freebsd-amd64

    echo "All builds and UPX compression completed successfully!"
    echo "Build artifacts:"
    ls -la bin/

# Install and setup lefthook git hooks
lefthook-install:
    #!/bin/zsh
    set -e
    if ! command -v lefthook >/dev/null 2>&1; then
        echo "lefthook not found. Installing..."
        # Install lefthook via package manager or binary download
        if command -v apt-get >/dev/null 2>&1; then
            curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.deb.sh' | sudo bash
            sudo apt-get update && sudo apt-get install lefthook -y
        elif command -v brew >/dev/null 2>&1; then
            brew install lefthook
        else
            echo "Please install lefthook manually from https://github.com/evilmartians/lefthook"
            exit 1
        fi
    fi
    echo "Installing lefthook hooks..."
    lefthook install
    echo "Lefthook hooks installed successfully!"

# Update lefthook hooks
lefthook-update:
    lefthook install

# Run lefthook hooks manually
lefthook-run hook:
    lefthook run {{hook}}

# Uninstall lefthook hooks
lefthook-uninstall:
    lefthook uninstall

# Check lefthook status and configuration
lefthook-status:
    @echo "=== Lefthook Status ==="
    @lefthook version
    @echo ""
    @echo "Installed hooks:"
    @ls -la .git/hooks/ 2>/dev/null || echo "No git hooks found"
    @echo ""
    @echo "Configuration file:"
    @ls -la lefthook.yml 2>/dev/null || echo "No lefthook.yml found"

# Show project status
status:
    @echo "=== Versionator Project Status ==="
    @go version
    @echo ""
    @echo "Current version:"
    @just version 2>/dev/null || echo "No VERSION file found"
    @echo ""
    @echo "Git Commands:"
    @echo "  commit          - Create git tag for current version"
    @echo "  commit-with-message MSG - Create git tag with custom message"
    @echo ""
    @echo "Lefthook Commands:"
    @echo "  lefthook-install - Install git hooks with lefthook"
    @echo "  lefthook-status  - Show lefthook status"
    @echo "  lefthook-run     - Run specific hook manually"
    @echo ""
    @echo "Module status:"
    @GO111MODULE=on go list -m all
    @echo ""
    @echo "Build status:"
    @ls -la bin/ 2>/dev/null || echo "No build artifacts found"

get-versionator:
    mkdir -p bin
    curl https://github.com/benjaminabbitt/versionator/releases/latest/versionator-windows-amd64.exe -o bin/versionator.exe
