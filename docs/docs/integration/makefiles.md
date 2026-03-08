---
title: Makefiles and Just
description: Integrating versionator with build tools
sidebar_position: 3
---

# Makefiles and Just

Versionator integrates seamlessly with Make and [Just](https://github.com/casey/just) command runners.

## Make

### Basic Integration

```makefile
# Get version from versionator
VERSION := $(shell versionator version)

# Build with version
build:
	go build -ldflags "-X main.VERSION=$(VERSION)" -o app

# Show version
version:
	@echo $(VERSION)
```

### Full Version String

```makefile
# With pre-release and metadata
FULL_VERSION := $(shell versionator version \
	-t "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
	--prefix \
	--metadata="{{ShortHash}}")

build:
	go build -ldflags "-X main.VERSION=$(FULL_VERSION)" -o app
```

### Version Bump Targets

```makefile
.PHONY: bump-major bump-minor bump-patch release

bump-major:
	versionator major increment
	@echo "Version: $(shell versionator version)"

bump-minor:
	versionator minor increment
	@echo "Version: $(shell versionator version)"

bump-patch:
	versionator patch increment
	@echo "Version: $(shell versionator version)"

release: bump-patch
	git add VERSION
	git commit -m "Release $(shell versionator version)"
	versionator tag
	git push
	git push --tags
```

### Multiple Languages

```makefile
VERSION := $(shell versionator version)

# Generate version files
generate-version:
	versionator emit python --output src/_version.py
	versionator emit json --output version.json

# Build all
build: generate-version
	python -m build
	docker build -t myapp:$(VERSION) .
```

### C/C++ Integration

```makefile
VERSION := $(shell versionator version)

# Pass version as compiler define
CFLAGS += -DVERSION="\"$(VERSION)\""

app: main.c
	$(CC) $(CFLAGS) -o $@ $<
```

## Just

[Just](https://github.com/casey/just) is a modern command runner alternative to Make.

### Basic Integration

```just
# Get version
version := `versionator version`

# Build with version
build:
    go build -ldflags "-X main.VERSION={{version}}" -o app

# Show version
show-version:
    @echo {{version}}
```

### Shell Interpolation

```just
# Using shell in recipe
build:
    #!/bin/bash
    VERSION=$(versionator version)
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

### Version Bumping

```just
# Bump and show new version
bump-major:
    versionator major increment
    @versionator version

bump-minor:
    versionator minor increment
    @versionator version

bump-patch:
    versionator patch increment
    @versionator version
```

### Release Recipe

```just
# Full release workflow
release bump="patch":
    #!/bin/bash
    set -e

    # Bump version
    versionator {{bump}} increment
    VERSION=$(versionator version)

    # Commit and tag
    git add VERSION
    git commit -m "Release $VERSION"
    versionator tag

    # Push
    git push
    git push --tags

    echo "Released $VERSION"
```

### Generate Version Files

```just
# Generate version files for all languages
generate-version:
    versionator emit python --output src/_version.py
    versionator emit json --output version.json
    versionator emit go --output internal/version/version.go

# Build (depends on generate)
build: generate-version
    go build -o app
```

### Dynamic Metadata

```just
# Build with dynamic metadata
build-ci:
    #!/bin/bash
    VERSION=$(versionator version \
        -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
        --metadata="{{BuildDateTimeCompact}}.{{ShortHash}}")
    echo "Building $VERSION"
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

### Cross-Platform Recipes

```just
# Platform-specific builds
build-linux:
    GOOS=linux GOARCH=amd64 go build \
        -ldflags "-X main.VERSION={{version}}" \
        -o dist/app-linux-amd64

build-darwin:
    GOOS=darwin GOARCH=arm64 go build \
        -ldflags "-X main.VERSION={{version}}" \
        -o dist/app-darwin-arm64

build-windows:
    GOOS=windows GOARCH=amd64 go build \
        -ldflags "-X main.VERSION={{version}}" \
        -o dist/app-windows-amd64.exe

# Build all platforms
build-all: build-linux build-darwin build-windows
```

### Documentation Tasks

```just
# Docs tasks
docs-dev:
    cd docs && npm start

docs-build:
    cd docs && npm run build

docs-generate:
    node docs/scripts/generate-command-docs.js
```

## Comparison

| Feature | Make | Just |
|---------|------|------|
| Variable syntax | `$(VAR)` | `{{var}}` |
| Shell recipe | Requires `@` | Native support |
| Cross-platform | Limited | Better |
| Recipe arguments | Complex | Simple |
| Dependencies | Tabs required | Flexible |

## Best Practices

1. **Cache version**: Capture version once per build, not per target
2. **Use variables**: Define version at top, reference everywhere
3. **Clean targets**: Include version in artifact names for clarity
4. **Document recipes**: Add comments explaining complex recipes
5. **Fail fast**: Use `set -e` in shell recipes

## See Also

- [CI/CD Integration](./cicd) - Pipeline integration
- [Language Integration](./languages) - Language-specific patterns
