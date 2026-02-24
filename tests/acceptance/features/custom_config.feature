Feature: Custom Variables and Configuration
  As a software developer
  I want to manage custom variables and configuration
  So that I can use project-specific values in version templates

  Background:
    Given a clean git repository
    And a VERSION file with version "1.0.0"

  # Custom variable management
  Scenario: Set custom variable
    When I run "versionator custom set AppName MyApplication"
    Then the exit code should be 0

  Scenario: Get custom variable
    When I run "versionator custom set BuildEnv production"
    And I run "versionator custom get BuildEnv"
    Then the output should contain "production"

  Scenario: List custom variables
    When I run "versionator custom set Var1 value1"
    And I run "versionator custom set Var2 value2"
    And I run "versionator custom list"
    Then the output should contain "Var1"
    And the output should contain "Var2"

  Scenario: Delete custom variable
    When I run "versionator custom set TempVar tempvalue"
    And I run "versionator custom delete TempVar"
    Then the exit code should be 0

  Scenario: Use custom variable in template
    When I run "versionator custom set Environment staging"
    And I run "versionator version -t '{{MajorMinorPatch}}-{{Environment}}'"
    Then the output should be "1.0.0-staging"

  Scenario: Override custom variable with --set flag
    When I run "versionator custom set Env dev"
    And I run "versionator version -t '{{MajorMinorPatch}}-{{Env}}' --set Env=prod"
    Then the output should be "1.0.0-prod"

  Scenario: Multiple --set flags
    When I run "versionator version -t '{{Var1}}-{{Var2}}' --set Var1=hello --set Var2=world"
    Then the output should be "hello-world"

  # Config dump
  Scenario: Dump default config to stdout
    When I run "versionator config dump"
    Then the exit code should be 0
    And the output should contain "prefix:"
    And the output should contain "prerelease:"
    And the output should contain "metadata:"
    And the output should contain "logging:"

  Scenario: Dump config to file
    When I run "versionator config dump --output .versionator.yaml"
    Then the exit code should be 0
    And the file ".versionator.yaml" should exist
    And the file ".versionator.yaml" should contain "prefix:"

  # Vars command
  Scenario: Show template variables
    When I run "versionator vars"
    Then the exit code should be 0
    And the output should contain "Major"
    And the output should contain "Minor"
    And the output should contain "Patch"

  Scenario: Vars shows VCS info
    When I run "versionator vars"
    Then the output should contain "ShortHash"
    And the output should contain "BranchName"

  Scenario: Vars shows build timestamps
    When I run "versionator vars"
    Then the output should contain "BuildDateTimeUTC"
    And the output should contain "BuildYear"
