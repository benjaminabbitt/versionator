Feature: Version Emission
  As a software developer
  I want to emit version information in various formats
  So that I can inject version data into my source code

  Background:
    Given a clean git repository
    And a VERSION file with prefix "v" and version "2.3.4"

  Scenario: Emit Python format
    When I run "versionator output file emit-python"
    Then the output should contain '__version__ = "2.3.4"'
    And the output should contain '__version_tuple__ = (2, 3, 4)'

  Scenario: Emit JSON format
    When I run "versionator output file emit-json"
    Then the output should contain '"version": "2.3.4"'
    And the output should contain '"major": 2'

  Scenario: Emit Go format
    When I run "versionator output file emit-go"
    Then the output should contain 'Version     = "2.3.4"'
    And the output should contain 'Major       = 2'

  Scenario: Emit to file
    When I run "versionator output file emit-python --output _version.py"
    Then the file "_version.py" should exist
    And the file "_version.py" should contain '__version__ = "2.3.4"'

  Scenario: Emit with prerelease template
    When I run "versionator output file emit-python --prerelease='alpha'"
    Then the output should contain '__version__ = "2.3.4-alpha"'

  Scenario: Emit with metadata template
    When I run "versionator output file emit-python --metadata='build.42'"
    Then the output should contain '__version__ = "2.3.4+build.42"'

  Scenario: Version with custom template
    When I run "versionator version -t 'v{{Major}}.{{Minor}}.{{Patch}}'"
    Then the output should be "v2.3.4"

  Scenario Outline: Emit all supported formats
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
