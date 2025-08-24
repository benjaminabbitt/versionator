# Build the sample application with version from versionator

SHELL := /bin/zsh

.PHONY: build run clean show-version

build:
	@echo "Getting version from versionator..."
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	echo "Building sample application with version: $$VERSION"; \
	javac Main.java; \
	echo "Build completed: Main.class"

# Run the sample application
run: build
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	java -DVERSION="$$VERSION" Main

# Clean build artifacts
clean:
	rm -f Main.class

# Show the version that would be used
show-version:
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	echo "BuildTime from versionator: $$VERSION"