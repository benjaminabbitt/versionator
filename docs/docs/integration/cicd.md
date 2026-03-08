---
title: CI/CD Integration
description: Using versionator in CI/CD pipelines
sidebar_position: 2
---

# CI/CD Integration

Versionator integrates seamlessly with CI/CD pipelines for automated version management and releases.

:::note
These examples are illustrative and may need tweaking based on your specific CI/CD environment and requirements. Patches with fixes are welcome at [github.com/benjaminabbitt/versionator](https://github.com/benjaminabbitt/versionator).
:::

## GitHub Actions

### Get Version

Capture the current version as an output:

```yaml
- name: Get version
  id: version
  run: echo "version=$(versionator version)" >> $GITHUB_OUTPUT

- name: Use version
  run: echo "Building version ${{ steps.version.outputs.version }}"
```

### Build with Version

Inject version at build time:

```yaml
- name: Build with version
  run: |
    VERSION=$(versionator version)
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

### Full Version Info

Get extended version information:

```yaml
- name: Get full version info
  id: version
  run: |
    echo "version=$(versionator version)" >> $GITHUB_OUTPUT
    echo "full=$(versionator version -t '{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}' --prefix --metadata='{{ShortHash}}')" >> $GITHUB_OUTPUT
```

### Release Workflow

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Get version from tag
        id: version
        run: echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Build
        run: |
          VERSION=${{ steps.version.outputs.version }}
          go build -ldflags "-X main.VERSION=$VERSION" -o app

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: app
          body: "Release ${{ steps.version.outputs.version }}"
```

## GitLab CI

### Basic Usage

```yaml
variables:
  VERSION: ""

before_script:
  - go install github.com/benjaminabbitt/versionator@latest
  - VERSION=$(versionator version)

build:
  script:
    - echo "Building version $VERSION"
    - go build -ldflags "-X main.VERSION=$VERSION" -o app
```

### Dynamic Version

```yaml
build:
  script:
    - |
      VERSION=$(versionator version \
        -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
        --metadata="${CI_COMMIT_SHORT_SHA}")
    - echo "Building $VERSION"
```

### Release Job

```yaml
release:
  stage: deploy
  rules:
    - if: $CI_COMMIT_TAG
  script:
    - VERSION=${CI_COMMIT_TAG#v}
    - echo "Releasing version $VERSION"
```

## Azure DevOps

### Pipeline Variables

```yaml
steps:
  - script: |
      go install github.com/benjaminabbitt/versionator@latest
      VERSION=$(versionator version)
      echo "##vso[task.setvariable variable=VERSION]$VERSION"
    displayName: 'Get Version'

  - script: |
      echo "Building version $(VERSION)"
    displayName: 'Build'
```

## Jenkins

### Declarative Pipeline

```groovy
pipeline {
    agent any

    environment {
        VERSION = ''
    }

    stages {
        stage('Get Version') {
            steps {
                script {
                    sh 'go install github.com/benjaminabbitt/versionator@latest'
                    env.VERSION = sh(script: 'versionator version', returnStdout: true).trim()
                }
            }
        }

        stage('Build') {
            steps {
                sh "go build -ldflags '-X main.VERSION=${env.VERSION}' -o app"
            }
        }
    }
}
```

## CircleCI

```yaml
version: 2.1

jobs:
  build:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run:
          name: Install versionator
          command: go install github.com/benjaminabbitt/versionator@latest
      - run:
          name: Get version
          command: |
            VERSION=$(versionator version)
            echo "export VERSION=$VERSION" >> $BASH_ENV
      - run:
          name: Build
          command: go build -ldflags "-X main.VERSION=$VERSION" -o app
```

## Common Patterns

### Conditional Release

Only release on version tags:

```yaml
# GitHub Actions
on:
  push:
    tags:
      - 'v*'
```

### Environment-specific Versions

```yaml
- name: Set version
  run: |
    if [ "${{ github.ref }}" = "refs/heads/main" ]; then
      # Production: clean version
      VERSION=$(versionator version)
    else
      # Development: add commit info
      VERSION=$(versionator version \
        -t "{{MajorMinorPatch}}-{{EscapedBranchName}}.{{CommitsSinceTag}}")
    fi
    echo "VERSION=$VERSION" >> $GITHUB_ENV
```

### Matrix Builds (Monorepo)

```yaml
jobs:
  build:
    strategy:
      matrix:
        package: [core, utils, cli]
    steps:
      - name: Get package version
        working-directory: packages/${{ matrix.package }}
        run: echo "VERSION=$(versionator version)" >> $GITHUB_OUTPUT
```

## Best Practices

1. **Install versionator in CI**: Include installation step or use a pre-built image
2. **Cache binary**: Cache the versionator binary to speed up builds
3. **Use output variables**: Capture version once, use everywhere
4. **Tag-based releases**: Trigger release workflows from Git tags
5. **Validate versions**: Check VERSION file validity in CI

## See Also

- [Git Integration](./git) - Local Git workflows
- [Makefiles and Just](./makefiles) - Build tool integration
