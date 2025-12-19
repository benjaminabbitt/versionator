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

# Increment versions
versionator major increment   # 0.0.0 -> 1.0.0
versionator minor increment   # 1.0.0 -> 1.1.0
versionator patch increment   # 1.1.0 -> 1.1.1

# Decrement versions
versionator patch decrement   # 1.1.1 -> 1.1.0

# Create git tag for current version
versionator commit            # Creates tag v1.1.0

# Full SemVer 2.0.0 with pre-release and metadata
versionator version \
  -t "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prefix \
  --prerelease="alpha-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"
# Output: v1.1.0-alpha-5+20241211103045.abc1234
```

## Git Integration

Versionator can automatically create git tags for your versions:

```bash
# Bump version and tag in one workflow
versionator patch increment
versionator commit

# This creates an annotated git tag (e.g., v1.0.1) pointing to HEAD
# Push tags to remote
git push --tags
```

The tag name respects your prefix configuration (e.g., `v1.0.0` with prefix, `1.0.0` without).

## Usage

```
versionator [command]

Available Commands:
  version     Show current version (with template support)
  emit        Emit version in various language formats
  major       Increment or decrement major version
  minor       Increment or decrement minor version
  patch       Increment or decrement patch version
  prefix      Manage version prefix in VERSION file
  prerelease  Manage pre-release identifier in VERSION file
  metadata    Manage build metadata in VERSION file
  config      Configuration management commands
  vars        Show all template variables and their values
  commit      Create git tag for current version
  help        Help about any command

Global Flags:
  --log-format string   Log output format (console, json, development) (default "console")
  -h, --help            Help for versionator
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

Versionator stores version information in a plain text `VERSION` file containing the full SemVer string:

```
v1.2.3-alpha.1+build.123
```

The format is: `[prefix]major.minor.patch[-prerelease][+metadata]`

Examples:
- `1.0.0` - Simple version without prefix
- `v2.5.3` - Version with "v" prefix
- `release-1.0.0-beta.1` - Custom prefix with pre-release
- `v3.0.0+20241212.abc1234` - With build metadata

**Important**: The VERSION file is the **source of truth**. Its content takes priority over any configuration in `.versionator.yaml`. The prefix is parsed directly from the VERSION file (everything before the first digit). Config settings only apply as defaults when creating a new VERSION file.

**Note**: Custom variables are stored in `.versionator.yaml` config file, not in the VERSION file.

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

# Pre-release template (Mustache syntax)
# Use DASHES (-) to separate identifiers per SemVer 2.0.0
prerelease:
  template: "alpha-{{CommitsSinceTag}}"   # e.g., "alpha-5"

# Metadata template (Mustache syntax)
# Use DOTS (.) to separate identifiers per SemVer 2.0.0
metadata:
  template: "{{BuildDateTimeCompact}}.{{ShortHash}}"   # e.g., "20241211103045.abc1234"
  git:
    hashLength: 12    # Length for {{MediumHash}}

logging:
  output: "console"   # console, json, or development

# Custom variables for templates
custom:
  AppName: "MyApp"
  Environment: "production"
```

See [docs/VERSION_TEMPLATES.md](docs/VERSION_TEMPLATES.md) for complete documentation on pre-release and metadata configuration.

## Go Projects

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

This creates a `.versionator.yaml` with a pre-release template optimized for Go:

```yaml
prefix: "v"
prerelease:
  template: "{{CommitsSinceTag}}.{{BuildDateTimeCompact}}.{{ShortHash}}"
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
| **Go** | Ignored; use pre-release | `versionator init --go` |
| **npm** | Stripped on publish | Use pre-release tags |
| **PyPI** | Rejected (local versions) | Strip before publish |
| **Cargo** | Preserved | Full support |
| **NuGet** | Normalized | Preserved but deduplicated |

See [resources/semver-suffixes.md](resources/semver-suffixes.md) for detailed ecosystem comparison.

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
| **Interpreted** (Python, Ruby, JS) | `versionator emit` | No compilation step; generates source file at build time |
| **Compiled** (Go, Rust, C, C++) | Build-time variables | Avoids generated files in source control; version injected at compile time |

### Interpreted Languages: Use `emit`

For Python, Ruby, JavaScript, etc., use `versionator emit` to generate a version file:

```bash
# Generate Python _version.py
versionator emit python --output mypackage/_version.py

# Generate with custom template
versionator emit --template-file version.py.tmpl --output _version.py

# Supported formats: python, json, js, ruby, rust, go
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
| `{{BranchName}}` | `feature/foo` | Current branch |
| `{{EscapedBranchName}}` | `feature-foo` | Branch with `/` → `-` |
| `{{CommitsSinceTag}}` | `42` | Commits since last tag |
| `{{BuildNumber}}` | `42` | Alias for CommitsSinceTag |
| `{{BuildNumberPadded}}` | `0042` | Padded to 4 digits |
| `{{UncommittedChanges}}` | `3` | Count of dirty files |
| `{{Dirty}}` | `dirty` | Non-empty if uncommitted changes |
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
