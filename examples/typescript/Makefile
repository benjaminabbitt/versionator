# Build the sample TypeScript package with version from versionator

SHELL := /bin/bash
VERSIONATOR := ../../bin/versionator

.PHONY: build run clean show-version version-file install

# Generate version.ts from versionator
version-file:
	@echo "Generating version.ts using versionator emit..."
	@$(VERSIONATOR) emit ts --output src/version.ts

# Build the package (generates version.ts first)
build: version-file
	@echo "Building TypeScript package..."
	npx tsc
	@echo "Build completed!"

# Install dependencies
install:
	npm install

# Run the sample application
run: build
	node dist/index.js

# Clean build artifacts
clean:
	rm -f src/version.ts
	rm -rf dist
	rm -rf node_modules

# Show the version that would be used
show-version:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "NOT SET"); \
	echo "Version from versionator: $$VERSION"

# Generate using custom template file
version-file-template:
	@echo "Generating version.ts using template file..."
	@$(VERSIONATOR) emit --template-file src/version.tmpl.ts --output src/version.ts
