# Semver build metadata: widely specified, inconsistently supported

**The semver 2.0.0 specification defines a `+` build metadata suffix that is explicitly ignored for version precedence**—yet this simple rule creates surprising divergence across package ecosystems. While npm strips build metadata entirely, Cargo preserves it; PyPI rejects it on the public registry; Go invented an entirely different mechanism; and Maven never adopted it at all. Understanding these differences is essential for anyone embedding git commits or timestamps in version strings.

## The specification says "ignore it"—but ecosystems interpret this differently

Semver 2.0.0 (section 10) states: **"Build metadata MAY be denoted by appending a plus sign and a series of dot separated identifiers immediately following the patch or pre-release version... Build metadata MUST be ignored when determining version precedence."**

The grammar permits only ASCII alphanumerics and hyphens (`[0-9A-Za-z-]`), with dots separating multiple identifiers. Unlike pre-release identifiers, build metadata **allows leading zeros**, making it suitable for timestamps like `+20231215143052`.

Valid examples from the specification include:
- `1.0.0+20130313144700` (timestamp)
- `1.0.0-beta+exp.sha.5114f85` (pre-release with git hash)
- `1.0.0+21AF26D3----117B344092BD` (multiple hyphens allowed)

The critical implication: **`1.0.0+build1` and `1.0.0+build2` have identical precedence**. Package managers interpret "ignore for precedence" differently—some store the metadata but skip it during comparison, while others strip it entirely.

## Ecosystem support varies from full adoption to outright rejection

### npm: silently strips metadata

The npm registry **removes build metadata during publication**. Publishing `1.2.3+abc1234` stores `1.2.3`; attempting to later publish `1.2.3+xyz5678` fails because `1.2.3` already exists. The node-semver library parses metadata into a `build` array property via `semver.parse()`, but `semver.valid()` and `semver.clean()` return versions without it.

```javascript
semver.parse('1.2.3+abc123').build  // ['abc123']
semver.valid('1.2.3+abc123')         // '1.2.3' (stripped)
```

npm maintainers consider this intentional: different builds of the same semantic version shouldn't exist as separate packages since each version should be immutable once published.

### PyPI: explicitly forbidden for public uploads

PEP 440 defines "local version identifiers" with the same `+suffix` syntax, but PyPI **rejects uploads containing them**. The specification states: *"As the Python Package Index is intended solely for indexing and hosting upstream projects, it MUST NOT allow the use of local version identifiers."*

Local versions are designed for downstream integrators—Linux distributions patching upstream packages—not for original package authors. Private package indexes may allow them, and pip handles local versions during resolution by ignoring the local portion when matching specifiers like `>=1.5`.

Tools like **setuptools-scm** generate versions with local identifiers during development (`1.0.0.dev35+gaa91980.dirty`) but provide configuration to strip them before PyPI upload:

```toml
[tool.setuptools_scm]
local_scheme = "no-local-version"  # Removes +local for publishing
```

### Cargo: full support with caveats

Crates.io **preserves build metadata** and several real-world packages use it effectively:

| Package | Version | Purpose |
|---------|---------|---------|
| `libgit2-sys` | `0.12.20+1.1.0` | Bundled C library version |
| `openssl-src` | `110.0.0+1.1.0f` | OpenSSL version reference |
| `mashup` | `0.1.13+deprecated` | Deprecation notice |
| `google-bigquery2` | `2.0.4+20210327` | API schema date |

However, Cargo's documentation warns: **"Version metadata is ignored and should not be used in version requirements."** A dependency specifying `1.0.0` matches `1.0.0+foo`, `1.0.0+bar`, and plain `1.0.0` interchangeably. This creates a documented issue: crates.io permits publishing versions differing only by metadata, potentially causing non-deterministic resolution.

### Go modules: a different paradigm entirely

Go **does not use standard build metadata**. Instead, it invented **pseudo-versions** that embed commit information in the pre-release field, ensuring it participates in version ordering:

```
v0.0.0-20231215120000-abc123def456
```

This format combines:
- A base version prefix (`v0.0.0` or derived from the last tag)
- A UTC timestamp (`20231215120000`)
- A 12-character commit hash (`abc123def456`)

Because this information sits in the pre-release position rather than build metadata, pseudo-versions sort chronologically. Go reserves the `+` suffix for special markers: `+incompatible` (for v2+ modules without proper go.mod) and `+dirty` (builds with uncommitted changes, added in Go 1.24).

### Maven: no native support

Maven predates semver 2.0.0 and uses its own versioning scheme. The `+` character has no special meaning—it would be parsed as part of a qualifier string with unpredictable ordering. Maven instead uses **SNAPSHOT** conventions for development builds:

```
1.0.0-SNAPSHOT              (development reference)
1.0.0-20231215.143052-1     (timestamped deployment)
```

