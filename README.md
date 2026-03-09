# Versionator

A semantic version management CLI tool that manages versions in a plain text `VERSION` file.

## Features

- **Single source of truth**: Plain text VERSION file
- **SemVer 2.0.0 compliant**: Full support for pre-release and build metadata
- **Automatic version bumping**: From commit messages (+semver: tags or Conventional Commits)
- **Code embedding**: Generate version files for 17+ languages
- **CI/CD integration**: Output version variables for GitHub Actions, GitLab CI, etc.
- **Git integration**: Create tags and release branches

## Documentation

Full documentation: **https://benjaminabbitt.github.io/versionator/**

## Quick Install

```bash
# Linux/macOS
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64
chmod +x versionator-linux-amd64
sudo mv versionator-linux-amd64 /usr/local/bin/versionator

# Verify
versionator version
```

See [Installation](https://benjaminabbitt.github.io/versionator/installation) for all platforms.

## Quick Start

```bash
versionator init                  # Create VERSION file (0.0.1)
versionator patch increment       # 0.0.1 -> 0.0.2
versionator release               # Create tag v0.0.2
```

## License

BSD-3-Clause
