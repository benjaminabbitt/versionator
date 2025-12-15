# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

# Target platform args (set by Docker Buildx)
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# Install git for VCS info (needed by go build)
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary for target platform
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X github.com/benjaminabbitt/versionator/internal/buildinfo.Version=${VERSION}" \
    -trimpath \
    -o versionator .

# Final stage - scratch (empty base image)
FROM scratch

# Copy CA certificates for HTTPS (if needed)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data (if needed)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/versionator /versionator

# Set the entrypoint
ENTRYPOINT ["/versionator"]
