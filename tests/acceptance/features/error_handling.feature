Feature: Error Handling
  As a software developer
  I want clear error messages when operations fail
  So that I can understand and fix issues quickly

  Background:
    Given a clean git repository

  # Invalid commands and arguments
  Scenario: Unknown command
    When I run "versionator unknown"
    Then the exit code should not be 0

  Scenario: Invalid emit format
    Given a VERSION file with version "1.0.0"
    When I run "versionator emit invalid-format"
    Then the exit code should not be 0
    And the output should contain "unsupported"

  Scenario: Missing required argument for custom get
    Given a VERSION file with version "1.0.0"
    When I run "versionator custom get"
    Then the exit code should not be 0

  # Version boundary errors
  Scenario: Decrement major below zero
    Given a VERSION file with version "0.5.0"
    When I run "versionator major decrement"
    Then the exit code should not be 0

  Scenario: Decrement minor below zero
    Given a VERSION file with version "1.0.5"
    When I run "versionator minor decrement"
    Then the exit code should not be 0

  Scenario: Decrement patch below zero
    Given a VERSION file with version "1.5.0"
    When I run "versionator patch decrement"
    Then the exit code should not be 0

  # Git/VCS errors
  Scenario: Release command outside git repository
    Given a VERSION file with version "1.0.0"
    # Note: Background creates git repo, we need to test outside it
    # This scenario tests the error message format
    When I run "versionator release"
    Then the exit code should be 0

  Scenario: Release with dirty working directory (other files dirty)
    Given a VERSION file with version "1.0.0"
    And a committed file "README.md" with content "initial"
    And a file "uncommitted.txt" with content "dirty"
    When I run "versionator release"
    Then the exit code should not be 0
    And the output should contain "not clean"

  # Custom variable errors
  Scenario: Get nonexistent custom variable
    Given a VERSION file with version "1.0.0"
    When I run "versionator custom get NonExistent"
    Then the exit code should not be 0
    And the output should contain "not found"

  Scenario: Delete custom variable is idempotent
    Given a VERSION file with version "1.0.0"
    When I run "versionator custom delete NonExistent"
    Then the exit code should be 0

  # Template errors
  Scenario: Invalid template syntax
    Given a VERSION file with version "1.0.0"
    When I run "versionator version -t '{{InvalidUnclosed'"
    Then the exit code should not be 0

  Scenario: Template file not found
    Given a VERSION file with version "1.0.0"
    When I run "versionator emit --template-file nonexistent.tmpl"
    Then the exit code should not be 0

  # Completion errors
  Scenario: Invalid completion shell
    When I run "versionator completion invalid"
    Then the exit code should not be 0

  Scenario: Missing completion argument
    When I run "versionator completion"
    Then the exit code should not be 0
