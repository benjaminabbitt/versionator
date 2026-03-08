# Build the sample Python package with custom template
# This Makefile is used inside the container where versionator is in PATH

.PHONY: build run clean show-version version-file

VERSION := $(shell versionator version 2>/dev/null || echo "NOT SET")

# Generate _version.py from custom template
version-file:
	@echo "Generating _version.py from custom template..."
	versionator emit --template-file mypackage/_version.tmpl.py --output mypackage/_version.py

# Run the sample application
run: version-file
	python3 -m mypackage.main

# Clean build artifacts
clean:
	rm -rf build/ dist/ *.egg-info/
	rm -f mypackage/_version.py
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true

# Show the version that would be used
show-version:
	@echo "Version from versionator: $(VERSION)"
