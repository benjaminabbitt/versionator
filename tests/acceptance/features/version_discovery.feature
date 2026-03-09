Feature: VERSION File Discovery
  As a developer working in a monorepo or nested project structure
  I want versionator to find the closest VERSION file
  So that each subproject can have its own version

  Background:
    Given a clean git repository

  Scenario: Find VERSION file in current directory
    Given a VERSION file with version "1.0.0"
    When I run "versionator output version"
    Then the output should be "1.0.0"
    And the exit code should be 0

  Scenario: Find VERSION file in parent directory
    Given a VERSION file with version "2.0.0"
    And a subdirectory "subproject"
    When I run "versionator output version" in subdirectory "subproject"
    Then the output should be "2.0.0"
    And the exit code should be 0

  Scenario: Find VERSION file in grandparent directory
    Given a VERSION file with version "3.0.0"
    And a subdirectory "packages/mylib"
    When I run "versionator output version" in subdirectory "packages/mylib"
    Then the output should be "3.0.0"
    And the exit code should be 0

  Scenario: Nested project has its own VERSION file
    Given a VERSION file with version "1.0.0"
    And a subdirectory "subproject"
    And a VERSION file with version "2.0.0" in subdirectory "subproject"
    When I run "versionator output version" in subdirectory "subproject"
    Then the output should be "2.0.0"
    And the exit code should be 0

  Scenario: Parent VERSION unchanged when subproject has its own
    Given a VERSION file with version "1.0.0"
    And a subdirectory "subproject"
    And a VERSION file with version "2.0.0" in subdirectory "subproject"
    When I run "versionator output version"
    Then the output should be "1.0.0"

  Scenario: Increment version in nested project
    Given a VERSION file with version "1.0.0"
    And a subdirectory "subproject"
    And a VERSION file with version "2.0.0" in subdirectory "subproject"
    When I run "versionator bump patch increment" in subdirectory "subproject"
    And I run "versionator output version" in subdirectory "subproject"
    Then the output should be "2.0.1"

  Scenario: Parent version unaffected by nested increment
    Given a VERSION file with version "1.0.0"
    And a subdirectory "subproject"
    And a VERSION file with version "2.0.0" in subdirectory "subproject"
    When I run "versionator bump patch increment" in subdirectory "subproject"
    And I run "versionator output version"
    Then the output should be "1.0.0"

  Scenario: Create VERSION in current directory when not found
    And a subdirectory "newproject"
    When I run "versionator output version" in subdirectory "newproject"
    Then the output should be "0.0.1"
    And the file "newproject/VERSION" should exist

  Scenario: Multiple levels of nested projects
    Given a VERSION file with version "1.0.0"
    And a subdirectory "packages"
    And a VERSION file with version "2.0.0" in subdirectory "packages"
    And a subdirectory "packages/core"
    And a VERSION file with version "3.0.0" in subdirectory "packages/core"
    When I run "versionator output version" in subdirectory "packages/core"
    Then the output should be "3.0.0"

  Scenario: Walk up finds closest VERSION in middle level
    Given a VERSION file with version "1.0.0"
    And a subdirectory "packages"
    And a VERSION file with version "2.0.0" in subdirectory "packages"
    And a subdirectory "packages/core/src"
    When I run "versionator output version" in subdirectory "packages/core/src"
    Then the output should be "2.0.0"
