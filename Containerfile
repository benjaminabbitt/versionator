# Minimal container for versionator
# Works with Docker, Podman, and other OCI-compliant runtimes

FROM scratch

# Copy the statically-linked binary
# ARG TARGETARCH is automatically set by buildx/buildah
ARG TARGETARCH
COPY versionator-linux-${TARGETARCH} /versionator

# Set the entrypoint
ENTRYPOINT ["/versionator"]
