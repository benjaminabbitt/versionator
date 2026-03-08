# Build the sample TypeScript package with custom template
# This Makefile is used inside the container where versionator is in PATH

.PHONY: build run clean show-version version-file install

VERSION := $(shell versionator version 2>/dev/null || echo "NOT SET")

# Generate version.ts from custom template
version-file:
	@echo "Generating version.ts from custom template..."
	versionator emit --template-file src/version.tmpl.ts --output src/version.ts

# Build the package (generates version.ts first)
build: version-file
	@echo "Building TypeScript package..."
	npx tsc
	@echo "Build completed!"

# Install dependencies
install:
	npm install

# Run the sample application
run: install build
	node dist/index.js

# Clean build artifacts
clean:
	rm -f src/version.ts
	rm -rf dist

# Show the version that would be used
show-version:
	@echo "Version from versionator: $(VERSION)"
