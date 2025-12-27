@slow
Feature: Emit Version Files
  Versionator generates version source files that compile and run correctly
  in each supported language.

  Scenario: Go emit generates valid Go code
    When I run the "go-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Python emit generates valid Python code
    When I run the "python-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Rust emit generates valid Rust code
    When I run the "rust-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: JavaScript emit generates valid JavaScript code
    When I run the "js-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: TypeScript emit generates valid TypeScript code
    When I run the "ts-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Java emit generates valid Java code
    When I run the "java-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C# emit generates valid C# code
    When I run the "csharp-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: PHP emit generates valid PHP code
    When I run the "php-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C emit generates valid C code
    When I run the "c-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: C++ emit generates valid C++ code
    When I run the "cpp-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Swift emit generates valid Swift code
    When I run the "swift-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Ruby emit generates valid Ruby code
    When I run the "ruby-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Java Gradle emit generates valid Java code
    When I run the "java-gradle-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Python setuptools emit generates valid Python code
    When I run the "python-setuptools-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Java Maven emit generates valid Java code
    When I run the "java-maven-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="

  Scenario: Kotlin emit generates valid Kotlin code
    When I run the "kotlin-emit" container test
    Then the container should exit successfully
    And the output should contain "Version: 1.2.3"
    And the output should contain "=== PASS ==="
