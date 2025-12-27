Feature: Emit Build - Build Flag Generation
  As a software developer
  I want to generate build flags for version injection
  So that I can inject version at build time without source files

  Background:
    Given a clean git repository
    And a VERSION file with version "1.2.3"

  # Go ldflags
  Scenario: Emit build flags for Go
    When I run "versionator emit build go"
    Then the exit code should be 0
    And the output should contain "-X"
    And the output should contain "1.2.3"

  Scenario: Emit build flags for Go with custom variable
    When I run "versionator emit build go --var main.Version"
    Then the exit code should be 0
    And the output should contain "-X main.Version=1.2.3"

  Scenario: Emit build flags for Go with multiple variables
    When I run "versionator emit build go --var main.Version --var main.GitHash={{ShortHash}}"
    Then the exit code should be 0
    And the output should contain "-X main.Version=1.2.3"
    And the output should contain "-X main.GitHash="

  # Rust env vars (use with: $(versionator emit build rust) cargo build)
  Scenario: Emit build flags for Rust
    When I run "versionator emit build rust"
    Then the exit code should be 0
    And the output should contain "VERSION=1.2.3"

  Scenario: Emit build flags for Rust with custom variable
    When I run "versionator emit build rust --var MY_VERSION"
    Then the exit code should be 0
    And the output should contain "MY_VERSION=1.2.3"

  # C/C++ defines
  Scenario: Emit build flags for C
    When I run "versionator emit build c"
    Then the exit code should be 0
    And the output should contain "-D"
    And the output should contain "VERSION"

  Scenario: Emit build flags for C with custom macro
    When I run "versionator emit build c --var APP_VERSION"
    Then the exit code should be 0
    And the output should contain "-DAPP_VERSION="

  Scenario: Emit build flags for C++
    When I run "versionator emit build cpp"
    Then the exit code should be 0
    And the output should contain "-D"

  # With prerelease and metadata
  Scenario: Emit build with prerelease
    When I run "versionator emit build go --var main.Version --prerelease='alpha'"
    Then the output should contain "1.2.3-alpha"

  Scenario: Emit build with metadata
    When I run "versionator emit build go --var main.Version --metadata='build.42'"
    Then the output should contain "1.2.3+build.42"

  # Custom template for build flags
  Scenario: Emit build with custom template
    When I run "versionator emit build go --template '-X main.Version={{MajorMinorPatch}} -X main.Commit={{ShortHash}}'"
    Then the exit code should be 0
    And the output should contain "-X main.Version=1.2.3"
    And the output should contain "-X main.Commit="

  # Error handling
  Scenario: Error for language that doesn't support link-time injection
    When I run "versionator emit build python"
    Then the exit code should be 1
    And the output should contain "does not support link-time version injection"

  Scenario: Error for unknown language
    When I run "versionator emit build unknown-lang"
    Then the exit code should be 1
    And the output should contain "is not supported"
