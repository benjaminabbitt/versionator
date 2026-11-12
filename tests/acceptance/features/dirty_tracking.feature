Feature: Tracked-only dirtiness
  As a developer whose build writes untracked files into its own checkout
  I want untracked files not to make my tree dirty when I say so
  So that {{Dirty}}, {{UncommittedChanges}} and {{DateTimeDirty}} mean what git describe --dirty means

  Background:
    Given a clean git repository
    And a VERSION file with version "1.2.3"
    And I commit the VERSION changes
    And a committed file "README.md" with content "# project"

  Scenario: By default an untracked file makes the tree dirty
    Given a file "scratch.txt" with content "untracked"
    When I run "versionator output version -t '[{{UncommittedChanges}}|{{Dirty}}]'"
    Then the exit code should be 0
    And the output should be "[1|dirty]"

  Scenario: dirty.ignoreUntracked leaves untracked files out
    Given a config file with:
      """
      prefix: ""
      dirty:
        ignoreUntracked: true
      """
    And a file "scratch.txt" with content "untracked"
    When I run "versionator output version -t '[{{UncommittedChanges}}|{{Dirty}}|{{DateTimeDirty}}]'"
    Then the exit code should be 0
    And the output should be "[0||]"

  Scenario: --ignore-untracked leaves untracked files out without config
    Given a file "scratch.txt" with content "untracked"
    When I run "versionator --ignore-untracked output version -t '[{{UncommittedChanges}}|{{Dirty}}|{{DateTimeDirty}}]'"
    Then the exit code should be 0
    And the output should be "[0||]"

  Scenario: A modified tracked file is still dirty
    Given a file "README.md" with content "# changed"
    When I run "versionator --ignore-untracked output version -t '[{{UncommittedChanges}}|{{Dirty}}]'"
    Then the exit code should be 0
    And the output should be "[1|dirty]"

  Scenario: A staged file is still dirty
    Given a file "added.txt" with content "staged"
    And I run "git add added.txt"
    When I run "versionator --ignore-untracked output version -t '[{{UncommittedChanges}}|{{Dirty}}]'"
    Then the exit code should be 0
    And the output should be "[1|dirty]"
