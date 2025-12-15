# Tasks - CLAUDE Todo

## In Progress


## To Do

- [ ] Simplify to basic VERSION file with just full rendered version string (parse components from it instead of JSON schema). Eliminate VERSION.json from codebase. Do not consider backwards compatibility and remove previous backwards compatibility functionality (migration from straight version to json)

## Backlog

## Done

- [x] Add prerelease and metadata set/clear commands for managing VERSION.json fields directly
- [x] Update all tests for new JSON schema (added acceptance tests for prerelease/metadata commands)
- [x] Update README and help text for new VERSION.json format
- [x] Add template variable aliases: CommitDateTime, CommitDateTimeCompact, CommitUser, CommitUserEmail
- [x] Credit GitVersion in README for variable naming inspiration
- [x] Extend VCS interface with new methods (GetBranchName, GetCommitDate, GetCommitsSinceTag, GetUncommittedChanges)
- [x] Implement new VCS methods in git/git_vcs.go
- [x] Add version parsing to extract Major/Minor/Patch components
- [x] Create internal/version/semver.go for semantic version parsing
- [x] Extend TemplateData struct with all new variables (Major, Minor, Patch, BranchName, etc.)
- [x] Update emit.go RenderTemplate to populate all new variables
- [x] Design SemVer format string variables (SemVer, FullSemVer, PreReleaseTag, BuildMetaData)
- [x] Implement pre-formatted variables with automatic +/. separators (e.g., PreReleaseTagWithDash, FullBuildMetaData)
- [x] Add composite variables: SemVer = Major.Minor.Patch[-PreRelease], FullSemVer = SemVer[+BuildMetaData]
- [x] Create helper functions for conditional separator injection (only add - if PreRelease exists, only add + if BuildMetaData exists)
- [x] Add InformationalVersion variable (full human-readable version string)
- [x] Add BuildDateTimeCompact variable (YYYYMMDDHHmmss format, e.g., 20250115103045)
- [x] Update cmd/emit.go help text with new template variables
- [x] Update README with new template variable documentation
- [x] Add unit tests for new VCS methods
- [x] Add unit tests for version parsing
- [x] Add unit tests for extended template variables
- [x] Add unit tests for SemVer format string generation with edge cases (no prerelease, no metadata, both, neither)
- [x] Design VERSION.json schema with prefix, major, minor, patch, prerelease, metadata fields
- [x] Create internal/version/schema.go for VERSION.json read/write
- [x] Add migration logic: detect VERSION file and convert to VERSION.json
- [x] Update version.go to use new JSON schema instead of plain text
- [x] Add --template flag to `versionator version` command
- [x] Update major/minor/patch increment/decrement to work with JSON schema
- [x] Update prefix commands to modify JSON prefix field
- [x] Update emit.go to read components from VERSION.json
- [x] Ensure default output is strict SemVer 2.0.0 (Major.Minor.Patch)
- [x] Add dump subcommand to output default .versionator.yaml config
- [x] Update VCS interface with new methods
- [x] Add CommitDate/CommitDateTime template variables
- [x] Add CommitAuthor/CommitUserEmail template variables
- [x] Add PreReleaseLabel/PreReleaseNumber template variables
- [x] Add VersionSourceSha template variable
- [x] Add BuildNumber/BuildMetaDataPadded (padded commit count) variables
- [x] Update emit.go TemplateData with new fields
- [x] Update default config documentation with new variables
- [x] Add custom key/value pairs support to VERSION.json schema
- [x] Add --set flag to inject key/value pairs from command line
- [x] Update emit.go to merge custom vars into TemplateData
