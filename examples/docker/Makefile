# Build the sample Docker image with version from versionator
#
# This demonstrates embedding version info in:
# 1. The compiled binary (via Go ldflags)
# 2. The container image labels (OCI standard)

SHELL := /bin/bash
VERSIONATOR := ../../bin/versionator
IMAGE_NAME := sample-app

.PHONY: docker-build docker-run docker-inspect clean show-version help

# Build Docker image with version info
docker-build:
	@echo "Getting version info from versionator..."
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "dev"); \
	COMMIT=$$($(VERSIONATOR) version -t "{{ShortHash}}" 2>/dev/null || echo "unknown"); \
	DATE=$$($(VERSIONATOR) version -t "{{BuildDateTimeUTC}}" 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ); \
	echo "Building Docker image with:"; \
	echo "  VERSION=$$VERSION"; \
	echo "  GIT_COMMIT=$$COMMIT"; \
	echo "  BUILD_DATE=$$DATE"; \
	docker build \
		--build-arg VERSION="$$VERSION" \
		--build-arg GIT_COMMIT="$$COMMIT" \
		--build-arg BUILD_DATE="$$DATE" \
		-t $(IMAGE_NAME):$$VERSION \
		-t $(IMAGE_NAME):latest \
		.

# Run the Docker container
docker-run:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "latest"); \
	echo "Running $(IMAGE_NAME):$$VERSION"; \
	docker run --rm $(IMAGE_NAME):$$VERSION

# Build and run
run: docker-build docker-run

# Inspect image labels to verify embedded version info
docker-inspect:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "latest"); \
	echo "Inspecting labels for $(IMAGE_NAME):$$VERSION"; \
	docker inspect $(IMAGE_NAME):$$VERSION --format '{{json .Config.Labels}}' | python3 -m json.tool 2>/dev/null || \
	docker inspect $(IMAGE_NAME):$$VERSION --format '{{json .Config.Labels}}'

# Show the version that would be used
show-version:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "NOT SET"); \
	COMMIT=$$($(VERSIONATOR) version -t "{{ShortHash}}" 2>/dev/null || echo "unknown"); \
	DATE=$$($(VERSIONATOR) version -t "{{BuildDateTimeUTC}}" 2>/dev/null || echo "unknown"); \
	echo "Version from versionator:"; \
	echo "  VERSION=$$VERSION"; \
	echo "  GIT_COMMIT=$$COMMIT"; \
	echo "  BUILD_DATE=$$DATE"

# Clean up Docker images
clean:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "latest"); \
	docker rmi $(IMAGE_NAME):$$VERSION $(IMAGE_NAME):latest 2>/dev/null || true

# Help
help:
	@echo "Docker example for versionator"
	@echo ""
	@echo "Targets:"
	@echo "  docker-build   Build Docker image with embedded version info"
	@echo "  docker-run     Run the Docker container"
	@echo "  run            Build and run (shortcut)"
	@echo "  docker-inspect Show image labels to verify version embedding"
	@echo "  show-version   Show version info that would be embedded"
	@echo "  clean          Remove Docker images"
	@echo ""
	@echo "Version info is embedded in:"
	@echo "  1. The Go binary (via -ldflags)"
	@echo "  2. OCI image labels (org.opencontainers.image.*)"
