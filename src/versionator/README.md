# Versionator Devcontainer Feature

Installs [versionator](https://github.com/benjaminabbitt/versionator) - a semantic versioning tool for CI/CD pipelines.

## Usage

Add to your `devcontainer.json`:

```json
{
    "features": {
        "ghcr.io/benjaminabbitt/versionator/versionator:1": {}
    }
}
```

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `version` | string | `latest` | Version of versionator to install (e.g., `0.1.0` or `latest`) |

### Example with specific version

```json
{
    "features": {
        "ghcr.io/benjaminabbitt/versionator/versionator:1": {
            "version": "0.1.0"
        }
    }
}
```

## What is Versionator?

Versionator is a semantic versioning tool that:

- Manages VERSION files with full semver support
- Generates version variables for CI/CD pipelines (GitHub Actions, GitLab CI, etc.)
- Creates release tags and branches
- Supports pre-release and metadata components
- Works with multiple programming language version files (package.json, Cargo.toml, etc.)

## Quick Start

After installing via devcontainer feature:

```bash
# Initialize versionator in your project
versionator init

# Check current version
versionator output version

# Bump patch version
versionator bump patch increment

# Create a release
versionator release
```

## Learn More

- [Documentation](https://github.com/benjaminabbitt/versionator)
- [Release Notes](https://github.com/benjaminabbitt/versionator/releases)
