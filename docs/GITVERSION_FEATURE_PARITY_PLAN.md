# GitVersion Feature Parity Plan for Versionator

## Executive Summary

This document outlines a plan to bring GitVersion's key features to versionator while preserving versionator's unique strengths. GitVersion is the industry-standard Git-based versioning tool with automatic version calculation from git history, while versionator focuses on explicit version control with rich multi-language code generation.

---

## Current State Comparison

### Versionator Strengths (To Preserve)

| Feature | Description |
|---------|-------------|
| **Simple VERSION file** | Plain text version storage - human readable and VCS-friendly |
| **Explicit version control** | `major/minor/patch increment/decrement` - deterministic versioning |
| **17-language code generation** | Python, Ruby, JS, TS, Go, Java, Kotlin, Rust, C, C++, C#, Swift, PHP, JSON, YAML |
| **Custom templates** | User-defined templates with `--template` and `--template-file` |
| **Template export** | `emit dump <format>` to customize embedded templates |
| **Prefix management** | Enable/disable/custom prefix with `prefix` commands |
| **Build metadata** | Configurable git hash and build info (SemVer 2.0.0 +metadata) |
| **Lightweight binary** | Single Go binary, no runtime dependencies |
| **Language agnostic** | Not tied to any specific ecosystem |
| **Git tag creation** | `commit` command with message, force, verbose options |

### GitVersion Features (To Implement)

| Feature | Priority | Description |
|---------|----------|-------------|
| **Auto-version from git history** | HIGH | Calculate version from tags + commits since tag |
| **Branching strategy support** | HIGH | GitFlow, GitHub Flow, TrunkBased workflows |
| **Pre-release labels** | HIGH | alpha, beta, rc based on branch type |
| **Commit message parsing** | MEDIUM | `+semver: major/minor/patch` in commits |
| **Deployment modes** | MEDIUM | Continuous Deployment, Continuous Delivery, Manual |
| **Rich output variables** | MEDIUM | 25+ version variables |
| **CI/CD output formats** | MEDIUM | json, dotenv, buildserver formats |
| **Version from branch name** | MEDIUM | Extract from `release-1.2.3` branches |
| **Commit count metadata** | LOW | CommitsSinceVersionSource |
| **Merge message tracking** | LOW | Extract versions from merge commits |

---

## Feature Implementation Plan

### Phase 1: Foundation - Auto-Versioning Engine

**Goal**: Add automatic version calculation from git history as an alternative to VERSION file.

#### 1.1 Version Source Detection
```
versionator auto [--source tags|file|branch]
```

- **Tags**: Find most recent semver tag, use as base version
- **File**: Current behavior (VERSION file)
- **Branch**: Extract version from branch name pattern

**New Config Options** (`.versionator.yaml`):
```yaml
versioning:
  mode: auto          # auto | manual (default: manual for backwards compat)
  source: tags        # tags | file | branch
  tagPattern: "v*"    # glob pattern for version tags
  branchPattern: "release-{version}"  # version extraction pattern
```

#### 1.2 Commits Since Version Source
Track commits since last version tag for build metadata.

**New Template Variables**:
- `{{.CommitsSinceTag}}` - Number of commits since version source
- `{{.VersionSourceSha}}` - SHA of version source commit
- `{{.VersionSourceTag}}` - Tag name of version source

**New Command**:
```
versionator info    # Display all version metadata
```

---

### Phase 2: Branching Strategies

**Goal**: Support GitFlow, GitHub Flow, and TrunkBased workflows.

#### 2.1 Workflow Presets
```
versionator init --workflow gitflow|githubflow|trunk|manual
```