Remote repositories store SNAPSHOT builds with timestamps in `maven-metadata.xml`, but this is Maven-specific infrastructure unrelated to semver.

### NuGet: supported but normalized

NuGet 4.3.0+ supports semver 2.0.0 including build metadata (`1.0.0+githash`), but with a critical constraint: **version normalization removes build metadata for comparison**. Nuget.org allows only one package per normalized version—uploading `1.0.0+build1` then `1.0.0+build2` fails as a collision.

## Common patterns for datetime and git commit encoding

Practitioners have developed conventions despite inconsistent ecosystem support:

**Datetime stamps** typically use compact formats for textual sortability:
- `+YYYYMMDDHHMMSS` → `1.0.0+20231215143052`
- `+YYYYMMDD` → `1.0.0+20231215`
- `+YYYY.MM.DD` → `1.0.0+2023.12.15`

**Git commit hashes** commonly use:
- Short hash: `+abc1234`
- Prefixed: `+git.abc1234` or `+sha.5114f85`
- Combined with datetime: `+20231215.abc1234`

**Platform/build configuration:**
- `+linux.amd64`
- `+debug.x86`

**Setuptools-scm's default format** for Python development builds:
```
1.0.0.dev35+gaa91980        # 35 commits after tag, git hash aa91980
1.0.0.dev35+gaa91980.dirty  # With uncommitted changes
```

The `g` prefix indicates Git (versus `h` for Mercurial).

## Non-standard extensions push beyond the specification

When build metadata proves insufficient, projects adopt workarounds:

**Pre-release tags with commit info** participate in version ordering unlike build metadata:
```
1.2.3-alpha.1.sha.abc123    # Sortable, unlike +abc123
1.2.3-dev.22.8eaec5d3       # Commits since tag + hash
```

**Date-based versioning hybrids** treat dates as version components:
```
2023.12.15                  # CalVer: pure date-based
1.0.0-20231215.1            # Semver with date pre-release
```

**Four-part versions** (non-semver-compliant) appear in .NET ecosystems:
```
1.2.3.4                     # Major.Minor.Patch.Revision
```

NuGet supports this for `System.Version` compatibility, though it violates strict semver.

## Package managers handle metadata differently in resolution and caching

**Version resolution** universally ignores build metadata per spec—a requirement for `^1.0.0` matches `1.0.5+anything` identically to `1.0.5`. However, the treatment of stored metadata varies:

| Operation | npm | PyPI | Cargo | NuGet |
|-----------|-----|------|-------|-------|
| Registry storage | Stripped | Rejected | Preserved | Normalized |
| Lock file | N/A | N/A | Full version | Full version |
| Cache key | Version only | Version only | Version only | Normalized |

**Lock files** provide deterministic builds regardless of metadata handling. npm's `package-lock.json`, Yarn's `yarn.lock`, and NuGet's `packages.lock.json` all record exact resolved versions with integrity hashes, ensuring reproducibility even if registries handle metadata differently.

**Caching and deduplication** treat versions differing only by metadata as equivalent. This is intentional—the spec considers `1.0.0+linux` and `1.0.0+darwin` to be the same *semantic* version, even if they're different binary artifacts.

## Tooling bridges the ecosystem gaps

Several tools generate versions with embedded VCS information:

**setuptools-scm** (Python): Extracts versions from git tags with configurable formats, produces `TAG.devDISTANCE+gHASH` for development builds, offers `no-local-version` scheme for PyPI publishing.

**Versioneer** (Python): Generates `_version.py` files from VCS metadata, supports multiple output styles including PEP 440 and raw git-describe.

**GitVersion** (multi-platform): Generates semver from git history analysis, integrates with CI/CD systems, outputs build metadata for commit tracing.

**NerdBank.GitVersioning** (.NET): Adds git commit info to assemblies and packages via `version.json` configuration.

The `semver` crate in Rust provides `BuildMetadata` struct for parsing and comparing metadata values, implementing total ordering within metadata (though Cargo ignores it for requirements matching).

## Conclusion: choose patterns based on your ecosystem

The fragmentation across ecosystems means **no universal solution exists** for embedding build information in versions. For npm packages, don't rely on build metadata—use pre-release tags or store build identifiers elsewhere. For PyPI, use setuptools-scm during development but strip local identifiers before publishing. For Cargo, build metadata works but understand its limitations in resolution. For Go, embrace pseudo-versions as the idiomatic approach. For Maven, use SNAPSHOT timestamps rather than attempting semver `+` syntax.

The underlying tension is fundamental: semver's build metadata was designed to be **semantically meaningless** while providing informational context. Package managers reasonably interpret this either as "preserve but ignore" or "strip entirely since it doesn't affect behavior." Projects needing build traceability should encode information in pre-release tags when version ordering matters, or use CI/CD metadata and provenance systems when it doesn't.