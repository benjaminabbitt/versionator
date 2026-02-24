Feature: Shell Completion Generation
  As a CLI user
  I want shell completion scripts
  So that I can tab-complete commands

  Scenario: Generate bash completion
    When I run "versionator completion bash"
    Then the exit code should be 0
    And the output should contain "__start_versionator"
    And the output should contain "versionator"

  Scenario: Generate zsh completion
    When I run "versionator completion zsh"
    Then the exit code should be 0
    And the output should contain "#compdef versionator"

  Scenario: Generate fish completion
    When I run "versionator completion fish"
    Then the exit code should be 0
    And the output should contain "complete -c versionator"

  Scenario: Generate powershell completion
    When I run "versionator completion powershell"
    Then the exit code should be 0
    And the output should contain "Register-ArgumentCompleter"
    And the output should contain "versionator"

  Scenario: Invalid shell returns error
    When I run "versionator completion invalid"
    Then the exit code should not be 0

  Scenario: No shell argument returns error
    When I run "versionator completion"
    Then the exit code should not be 0
