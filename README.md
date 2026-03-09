# Versionator

A semantic version management CLI tool that manages versions in a plain text `VERSION` file.

## Features

- **Single source of truth**: Plain text VERSION file
- **SemVer 2.0.0 compliant**: Full support for pre-release and build metadata
- **Deliberate version control**: Explicit `major`/`minor`/`patch` commands for teams who prefer manual versioning
- **Auto-bump available**: Optional commit message parsing via `bump` command (+semver: tags or Conventional Commits)
- **Code embedding**: Generate version constants for 10+ languages
- **CI/CD integration**: Output version variables for GitHub Actions, GitLab CI, etc.
- **Git integration**: Create tags and release branches

## Language Support

Generate version constants for your codebase (sorted by [TIOBE Index](https://www.tiobe.com/tiobe-index/)):

| Language | Format | Documentation |
|----------|--------|---------------|
| [Python](https://benjaminabbitt.github.io/versionator/commands/emit#python) | `python` | `_version.py` |
| [C](https://benjaminabbitt.github.io/versionator/commands/emit#c) | `c`, `c-header` | `version.c`, `version.h` |
| [C++](https://benjaminabbitt.github.io/versionator/commands/emit#cpp) | `cpp`, `cpp-header` | `version.cpp`, `version.hpp` |
| [Java](https://benjaminabbitt.github.io/versionator/commands/emit#java) | `java` | `Version.java` |
| [C#](https://benjaminabbitt.github.io/versionator/commands/emit#csharp) | `csharp` | `Version.cs` |
| [JavaScript](https://benjaminabbitt.github.io/versionator/commands/emit#javascript) | `js` | `version.js` |
| [Go](https://benjaminabbitt.github.io/versionator/commands/emit#go) | `go` | `version.go` |
| [TypeScript](https://benjaminabbitt.github.io/versionator/commands/emit#typescript) | `ts` | `version.ts` |
| [PHP](https://benjaminabbitt.github.io/versionator/commands/emit#php) | `php` | `Version.php` |
| [Swift](https://benjaminabbitt.github.io/versionator/commands/emit#swift) | `swift` | `Version.swift` |
| [Kotlin](https://benjaminabbitt.github.io/versionator/commands/emit#kotlin) | `kotlin` | `Version.kt` |
| [Rust](https://benjaminabbitt.github.io/versionator/commands/emit#rust) | `rust` | `version.rs` |
| [Ruby](https://benjaminabbitt.github.io/versionator/commands/emit#ruby) | `ruby` | `version.rb` |

**Data formats:** [JSON](https://benjaminabbitt.github.io/versionator/commands/emit#json), [YAML](https://benjaminabbitt.github.io/versionator/commands/emit#yaml)

**Container files:** [Containerfile/Dockerfile](https://benjaminabbitt.github.io/versionator/commands/emit#containers), [compose.yml](https://benjaminabbitt.github.io/versionator/commands/emit#containers)

## Documentation

Full documentation: **https://benjaminabbitt.github.io/versionator/**

## Quick Install

```bash
# Linux (x64)
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64.tar.gz
tar xzf versionator-linux-amd64.tar.gz
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
