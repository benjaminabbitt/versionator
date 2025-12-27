Feature: Emit File - Source File Generation
  As a software developer
  I want to generate version source files for my language
  So that I can compile version info into my application

  Background:
    Given a clean git repository
    And a VERSION file with version "2.3.4"

  # This restructures "emit <language>" to "emit file <language>"
  # The old syntax should remain as an alias for backwards compatibility

  Scenario: Emit file for Python
    When I run "versionator emit file python"
    Then the exit code should be 0
    And the output should contain '__version__ = "2.3.4"'
    And the output should contain '__version_tuple__ = (2, 3, 4)'

  Scenario: Emit file for Go
    When I run "versionator emit file go"
    Then the exit code should be 0
    And the output should contain 'Version     = "2.3.4"'
    And the output should contain 'Major       = 2'

  Scenario: Emit file for Rust
    When I run "versionator emit file rust"
    Then the exit code should be 0
    And the output should contain 'pub const VERSION'
    And the output should contain '"2.3.4"'

  Scenario: Emit file for JavaScript
    When I run "versionator emit file js"
    Then the exit code should be 0
    And the output should contain 'export const VERSION'

  Scenario: Emit file for TypeScript
    When I run "versionator emit file ts"
    Then the exit code should be 0
    And the output should contain 'export const VERSION: string'

  Scenario: Emit file to output path
    When I run "versionator emit file python --output _version.py"
    Then the exit code should be 0
    And the file "_version.py" should exist
    And the file "_version.py" should contain '__version__ = "2.3.4"'

  Scenario: Emit file with prerelease
    When I run "versionator emit file python --prerelease='beta'"
    Then the output should contain '__version__ = "2.3.4-beta"'

  Scenario: Emit file with metadata
    When I run "versionator emit file python --metadata='build.123'"
    Then the output should contain '__version__ = "2.3.4+build.123"'

  Scenario Outline: Emit file for all supported languages
    When I run "versionator emit file <format>"
    Then the exit code should be 0
    And the output should contain "<expected>"

    Examples:
      | format    | expected                         |
      | python    | __version__                      |
      | json      | "version"                        |
      | yaml      | version:                         |
      | go        | Version     =                    |
      | c         | #define VERSION                  |
      | c-header  | #ifndef                          |
      | cpp       | namespace version                |
      | js        | export const VERSION             |
      | ts        | export const VERSION: string     |
      | java      | public static final String       |
      | kotlin    | const val VERSION                |
      | csharp    | public const string Version      |
      | php       | const VERSION                    |
      | swift     | public let VERSION               |
      | ruby      | VERSION =                        |
      | rust      | pub const VERSION                |

  # Backwards compatibility - old syntax should still work
  Scenario: Old emit syntax still works for backwards compatibility
    When I run "versionator emit python"
    Then the exit code should be 0
    And the output should contain '__version__ = "2.3.4"'
