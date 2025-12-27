# Default recipe to display help
default:
    @just --list

# Fix Go module cache permissions (try different approaches)
fix-perms:
    #!/bin/zsh
    set -e

    echo "Fixing Go module permissions..."

    # Fix /go directory permissions if it exists
    if [ -d "/go" ]; then
        echo "Fixing /go directory permissions..."
        if command -v sudo >/dev/null 2>&1; then
            sudo chown -R $(whoami):$(whoami) /go 2>/dev/null || true
            sudo chmod -R u+w /go 2>/dev/null || true
        else
            chown -R $(whoami):$(whoami) /go 2>/dev/null || true
            chmod -R u+w /go 2>/dev/null || true
        fi
    fi

    # Fix local GOPATH if it exists
    if [ -d "$HOME/go" ]; then
        echo "Fixing local GOPATH permissions..."
        chmod -R u+w $HOME/go 2>/dev/null || true
    fi

    # Create directories with proper permissions if they don't exist
    mkdir -p /go/pkg/mod 2>/dev/null || true
    mkdir -p /go/pkg/sumdb 2>/dev/null || true
    mkdir -p $HOME/go/pkg/mod 2>/dev/null || true

    # Set proper ownership for created directories
    if command -v sudo >/dev/null 2>&1; then
        sudo chown -R $(whoami):$(whoami) /go 2>/dev/null || true
    fi

# Clean Go module cache and fix permissions
clean-cache: fix-perms
    #!/bin/zsh
    echo "Cleaning Go module cache..."
    GO111MODULE=on go clean -modcache || true

# Download dependencies with permission fix
deps: fix-perms
    #!/bin/zsh
    set -e
    echo "Fixing permissions before downloading dependencies..."
    echo "Downloading Go modules..."
    GO111MODULE=on go mod download
    echo "Tidying go.mod..."
    GO111MODULE=on go mod tidy
    echo "Dependencies downloaded successfully!"

# Build the application (static binary)
build: fix-git-dubious-ownership-warning
    #!/bin/zsh
    set -e
    just fix-perms
    mkdir -p bin/
    VERSION=$(cat VERSION 2>/dev/null || echo "dev")
    echo "Building versionator $VERSION (static binary)..."
    CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-s -w -X github.com/benjaminabbitt/versionator/internal/buildinfo.Version=$VERSION" -trimpath -o bin/versionator .
    echo "Build completed: bin/versionator"

# Build with verbose output for debugging (static binary)
build-verbose: fix-git-dubious-ownership-warning
    #!/bin/zsh
    set -e
    just fix-perms
    mkdir -p bin/
    VERSION=$(cat VERSION 2>/dev/null || echo "dev")
    echo "Building versionator $VERSION (verbose, static binary)..."
    CGO_ENABLED=0 GO111MODULE=on go build -v -ldflags="-s -w -X github.com/benjaminabbitt/versionator/internal/buildinfo.Version=$VERSION" -trimpath -o bin/versionator .

# Run the application with arguments
run *args:
    @just build
    ./bin/versionator {{args}}

# Install the binary to ~/.local/bin
install:
    @just build
    mkdir -p ~/.local/bin
    cp bin/versionator ~/.local/bin/
    @echo "Installed to ~/.local/bin/versionator"
    @echo "Ensure ~/.local/bin is in your PATH"

# Uninstall the binary from ~/.local/bin
uninstall:
    rm -f ~/.local/bin/versionator
    @echo "Removed ~/.local/bin/versionator"

# Clean build artifacts
clean:
    rm -rf bin/
    GO111MODULE=on go clean

# Run tests
test:
    @just fix-perms
    GO111MODULE=on go test ./...

# Run tests with coverage
test-coverage:
    @just fix-perms
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
build-all: fix-perms fix-git-dubious-ownership-warning
    #!/bin/zsh
    set -e
    VERSION=$(cat VERSION 2>/dev/null || echo "dev")
    LDFLAGS="-s -w -X github.com/benjaminabbitt/versionator/internal/buildinfo.Version=$VERSION"
    echo "Building versionator $VERSION for all platforms with static linking..."
    mkdir -p bin/

    # Linux amd64
    echo "Building for Linux amd64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-linux-amd64 .

    # Linux arm64
    echo "Building for Linux arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-linux-arm64 .

    # macOS amd64 (Intel)
    echo "Building for macOS amd64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-darwin-amd64 .

    # macOS arm64 (Apple Silicon)
    echo "Building for macOS arm64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-darwin-arm64 .

    # Windows amd64
    echo "Building for Windows amd64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-windows-amd64.exe .

    # Windows arm64
    echo "Building for Windows arm64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=arm64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-windows-arm64.exe .

    # FreeBSD amd64
    echo "Building for FreeBSD amd64..."
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 GO111MODULE=on go build -ldflags="$LDFLAGS" -trimpath -o bin/versionator-freebsd-amd64 .

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

