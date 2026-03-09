Feature: Pre-release and Metadata Management
  As a software developer
  I want to manage pre-release identifiers and build metadata
  So that I can follow SemVer 2.0.0 specification

  Background:
    Given a clean git repository
    And a VERSION file with version "1.0.0"

  # Pre-release enable/disable/status
  Scenario: Enable prerelease from config template
    Given a config file with:
      """
      prerelease:
        template: "beta"
      """
    When I run "versionator config prerelease enable"
    Then the VERSION should have prerelease "beta"

  Scenario: Disable prerelease preserves config
    Given a VERSION file with prefix "v", version "1.0.0" and prerelease "alpha"
    When I run "versionator config prerelease disable"
    Then the VERSION should have prerelease ""
    And the exit code should be 0

  Scenario: Prerelease status when enabled
    Given a VERSION file with prefix "v", version "1.0.0" and prerelease "rc.1"
    When I run "versionator config prerelease status"
    Then the output should contain "ENABLED"
    And the output should contain "rc.1"

  Scenario: Prerelease status when disabled
    When I run "versionator config prerelease status"
    Then the output should contain "DISABLED"

  Scenario: Set prerelease template in config
    When I run "versionator config prerelease template alpha"
    And I run "versionator config prerelease enable"
    Then the VERSION should have prerelease "alpha"

  # Metadata enable/disable/status
  Scenario: Enable metadata from config template
    Given a config file with:
      """
      metadata:
        template: "build42"
      """
    When I run "versionator config metadata enable"
    Then the exit code should be 0

  Scenario: Disable metadata preserves config
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "build.123"
    When I run "versionator config metadata disable"
    Then the VERSION should have metadata ""
    And the exit code should be 0

  Scenario: Metadata status when enabled
    Given a VERSION file with prefix "v", version "1.0.0" and metadata "20241212"
    When I run "versionator config metadata status"
    Then the output should contain "ENABLED"
    And the output should contain "20241212"

  Scenario: Metadata status when disabled
    When I run "versionator config metadata status"
    Then the output should contain "DISABLED"

  Scenario: Set metadata template in config
    When I run "versionator config metadata template build123"
    Then the exit code should be 0

  # Combined scenarios
  Scenario: Full SemVer with prerelease and metadata
    When I run "versionator config prerelease set alpha.1"
    And I run "versionator config metadata set build.42"
    And I run "versionator output version"
    Then the output should be "1.0.0-alpha.1+build.42"

  Scenario: Prerelease with dots separator
    When I run "versionator config prerelease set alpha.1.test"
    And I run "versionator output version"
    Then the output should be "1.0.0-alpha.1.test"

  Scenario: Metadata with dots separator
    When I run "versionator config metadata set 2024.12.25.abc1234"
    And I run "versionator output version"
    Then the output should be "1.0.0+2024.12.25.abc1234"
