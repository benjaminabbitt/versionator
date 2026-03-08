# Docker Example

This example demonstrates embedding version info into Docker container images.

## What Gets Embedded

Version info is embedded in **two places**:

1. **The compiled binary** - Using Go's `-ldflags` to inject version at compile time
2. **OCI image labels** - Standard labels that can be inspected without running the container

## Usage

```bash
# Build the image with version from versionator
make docker-build

# Run the container
make docker-run

# Output:
# Sample Docker Application
# Version: v0.0.12 (commit: abc1234, built: 2024-01-15T10:30:00Z)

# Inspect the image labels
make docker-inspect

# Output:
# {
#     "org.opencontainers.image.version": "v0.0.12",
#     "org.opencontainers.image.revision": "abc1234",
#     "org.opencontainers.image.created": "2024-01-15T10:30:00Z",
#     ...
# }
```

## How It Works

The Dockerfile accepts build arguments:

```dockerfile
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
```

These are passed during `docker build`:

```bash
docker build \
    --build-arg VERSION=$(versionator version) \
    --build-arg GIT_COMMIT=$(versionator version -t "{{ShortHash}}") \
    --build-arg BUILD_DATE=$(versionator version -t "{{BuildDateTimeUTC}}") \
    -t myapp:$VERSION .
```

Inside the Dockerfile:
- **Build stage**: Args are passed to `go build -ldflags` to embed in binary
- **Runtime stage**: Args are used in `LABEL` instructions for image metadata

## OCI Labels

The image uses [OCI standard labels](https://github.com/opencontainers/image-spec/blob/main/annotations.md):

| Label | Description |
|-------|-------------|
| `org.opencontainers.image.version` | Semantic version |
| `org.opencontainers.image.revision` | Git commit hash |
| `org.opencontainers.image.created` | Build timestamp |

These can be queried without running the container:

```bash
docker inspect myapp:v1.2.3 --format '{{index .Config.Labels "org.opencontainers.image.version"}}'
```