# Force rebuild everything
rebuild:
    #!/bin/zsh
    set -e
    echo "Rebuilding everything..."
    just clean
    just clean-cache
    just init
    just build
    echo "Rebuild complete!"

fix-git-dubious-ownership-warning:
    git config --global --add safe.directory /workspace

# Run Claude Code with permissions bypassed
claude:
    claude --dangerously-skip-permissions

# Run acceptance tests locally (requires versionator in PATH or bin/)
acceptance-test: build
    #!/bin/zsh
    set -e
    echo "Running acceptance tests locally..."
    # Tests look for $VERSIONATOR_PROJECT_ROOT/versionator
    export VERSIONATOR_PROJECT_ROOT=$(pwd)/bin
    GO111MODULE=on go test -v ./tests/acceptance/...

# Build acceptance test container image
acceptance-test-build:
    #!/bin/zsh
    set -e
    echo "Building acceptance test container..."
    docker build -f tests/acceptance/Dockerfile -t versionator-acceptance:latest .

# Run acceptance tests in container (fast tests only)
acceptance-test-container: acceptance-test-build
    #!/bin/zsh
    set -e
    echo "Running acceptance tests in container..."
    docker run --rm \
      -v "$(pwd)/tests:/app/tests:ro" \
      -e VERSIONATOR_PROJECT_ROOT=/usr/local/bin \
      versionator-acceptance:latest \
      sh -c "cd /app && go test -v ./tests/acceptance/..."

# Run ALL acceptance tests in container (including slow)
acceptance-test-container-all: acceptance-test-build
    #!/bin/zsh
    set -e
    echo "Running ALL acceptance tests in container (including slow)..."
    docker run --rm \
      -v "$(pwd)/tests:/app/tests:ro" \
      -e VERSIONATOR_PROJECT_ROOT=/usr/local/bin \
      versionator-acceptance:latest \
      sh -c "cd /app && go test -v ./tests/acceptance/... -run '.*'"

# Run acceptance tests via docker-compose
acceptance-test-compose:
    #!/bin/zsh
    set -e
    echo "Running acceptance tests via docker-compose..."
    docker compose -f tests/acceptance/docker-compose.yml up --build --abort-on-container-exit
    docker compose -f tests/acceptance/docker-compose.yml down

# Run slow acceptance tests via docker-compose
acceptance-test-compose-slow:
    #!/bin/zsh
    set -e
    echo "Running slow acceptance tests via docker-compose..."
    docker compose -f tests/acceptance/docker-compose.yml run --build acceptance-tests-slow
    docker compose -f tests/acceptance/docker-compose.yml down

# === Container Language Tests ===

# Build the versionator-builder base image (required before other containers)
container-builder:
    #!/bin/zsh
    set -e
    echo "Building versionator-builder base image..."
    docker build -t versionator-builder:latest -f tests/containers/images/versionator-builder.Dockerfile .

# Build a specific container test (e.g., just container-build go-emit)
container-build name: container-builder
    #!/bin/zsh
    set -e
    echo "Building versionator-test-{{name}}..."
    docker build -t versionator-test-{{name}}:latest -f tests/containers/images/{{name}}.Dockerfile .

# Run a specific container test (e.g., just container-test go-emit)
container-test name: (container-build name)
    #!/bin/zsh
    set -e
    echo "Running versionator-test-{{name}}..."
    docker run --rm versionator-test-{{name}}:latest

# Build all container tests
container-build-all: container-builder
    #!/bin/zsh
    set -e
    echo "Building all container tests..."
    for f in tests/containers/images/*.Dockerfile; do
        name=$(basename "$f" .Dockerfile)
        if [ "$name" != "versionator-builder" ]; then
            echo "Building $name..."
            docker build -t "versionator-test-$name:latest" -f "$f" .
        fi
    done

# Run all container tests
container-test-all: container-build-all
    #!/bin/zsh
    set -e
    echo "Running all container tests..."
    failed=0
    for f in tests/containers/images/*.Dockerfile; do
        name=$(basename "$f" .Dockerfile)
        if [ "$name" != "versionator-builder" ]; then
            echo "=== Testing $name ==="
            if docker run --rm "versionator-test-$name:latest"; then
                echo "✓ $name passed"
            else
                echo "✗ $name failed"
                failed=1
            fi
            echo ""
        fi
    done
    if [ $failed -eq 1 ]; then
        echo "Some container tests failed!"
        exit 1
    fi
    echo "All container tests passed!"

# List available container tests
container-list:
    @ls tests/containers/images/*.Dockerfile | xargs -n1 basename | sed 's/.Dockerfile//' | grep -v versionator-builder

# Clean all test container images
container-clean:
    #!/bin/zsh
    echo "Cleaning test container images..."
    docker images --format '{{"{{"}}.Repository{{"}}"}}' | grep '^versionator-test-' | xargs -r docker rmi -f 2>/dev/null || true
    docker rmi -f versionator-builder:latest 2>/dev/null || true
    echo "Done"
