# PHP Emit Test Container
# Tests: versionator emit php → version.php → php require
# Requires: docker build -t versionator-builder -f tests/containers/images/versionator-builder.Dockerfile .

# PHP CLI runtime with test
FROM php:8.3-cli

# Install git for VCS info
RUN apt-get update && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*

# Configure git
RUN git config --global user.email "test@test.com" && \
    git config --global user.name "Test" && \
    git config --global init.defaultBranch main

# Copy versionator binary from pre-built image
COPY --from=versionator-builder:latest /versionator /usr/local/bin/versionator

# Copy test project
WORKDIR /test
COPY tests/containers/projects/php/emit/ ./

# Initialize git repo with VERSION file
RUN git init && \
    echo "1.2.3" > VERSION && \
    git add . && \
    git commit -m "init"

# Copy test script
COPY tests/containers/scripts/php-emit.sh ./test.sh
RUN chmod +x test.sh

CMD ["./test.sh"]
