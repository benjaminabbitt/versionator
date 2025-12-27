# Language Container Testing Plan

## Overview

Set up Docker containers for each supported language to verify that versionator-generated code compiles and runs correctly. Versionator is **copied into each container** and used as part of the precompilation/compilation process.

## Key Design Decisions

1. **Docker Official Images only** - Use verified/trusted base images from Docker Hub
2. **Plain Docker (no Compose)** - Simple `docker build` + `docker run` commands
3. **Versionator inside containers** - Binary built once, copied into each language container
4. **Integrated build process** - versionator emit/link/patch runs inside container
5. **Cucumber orchestration from host** - godog tests run containers and verify output
6. **Validate all methods** - Test emit, link (where applicable), and patch

## Directory Structure

```
tests/
├── acceptance/                    # Existing acceptance tests
└── containers/
    ├── features/
    │   ├── emit.feature           # All emit approach scenarios
    │   ├── link.feature           # All link approach scenarios
    │   └── patch.feature          # All patch approach scenarios
    ├── projects/                  # Test projects organized by approach
    │   ├── go/
    │   │   ├── emit/              # For go-emit container
    │   │   │   ├── go.mod
    │   │   │   └── main.go
    │   │   └── link/              # For go-link container
    │   │       ├── go.mod
    │   │       └── main.go
    │   ├── python/
    │   │   ├── emit/              # For python-emit container
    │   │   │   └── main.py
    │   │   └── patch/             # For python-patch container
    │   │       ├── pyproject.toml
    │   │       └── main.py
    │   ├── rust/
    │   │   ├── emit/
    │   │   ├── link/
    │   │   └── patch/
    │   └── ... (per language/approach)
    ├── images/                    # Multi-stage Dockerfiles per approach
    │   ├── go-emit.Dockerfile
    │   ├── go-link.Dockerfile
    │   ├── python-emit.Dockerfile
    │   ├── python-patch.Dockerfile
    │   └── ...
    ├── scripts/                   # Test scripts per approach
    │   ├── go-emit.sh
    │   ├── go-link.sh
    │   ├── python-emit.sh
    │   ├── python-patch.sh
    │   └── ...
    └── container_test.go          # godog step definitions
```

## Docker Official Images

| Language | Official Image | Verified |
|----------|----------------|----------|
| Go | `golang:1.23-bookworm` | Docker Official |
| Python | `python:3.12-slim-bookworm` | Docker Official |
| Rust | `rust:1.83-bookworm` | Docker Official |
| Node.js (JS/TS) | `node:22-bookworm-slim` | Docker Official |
| Java | `eclipse-temurin:21-jdk` | Docker Official (Adoptium) |
| Maven | `maven:3.9-eclipse-temurin-21` | Docker Official |
| Gradle | `gradle:8-jdk21` | Docker Official |
| .NET/C# | `mcr.microsoft.com/dotnet/sdk:8.0` | Microsoft Verified |
| PHP | `php:8.3-cli` | Docker Official |
| GCC (C/C++) | `gcc:14-bookworm` | Docker Official |
| Swift | `swift:5.10-bookworm` | Docker Official |
| Ruby | `ruby:3.3-slim-bookworm` | Docker Official |

## Test Pattern (Cucumber from host, versionator inside container)

```gherkin
Feature: Go Emit and Link Verification

  Scenario: Generated Go code compiles and runs
    When I run container "go" with test
    Then the container should exit with code 0
    And the container output should contain "Version: 1.2.3"

  Scenario: Go link injection works
    When I run container "go-link" with test
    Then the container should exit with code 0
    And the container output should contain "Version: 1.2.3"
```

## Step Definition Approach

```go
// container_test.go
const imagePrefix = "versionator-test"

func iRunContainerWithTest(container string) error {
    imageName := fmt.Sprintf("%s-%s:latest", imagePrefix, container)

    // Build the container
    buildCmd := exec.Command("docker", "build",
        "-t", imageName,
        "-f", fmt.Sprintf("tests/containers/images/%s.Dockerfile", container),
        ".")
    if out, err := buildCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("build failed: %s\n%s", err, out)
    }

    // Run the container
    runCmd := exec.Command("docker", "run", "--rm", imageName)
    output, err := runCmd.CombinedOutput()
    testContext.containerOutput = string(output)
    if exitErr, ok := err.(*exec.ExitError); ok {
        testContext.exitCode = exitErr.ExitCode()
    } else if err != nil {
        return err
    } else {
        testContext.exitCode = 0
    }
    return nil
}
```

## Multi-Stage Container Dockerfiles

