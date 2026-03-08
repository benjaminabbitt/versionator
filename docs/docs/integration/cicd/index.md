---
title: CI/CD Integration
description: Using versionator in CI/CD pipelines
sidebar_position: 0
---

# CI/CD Integration

Versionator integrates seamlessly with CI/CD pipelines for automated version management and releases.

:::note
These examples are illustrative and may need tweaking based on your specific CI/CD environment and requirements. Patches with fixes are welcome at [github.com/benjaminabbitt/versionator](https://github.com/benjaminabbitt/versionator).
:::

## Supported Platforms

| Platform | Description |
|----------|-------------|
| [GitHub Actions](./github-actions) | Workflow automation for GitHub repositories |
| [GitLab CI](./gitlab-ci) | Built-in CI/CD for GitLab |
| [Azure DevOps](./azure-devops) | Microsoft's DevOps platform |
| [Jenkins](./jenkins) | Open-source automation server |
| [CircleCI](./circleci) | Cloud-native CI/CD platform |

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

- [Git Integration](../git) - Local Git workflows
- [Makefiles and Just](../makefiles) - Build tool integration
