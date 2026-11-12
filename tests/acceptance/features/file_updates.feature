Feature: Structured File Updates
  As a software developer
  I want versionator to update version fields in my project files
  So that all my manifest files stay in sync with the VERSION file

  Background:
    Given a clean git repository
    And a VERSION file with prefix "v" and version "1.0.0"
    And a committed file "README.md" with content "# My Project"

  Scenario: Update package.json version on release
    Given a file "package.json" with content:
      """
      {
        "name": "myapp",
        "version": "0.0.1"
      }
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: package.json
          path: version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then the exit code should be 0
    And a git tag "v1.0.0" should exist
    And the file "package.json" should contain "1.0.0"

  Scenario: Update Cargo.toml version on release
    Given a file "Cargo.toml" with content:
      """
      [package]
      name = "myapp"
      version = "0.0.1"
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: Cargo.toml
          path: package.version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then the exit code should be 0
    And a git tag "v1.0.0" should exist
    And the file "Cargo.toml" should contain "1.0.0"

  Scenario: Update Helm Chart.yaml version and appVersion
    Given a file "Chart.yaml" with content:
      """
      apiVersion: v2
      name: myapp
      version: 0.0.1
      appVersion: 0.0.1
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: Chart.yaml
          path: version
          template: "{{MajorMinorPatch}}"
        - file: Chart.yaml
          path: appVersion
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then the exit code should be 0
    And a git tag "v1.0.0" should exist
    And the file "Chart.yaml" should contain "version: 1.0.0"
    And the file "Chart.yaml" should contain "appVersion: 1.0.0"

  Scenario: Update multiple files on release
    Given a file "package.json" with content:
      """
      {"version": "0.0.1"}
      """
    Given a file "Chart.yaml" with content:
      """
      version: 0.0.1
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: package.json
          path: version
          template: "{{MajorMinorPatch}}"
        - file: Chart.yaml
          path: version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then the exit code should be 0
    And the file "package.json" should contain "1.0.0"
    And the file "Chart.yaml" should contain "1.0.0"
    And the output should contain "Updated 2 file(s)"

  Scenario: Updated files are committed with VERSION
    Given a file "package.json" with content:
      """
      {"version": "0.0.1"}
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: package.json
          path: version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator bump patch increment"
    And I run "versionator release"
    Then the exit code should be 0
    And a git tag "v1.0.1" should exist
    And the output should contain "Committed:"
    And the output should contain "package.json"

  Scenario: Release fails if update file not found
    And a config file with:
      """
      prefix: v
      updates:
        - file: nonexistent.json
          path: version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator release"
    Then the exit code should not be 0
    And the output should contain "file not found"

  Scenario: set keeps a Cargo workspace version in sync, byte for byte
    Given a file "Cargo.toml" with content:
      """
      [workspace]
      members = ["crates/*"]

      [workspace.dependencies]
      core-lib = { path = "crates/core-lib", version = "1.0.0" }

      [workspace.package]
      version = "1.0.0" # stamped
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: Cargo.toml
          path: workspace.package.version
          template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}"
      """
    When I run "versionator set 0.7.0-pre1"
    Then the exit code should be 0
    And the file "Cargo.toml" should contain 'version = "0.7.0-pre1" # stamped'
    And the file "Cargo.toml" should contain 'core-lib = { path = "crates/core-lib", version = "1.0.0" }'

  Scenario: bump keeps a crate's package version in sync
    Given a file "Cargo.toml" with content:
      """
      [package]
      name = "app"
      version = "1.0.0"
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: Cargo.toml
          path: package.version
          template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}"
      """
    When I run "versionator bump minor increment"
    Then the exit code should be 0
    And the file "Cargo.toml" should contain 'version = "1.1.0"'

  Scenario: A four-part version is refused for Cargo.toml
    Given a file "Cargo.toml" with content:
      """
      [package]
      name = "app"
      version = "1.0.0"
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: Cargo.toml
          path: package.version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator set 1.2.3.4"
    # The refusal is said on stderr (unit-tested); stdout reports the VERSION write.
    Then the exit code should not be 0
    And the file "Cargo.toml" should contain 'version = "1.0.0"'

  Scenario: A Cargo.toml without the configured version key is an error
    Given a file "Cargo.toml" with content:
      """
      [package]
      name = "app"
      """
    And a config file with:
      """
      prefix: v
      updates:
        - file: Cargo.toml
          path: package.version
          template: "{{MajorMinorPatch}}"
      """
    When I run "versionator set 1.2.3"
    Then the exit code should not be 0
    And the file "Cargo.toml" should contain 'name = "app"'
