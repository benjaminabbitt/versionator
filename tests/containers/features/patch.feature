@slow
Feature: Patch Manifest Files
  Versionator patches version fields in language-specific manifest files.

  Scenario: Python pyproject.toml patching works
    When I run the "python-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: JavaScript package.json patching works
    When I run the "js-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: TypeScript package.json patching works
    When I run the "ts-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Rust Cargo.toml patching works
    When I run the "rust-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C# csproj patching works
    When I run the "csharp-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: PHP composer.json patching works
    When I run the "php-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Swift Package.swift patching works
    When I run the "swift-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Ruby gemspec patching works
    When I run the "ruby-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Java Maven pom.xml patching works
    When I run the "java-maven-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Java Gradle build.gradle patching works
    When I run the "java-gradle-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Python setuptools setup.py patching works
    When I run the "python-setuptools-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Kotlin build.gradle.kts patching works
    When I run the "kotlin-patch" container test
    Then the container should exit successfully
    And the output should contain "Patched 1 file(s) to version 1.2.3"
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="
