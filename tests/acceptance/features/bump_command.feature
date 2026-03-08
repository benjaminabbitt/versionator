Feature: Auto-bump version based on commit messages
  As a software developer
  I want to automatically bump versions based on commits
  So that I can follow semantic versioning without manual work

  Background:
    Given a clean git repository
    And a VERSION file with version "1.0.0"
    And I create a git tag "v1.0.0"

  Scenario: Bump minor version from feat: commit
    When I create a commit with message "feat: add new feature"
    And I run "versionator bump"
    Then the VERSION should have version "1.1.0"

  Scenario: Bump patch version from fix: commit
    When I create a commit with message "fix: resolve bug"
    And I run "versionator bump"
    Then the VERSION should have version "1.0.1"

  Scenario: Bump major version from breaking change
    When I create a commit with message "feat!: breaking change"
    And I run "versionator bump"
    Then the VERSION should have version "2.0.0"

  Scenario: Bump with +semver:major marker
    When I create a commit with message "chore: stuff +semver:major"
    And I run "versionator bump"
    Then the VERSION should have version "2.0.0"

  Scenario: Bump with +semver:minor marker
    When I create a commit with message "chore: stuff +semver:minor"
    And I run "versionator bump"
    Then the VERSION should have version "1.1.0"

  Scenario: Bump with +semver:patch marker
    When I create a commit with message "chore: stuff +semver:patch"
    And I run "versionator bump"
    Then the VERSION should have version "1.0.1"

  Scenario: Skip bump with +semver:skip
    When I create a commit with message "feat: new feature +semver:skip"
    And I run "versionator bump"
    Then the VERSION should have version "1.0.0"

  Scenario: Highest level wins with multiple commits
    When I create a commit with message "fix: bug fix"
    And I create a commit with message "feat: new feature"
    And I create a commit with message "fix: another fix"
    And I run "versionator bump"
    Then the VERSION should have version "1.1.0"

  Scenario: No bump for non-semantic commits
    When I create a commit with message "chore: cleanup code"
    And I run "versionator bump"
    Then the VERSION should have version "1.0.0"

  Scenario: Dry run shows what would happen
    When I create a commit with message "feat: new feature"
    And I run "versionator bump --dry-run"
    Then the output should contain "Would bump"
    And the output should contain "minor"
    And the VERSION should have version "1.0.0"

  Scenario: Bump respects semver-only mode
    When I create a commit with message "feat: should be ignored"
    And I create a commit with message "chore: stuff +semver:patch"
    And I run "versionator bump --mode=semver"
    Then the VERSION should have version "1.0.1"

  Scenario: Bump respects conventional-only mode
    When I create a commit with message "chore: stuff +semver:major"
    And I create a commit with message "fix: bug fix"
    And I run "versionator bump --mode=conventional"
    Then the VERSION should have version "1.0.1"
