Feature: Stability Settings
  As a software developer
  I want to control whether pre-release and metadata are stored in VERSION or generated dynamically
  So that I can support both release and continuous-delivery workflows

  Background:
    Given a clean git repository
    And a VERSION file with version "1.0.0"
    And I create a git tag "v1.0.0"
    And I create a commit with message "feat: new feature"

  # Default behavior (stable: false)
  Scenario: Default stability is false for both prerelease and metadata
    When I run "versionator config prerelease stable"
    Then the output should contain "false"
    When I run "versionator config metadata stable"
    Then the output should contain "false"

  Scenario: Default templates are applied when stable is false
    Given a config file with:
      """
      prerelease:
        template: "alpha"
        stable: false
      metadata:
        template: "build42"
        stable: false
      """
    When I run "versionator output ci --format=shell"
    Then the output should contain 'VERSION_PRERELEASE="alpha"'
    And the output should contain 'VERSION_METADATA="build42"'

  # Setting stability
  Scenario: Set prerelease stability to true
    When I run "versionator config prerelease stable true"
    Then the exit code should be 0
    When I run "versionator config prerelease stable"
    Then the output should contain "true"

  Scenario: Set metadata stability to true
    When I run "versionator config metadata stable true"
    Then the exit code should be 0
    When I run "versionator config metadata stable"
    Then the output should contain "true"

  # Set command behavior with stability
  Scenario: Set prerelease succeeds when stable is true
    Given a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease set alpha"
    Then the exit code should be 0
    And the VERSION should have prerelease "alpha"

  Scenario: Set prerelease fails when stable is false
    Given a config file with:
      """
      prerelease:
        template: "dynamic"
        stable: false
      """
    When I run "versionator config prerelease set alpha"
    Then the exit code should not be 0
    And the output should contain "stable: false"

  Scenario: Set prerelease with force sets template when stable is false
    Given a config file with:
      """
      prerelease:
        template: "old"
        stable: false
      """
    When I run "versionator config prerelease set newvalue --force"
    Then the exit code should be 0
    When I run "versionator config prerelease template"
    Then the output should contain "newvalue"

  Scenario: Set metadata succeeds when stable is true
    Given a config file with:
      """
      metadata:
        stable: true
      """
    When I run "versionator config metadata set build123"
    Then the exit code should be 0
    And the VERSION should have metadata "build123"

  Scenario: Set metadata fails when stable is false
    Given a config file with:
      """
      metadata:
        template: "dynamic"
        stable: false
      """
    When I run "versionator config metadata set build123"
    Then the exit code should not be 0
    And the output should contain "stable: false"

  Scenario: Set metadata with force sets template when stable is false
    Given a config file with:
      """
      metadata:
        template: "old"
        stable: false
      """
    When I run "versionator config metadata set newvalue --force"
    Then the exit code should be 0
    When I run "versionator config metadata template"
    Then the output should contain "newvalue"

  # Clear command behavior
  Scenario: Clear prerelease succeeds when stable is true
    Given a VERSION file with prefix "", version "1.0.0" and prerelease "alpha"
    And a config file with:
      """
      prerelease:
        stable: true
      """
    When I run "versionator config prerelease clear"
    Then the exit code should be 0
    And the VERSION should have prerelease ""

  Scenario: Clear prerelease fails when stable is false
    Given a config file with:
      """
      prerelease:
        template: "dynamic"
        stable: false
      """
    When I run "versionator config prerelease clear"
    Then the exit code should not be 0

  # Emit command with stability
  Scenario: Emit uses templates when stable is false
    Given a config file with:
      """
      prerelease:
        template: "dev"
        stable: false
      metadata:
        template: "hash123"
        stable: false
      """
    When I run "versionator output emit --template '{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}'"
    Then the output should contain "1.0.0-dev+hash123"

  Scenario: Emit uses VERSION file when stable is true
    Given a VERSION file with prefix "", version "1.0.0" and prerelease "rc.1"
    And a config file with:
      """
      prerelease:
        template: "ignored"
        stable: true
      metadata:
        template: "ignored"
        stable: true
      """
    When I run "versionator output emit --template '{{MajorMinorPatch}}{{PreReleaseWithDash}}'"
    Then the output should contain "1.0.0-rc.1"

  # Mixed stability
  Scenario: Mixed stability - prerelease stable, metadata dynamic
    Given a VERSION file with prefix "", version "2.0.0" and prerelease "beta"
    And a config file with:
      """
      prerelease:
        stable: true
      metadata:
        template: "dynamic-meta"
        stable: false
      """
    When I run "versionator output emit --template '{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}'"
    Then the output should contain "2.0.0-beta+dynamic-meta"