Each Dockerfile builds versionator in stage 1, then uses official image in stage 2:

```dockerfile
# images/go.Dockerfile
# Stage 1: Build versionator
FROM golang:1.23-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /versionator .

# Stage 2: Go runtime with versionator
FROM golang:1.23-bookworm
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*
RUN git config --global user.email "test@test.com" && \
    git config --global user.name "Test" && \
    git config --global init.defaultBranch main
COPY --from=builder /versionator /usr/local/bin/versionator
COPY tests/containers/projects/go /test
WORKDIR /test
RUN git init && echo "1.2.3" > VERSION && git add . && git commit -m "init"
COPY tests/containers/scripts/go-test.sh /test/test.sh
RUN chmod +x /test/test.sh
CMD ["/test/test.sh"]
```

```dockerfile
# images/python.Dockerfile
FROM golang:1.23-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /versionator .

FROM python:3.12-slim-bookworm
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*
RUN git config --global user.email "test@test.com" && \
    git config --global user.name "Test" && \
    git config --global init.defaultBranch main
COPY --from=builder /versionator /usr/local/bin/versionator
COPY tests/containers/projects/python /test
WORKDIR /test
RUN git init && echo "1.2.3" > VERSION && git add . && git commit -m "init"
COPY tests/containers/scripts/python-test.sh /test/test.sh
RUN chmod +x /test/test.sh
CMD ["/test/test.sh"]
```

## Test Scripts (run inside container)

```bash
# scripts/go-test.sh
#!/bin/bash
set -e
echo "=== Go Emit Test ==="
mkdir -p version
versionator emit go --output version/version.go
cat version/version.go
go build -o app .
./app
echo "=== PASS ==="
```

```bash
# scripts/python-test.sh
#!/bin/bash
set -e
echo "=== Python Emit Test ==="
versionator emit python --output _version.py
cat _version.py
python main.py
echo "=== PASS ==="
```

## Docker Commands (Plain Docker)

```bash
# Build a container
docker build -t versionator-test-go-emit -f tests/containers/images/go-emit.Dockerfile .

# Run a container test
docker run --rm versionator-test-go-emit

# Build and run in one command
docker build -t versionator-test-python-emit -f tests/containers/images/python-emit.Dockerfile . && \
docker run --rm versionator-test-python-emit
```

## Container Matrix (Separate Containers Per Approach)

Each language gets separate containers for each supported approach:

| Language | Emit Container | Link Container | Patch Container |
|----------|----------------|----------------|-----------------|
| Go | `go-emit` | `go-link` | - |
| Python | `python-emit` | - | `python-patch` |
| Python-setuptools | `python-setuptools-emit` | - | `python-setuptools-patch` |
| Rust | `rust-emit` | `rust-link` | `rust-patch` |
| JavaScript | `js-emit` | - | `js-patch` |
| TypeScript | `ts-emit` | - | `ts-patch` |
| Java Maven | `java-maven-emit` | - | `java-maven-patch` |
| Java Gradle | `java-gradle-emit` | - | `java-gradle-patch` |
| Kotlin | `kotlin-emit` | - | `kotlin-patch` |
| C# | `csharp-emit` | - | `csharp-patch` |
| PHP | `php-emit` | - | `php-patch` |
| C | `c-emit` | `c-link` | - |
| C++ | `cpp-emit` | `cpp-link` | - |
| Swift | `swift-emit` | - | `swift-patch` |
| Ruby | `ruby-emit` | - | `ruby-patch` |

**Total: ~35 container services**

### Approach Details

**Emit Containers** - Generate version source file, compile/import, verify:
- Go: `versionator emit go` → `version/version.go` → `go build` → run
- Python: `versionator emit python` → `_version.py` → `python -c "from _version import __version__"`
- Rust: `versionator emit rust` → `src/version.rs` → `cargo build` → run

**Link Containers** - Inject version via linker flags:
- Go: `versionator link go` → get ldflags → `go build -ldflags "..."` → run
- Rust: `versionator link rust` → get rustflags → `RUSTFLAGS="..." cargo build` → run
- C/C++: `versionator link c` → get defines → `gcc -DVERSION="..."` → run

**Patch Containers** - Update manifest files:
- Python: `versionator patch` → updates `pyproject.toml` → `pip install -e .` → import version
- JavaScript: `versionator patch` → updates `package.json` → verify with `node -e "require('./package.json').version"`
- Rust: `versionator patch` → updates `Cargo.toml` → `cargo build` → run

## Implementation Phases

### Phase 1: Infrastructure
- [ ] Create `tests/containers/` directory structure
- [ ] Create base docker-compose.yml structure
- [ ] Set up godog test harness with container step definitions
- [ ] Create common test script templates

