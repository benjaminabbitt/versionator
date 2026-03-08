# Build the sample Ruby gem with version from versionator
# This Makefile is used inside the container where versionator is in PATH

.PHONY: build run clean show-version version-file

VERSION := $(shell versionator version 2>/dev/null || echo "NOT SET")

# Generate version.rb from versionator
version-file:
	@echo "Generating version.rb using versionator emit..."
	versionator emit ruby --output lib/mypackage/version.rb

# Run the sample application
run: version-file
	ruby -I lib -e "require 'mypackage'; Mypackage.hello"

# Clean build artifacts
clean:
	rm -f lib/mypackage/version.rb
	rm -f *.gem

# Show the version that would be used
show-version:
	@echo "Version from versionator: $(VERSION)"
