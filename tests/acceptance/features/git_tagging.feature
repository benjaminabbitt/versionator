Feature: Git Tag Creation
  As a software developer
  I want to create git tags from my semantic version
  So that I can mark releases in version control

  Background:
    Given a clean git repository
    And a VERSION file with prefix "v" and version "1.0.0"
    And a committed file "README.md" with content "# My Project"

  Scenario: Create version tag
    When I run "versionator commit"
    Then a git tag "v1.0.0" should exist
    And the exit code should be 0

  Scenario: Create tag after version increment
    When I run "versionator minor increment"
    And I commit the VERSION changes
    And I run "versionator commit"
    Then a git tag "v1.1.0" should exist

  Scenario: Create tag with prerelease
    Given a VERSION file with prefix "v", version "1.0.0" and prerelease "alpha"
    When I run "versionator commit"
    Then a git tag "v1.0.0-alpha" should exist

  Scenario: Create tag with metadata
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "build.123"
    When I run "versionator commit"
    Then a git tag "v1.0.0+build.123" should exist

  Scenario: Tag with custom message
    When I run "versionator commit -m 'Release version 1.0.0'"
    Then a git tag "v1.0.0" should exist
    And the tag "v1.0.0" should have message "Release version 1.0.0"

  Scenario: Prevent duplicate tag
    When I run "versionator commit"
    And I run "versionator commit"
    Then the exit code should not be 0
    And the output should contain "already exists"

  Scenario: Tag points to correct commit
    When I run "versionator commit"
    Then the tag "v1.0.0" should point to HEAD

  Scenario: Multiple releases
    When I run "versionator commit"
    And I run "versionator patch increment"
    And I commit the VERSION changes
    And I run "versionator commit"
    Then a git tag "v1.0.0" should exist
    And a git tag "v1.0.1" should exist
