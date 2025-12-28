# Default recipe to display help
default:
    @just --list

# Clean Go module cache
clean-cache:
    #!/bin/bash
    echo "Cleaning Go module cache..."
    go clean -modcache || true

# Download dependencies
deps:
    #!/bin/bash
    set -e
    echo "Downloading Go modules..."
    go mod download
    echo "Tidying go.mod..."
    go mod tidy
    echo "Dependencies downloaded successfully!"

# Compile protocol buffer definitions
proto:
    #!/bin/bash
    set -e
    echo "Compiling protocol buffer definitions..."
    PROTO_DIR="pkg/plugin/proto"
    if [ ! -f "$PROTO_DIR/plugin.proto" ]; then
        echo "No proto files found, skipping..."
        exit 0
    fi
    if ! command -v protoc >/dev/null 2>&1; then
        echo "Error: protoc not found. Install protobuf compiler:"
        echo "  macOS: brew install protobuf"
        echo "  Linux: apt install protobuf-compiler"
        exit 1
    fi
    if ! command -v protoc-gen-go >/dev/null 2>&1 || ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
        echo "Installing Go protobuf plugins..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           "$PROTO_DIR/plugin.proto"
    echo "Proto compilation complete!"

# Build the application (static binary)
build: proto
    #!/bin/bash
    set -e
    mkdir -p bin/
    VERSION=$(cat VERSION 2>/dev/null || echo "dev")
    echo "Building versionator $VERSION (static binary)..."
    CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-s -w -X github.com/benjaminabbitt/versionator/internal/buildinfo.Version=$VERSION" -trimpath -o bin/versionator .
    echo "Build completed: bin/versionator"

# Build with verbose output for debugging (static binary)
build-verbose: proto
    #!/bin/bash
    set -e
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
    go test ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Format code
fmt:
    GO111MODULE=on go fmt ./...

# Run linter (if golangci-lint is available)
lint:
    #!/bin/bash
    if command -v golangci-lint >/dev/null 2>&1; then
        GO111MODULE=on golangci-lint run
    else
        echo "golangci-lint not found, running go vet instead..."
        GO111MODULE=on go vet ./...
    fi

# Initialize project (run once after cloning)
init:
    #!/bin/bash
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
    #!/bin/bash
    set -e
    echo "Setting up development environment..."
    just init
    echo "Building application..."
    just build
    echo "Development environment ready!"

# Build for all platforms with static linking
build-all: proto
    #!/bin/bash
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
    #!/bin/bash
    set -e
    echo "Rebuilding everything..."
    just clean
    just clean-cache
    just init
    just build
    echo "Rebuild complete!"

# Run acceptance tests locally (requires versionator in PATH or bin/)
acceptance-test: build
    #!/bin/bash
    set -e
    echo "Running acceptance tests locally..."
    # Tests look for $VERSIONATOR_PROJECT_ROOT/versionator
    export VERSIONATOR_PROJECT_ROOT=$(pwd)/bin
    GO111MODULE=on go test -v ./tests/acceptance/...

# Build acceptance test container image
acceptance-test-build:
    #!/bin/bash
    set -e
    echo "Building acceptance test container..."
    docker build -f tests/acceptance/Dockerfile -t versionator-acceptance:latest .

# Run acceptance tests in container (fast tests only)
acceptance-test-container: acceptance-test-build
    #!/bin/bash
    set -e
    echo "Running acceptance tests in container..."
    docker run --rm \
      -v "$(pwd)/tests:/app/tests:ro" \
      -e VERSIONATOR_PROJECT_ROOT=/usr/local/bin \
      versionator-acceptance:latest \
      sh -c "cd /app && go test -v ./tests/acceptance/..."

# Run ALL acceptance tests in container (including slow)
acceptance-test-container-all: acceptance-test-build
    #!/bin/bash
    set -e
    echo "Running ALL acceptance tests in container (including slow)..."
    docker run --rm \
      -v "$(pwd)/tests:/app/tests:ro" \
      -e VERSIONATOR_PROJECT_ROOT=/usr/local/bin \
      versionator-acceptance:latest \
      sh -c "cd /app && go test -v ./tests/acceptance/... -run '.*'"

# Run acceptance tests via docker-compose
acceptance-test-compose:
    #!/bin/bash
    set -e
    echo "Running acceptance tests via docker-compose..."
    docker compose -f tests/acceptance/docker-compose.yml up --build --abort-on-container-exit
    docker compose -f tests/acceptance/docker-compose.yml down

# Run slow acceptance tests via docker-compose
acceptance-test-compose-slow:
    #!/bin/bash
    set -e
    echo "Running slow acceptance tests via docker-compose..."
    docker compose -f tests/acceptance/docker-compose.yml run --build acceptance-tests-slow
    docker compose -f tests/acceptance/docker-compose.yml down

# === Container Language Tests ===

# Build the versionator-builder base image (required before other containers)
container-builder:
    #!/bin/bash
    set -e
    echo "Building versionator-builder base image..."
    docker build -t versionator-builder:latest -f tests/containers/images/versionator-builder.Dockerfile .

