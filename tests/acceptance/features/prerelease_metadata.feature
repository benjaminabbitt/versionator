Feature: Pre-release and Metadata Management
  As a software developer
  I want to manage pre-release identifiers and build metadata
  So that I can follow SemVer 2.0.0 specification

  Background:
    Given a clean git repository
    And a VERSION file with version "1.0.0"

  # Pre-release with stable: true
  Scenario: Set prerelease when stable is true
    Given a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease set alpha"
    Then the VERSION should have prerelease "alpha"

  Scenario: Clear prerelease when stable is true
    Given a VERSION file with prefix "v", version "1.0.0" and prerelease "alpha"
    And a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease clear"
    Then the VERSION should have prerelease ""
    And the exit code should be 0

  Scenario: Prerelease status shows stable mode
    Given a VERSION file with prefix "v", version "1.0.0" and prerelease "rc.1"
    And a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease status"
    Then the output should contain "Stable: true"
    And the output should contain "rc.1"

  Scenario: Prerelease status shows dynamic mode
    Given a config file with:
      """
      prerelease:
        template: "build-{{CommitsSinceTag}}"
        stable: false
      """
    When I run "versionator config prerelease status"
    Then the output should contain "Stable: false"
    And the output should contain "Template:"

  Scenario: Set prerelease template
    Given a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease template alpha"
    Then the VERSION should have prerelease "alpha"

  # Metadata with stable: true
  Scenario: Set metadata when stable is true
    Given a config file with:
      """
      metadata:
        stable: true
      """
    When I run "versionator config metadata set build42"
    Then the exit code should be 0

  Scenario: Clear metadata when stable is true
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "build.123"
    And a config file with:
      """
      metadata:
        stable: true
      """
    When I run "versionator config metadata clear"
    Then the VERSION should have metadata ""
    And the exit code should be 0

  Scenario: Metadata status shows stable mode
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "20241212"
    And a config file with:
      """
      metadata:
        stable: true
      """
    When I run "versionator config metadata status"
    Then the output should contain "Stable: true"
    And the output should contain "20241212"

  Scenario: Metadata status shows dynamic mode
    Given a config file with:
      """
      metadata:
        template: "{{ShortHash}}"
        stable: false
      """
    When I run "versionator config metadata status"
    Then the output should contain "Stable: false"
    And the output should contain "Template:"

  Scenario: Set metadata template when stable is true
    Given a config file with:
      """
      metadata:
        stable: true
      """
    When I run "versionator config metadata template build123"
    Then the exit code should be 0

  # Combined scenarios with stable: true
  Scenario: Full SemVer with prerelease and metadata
    Given a config file with:
      """
      prerelease:
        stable: true
      metadata:
        stable: true
      """
    When I run "versionator config prerelease set alpha.1"
    And I run "versionator config metadata set build.42"
    And I run "versionator output version"
    Then the output should be "1.0.0-alpha.1+build.42"

  Scenario: Prerelease with dots separator
    Given a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease set alpha.1.test"
    And I run "versionator output version"
    Then the output should be "1.0.0-alpha.1.test"

  Scenario: Metadata with dots separator
    Given a config file with:
      """
      metadata:
        stable: true
      """
    When I run "versionator config metadata set 2024.12.25.abc1234"
    And I run "versionator output version"
    Then the output should be "1.0.0+2024.12.25.abc1234"
