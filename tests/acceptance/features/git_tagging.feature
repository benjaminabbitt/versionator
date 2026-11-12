Feature: Git Release Creation
  As a software developer
  I want to create git tags and release branches from my semantic version
  So that I can mark releases in version control

  Background:
    Given a clean git repository
    And a VERSION file with prefix "v" and version "1.0.0"
    And a committed file "README.md" with content "# My Project"

  Scenario: Create version tag and release branch
    When I run "versionator release"
    Then a git tag "v1.0.0" should exist
    And a git branch "release/v1.0.0" should exist
    And the exit code should be 0

  Scenario: Create release after version increment with auto-commit
    When I run "versionator bump minor increment"
    And I run "versionator release"
    Then a git tag "v1.1.0" should exist
    And a git branch "release/v1.1.0" should exist
    And the output should contain "Committed VERSION file"

  Scenario: Create release with prerelease
    Given a VERSION file with prefix "v", version "1.0.0" and prerelease "alpha"
    When I run "versionator release"
    Then a git tag "v1.0.0-alpha" should exist

  # Build metadata is deliberately absent from the tag: Go module versions
  # reject SemVer build metadata, so "v1.0.0+build.123" cannot be resolved as a
  # module version. The metadata still reaches the annotated tag message, and a
  # project that wants it in the tag can add {{MetadataWithPlus}} to its own
  # release tag template.
  Scenario: Create release with metadata
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "build.123"
    When I run "versionator release"
    Then a git tag "v1.0.0" should exist
    And the tag "v1.0.0" should have message "Release 1.0.0+build.123"

  Scenario: Metadata reaches the tag when the template asks for it
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "build.123"
    And a committed config file with:
      """
      release:
        createBranch: false
        tagTemplate: "{{Prefix}}{{MajorMinorPatch}}{{MetadataWithPlus}}"
      """
    When I run "versionator release"
    Then a git tag "v1.0.0+build.123" should exist

  # {{ModulePath}} is empty at the repository root, so one template serves both
  # a root module and a submodule. That is why it carries its own trailing
  # slash rather than the template supplying one.
  Scenario: Module path template tags a root module unchanged
    Given a committed config file with:
      """
      release:
        createBranch: false
        tagTemplate: "{{ModulePath}}{{Prefix}}{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then a git tag "v1.0.0" should exist

  Scenario: Tag template with a literal component prefix
    Given a committed config file with:
      """
      release:
        createBranch: false
        tagTemplate: "tools-{{Prefix}}{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then a git tag "tools-v1.0.0" should exist

  Scenario: Release message rendered from a template
    Given a committed config file with:
      """
      release:
        createBranch: false
        messageTemplate: "chore(release): {{Tag}}"
      """
    When I run "versionator release"
    Then the tag "v1.0.0" should have message "chore(release): v1.0.0"

  Scenario: Release branch rendered from a template
    Given a committed config file with:
      """
      release:
        createBranch: true
        branchTemplate: "releases/{{MajorMinor}}"
      """
    When I run "versionator release"
    Then a git branch "releases/1.0" should exist

  Scenario: A tag template that renders an unusable name is rejected
    Given a committed config file with:
      """
      release:
        createBranch: false
        tagTemplate: "release {{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then the exit code should not be 0
    And the output should contain "invalid tag name"

  Scenario: Release with custom message
    When I run "versionator release -m 'Release version 1.0.0'"
    Then a git tag "v1.0.0" should exist
    And the tag "v1.0.0" should have message "Release version 1.0.0"

  # Re-running release with the tag already at HEAD is the documented
  # "release then push" flow and must succeed rather than demand --force.
  Scenario: Re-releasing the same commit is idempotent
    When I run "versionator release"
    And I run "versionator release"
    Then the exit code should be 0

  # A tag that exists at a DIFFERENT commit is still refused: silently moving
  # a published tag is what --force is for.
  Scenario: Prevent retagging a different commit
    When I run "versionator release"
    And I create a commit with message "chore: another commit"
    And I run "versionator release"
    Then the exit code should not be 0
    And the output should contain "already exists"

  Scenario: Tag points to correct commit
    When I run "versionator release"
    Then the tag "v1.0.0" should point to HEAD

  Scenario: Multiple releases
    When I run "versionator release"
    And I run "versionator bump patch increment"
    And I run "versionator release"
    Then a git tag "v1.0.0" should exist
    And a git tag "v1.0.1" should exist

  Scenario: Release without branch creation
    When I run "versionator release --no-branch"
    Then a git tag "v1.0.0" should exist
    And a git branch "release/v1.0.0" should not exist

  Scenario: Fail release when other files are dirty
    Given an uncommitted file "dirty.txt" with content "dirty"
    When I run "versionator release"
    Then the exit code should not be 0
    And the output should contain "working directory is not clean"
