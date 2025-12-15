# Build the sample Python package with version from versionator
# Demonstrates two approaches: generated file vs environment variable

SHELL := /bin/bash
VERSIONATOR := ../../bin/versionator

.PHONY: build run clean show-version version-file install test env-run

#
# Approach 1: Generated Source File (Recommended)
#

# Generate _version.py from versionator
version-file:
	@echo "Generating _version.py using versionator emit..."
	@$(VERSIONATOR) emit python --output mypackage/_version.py

# Build the package (generates _version.py first)
build: version-file
	@echo "Building Python package..."
	python -m build
	@echo "Build completed: dist/"

# Install in development mode
install: version-file
	pip install -e .

# Run the sample application (Approach 1)
run: version-file
	python -m mypackage.main

#
# Approach 2: Environment Variable
#

# Run using environment variable (no file generation)
env-run:
	@echo "Running with version from environment variable..."
	APP_VERSION=$$($(VERSIONATOR) version) python -c "\
		import os; \
		print('Sample Python Application'); \
		print(f'Version: {os.environ.get(\"APP_VERSION\", \"unknown\")}')"

#
# Utility targets
#

# Clean build artifacts
clean:
	rm -rf build/ dist/ *.egg-info/
	rm -f mypackage/_version.py
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true

# Show the version that would be used
show-version:
	@VERSION=$$($(VERSIONATOR) version 2>/dev/null || echo "NOT SET"); \
	echo "Version from versionator: $$VERSION"

# Run tests (if any)
test: version-file
	python -m pytest tests/ -v 2>/dev/null || echo "No tests found"

#
# Custom template example
#

# Generate using custom template file
version-file-template:
	@echo "Generating _version.py using template file..."
	@$(VERSIONATOR) emit --template-file mypackage/_version.py.tmpl --output mypackage/_version.py
