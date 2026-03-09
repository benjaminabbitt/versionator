---
title: ci
description: Output version variables for CI/CD systems
---

# ci

Output version variables for CI/CD systems

Output version variables in CI/CD-specific formats.

Auto-detects CI environment or use --format to specify:
  github   - GitHub Actions ($GITHUB_OUTPUT, $GITHUB_ENV)
  gitlab   - GitLab CI (dotenv artifact format)
  azure    - Azure DevOps (##vso[task.setvariable])
  circleci - CircleCI ($BASH_ENV)
  jenkins  - Jenkins (properties file format)
  shell    - Generic shell exports

Variables exported:
  VERSION, VERSION_SEMVER, VERSION_CORE,
  VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH,
  VERSION_PRERELEASE, VERSION_METADATA,
  GIT_SHA, GIT_SHA_SHORT, GIT_BRANCH, BUILD_NUMBER, DIRTY

Examples:
  versionator ci                    # Auto-detect CI and set vars
  versionator ci --format=github    # Force GitHub Actions format
  versionator ci --format=shell     # Print shell exports to stdout
  versionator ci --output=vars.env  # Write to file
  versionator ci --prefix=MYAPP_    # Variable prefix (MYAPP_VERSION, etc.)

## Installation in CI

The `ci` command renders dynamic content (git hashes, timestamps, commit counts) at runtime. This requires versionator to be installed in your CI environment.

Versionator is a **static binary** with no dependencies:

```bash
# Add to your CI pipeline
curl -sSL https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64 -o /usr/local/bin/versionator
chmod +x /usr/local/bin/versionator
```

:::tip
If you're not using dynamic pre-release or metadata templates, you can simply read the VERSION file directly without installing versionator. See [VERSION File - Static vs Dynamic Content](../concepts/version-file#static-vs-dynamic-content).
:::

## Usage

```bash
versionator ci [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --format` | string | - | Output format (github, gitlab, azure, circleci, jenkins, shell) |
| `-o, --output` | string | - | Output file (default: stdout or CI-specific location) |
| `--prefix` | string | - | Variable name prefix (e.g., 'MYAPP_' -\> MYAPP_VERSION) |

