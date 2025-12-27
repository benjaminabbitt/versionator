# Swift Emit Test Container
# Tests: versionator emit swift → version.swift → swiftc → run
# Requires: docker build -t versionator-builder -f tests/containers/images/versionator-builder.Dockerfile .

# Swift runtime with test
FROM swift:5.10-bookworm

# Install git for VCS info (Swift image already has git, but ensure config)
RUN git config --global user.email "test@test.com" && \
    git config --global user.name "Test" && \
    git config --global init.defaultBranch main

# Copy versionator binary from pre-built image
COPY --from=versionator-builder:latest /versionator /usr/local/bin/versionator

# Copy test project
WORKDIR /test
COPY tests/containers/projects/swift/emit/ ./

# Initialize git repo with VERSION file
RUN git init && \
    echo "1.2.3" > VERSION && \
    git add . && \
    git commit -m "init"

# Copy test script
COPY tests/containers/scripts/swift-emit.sh ./test.sh
RUN chmod +x test.sh

CMD ["./test.sh"]
