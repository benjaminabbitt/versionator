---
title: Monorepo Support
description: Managing independent versions in monorepos
sidebar_position: 3
---

# Monorepo Support

Versionator supports monorepos with multiple packages that need independent versioning. Each package can have its own VERSION file, and versionator automatically discovers the correct one based on your working directory.

## How It Works

Versionator walks up the directory tree from your current working directory to find the nearest VERSION file. This enables:

- **Independent versions**: Each package has its own version
- **Workspace versions**: Groups of related packages can share a version
- **Root version**: A project-wide version at the repository root

## Directory Structure Example

```
myproject/
├── VERSION                    # 1.0.0 (root project version)
├── .versionator.yaml          # Root config
│
├── packages/
│   ├── VERSION                # 2.0.0 (packages workspace)
│   ├── .versionator.yaml      # Packages config
│   │
│   ├── core/
│   │   ├── VERSION            # 3.0.0
│   │   ├── package.json
│   │   └── src/
│   │
│   ├── utils/
│   │   ├── VERSION            # 1.5.0
│   │   ├── package.json
│   │   └── src/
│   │
│   └── cli/                   # No VERSION - inherits from packages/
│       ├── package.json
│       └── src/
│
└── apps/
    └── web/
        ├── VERSION            # 0.1.0
        └── src/
```

## Working with Packages

### Check Version

```bash
# From myproject/packages/core/
versionator version
# Output: 3.0.0

# From myproject/packages/cli/ (no VERSION file)
versionator version
# Output: 2.0.0 (from packages/VERSION)

# From myproject/
versionator version
# Output: 1.0.0
```

### Bump Package Version

```bash
cd packages/core
versionator minor increment
# packages/core/VERSION: 3.1.0

cd ../utils
versionator patch increment
# packages/utils/VERSION: 1.5.1
```

### Release Individual Packages

```bash
cd packages/core
versionator release
# Creates tag: core-v3.1.0 and branch release/core-v3.1.0 (if prefix configured)
```

## Configuration Inheritance

Each directory level can have its own `.versionator.yaml`:

```yaml
# packages/.versionator.yaml
prefix: "v"
prerelease:
  template: "alpha-{{CommitsSinceTag}}"
```

The nearest config file is used for that directory.

## Common Patterns

### Workspace-Level Versioning

Group related packages under a single VERSION:

```
packages/
├── VERSION                # Shared version: 2.0.0
├── @myorg/core/
├── @myorg/utils/
└── @myorg/cli/
```

All packages in this workspace share the same version.

### Independent Package Versioning

Each package has its own VERSION:

```
packages/
├── core/
│   └── VERSION            # 3.0.0
├── utils/
│   └── VERSION            # 1.5.0
└── cli/
    └── VERSION            # 2.1.0
```

### Root + Package Versioning

Root version for the overall project, separate versions for packages:

```
myproject/
├── VERSION                # 1.0.0 (overall project)
├── packages/
│   ├── core/
│   │   └── VERSION        # Independent: 3.0.0
│   └── utils/
│       └── VERSION        # Independent: 1.5.0
└── apps/                  # Uses root VERSION
    └── web/
```

## CI/CD Integration

### Matrix Builds

Run version commands per package in CI:

```yaml
# GitHub Actions
jobs:
  version-packages:
    strategy:
      matrix:
        package: [core, utils, cli]
    steps:
      - uses: actions/checkout@v4
      - name: Get version
        working-directory: packages/${{ matrix.package }}
        run: echo "VERSION=$(versionator version)" >> $GITHUB_OUTPUT
```

### Conditional Releases

Only release packages that changed:

```yaml
- name: Check for version change
  run: |
    if git diff HEAD~1 --name-only | grep -q "packages/core/VERSION"; then
      echo "RELEASE_CORE=true" >> $GITHUB_ENV
    fi
```

## Best Practices

1. **Consistent prefixes**: Use the same prefix convention across packages
2. **Clear directory structure**: Group related packages together
3. **Version at the right level**: Don't create unnecessary VERSION files
4. **Document versioning strategy**: Make clear which packages are independently versioned

## Generating Version Files

Generate version files for each package:

```bash
# Generate Python version file for each package
for pkg in core utils cli; do
  cd packages/$pkg
  versionator emit python --output src/_version.py
  cd ../..
done
```

## See Also

- [VERSION File](./version-file) - File format and discovery
- [CI/CD Integration](../integration/cicd) - Automation patterns
