c# Build the sample application with version from application

SHELL := /bin/zsh

.PHONY: build run clean show-version

build:
	@echo "Getting version from versionator..."
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	echo "Generating BuildTime.java from template..."; \
	sed -e "s/@VERSION@/$${VERSION}/g" BuildTime.java.tmpl > BuildTime.java; \
	echo "Building sample application with version: $$VERSION"; \
	javac Main.java BuildTime.java; \
	echo "Build completed: app/Main.class app/BuildTime.class"

# Run the sample application
run: build
	@java app.Main

# Clean build artifacts
clean:
	rm -rf app BuildTime.java

# Show the version that would be used
show-version:
	@VERSION=$$(../../bin/versionator version 2>/dev/null || echo "NOT SET"); \
	echo "BuildTime from versionator: $$VERSION"