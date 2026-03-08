# Build the sample application with version from versionator
# This Makefile is used inside the container where versionator is in PATH

.PHONY: build run clean show-version

VERSION := $(shell versionator version 2>/dev/null || echo "NOT SET")

build:
	@echo "Getting version from versionator..."
	@echo "Building sample application with version: $(VERSION)"
	VERSION="$(VERSION)" rustc -o sample-app main.rs
	@echo "Build completed: sample-app"

run: build
	./sample-app

clean:
	rm -f sample-app

show-version:
	@echo "Version from versionator: $(VERSION)"
