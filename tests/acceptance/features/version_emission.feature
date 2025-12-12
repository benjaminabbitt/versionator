Feature: Version Emission
  As a software developer
  I want to emit version information in various formats
  So that I can inject version data into my source code

  Background:
    Given a clean git repository
    And a VERSION.json file with prefix "v" and version "2.3.4"

  Scenario: Emit Python format
    When I run "versionator emit python"
    Then the output should contain '__version__ = "2.3.4"'
    And the output should contain '__version_tuple__ = (2, 3, 4)'

  Scenario: Emit JSON format
    When I run "versionator emit json"
    Then the output should contain '"version": "2.3.4"'
    And the output should contain '"major": 2'

  Scenario: Emit Go format
    When I run "versionator emit go"
    Then the output should contain 'Version     = "2.3.4"'
    And the output should contain 'Major       = 2'

  Scenario: Emit to file
    When I run "versionator emit python --output _version.py"
    Then the file "_version.py" should exist
    And the file "_version.py" should contain '__version__ = "2.3.4"'

  Scenario: Emit with prerelease template
    When I run "versionator emit python --prerelease='alpha'"
    Then the output should contain '__version__ = "2.3.4-alpha"'

  Scenario: Emit with metadata template
    When I run "versionator emit python --metadata='build.42'"
    Then the output should contain '__version__ = "2.3.4+build.42"'

  Scenario: Emit custom template
    When I run "versionator emit -t 'v{{Major}}.{{Minor}}.{{Patch}}'"
    Then the output should be "v2.3.4"

  Scenario: Emit dump template
    When I run "versionator emit dump python"
    Then the output should contain '{{MajorMinorPatch}}'
    And the exit code should be 0

  Scenario: Emit dump to file
    When I run "versionator emit dump python --output _version.tmpl.py"
    Then the file "_version.tmpl.py" should exist
    And the file "_version.tmpl.py" should contain '{{MajorMinorPatch}}'

  Scenario: Emit with template file
    Given a template file "custom.tmpl" with content "Version: {{MajorMinorPatch}}"
    When I run "versionator emit --template-file custom.tmpl"
    Then the output should be "Version: 2.3.4"

  Scenario Outline: Emit all supported formats
    When I run "versionator emit <format>"
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