### Phase 2: Go (All Approaches)
- [ ] `go-emit`: version.go generation + compile
- [ ] `go-link`: -ldflags injection + compile

### Phase 3: Python (All Approaches)
- [ ] `python-emit`: _version.py generation + import
- [ ] `python-patch`: pyproject.toml patching + pip install

### Phase 4: Rust (All Approaches)
- [ ] `rust-emit`: version.rs generation + cargo build
- [ ] `rust-link`: RUSTFLAGS injection (if supported)
- [ ] `rust-patch`: Cargo.toml patching + cargo build

### Phase 5: JavaScript/TypeScript
- [ ] `js-emit` + `js-patch`
- [ ] `ts-emit` + `ts-patch`

### Phase 6: JVM Languages
- [ ] `java-maven-emit` + `java-maven-patch`
- [ ] `java-gradle-emit` + `java-gradle-patch`
- [ ] `kotlin-emit` + `kotlin-patch`

### Phase 7: Systems Languages
- [ ] `c-emit` + `c-link`
- [ ] `cpp-emit` + `cpp-link`
- [ ] `swift-emit` + `swift-patch`
- [ ] `csharp-emit` + `csharp-patch`

### Phase 8: Scripting Languages
- [ ] `php-emit` + `php-patch`
- [ ] `ruby-emit` + `ruby-patch`

### Phase 9: Integration
- [ ] Add just tasks for running individual/all containers
- [ ] GitHub Actions workflow with matrix
- [ ] Documentation

## Just Tasks

```just
# Run all container language tests (via godog)
container-test:
    cd tests/containers && go test -v ./...

# Run specific container test (e.g., just container-test-one go-emit)
container-test-one name:
    docker build -t versionator-test-{{name}} -f tests/containers/images/{{name}}.Dockerfile .
    docker run --rm versionator-test-{{name}}

# Build a specific container
container-build name:
    docker build -t versionator-test-{{name}} -f tests/containers/images/{{name}}.Dockerfile .

# Build all containers
container-build-all:
    #!/bin/bash
    for f in tests/containers/images/*.Dockerfile; do
        name=$(basename "$f" .Dockerfile)
        echo "Building $name..."
        docker build -t "versionator-test-$name" -f "$f" .
    done

# Clean all test containers
container-clean:
    docker images --format '{{{{.Repository}}' | grep '^versionator-test-' | xargs -r docker rmi -f

# List available containers
container-list:
    @ls tests/containers/images/*.Dockerfile | xargs -n1 basename | sed 's/.Dockerfile//'
```

## Files to Create

**Infrastructure:**
1. `tests/containers/container_test.go` (godog step definitions)
2. `tests/containers/features/emit.feature`
3. `tests/containers/features/link.feature`
4. `tests/containers/features/patch.feature`

**Per Language (~35 Dockerfiles, ~35 scripts, ~35 project dirs):**

| Language | Dockerfiles | Scripts | Projects |
|----------|-------------|---------|----------|
| Go | go-emit, go-link | go-emit.sh, go-link.sh | go/emit, go/link |
| Python | python-emit, python-patch | python-emit.sh, python-patch.sh | python/emit, python/patch |
| Rust | rust-emit, rust-link, rust-patch | 3 scripts | rust/emit, rust/link, rust/patch |
| JavaScript | js-emit, js-patch | 2 scripts | js/emit, js/patch |
| TypeScript | ts-emit, ts-patch | 2 scripts | ts/emit, ts/patch |
| Java Maven | java-maven-emit, java-maven-patch | 2 scripts | java-maven/emit, java-maven/patch |
| Java Gradle | java-gradle-emit, java-gradle-patch | 2 scripts | java-gradle/emit, java-gradle/patch |
| Kotlin | kotlin-emit, kotlin-patch | 2 scripts | kotlin/emit, kotlin/patch |
| C# | csharp-emit, csharp-patch | 2 scripts | csharp/emit, csharp/patch |
| PHP | php-emit, php-patch | 2 scripts | php/emit, php/patch |
| C | c-emit, c-link | 2 scripts | c/emit, c/link |
| C++ | cpp-emit, cpp-link | 2 scripts | cpp/emit, cpp/link |
| Swift | swift-emit, swift-patch | 2 scripts | swift/emit, swift/patch |
| Ruby | ruby-emit, ruby-patch | 2 scripts | ruby/emit, ruby/patch |

**CI/CD:**
6. `.github/workflows/container-tests.yml`
7. Updates to `justfile`
