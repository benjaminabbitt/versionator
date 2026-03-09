Feature: CI/CD Output Formats
  As a CI/CD pipeline
  I want to export version variables in platform-specific formats
  So that I can use them in build and deployment steps

  Background:
    Given a clean git repository
    And a VERSION file with prefix "v" and version "2.5.3"

  Scenario: Shell format output
    When I run "versionator output ci --format=shell"
    Then the exit code should be 0
    And the output should contain 'export VERSION="v2.5.3"'
    And the output should contain 'export VERSION_MAJOR="2"'
    And the output should contain 'export VERSION_MINOR="5"'
    And the output should contain 'export VERSION_PATCH="3"'

  Scenario: GitLab CI format output
    When I run "versionator output ci --format=gitlab"
    Then the exit code should be 0
    And the output should contain "VERSION=v2.5.3"
    And the output should contain "VERSION_MAJOR=2"

  Scenario: Azure DevOps format output
    When I run "versionator output ci --format=azure"
    Then the exit code should be 0
    And the output should contain "##vso[task.setvariable variable=VERSION]v2.5.3"

  Scenario: Jenkins format output
    When I run "versionator output ci --format=jenkins"
    Then the exit code should be 0
    And the output should contain "VERSION=v2.5.3"

  Scenario: GitHub Actions format output
    When I run "versionator output ci --format=github"
    Then the exit code should be 0
    And the output should contain "VERSION=v2.5.3"

  Scenario: CircleCI format output
    When I run "versionator output ci --format=circleci"
    Then the exit code should be 0
    And the output should contain 'export VERSION="v2.5.3"'

  Scenario: Variable prefix
    When I run "versionator output ci --format=shell --prefix=MYAPP_"
    Then the exit code should be 0
    And the output should contain 'export MYAPP_VERSION="v2.5.3"'
    And the output should contain 'export MYAPP_VERSION_MAJOR="2"'

  Scenario: Output to file
    When I run "versionator output ci --format=shell --output=version.env"
    Then the exit code should be 0
    And the file "version.env" should exist
    And the file "version.env" should contain 'export VERSION="v2.5.3"'

  Scenario: CI includes git info
    When I run "versionator output ci --format=shell"
    Then the output should contain "GIT_SHA="
    And the output should contain "GIT_SHA_SHORT="
    And the output should contain "GIT_BRANCH="

  Scenario: CI includes build number
    Given I create a git tag "v2.5.3"
    When I create a commit with message "feat: new commit"
    And I run "versionator output ci --format=shell"
    Then the output should contain 'BUILD_NUMBER="1"'

  Scenario: CI with prerelease version
    Given a VERSION file with prefix "v", version "2.5.3" and prerelease "alpha.1"
    When I run "versionator output ci --format=shell"
    Then the output should contain 'VERSION_PRERELEASE="alpha.1"'

  Scenario: CI with metadata version
    Given a VERSION file with prefix "v", version "2.5.3" and metadata "build.42"
    When I run "versionator output ci --format=shell"
    Then the output should contain 'VERSION_METADATA="build.42"'
