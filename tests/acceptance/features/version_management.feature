Feature: Version Management
  As a software developer
  I want to manage semantic versions in my project
  So that I can track releases and communicate changes to users

  Background:
    Given a clean git repository
    And a VERSION.json file with version "1.0.0"

  Scenario: Display current version
    When I run "versionator version"
    Then the output should be "1.0.0"
    And the exit code should be 0

  Scenario: Display version with prefix
    Given a VERSION.json file with prefix "v" and version "1.2.3"
    When I run "versionator version -t '{{Prefix}}{{MajorMinorPatch}}' --prefix"
    Then the output should be "v1.2.3"
    And the exit code should be 0

  Scenario: Increment major version
    When I run "versionator major increment"
    And I run "versionator version"
    Then the output should be "2.0.0"

  Scenario: Increment minor version
    When I run "versionator minor increment"
    And I run "versionator version"
    Then the output should be "1.1.0"

  Scenario: Increment patch version
    When I run "versionator patch increment"
    And I run "versionator version"
    Then the output should be "1.0.1"

  Scenario: Multiple version increments
    When I run "versionator patch increment"
    And I run "versionator patch increment"
    And I run "versionator patch increment"
    And I run "versionator version"
    Then the output should be "1.0.3"

  Scenario: Major reset resets minor and patch
    Given a VERSION.json file with version "1.5.3"
    When I run "versionator major increment"
    And I run "versionator version"
    Then the output should be "2.0.0"

  Scenario: Minor reset resets patch
    Given a VERSION.json file with version "1.5.3"
    When I run "versionator minor increment"
    And I run "versionator version"
    Then the output should be "1.6.0"

  Scenario: Version with custom template
    Given a VERSION.json file with version "2.1.0"
    When I run "versionator version -t '{{Major}}.{{Minor}}'"
    Then the output should be "2.1"

  Scenario: Set prefix
    When I run "versionator prefix set release-"
    Then the VERSION.json should have prefix "release-"

  Scenario: Clear prefix
    Given a VERSION.json file with prefix "v" and version "1.0.0"
    When I run "versionator prefix disable"
    Then the VERSION.json should have prefix ""
