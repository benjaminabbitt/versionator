# Build the sample JavaScript package with version from versionator

SHELL := /bin/bash
VERSIONATOR := ../../bin/versionator

.PHONY: build run clean show-version version-file install test

# Generate version.js from versionator
version-file:
	@echo "Generating version.js using versionator emit..."
	@$(VERSIONATOR) emit js --output src/version.js

# Build the package (generates version.js first)
build: version-file
	@echo "Building JavaScript package..."
	npm run build
	@echo "Build completed!"

# Install dependencies
install:
	npm install

# Run the sample application
run: version-file
	node src/index.js

# Clean build artifacts
clean:
	rm -f src/version.js
	rm -rf node_modules

# Show the version that would be used
show-version:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "NOT SET"); \
	echo "Version from versionator: $$VERSION"

# Generate using custom template file
version-file-template:
	@echo "Generating version.js using template file..."
	@$(VERSIONATOR) emit --template-file src/version.tmpl.js --output src/version.js
