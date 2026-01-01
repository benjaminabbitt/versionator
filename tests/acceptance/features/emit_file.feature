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
    When I run "versionator output file emit-python"
    Then the exit code should be 0
    And the output should contain '__version__ = "2.3.4"'
    And the output should contain '__version_tuple__ = (2, 3, 4)'

  Scenario: Emit file for Go
    When I run "versionator output file emit-go"
    Then the exit code should be 0
    And the output should contain 'Version     = "2.3.4"'
    And the output should contain 'Major       = 2'

  Scenario: Emit file for Rust
    When I run "versionator output file emit-rust"
    Then the exit code should be 0
    And the output should contain 'pub const VERSION'
    And the output should contain '"2.3.4"'

  Scenario: Emit file for JavaScript
    When I run "versionator output file emit-javascript"
    Then the exit code should be 0
    And the output should contain 'export const VERSION'

  Scenario: Emit file for TypeScript
    When I run "versionator output file emit-typescript"
    Then the exit code should be 0
    And the output should contain 'export const VERSION: string'

  Scenario: Emit file to output path
    When I run "versionator output file emit-python --output _version.py"
    Then the exit code should be 0
    And the file "_version.py" should exist
    And the file "_version.py" should contain '__version__ = "2.3.4"'

  Scenario: Emit file with prerelease
    When I run "versionator output file emit-python --prerelease='beta'"
    Then the output should contain '__version__ = "2.3.4-beta"'

  Scenario: Emit file with metadata
    When I run "versionator output file emit-python --metadata='build.123'"
    Then the output should contain '__version__ = "2.3.4+build.123"'

  Scenario Outline: Emit file for all supported languages
    When I run "versionator output file <plugin>"
    Then the exit code should be 0
    And the output should contain "<expected>"

    Examples:
      | plugin          | expected                         |
      | emit-python     | __version__                      |
      | emit-json       | "version"                        |
      | emit-yaml       | version:                         |
      | emit-go         | Version     =                    |
      | emit-c          | #define VERSION                  |
      | emit-cpp        | namespace version                |
      | emit-javascript | export const VERSION             |
      | emit-typescript | export const VERSION: string     |
      | emit-java       | public static final String       |
      | emit-kotlin     | const val VERSION                |
      | emit-csharp     | public const string Version      |
      | emit-php        | const VERSION                    |
      | emit-swift      | public static let version        |
      | emit-ruby       | VERSION =                        |
      | emit-rust       | pub const VERSION                |
