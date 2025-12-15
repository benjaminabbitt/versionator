Feature: End-to-End Integration
  As a software developer
  I want versionator to work seamlessly in my build workflow
  So that I can automate version management in CI/CD pipelines

  Background:
    Given a clean git repository
    And versionator is installed

  Scenario: Complete release workflow
    Given a VERSION file with prefix "v" and version "1.0.0"
    And a committed file "README.md" with content "# My Project"
    When I run "versionator patch increment"
    And I commit the VERSION changes
    And I run "versionator commit -m 'Release v1.0.1'"
    Then a git tag "v1.0.1" should exist
    And the VERSION should have version "1.0.1"

  Scenario: Version injection for Python project
    Given a VERSION file with version "1.2.3"
    And a file "mypackage/__init__.py" with content "# placeholder"
    When I run "versionator emit python --output mypackage/_version.py"
    Then the file "mypackage/_version.py" should contain '__version__ = "1.2.3"'

  Scenario: Version injection for Go project
    Given a VERSION file with version "0.5.0"
    When I run "versionator emit go --output version_info.go"
    Then the file "version_info.go" should contain 'Version     = "0.5.0"'

  Scenario: Pre-release workflow
    Given a VERSION file with prefix "v", version "2.0.0" and prerelease "beta"
    And a committed file "README.md" with content "# Beta Release"
    When I run "versionator commit"
    Then a git tag "v2.0.0-beta" should exist

  Scenario: Build metadata workflow
    Given a VERSION file with prefix "v" and version "1.0.0"
    And a committed file "README.md" with content "# Initial commit"
    When I run "versionator version -t '{{MajorMinorPatch}}+{{ShortHash}}' --metadata"
    Then the output should match pattern "1.0.0\+[a-f0-9]{7}"

  Scenario: Configuration file usage
    Given a config file with:
      """
      metadata:
        template: "{{BuildDateUTC}}"
      """
    And a VERSION file with version "3.0.0"
    When I run "versionator version -t '{{Prefix}}{{MajorMinorPatch}}{{MetadataWithPlus}}' --prefix --metadata"
    Then the output should match pattern "v3.0.0\+\d{4}-\d{2}-\d{2}"

  Scenario: Custom variables in template
    Given a VERSION file with version "1.0.0" and custom variable "BuildEnv" set to "production"
    When I run "versionator version -t '{{MajorMinorPatch}}-{{BuildEnv}}'"
    Then the output should be "1.0.0-production"

  Scenario: Override custom variables via CLI
    Given a VERSION file with version "1.0.0" and custom variable "Env" set to "dev"
    When I run "versionator version -t '{{MajorMinorPatch}}-{{Env}}' --set Env=staging"
    Then the output should be "1.0.0-staging"

  @slow
  Scenario: Multi-commit release cycle
    Given a VERSION file with prefix "v" and version "1.0.0"
    And a committed file "file1.txt" with content "Initial"
    When I run "versionator commit -m 'Initial release'"
    And I create 5 commits with message prefix "feat:"
    And I run "versionator minor increment"
    And I commit the VERSION changes
    And I run "versionator commit -m 'Feature release'"
    Then a git tag "v1.0.0" should exist
    And a git tag "v1.1.0" should exist
    And the tag "v1.1.0" should be 6 commits ahead of "v1.0.0"
