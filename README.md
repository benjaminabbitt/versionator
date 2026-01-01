# Versionator

A semantic version management CLI tool that manages versions in a plain text `VERSION` file, following [SemVer 2.0.0](https://semver.org/).

## Installation

### Go Install (Recommended for Go developers)

```bash
go install github.com/benjaminabbitt/versionator@latest
```

### Homebrew (macOS/Linux)

```bash
brew tap benjaminabbitt/tap
brew install versionator
```

### Chocolatey (Windows)

```powershell
choco install versionator
```

### Debian/Ubuntu (.deb)

```bash
# Download the latest .deb package (amd64)
VERSION="1.0.0"  # Replace with desired version
wget https://github.com/benjaminabbitt/versionator/releases/download/v${VERSION}/versionator_${VERSION}_amd64.deb
sudo dpkg -i versionator_${VERSION}_amd64.deb

# Or for arm64
wget https://github.com/benjaminabbitt/versionator/releases/download/v${VERSION}/versionator_${VERSION}_arm64.deb
sudo dpkg -i versionator_${VERSION}_arm64.deb
```

### Manual Installation

Download the pre-compiled binary for your platform from [GitHub Releases](https://github.com/benjaminabbitt/versionator/releases).

#### Linux/macOS

```bash
# Download (example for Linux amd64)
curl -LO https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64

# Make executable
chmod +x versionator-linux-amd64

# Move to PATH
sudo mv versionator-linux-amd64 /usr/local/bin/versionator

# Verify installation
versionator version
```

#### Windows

1. Download `versionator-windows-amd64.exe` from [Releases](https://github.com/benjaminabbitt/versionator/releases)
2. Rename to `versionator.exe`
3. Move to a directory in your PATH (e.g., `C:\Users\<username>\bin`)
4. Or add the download location to your PATH environment variable

### Available Binaries

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | x64 | `versionator-linux-amd64` |
| Linux | arm64 | `versionator-linux-arm64` |
| macOS | Intel | `versionator-darwin-amd64` |
| macOS | Apple Silicon | `versionator-darwin-arm64` |
| Windows | x64 | `versionator-windows-amd64.exe` |
| Windows | arm64 | `versionator-windows-arm64.exe` |
| FreeBSD | x64 | `versionator-freebsd-amd64` |

All binaries are statically compiled - no dependencies required.

## Quick Start

```bash
# Initialize a new project
versionator init              # Creates VERSION and .versionator.yaml

# Initialize a Go project (enables prerelease for pseudo-version compatibility)
versionator init --go         # Creates VERSION and Go-optimized config

# Increment versions (aliases: increment, inc, bump, up, +)
versionator version major bump   # 0.0.0 -> 1.0.0
versionator ver minor up         # 1.0.0 -> 1.1.0
versionator ver patch bump       # 1.1.0 -> 1.1.1

# Decrement versions (aliases: decrement, dec, down, -)
versionator ver patch down       # 1.1.1 -> 1.1.0

# Render VERSION with dynamic values from config elements
versionator version render    # Applies prerelease/metadata from config

# Create git tag from VERSION file
versionator commit            # Creates tag with full version from VERSION

# Emit version to source files
versionator out file emit-go --output version/version.go
```

## Git Integration

Versionator can create git tags from the VERSION file:

```bash
# Bump version (automatically renders from config elements)
versionator version patch bump

# Create git tag from VERSION file
versionator output tag

# This creates an annotated git tag (e.g., v1.0.1-5-abc1234) pointing to HEAD
# Push tags to remote
git push --tags
```

The tag name is taken directly from the VERSION file. Use `versionator version render` to update VERSION with fresh dynamic values before tagging.

## Usage

```
versionator [command]

Available Commands:
  version     Show current version or manage version components (alias: ver)
    major       Manage major version (bump/up/down)
    minor       Manage minor version (bump/up/down)
    patch       Manage patch version (bump/up/down)
    revision    Manage revision version for .NET (bump/up/down)
    render      Render VERSION with fresh config elements (convenience)
    sync        Sync config from VERSION file
  output      Output version information (alias: out)
    tag         Create git tag from VERSION file
    patch       Patch version in manifest files
    build       Generate build flags for version injection
    file        Generate version source files
  prefix      Manage version prefix configuration
  prerelease  Manage pre-release configuration
  metadata    Manage build metadata configuration
  config      Configuration management commands
  vars        Show all template variables and their values
  init        Initialize versionator in current directory
  help        Help about any command

Global Flags:
  --log-format string   Log output format (console, json, development) (default "console")
  -h, --help            Help for versionator
  --version             Show versionator version
```

### Version Command Flags

```bash
versionator version [flags]

Flags:
  -t, --template string   Output template (Mustache syntax)
  -p, --prefix[=VALUE]    Enable prefix ("v" if no value, or custom)
      --prerelease[=TPL]  Enable pre-release (config default or custom template)
      --metadata[=TPL]    Enable metadata (config default or custom template)
```

**Important**: Use `=` syntax when providing values: `--prefix=release-`

## VERSION File Format

The VERSION file is the **source of truth** and contains the full version:

```
v1.2.3-alpha-5+20241228.abc1234
```

**What the VERSION file contains:**
- Prefix (e.g., `v`, `release-`)
- Core version: `Major.Minor.Patch` (or `.Revision` for .NET)
- Pre-release tag (e.g., `-alpha`, `-5-20241228-abc1234`)
- Build metadata (e.g., `+build.123`, `+20241228.abc1234`)

**Config file (.versionator.yaml) stores:**
- Default prefix for new projects
- Pre-release element templates (e.g., `["CommitsSinceTag", "ShortHash"]`)
- Metadata element templates (rendered dynamically)

**Workflow:**
```bash
# 1. Bump version (clears prerelease per SemVer)
versionator version patch bump

# 2. Render with dynamic values from config elements
versionator version render
# VERSION now contains: v1.2.4-5-20241228-abc1234

# 3. Create git tag from VERSION
versionator commit
# Creates tag: v1.2.4-5-20241228-abc1234

# 4. Emit version to other files
versionator out file emit-go --output version/version.go
```

**Note**: Dynamic values (CommitsSinceTag, ShortHash, etc.) are rendered at `version render` time and saved to VERSION. Subsequent outputs use the snapshot in VERSION.

### Managing Pre-release and Metadata

```bash
# Set pre-release directly in VERSION file
versionator prerelease set alpha
versionator prerelease set beta.1
versionator prerelease set rc.2

# Clear pre-release
versionator prerelease clear

# Set build metadata
versionator metadata set build.123
versionator metadata set 20241212

# Clear metadata
versionator metadata clear

# View current version (includes prerelease and metadata)
versionator version
# Output: 1.2.3-alpha.1+build.123
```

**Note**: Incrementing major, minor, or patch versions automatically clears the pre-release field (per SemVer 2.0.0 spec).

### Pre-release vs Metadata Templates

The `prerelease set` and `metadata set` commands set **static values** in the VERSION file. For **dynamic values** (like commit count or git hash), use templates via `.versionator.yaml` or the `--prerelease` and `--metadata` flags with the `version` command.

### Custom Variables

Custom variables are stored in the `.versionator.yaml` config file:

```yaml
# .versionator.yaml
custom:
  AppName: "MyApp"
  Environment: "production"
```

Manage custom variables with:
```bash
versionator custom set AppName "MyApp"
versionator custom get AppName
versionator custom list
versionator custom delete AppName
```

## Configuration

Versionator can be configured via a `.versionator.yaml` file in your project root.

### Generate Default Config

```bash
# Print default config to stdout
versionator config dump

# Write default config to file
versionator config dump --output .versionator.yaml
```

### Config File Format

```yaml
# .versionator.yaml
prefix: "v"

# Pre-release configuration (SemVer: appended with dash)
# Elements are variable names joined with DASHES automatically
prerelease:
  elements: ["alpha", "CommitsSinceTag"]   # → "alpha-5"

# Metadata configuration (SemVer: appended with plus)
# Elements are variable names joined with DOTS automatically
metadata:
  elements: ["BuildDateTimeCompact", "ShortHash", "Dirty"]   # → "20241211103045.abc1234.dirty"
  git:
    hashLength: 12    # Length for MediumHash variable

logging:
  output: "console"   # console, json, or development

# Custom variables for templates
custom:
  AppName: "MyApp"
  Environment: "production"
```

**Element Lists vs Templates**: The `elements` list approach automatically handles joining (dashes for prerelease, dots for metadata) and skips empty values. Each element can be a variable name (e.g., `"CommitsSinceTag"`) or a literal string (e.g., `"alpha"`).

See [docs/VERSION_TEMPLATES.md](docs/VERSION_TEMPLATES.md) for complete documentation on pre-release and metadata configuration.

## Go Projects (Pre-release Canonical Use Case)

**Pre-release is the canonical versioning feature for Go projects.** While pre-release can be used in any ecosystem, it was designed primarily to address Go's unique versioning requirements.

Go modules have unique versioning requirements that differ from other ecosystems. While SemVer 2.0.0 defines `+metadata` suffixes for build information, **Go ignores build metadata entirely** and instead uses **pseudo-versions** that embed commit information in the pre-release field.

### Why Go Uses Pre-release Instead of Metadata

Per SemVer 2.0.0, build metadata (`+suffix`) is ignored for version precedence—meaning `1.0.0+build1` and `1.0.0+build2` are considered equal. Go needs commit information to participate in version ordering for proper dependency resolution, so it places this data in the pre-release field instead:

```
v0.0.0-20231215120000-abc123def456
       │              │
       │              └── 12-character commit hash
       └── UTC timestamp (YYYYMMDDHHmmss)
```

This format ensures versions sort chronologically, which is essential for Go's module system.

### Initializing for Go Projects

Use the `--go` flag to configure versionator for Go module compatibility:

```bash
versionator init --go
```

This creates a `.versionator.yaml` with elements optimized for Go:

```yaml
prefix: "v"
prerelease:
  elements: ["CommitsSinceTag", "BuildDateTimeCompact", "ShortHash", "Dirty"]
  # → "5-20241215143052-abc1234-dirty"
metadata:
  elements: ["BuildDateTimeCompact", "ShortHash", "Dirty"]
  # → "20241215143052.abc1234.dirty"
```

### Generating Go-Compatible Versions

After initializing with `--go`, generate versions with pre-release information:

```bash
# Output version with prerelease (Go pseudo-version style)
versionator version -t "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}" --prerelease
# Example output: v1.2.3-5.20241215143052.abc1234

# For development builds (similar to Go pseudo-versions)
versionator version -t "{{Prefix}}{{MajorMinorPatch}}-{{CommitsSinceTag}}.{{BuildDateTimeCompact}}.{{MediumHash}}"
# Example output: v1.2.3-42.20241215143052.abc1234def012
```

### Go's Reserved Metadata Suffixes

Go reserves the `+` suffix for special markers only:
- `+incompatible` — for v2+ modules without proper go.mod
- `+dirty` — builds with uncommitted changes (Go 1.24+)

For build traceability in Go projects, always use pre-release identifiers rather than build metadata.

### Comparison with Other Ecosystems

| Ecosystem | Build Metadata (`+suffix`) | Recommendation |
|-----------|---------------------------|----------------|
| **Go** | Ignored; use pre-release | `versionator init --go` (canonical) |
| **npm** | Stripped on publish | Use pre-release tags |
| **PyPI** | Rejected (local versions) | Strip before publish |
| **Cargo** | Preserved | Full support |
| **NuGet** | Normalized | Preserved but deduplicated |

See [resources/semver-suffixes.md](resources/semver-suffixes.md) for detailed ecosystem comparison.

### Using Pre-release in Non-Go Projects

While pre-release is canonical for Go, you can use it in any project that benefits from Go-style version ordering. Use the `--go` flag with any language:

```bash
# Rust project with Go-compatible versioning
versionator init rust --go

# Python project with Go-compatible versioning
versionator init python --go
```

This is useful when:
- Your project is consumed by Go modules (e.g., a library with Go bindings)
- You prefer commit-sortable versions over metadata-based traceability
- Your ecosystem strips or ignores build metadata (npm, PyPI)

## Source of Truth Architecture

**The VERSION file is the single source of truth** for the current version. The `.versionator.yaml` config file stores templates for restoration.

### How Enable/Disable Commands Work

| Command | Effect |
|---------|--------|
| `prerelease set <value>` | Sets static value in VERSION file, saves as template in config |
| `prerelease template <tpl>` | Saves template in config, renders and sets in VERSION file |
| `prerelease enable` | Renders config template, sets result in VERSION file |
| `prerelease disable` | Clears pre-release from VERSION file (preserves config template) |
| `prerelease clear` | Clears pre-release from VERSION file |
| `metadata set <value>` | Sets static value in VERSION file, saves as template in config |
| `metadata template <tpl>` | Saves template in config, renders and sets in VERSION file |
| `metadata enable` | Renders config template, sets result in VERSION file |
| `metadata disable` | Clears metadata from VERSION file (preserves config template) |
| `metadata clear` | Clears metadata from VERSION file |

**Key Points:**
- **VERSION file** contains the actual current version (e.g., `v1.2.3-alpha.5+build.123`)
- **Config file** stores templates for use with `enable` and `--prerelease`/`--metadata` flags
- `disable` commands only affect the VERSION file, not the config templates
- `enable` commands render config templates and write the result to VERSION
- The `--prerelease` and `--metadata` flags on `version` command render templates dynamically without modifying files

## Integration Examples

See the `examples/` directory for complete integration examples in multiple languages.

### Choosing an Approach

| Language Type | Recommended Approach | Why |
|---------------|---------------------|-----|
| **Interpreted** (Python, Ruby, JS) | `versionator out file` | No compilation step; generates source file at build time |
| **Compiled** (Go, Rust, C, C++) | Build-time variables | Avoids generated files in source control; version injected at compile time |

### Interpreted Languages: Use `out file`

For Python, Ruby, JavaScript, etc., use `versionator out file` to generate a version file:

```bash
# Generate Python _version.py
versionator out file emit-python --output mypackage/_version.py

# Generate Go version file
versionator out file emit-go --output version/version.go

# List available emit plugins
versionator plugin list emit
```

Add the generated file to `.gitignore` to avoid polluting source control.

### Template Variables

Templates use Mustache syntax. Use `versionator vars` to see all variables with current values.

| Variable | Example | Description |
|----------|---------|-------------|
| **Version Components** | | |
| `{{Major}}` | `1` | Major version |
| `{{Minor}}` | `2` | Minor version |
| `{{Patch}}` | `3` | Patch version |
| `{{MajorMinorPatch}}` | `1.2.3` | Core version |
| `{{Prefix}}` | `v` | Version prefix |
| **Pre-release** (from `--prerelease` template) | | |
| `{{PreRelease}}` | `alpha-5` | Rendered pre-release |
| `{{PreReleaseWithDash}}` | `-alpha-5` | With auto dash prefix |
| **Metadata** (from `--metadata` template) | | |
| `{{Metadata}}` | `20241211.abc1234` | Rendered metadata |
| `{{MetadataWithPlus}}` | `+20241211.abc1234` | With auto plus prefix |
| **VCS/Git** | | |
| `{{Hash}}` | `abc1234...` | Full commit hash (40 chars for git) |
| `{{ShortHash}}` | `abc1234` | Short hash (7 chars) |
| `{{MediumHash}}` | `abc1234def01` | Medium hash (12 chars) |
| `{{ShortHashWithDot}}` | `.abc1234` | Short hash with leading dot (Go prerelease) |
| `{{MediumHashWithDot}}` | `.abc1234def01` | Medium hash with leading dot (Go prerelease) |
| `{{BranchName}}` | `feature/foo` | Current branch |
| `{{EscapedBranchName}}` | `feature-foo` | Branch with `/` → `-` |
| `{{CommitsSinceTag}}` | `42` | Commits since last tag |
| `{{BuildNumber}}` | `42` | Alias for CommitsSinceTag |
| `{{BuildNumberPadded}}` | `0042` | Padded to 4 digits |
| `{{UncommittedChanges}}` | `3` | Count of dirty files |
| `{{Dirty}}` | `dirty` | Non-empty if uncommitted changes |
| `{{DirtyWithDot}}` | `.dirty` | With leading dot (Go prerelease) |
| `{{VersionSourceHash}}` | `abc1234...` | Hash of last tag's commit |
| **Commit Author** | | |
| `{{CommitAuthor}}` | `John Doe` | Commit author name |
| `{{CommitUser}}` | `John Doe` | Alias for CommitAuthor |
| `{{CommitAuthorEmail}}` | `john@example.com` | Commit author email |
| `{{CommitUserEmail}}` | `john@example.com` | Alias for CommitAuthorEmail |
| **Commit Timestamp (UTC)** | | |
| `{{CommitDate}}` | `2024-12-11T10:30:45Z` | ISO 8601 |
| `{{CommitDateTime}}` | `2024-12-11T10:30:45Z` | Alias for CommitDate |
| `{{CommitDateCompact}}` | `20241211103045` | YYYYMMDDHHmmss |
| `{{CommitDateTimeCompact}}` | `20241211103045` | Alias for CommitDateCompact |
| `{{CommitDateShort}}` | `2024-12-11` | Date only |
| **Build Timestamp (UTC)** | | |
| `{{BuildDateTimeUTC}}` | `2024-01-15T10:30:00Z` | ISO 8601 |
| `{{BuildDateTimeCompact}}` | `20240115103045` | YYYYMMDDHHmmss |
| `{{BuildDateTimeCompactWithDot}}` | `.20240115103045` | With leading dot (Go prerelease) |
| `{{BuildDateUTC}}` | `2024-01-15` | Date only |
| **Plugin Variables (git)** | | |
| `{{GitShortHash}}` | `git.abc1234` | Prefixed short hash |
| `{{ShaShortHash}}` | `sha.abc1234` | Prefixed short hash |

See [docs/VERSION_TEMPLATES.md](docs/VERSION_TEMPLATES.md) for complete variable reference.

### Compiled Languages: Use Build-Time Variables

For Go, Rust, C, C++, inject the version at compile time:

### Using Make

```makefile
# Makefile example - inject version into Go binary
VERSION := $(shell versionator version)
build:
	go build -ldflags "-X main.VERSION=$(VERSION)" -o app
```

```makefile
# Makefile example - inject version into C++ binary
VERSION := $(shell versionator version)
build:
	g++ -DVERSION="\"$(VERSION)\"" -o app main.cpp
```

### Using Just

[Just](https://github.com/casey/just) is a modern command runner alternative to Make.

```just
# justfile example - inject version into Go binary
build:
    #!/bin/bash
    VERSION=$(versionator version)
    go build -ldflags "-X main.VERSION=$VERSION" -o app
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Get version
  id: version
  run: echo "version=$(versionator version)" >> $GITHUB_OUTPUT

- name: Build with version
  run: go build -ldflags "-X main.VERSION=${{ steps.version.outputs.version }}" -o app
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/benjaminabbitt/versionator.git
cd versionator

# Build
go build -o versionator .

# Or use just
just build
```

### Running Tests

```bash
go test ./...

# Or use just
just test
```

## Acknowledgments

The extended template variable naming in versionator draws inspiration from [GitVersion](https://gitversion.net/), an excellent Git-based semantic versioning tool. While versionator takes a different approach (explicit version management via VERSION file vs. automatic calculation from git history), we've adopted GitVersion's well-designed variable naming conventions (Major, Minor, Patch, BranchName, CommitsSinceTag, etc.) to provide familiarity for users migrating between tools and to benefit from their real-world experience in versioning workflows.

## License

BSD 3-Clause License - see [LICENSE](LICENSE) for details.
