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

## Shell Completion

Generate shell completion scripts for tab-completion support:

### Bash

```bash
# Load completions for current session
source <(versionator completion bash)

# Install permanently (Linux)
versionator completion bash > /etc/bash_completion.d/versionator

# Install permanently (macOS with Homebrew)
versionator completion bash > $(brew --prefix)/etc/bash_completion.d/versionator
```

### Zsh

```zsh
# Enable completion (if not already)
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
versionator completion zsh > "${fpath[1]}/_versionator"
```

### Fish

```fish
# Load for current session
versionator completion fish | source

# Install permanently
versionator completion fish > ~/.config/fish/completions/versionator.fish
```

### PowerShell

```powershell
# Load for current session
versionator completion powershell | Out-String | Invoke-Expression

# Install permanently (add to profile)
versionator completion powershell >> $PROFILE
```

## Quick Start

```bash
# Initialize (creates VERSION file with 0.0.1)
versionator init

# Or with specific version and prefix
versionator init --version 1.0.0 --prefix v

# Increment versions
versionator major increment   # 0.0.1 -> 1.0.0
versionator minor increment   # 1.0.0 -> 1.1.0
versionator patch increment   # 1.1.0 -> 1.1.1

# Decrement versions
versionator patch decrement   # 1.1.1 -> 1.1.0

# Short aliases work too
versionator patch inc         # increment
versionator minor dec         # decrement

# Create release (commits VERSION if dirty, creates tag and branch)
versionator release           # Creates tag v1.1.0 and branch release/v1.1.0

# Full SemVer 2.0.0 with pre-release and metadata
versionator version \
  -t "{{Prefix}}{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}" \
  --prefix \
  --prerelease="alpha-{{CommitsSinceTag}}" \
  --metadata="{{BuildDateTimeCompact}}.{{ShortSha}}"
# Output: v1.1.0-alpha-5+20241211103045.abc1234
```

## Git Integration

Versionator can automatically create git tags and release branches:

```bash
# Bump version and release in one workflow
versionator patch increment
versionator release

# This will:
# 1. Auto-commit the VERSION file (if it's the only dirty file)
# 2. Create an annotated git tag (e.g., v1.0.1)
# 3. Create a release branch (e.g., release/v1.0.1)

# Push tags and branches to remote
git push --tags
git push origin release/v1.0.1
```

The tag name respects your prefix configuration (e.g., `v1.0.0` with prefix, `1.0.0` without).

For complete documentation, see the [Versionator Documentation](https://benjaminabbitt.github.io/versionator/).

## Usage

```
versionator [command]

Available Commands:
  version     Show current version (with template support)
  init        Initialize versionator in this directory
  major       Increment or decrement major version
  minor       Increment or decrement minor version
  patch       Increment or decrement patch version
  prefix      Manage version prefix in VERSION file
  prerelease  Manage pre-release identifier in VERSION file
  metadata    Manage build metadata in VERSION file
  release     Create git tag and release branch for current version
  bump        Auto-bump version based on commit messages
  emit        Emit version in various language formats
  ci          Output version variables for CI/CD systems
  mode        Manage versioning mode (release or continuous-delivery)
  custom      Manage custom key-value pairs
  vars        Show all template variables and their values
  completion  Generate shell completion scripts
  schema      Generate machine-readable CLI schema (JSON)
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

### Schema Command (AI/Tooling Integration)

Generate a JSON schema describing all commands, flags, and template variables:

```bash
# Print schema to stdout
versionator schema

# Write schema to file
versionator schema --output cli-schema.json
```

The schema includes:
- All commands and subcommands with descriptions
- Flag definitions with types and defaults
- Template variable documentation grouped by category
- Command aliases and usage patterns

Use cases:
- **AI assistants**: Provide context about available commands
- **IDE plugins**: Enable intelligent completion
- **Documentation**: Auto-generate command references
- **CI/CD**: Validate command usage in scripts

### Status Subcommands

Check the current state of prefix, prerelease, and metadata:

```bash
# Check prefix status
versionator prefix status      # Shows ENABLED/DISABLED and current value

# Check prerelease status
versionator prerelease status  # Shows ENABLED/DISABLED and current value

# Check metadata status
versionator metadata status    # Shows ENABLED/DISABLED and current value
```

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

### VERSION File Discovery (Monorepo Support)

Versionator walks up from the current directory looking for a VERSION file, enabling nested projects with independent versions:

```
myproject/
├── VERSION              # 1.0.0 (root project)
├── packages/
│   ├── VERSION          # 2.0.0 (packages workspace)
│   └── core/
│       ├── VERSION      # 3.0.0 (core package)
│       └── src/
└── apps/
    └── web/             # No VERSION - uses packages/VERSION (2.0.0)
```

```bash
# From myproject/
versionator version          # 1.0.0

# From myproject/packages/core/
versionator version          # 3.0.0

# From myproject/packages/core/src/
versionator version          # 3.0.0 (walks up to packages/core/)

# From myproject/apps/web/
versionator version          # Creates new VERSION with 0.0.1
```

This enables:
- **Monorepos**: Each package can have its own version
- **Workspaces**: Shared VERSION for related packages
- **Isolation**: Subprojects don't affect parent versions

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

### Create Config File

```bash
# Initialize with VERSION and config file
versionator init --config

# Initialize with specific version and prefix
versionator init --config --version 1.0.0 --prefix v
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

# Supported formats (17 languages):
# python, json, yaml, go, c, c-header, cpp, cpp-header,
# js, ts, java, kotlin, csharp, php, swift, ruby, rust
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
