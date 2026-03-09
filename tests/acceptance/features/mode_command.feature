Feature: Versioning Modes
  As a software developer
  I want to configure versioning modes
  So that I can support different release strategies

  Background:
    Given a clean git repository
    And a VERSION file with version "1.2.3"
    And I create a git tag "v1.2.3"
    And I create a commit with message "feat: new feature"

  Scenario: Default mode is release
    When I run "versionator config mode"
    Then the output should contain "release"
    And the exit code should be 0

  Scenario: Set continuous-delivery mode
    When I run "versionator config mode cd"
    Then the output should contain "continuous-delivery"
    And the exit code should be 0

  Scenario: Set release mode explicitly
    When I run "versionator config mode cd"
    And I run "versionator config mode release"
    Then the output should contain "release"
    And the exit code should be 0

  Scenario: CD mode with custom prerelease template
    When I run "versionator config mode cd --prerelease 'dev-{{CommitsSinceTag}}'"
    And I run "versionator output ci --format=shell"
    Then the output should contain 'VERSION_PRERELEASE="dev-1"'

  Scenario: CD mode with custom metadata template
    When I run "versionator config mode cd --metadata '{{ShortHash}}'"
    And I run "versionator output ci --format=shell"
    Then the output should contain "VERSION_METADATA="

  Scenario: CD mode shows templates in status
    When I run "versionator config mode cd --prerelease 'build-{{CommitsSinceTag}}' --metadata '{{ShortHash}}'"
    And I run "versionator config mode"
    Then the output should contain "continuous-delivery"
    And the output should contain "Pre-release template"
    And the output should contain "Metadata template"

  Scenario: CD mode generates unique version for CI
    When I run "versionator config mode cd"
    And I run "versionator output ci --format=shell"
    Then the output should contain "VERSION_PRERELEASE="
    And the output should contain "VERSION_METADATA="

  Scenario: Release mode uses VERSION file values
    Given a VERSION file with prefix "v", version "1.2.3" and prerelease "alpha.1"
    When I run "versionator config mode release"
    And I run "versionator output ci --format=shell"
    Then the output should contain 'VERSION_PRERELEASE="alpha.1"'

  Scenario: Invalid template shows error
    When I run "versionator config mode cd --prerelease '{{Invalid'"
    Then the exit code should not be 0
