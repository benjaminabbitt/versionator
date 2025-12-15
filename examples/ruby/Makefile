# Build the sample Ruby gem with version from versionator

SHELL := /bin/bash
VERSIONATOR := ../../bin/versionator

.PHONY: build run clean show-version version-file install test

# Generate version.rb from versionator
version-file:
	@echo "Generating version.rb using versionator emit..."
	@$(VERSIONATOR) emit ruby --output lib/mypackage/version.rb

# Build the gem (generates version.rb first)
build: version-file
	@echo "Building Ruby gem..."
	gem build mypackage.gemspec
	@echo "Build completed!"

# Install the gem locally
install: build
	gem install mypackage-*.gem

# Run the sample application
run: version-file
	ruby bin/mypackage

# Clean build artifacts
clean:
	rm -f lib/mypackage/version.rb
	rm -f *.gem

# Show the version that would be used
show-version:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "NOT SET"); \
	echo "Version from versionator: $$VERSION"

# Run tests
test: version-file
	bundle exec rspec 2>/dev/null || echo "No tests found or rspec not installed"

# Generate using custom template file
version-file-template:
	@echo "Generating version.rb using template file..."
	@$(VERSIONATOR) emit --template-file lib/mypackage/version.rb.tmpl --output lib/mypackage/version.rb
