# Build the sample application with version from versionator
# This Makefile is used inside the container where versionator is in PATH

.PHONY: build run clean show-version

VERSION := $(shell versionator version 2>/dev/null || echo "NOT SET")

build:
	@echo "Getting version from versionator..."
	@echo "Generating BuildTime.java from template..."
	sed -e "s/@VERSION@/$(VERSION)/g" app/BuildTime.tmpl.java > app/BuildTime.java
	@echo "Building sample application with version: $(VERSION)"
	javac app/Main.java app/BuildTime.java
	@echo "Build completed: app/Main.class app/BuildTime.class"

run: build
	java app.Main

clean:
	rm -f app/*.class app/BuildTime.java

show-version:
	@echo "Version from versionator: $(VERSION)"