**GitFlow Branch Handling**:
| Branch | Label | Increment | Example Output |
|--------|-------|-----------|----------------|
| main/master | (none) | from merge | `1.2.3` |
| develop | alpha | minor | `1.3.0-alpha.5` |
| feature/* | alpha.{branch} | minor | `1.3.0-alpha.feature-foo.3` |
| release/* | beta | from branch | `1.3.0-beta.2` |
| hotfix/* | beta | patch | `1.2.4-beta.1` |

**GitHub Flow Branch Handling**:
| Branch | Label | Increment | Example Output |
|--------|-------|-----------|----------------|
| main/master | (none) | from tag | `1.2.3` |
| feature/* | alpha | patch | `1.2.4-alpha.feature-foo.3` |
| PR branches | pr.{number} | patch | `1.2.4-pr.42.1` |

**Config Structure**:
```yaml
workflow: gitflow    # gitflow | githubflow | trunk | manual

branches:
  main:
    pattern: "^(main|master)$"
    label: ""
    increment: none
    isRelease: true
  develop:
    pattern: "^develop$"
    label: "alpha"
    increment: minor
  feature:
    pattern: "^feature/.*"
    label: "alpha.{branch}"
    increment: minor
    sourceBranches: [develop]
  release:
    pattern: "^release/.*"
    label: "beta"
    increment: none
    versionFromBranch: true
  hotfix:
    pattern: "^hotfix/.*"
    label: "beta"
    increment: patch
```

#### 2.2 Branch Detection
Automatically detect current branch and apply rules:
```go
// internal/workflow/detector.go
type BranchConfig struct {
    Pattern        string
    Label          string
    Increment      string  // major|minor|patch|none
    SourceBranches []string
    IsRelease      bool
    VersionFromBranch bool
}
```

---

### Phase 3: Pre-Release Labels

**Goal**: Support pre-release labels with automatic numbering.

#### 3.1 Label Management Commands
```
versionator label set alpha       # Set pre-release label
versionator label set rc.1        # Set with number
versionator label clear           # Remove pre-release label
versionator label increment       # alpha.1 -> alpha.2
versionator label status          # Show current label
```

#### 3.2 Template Variables
Add to existing template system:
- `{{.PreReleaseLabel}}` - Label without number (e.g., "alpha")
- `{{.PreReleaseNumber}}` - Just the number (e.g., "5")
- `{{.PreReleaseTag}}` - Full pre-release (e.g., "alpha.5")
- `{{.PreReleaseTagWithDash}}` - With dash prefix (e.g., "-alpha.5")
- `{{.SemVer}}` - Full semver with pre-release (e.g., "1.2.3-alpha.5")
- `{{.FullSemVer}}` - SemVer + build metadata (e.g., "1.2.3-alpha.5+42")

#### 3.3 VERSION File Format Extension
Support optional pre-release in VERSION file:
```
1.2.3-alpha.5
```

Or keep VERSION simple and store pre-release in config:
```yaml
# .versionator.yaml
preRelease:
  label: alpha
  number: 5
```

---

### Phase 4: Commit Message Parsing

**Goal**: Detect version bumps from commit messages.

#### 4.1 Conventional Commits Support
```
versionator auto --conventional-commits
```

**Recognized Patterns** (configurable):
| Pattern | Action |
|---------|--------|
| `+semver: major` or `+semver: breaking` | Bump major |
| `+semver: minor` or `+semver: feature` | Bump minor |
| `+semver: patch` or `+semver: fix` | Bump patch |
| `+semver: none` or `+semver: skip` | No bump |
| `feat!:` or `BREAKING CHANGE:` | Bump major (conventional commits) |
| `feat:` | Bump minor |
| `fix:` | Bump patch |

**Config**:
```yaml
commitMessages:
  enabled: true
  majorPattern: '\+semver:\s*(major|breaking)'
  minorPattern: '\+semver:\s*(minor|feature)'
  patchPattern: '\+semver:\s*(patch|fix)'
  noBumpPattern: '\+semver:\s*(none|skip)'
  conventionalCommits: true
```

#### 4.2 Scan Command
```
versionator scan [--from <ref>] [--to <ref>]
```
Output detected version bumps from commit range.

---

### Phase 5: Output Formats & CI/CD Integration

**Goal**: Add multiple output formats for CI/CD integration.

#### 5.1 Output Format Command
```
versionator output [--format json|dotenv|yaml|shell|github|gitlab]
```

**JSON Output** (default):
```json
{
  "major": 1,
  "minor": 2,
  "patch": 3,
  "preReleaseLabel": "alpha",
  "preReleaseNumber": 5,
  "semVer": "1.2.3-alpha.5",
  "fullSemVer": "1.2.3-alpha.5+42",
  "sha": "abc1234567890",
  "shortSha": "abc1234",
  "branchName": "feature/foo",
  "commitsSinceTag": 42,
  "commitDate": "2025-01-15T10:30:00Z"
}
```

**Dotenv Output**:
```
VERSIONATOR_MAJOR=1
VERSIONATOR_MINOR=2
VERSIONATOR_PATCH=3
VERSIONATOR_SEMVER=1.2.3-alpha.5
VERSIONATOR_SHA=abc1234567890
```

**GitHub Actions Output**:
```
::set-output name=semver::1.2.3-alpha.5
::set-output name=major::1
```

**GitLab CI Output**:
```
export SEMVER="1.2.3-alpha.5"
export MAJOR="1"
```

#### 5.2 Show Variable Command
```
versionator show semver           # Output: 1.2.3-alpha.5
versionator show major            # Output: 1
versionator show --format shell   # Output: export SEMVER="1.2.3"
```

---

### Phase 6: Deployment Modes

**Goal**: Support different versioning strategies for different release workflows.

#### 6.1 Mode Configuration
```yaml
deployment:
  mode: continuous-delivery  # manual | continuous-delivery | continuous-deployment

# Manual: Version only changes when explicitly tagged
# Continuous Delivery: Pre-release versions increment on each commit
# Continuous Deployment: Every commit gets unique version
```

#### 6.2 Mode-Specific Behavior

**Manual Mode** (current versionator behavior):
- Version only changes via explicit commands
- Pre-release labels are static until changed

**Continuous Delivery Mode**:
- Pre-release number increments with each commit
- Release versions require explicit tagging
- Example: `1.2.3-alpha.1` → `1.2.3-alpha.2` → ... → `1.2.3`

**Continuous Deployment Mode**:
- Every commit gets a unique version
- Build metadata always included
- Example: `1.2.3+42.abc1234`

---

### Phase 7: Extended Variables

**Goal**: Match GitVersion's rich variable set.

#### 7.1 New Template Variables

| Variable | Description |
|----------|-------------|
| `{{.Major}}` | Major version number |
| `{{.Minor}}` | Minor version number |
| `{{.Patch}}` | Patch version number |
| `{{.MajorMinorPatch}}` | `1.2.3` format |
| `{{.SemVer}}` | Full semver with pre-release |
| `{{.FullSemVer}}` | SemVer + build metadata |
| `{{.PreReleaseLabel}}` | Pre-release label (alpha, beta) |
| `{{.PreReleaseNumber}}` | Pre-release increment |
| `{{.PreReleaseTag}}` | Full pre-release tag |
| `{{.BuildMetaData}}` | Build metadata (commit count) |
| `{{.BranchName}}` | Current branch |
| `{{.EscapedBranchName}}` | Branch with slashes replaced |
| `{{.Sha}}` | Full commit SHA |
| `{{.ShortSha}}` | 7-char SHA |
| `{{.CommitsSinceVersionSource}}` | Commits since tag |
| `{{.CommitDate}}` | ISO-8601 commit date |
| `{{.VersionSourceSha}}` | Tag commit SHA |
| `{{.UncommittedChanges}}` | Count of uncommitted changes |

---

## Implementation Priority & Dependencies

```
Phase 1 (Foundation) ─────────────────────────────────────┐
    │                                                      │
    ├── 1.1 Version Source Detection                       │
    └── 1.2 Commits Since Tag                              │
                                                           │
Phase 2 (Branching) ──────────────────────────────────────┤
    │                                                      │
    ├── 2.1 Workflow Presets                               │
    └── 2.2 Branch Detection                               │
                                                           │
Phase 3 (Pre-Release) ─────────────────────────────┐       │
    │                                               │       │
    ├── 3.1 Label Commands                          │       │
    ├── 3.2 Template Variables                      │       │
    └── 3.3 VERSION Format                          │       │
                                                    │       │
Phase 4 (Commit Messages) ─────────────────────────┤       │
    │                                               │       │
    ├── 4.1 Conventional Commits                    ├───────┤
    └── 4.2 Scan Command                            │       │
                                                    │       │
Phase 5 (Output Formats) ──────────────────────────┤       │
    │                                               │       │
    ├── 5.1 Output Format Command                   │       │
    └── 5.2 Show Variable Command                   │       │
                                                    │       │
Phase 6 (Deployment Modes) ────────────────────────┤       │
    │                                               │       │
    └── Mode-specific behaviors                     │       │
                                                    │       │
Phase 7 (Extended Variables) ──────────────────────┘       │
    │                                                      │
    └── All template variables                             │
                                                           │
                                           ────────────────┘
```

---

## Backwards Compatibility Strategy

### Preserve Current Behavior
1. **Default mode**: Manual versioning (VERSION file)
2. **Existing commands**: All current commands work unchanged
3. **Configuration**: New features are opt-in via config
4. **VERSION file**: Remains the source of truth unless `versioning.mode: auto`

### Migration Path
```yaml
# Minimal config for existing users (no changes needed)
# .versionator.yaml
prefix: v

# Opt-in to auto-versioning
versioning:
  mode: auto
  source: tags

# Opt-in to workflows
workflow: githubflow
```

---

## Features NOT to Implement

These GitVersion features are out of scope:

| Feature | Reason |
|---------|--------|
| **.NET Assembly patching** | Versionator is language-agnostic; use `emit` instead |
| **MSBuild integration** | .NET-specific; not aligned with versionator's goals |
| **NuGet package** | Go binary distribution is simpler |
| **Remote repository cloning** | Out of scope; users should clone locally |

---

## New Command Summary

### New Commands
```
versionator auto              # Calculate version from git history
versionator info              # Display all version metadata
versionator init              # Initialize with workflow preset
versionator label             # Manage pre-release labels
versionator scan              # Scan commits for version bumps
versionator output            # Output version in various formats
versionator show <var>        # Show specific variable
```

### Enhanced Commands
```
versionator version           # Add --format flag
versionator emit              # Add new template variables
versionator commit            # Support pre-release in tags
```

---

## Configuration Schema (Complete)

```yaml
# .versionator.yaml - Full schema after feature parity

# Existing options (preserved)
prefix: "v"
prerelease:
  template: ""                # Mustache template, e.g., "alpha-{{CommitsSinceTag}}"
metadata:
  template: ""                # Mustache template, e.g., "{{BuildDateTimeCompact}}.{{ShortHash}}"
  git:
    hashLength: 7
logging:
  output: console

# New: Versioning mode
versioning:
  mode: manual              # manual | auto
  source: file              # file | tags | branch
  tagPattern: "v*"
  branchPattern: "release-{version}"

# New: Workflow configuration
workflow: manual            # manual | gitflow | githubflow | trunk

# New: Branch configurations
branches:
  main:
    pattern: "^(main|master)$"
    label: ""
    increment: none
    isRelease: true
  develop:
    pattern: "^develop$"
    label: "alpha"
    increment: minor
  feature:
    pattern: "^feature/.*"
    label: "alpha.{branch}"
    increment: minor
  release:
    pattern: "^release/.*"
    label: "beta"
    versionFromBranch: true
  hotfix:
    pattern: "^hotfix/.*"
    label: "beta"
    increment: patch

# New: Pre-release configuration
preRelease:
  label: ""
  number: 0

# New: Commit message parsing
commitMessages:
  enabled: false
  majorPattern: '\+semver:\s*(major|breaking)'
  minorPattern: '\+semver:\s*(minor|feature)'
  patchPattern: '\+semver:\s*(patch|fix)'
  conventionalCommits: false

# New: Deployment mode
deployment:
  mode: manual              # manual | continuous-delivery | continuous-deployment

# New: Output configuration
output:
  format: json              # json | dotenv | yaml | shell
  variables: []             # specific variables to output (empty = all)
```

---

## Success Criteria

### Feature Parity Achieved When:
- [ ] `versionator auto` calculates version from git tags
- [ ] GitFlow workflow produces same versions as GitVersion
- [ ] GitHub Flow workflow produces same versions as GitVersion
- [ ] Pre-release labels work with auto-incrementing
- [ ] Commit message parsing detects `+semver:` directives
- [ ] JSON/dotenv/shell output formats match GitVersion
- [ ] All template variables from Phase 7 are available
- [ ] Existing versionator users see no behavior changes

### Versionator Advantages Maintained:
- [ ] 17-language code generation still works
- [ ] Custom templates still supported
- [ ] VERSION file still works as primary source (manual mode)
- [ ] Single binary, no dependencies
- [ ] Fast startup time (<100ms)
- [ ] Simple configuration for simple use cases

---

## Estimated Effort by Phase

| Phase | Complexity | New Files | Modified Files |
|-------|------------|-----------|----------------|
| 1. Foundation | Medium | 3-4 | 2-3 |
| 2. Branching | High | 4-5 | 3-4 |
| 3. Pre-Release | Medium | 2-3 | 4-5 |
| 4. Commit Messages | Medium | 2-3 | 2-3 |
| 5. Output Formats | Low | 2-3 | 2-3 |
| 6. Deployment Modes | Medium | 1-2 | 3-4 |
| 7. Extended Variables | Low | 1 | 2-3 |

---

## Appendix: GitVersion Variable Mapping

| GitVersion Variable | Versionator Equivalent | Status |
|--------------------|------------------------|--------|
| Major | `{{.Major}}` | To Add |
| Minor | `{{.Minor}}` | To Add |
| Patch | `{{.Patch}}` | To Add |
| PreReleaseTag | `{{.PreReleaseTag}}` | To Add |
| PreReleaseLabel | `{{.PreReleaseLabel}}` | To Add |
| PreReleaseNumber | `{{.PreReleaseNumber}}` | To Add |
| BuildMetaData | `{{.CommitsSinceTag}}` | To Add |
| MajorMinorPatch | `{{.Version}}` | EXISTS |
| SemVer | `{{.SemVer}}` | To Add |
| FullSemVer | `{{.FullSemVer}}` | To Add |
| BranchName | `{{.BranchName}}` | To Add |
| EscapedBranchName | `{{.EscapedBranchName}}` | To Add |
| Sha | `{{.Identifier}}` | EXISTS |
| ShortSha | `{{.IdentifierShort}}` | EXISTS |
| CommitsSinceVersionSource | `{{.CommitsSinceTag}}` | To Add |
| CommitDate | `{{.CommitDate}}` | To Add |
| VersionSourceSha | `{{.VersionSourceSha}}` | To Add |
| UncommittedChanges | `{{.UncommittedChanges}}` | To Add |
| AssemblySemVer | N/A | Out of Scope |
| AssemblySemFileVer | N/A | Out of Scope |
| InformationalVersion | N/A | Out of Scope |
