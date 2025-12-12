# Dockerfile for running acceptance tests
# Includes versionator binary and all necessary tools for testing

FROM golang:1.23-bookworm

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    tzdata \
    bash \
    make \
    && rm -rf /var/lib/apt/lists/*

# Set up git configuration for tests
RUN git config --global user.email "test@example.com" && \
    git config --global user.name "Test User" && \
    git config --global init.defaultBranch main

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build versionator binary
RUN CGO_ENABLED=0 go build -o /usr/local/bin/versionator .

# Set environment variable for test runner to find the binary
ENV VERSIONATOR_PROJECT_ROOT=/app
ENV PATH="/usr/local/bin:${PATH}"

# Create a directory for test runs
RUN mkdir -p /test-workspace

WORKDIR /test-workspace

# Default command runs the acceptance tests
CMD ["go", "test", "-v", "/app/tests/acceptance/..."]
