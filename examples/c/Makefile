# Build the sample application with version from versionator

SHELL := /bin/zsh

.PHONY: build run clean show-version

build:
	@echo "Getting version from versionator..."
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	echo "Building sample application with version: $$VERSION"; \
	gcc -DVERSION="\"$$VERSION\"" -o sample-app main.c; \
	echo "Build completed: sample-app"

# Run the sample application
run: build
	./sample-app

# Clean build artifacts
clean:
	rm -f sample-app

# Show the version that would be used
show-version:
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	echo "Version from versionator: $$VERSION"