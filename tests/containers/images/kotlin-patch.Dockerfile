# Kotlin Patch Test Container
# Tests: versionator emit patch → build.gradle.kts → gradle build → run
# Requires: docker build -t versionator-builder -f tests/containers/images/versionator-builder.Dockerfile .

# Gradle with JDK for Kotlin
FROM gradle:8-jdk21

# Install git for VCS info
USER root
RUN apt-get update && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*

# Configure git
RUN git config --global user.email "test@test.com" && \
    git config --global user.name "Test" && \
    git config --global init.defaultBranch main && \
    git config --global --add safe.directory /test

# Copy versionator binary from pre-built image
COPY --from=versionator-builder:latest /versionator /usr/local/bin/versionator

# Copy test project
WORKDIR /test
COPY tests/containers/projects/kotlin/patch/ ./

# Initialize git repo with VERSION file
RUN git init && \
    echo "1.2.3" > VERSION && \
    git add . && \
    git commit -m "init"

# Copy test script and fix ownership
COPY tests/containers/scripts/kotlin-patch.sh ./test.sh
RUN chmod +x test.sh && chown -R gradle:gradle /test

USER gradle
CMD ["./test.sh"]
