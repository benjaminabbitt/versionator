---
title: Docker
description: Embed version in Docker containers
sidebar_position: 13
---

# Docker / Containers

**Location:** [`examples/docker/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/docker)

Container images embed version info in two places:
1. **The binary inside** (using the language-specific approach)
2. **OCI image labels** (for image inspection without running)

```dockerfile title="examples/docker/Dockerfile"
# Build arguments
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

FROM golang:1.21-alpine AS builder

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

WORKDIR /app
COPY . .

# Inject version at compile time
RUN go build -ldflags "\
    -X main.Version=${VERSION} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.BuildDate=${BUILD_DATE}" \
    -o /app/sample-app

FROM alpine:3.19

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# OCI Image Labels
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"

COPY --from=builder /app/sample-app /usr/local/bin/sample-app

ENTRYPOINT ["sample-app"]
```

```makefile title="examples/docker/Makefile (excerpt)"
docker-build:
    VERSION=$$(versionator version); \
    COMMIT=$$(versionator version -t "{{ShortHash}}"); \
    DATE=$$(versionator version -t "{{BuildDateTimeUTC}}"); \
    docker build \
        --build-arg VERSION=$$VERSION \
        --build-arg GIT_COMMIT=$$COMMIT \
        --build-arg BUILD_DATE=$$DATE \
        -t sample-app:$$VERSION .
```

## Run it

```bash
$ cd examples/docker && just show-version
Version from versionator:
  VERSION=0.0.16
  GIT_COMMIT=ba4ecb3
  BUILD_DATE=2026-03-08T18:52:29Z

$ just docker-build
Building Docker image with:
  VERSION=0.0.16
  GIT_COMMIT=ba4ecb3
  BUILD_DATE=2026-03-08T18:52:29Z
...

$ just docker-run
Running sample-app:0.0.16
Sample Docker Application
Version: 0.0.16 (commit: ba4ecb3, built: 2026-03-08T18:52:29Z)
```

## Source Code

- [`main.go`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/main.go)
- [`Dockerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/Dockerfile)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/Makefile)
