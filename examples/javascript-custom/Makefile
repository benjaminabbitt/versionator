# Build the sample JavaScript package with custom template
# This Makefile is used inside the container where versionator is in PATH

.PHONY: build run clean show-version version-file

VERSION := $(shell versionator version 2>/dev/null || echo "NOT SET")

# Generate version.js from custom template
version-file:
	@echo "Generating version.js from custom template..."
	versionator emit --template-file src/version.tmpl.js --output src/version.js

# Run the sample application
run: version-file
	node src/index.js

# Clean build artifacts
clean:
	rm -f src/version.js

# Show the version that would be used
show-version:
	@echo "Version from versionator: $(VERSION)"
