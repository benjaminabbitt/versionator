# Build arguments for version injection
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

# Build stage
FROM golang:1.21-alpine AS builder

# Re-declare args to make them available in this stage
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

WORKDIR /app

# Copy source
COPY go.mod main.go ./

# Build with version info injected via ldflags
RUN go build -ldflags "\
    -X main.Version=${VERSION} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.BuildDate=${BUILD_DATE}" \
    -o sample-app .

# Runtime stage
FROM alpine:3.19

# Re-declare args for labels
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# OCI Image Labels - version info embedded in image metadata
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.title="sample-app"
LABEL org.opencontainers.image.description="Sample Docker application demonstrating versionator integration"
LABEL org.opencontainers.image.source="https://github.com/benjaminabbitt/versionator"

# Copy binary from builder
COPY --from=builder /app/sample-app /usr/local/bin/sample-app

ENTRYPOINT ["sample-app"]