# Build a specific container test (e.g., just container-build go-emit)
container-build name: container-builder
    #!/bin/bash
    set -e
    echo "Building versionator-test-{{name}}..."
    docker build -t versionator-test-{{name}}:latest -f tests/containers/images/{{name}}.Dockerfile .

# Run a specific container test (e.g., just container-test go-emit)
container-test name: (container-build name)
    #!/bin/bash
    set -e
    echo "Running versionator-test-{{name}}..."
    docker run --rm versionator-test-{{name}}:latest

# Build all container tests
container-build-all: container-builder
    #!/bin/bash
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
    #!/bin/bash
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
    #!/bin/bash
    echo "Cleaning test container images..."
    docker images --format '{{"{{"}}.Repository{{"}}"}}' | grep '^versionator-test-' | xargs -r docker rmi -f 2>/dev/null || true
    docker rmi -f versionator-builder:latest 2>/dev/null || true
    echo "Done"

# === Plugin Build Targets ===

# Build all external plugins
plugins: proto
    #!/bin/bash
    set -e
    echo "Building all versionator plugins..."
    mkdir -p bin/plugins

    # Build emit plugins
    for d in plugins/emit/*/; do
        name=$(basename "$d")
        echo "Building emit plugin: $name"
        CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "bin/plugins/versionator-plugin-emit-$name" "./$d"
    done

    # Build build plugins
    for d in plugins/build/*/; do
        name=$(basename "$d")
        echo "Building build plugin: $name"
        CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "bin/plugins/versionator-plugin-build-$name" "./$d"
    done

    # Build patch plugins
    for d in plugins/patch/*/; do
        name=$(basename "$d")
        echo "Building patch plugin: $name"
        CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "bin/plugins/versionator-plugin-patch-$name" "./$d"
    done

    echo "All plugins built successfully!"
    ls -lh bin/plugins/

# Build all plugins with UPX compression (smaller binaries)
plugins-compressed: plugins
    #!/bin/bash
    set -e
    echo "Compressing plugins with UPX..."

    # Check for UPX
    UPX_BIN=""
    if command -v upx >/dev/null 2>&1; then
        UPX_BIN="upx"
    elif [ -x "/tmp/upx-4.2.2-amd64_linux/upx" ]; then
        UPX_BIN="/tmp/upx-4.2.2-amd64_linux/upx"
    else
        echo "UPX not found. Install with: apt install upx"
        echo "Or download from: https://github.com/upx/upx/releases"
        exit 1
    fi

    echo "Using UPX: $UPX_BIN"
    echo ""

    # Get size before compression
    SIZE_BEFORE=$(du -sh bin/plugins/ | cut -f1)

    # Compress all plugins
    for f in bin/plugins/versionator-plugin-*; do
        echo "Compressing $(basename $f)..."
        $UPX_BIN -q --best "$f" || true
    done

    # Get size after compression
    SIZE_AFTER=$(du -sh bin/plugins/ | cut -f1)

    echo ""
    echo "Compression complete!"
    echo "Before: $SIZE_BEFORE"
    echo "After:  $SIZE_AFTER"
    echo ""
    ls -lh bin/plugins/

# Build a single plugin (e.g., just plugin emit-go, just plugin patch-maven)
plugin name: proto
    #!/bin/bash
    set -e
    mkdir -p bin/plugins

    # Parse plugin type and name from input (e.g., "emit-go" -> type="emit", name="go")
    plugin_type=$(echo "{{name}}" | cut -d'-' -f1)
    plugin_name=$(echo "{{name}}" | cut -d'-' -f2-)

    if [ -d "plugins/$plugin_type/$plugin_name" ]; then
        echo "Building plugin: {{name}}"
        CGO_ENABLED=0 go build -trimpath -o "bin/plugins/versionator-plugin-{{name}}" "./plugins/$plugin_type/$plugin_name"
        echo "Built: bin/plugins/versionator-plugin-{{name}}"
    else
        echo "Error: Plugin 'plugins/$plugin_type/$plugin_name' not found"
        exit 1
    fi

# Install plugins to user plugin directory
plugins-install: plugins
    #!/bin/bash
    set -e
    PLUGIN_DIR="${HOME}/.versionator/plugins"
    echo "Installing plugins to $PLUGIN_DIR..."
    mkdir -p "$PLUGIN_DIR"
    cp bin/plugins/versionator-plugin-* "$PLUGIN_DIR/"
    chmod +x "$PLUGIN_DIR"/*
    echo "Installed $(ls bin/plugins/ | wc -l | tr -d ' ') plugins to $PLUGIN_DIR"
    ls "$PLUGIN_DIR/"

# List available plugins
plugins-list:
    @echo "=== Available Plugins ==="
    @echo ""
    @echo "Emit plugins (generate version source files):"
    @ls -1 plugins/emit/ 2>/dev/null | sed 's/^/  emit-/'
    @echo ""
    @echo "Build plugins (generate build/linker flags):"
    @ls -1 plugins/build/ 2>/dev/null | sed 's/^/  build-/'
    @echo ""
    @echo "Patch plugins (patch manifest files):"
    @ls -1 plugins/patch/ 2>/dev/null | sed 's/^/  patch-/'

# Clean built plugins
plugins-clean:
    rm -rf bin/plugins/
    @echo "Cleaned bin/plugins/"
