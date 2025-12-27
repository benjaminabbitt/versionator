@slow
Feature: Link-time Version Injection
  Versionator provides build flags for link-time version injection
  in compiled languages.

  Scenario: Go link injection works
    When I run the "go-link" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Rust link injection works
    When I run the "rust-link" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C link injection works
    When I run the "c-link" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C++ link injection works
    When I run the "cpp-link" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C# link injection works
    When I run the "csharp-link" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="
