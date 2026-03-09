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

  Scenario: Create release with metadata
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "build.123"
    When I run "versionator release"
    Then a git tag "v1.0.0+build.123" should exist

  Scenario: Release with custom message
    When I run "versionator release -m 'Release version 1.0.0'"
    Then a git tag "v1.0.0" should exist
    And the tag "v1.0.0" should have message "Release version 1.0.0"

  Scenario: Prevent duplicate tag
    When I run "versionator release"
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
