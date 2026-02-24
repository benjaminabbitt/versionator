Feature: CLI Schema Generation
  As an AI assistant or tooling developer
  I want machine-readable command documentation
  So that I can understand available CLI options

  Scenario: Generate JSON schema
    When I run "versionator schema"
    Then the exit code should be 0
    And the output should contain "\"name\": \"versionator\""
    And the output should contain "\"commands\":"
    And the output should contain "\"templateVariables\":"

  Scenario: Schema includes version command
    When I run "versionator schema"
    Then the exit code should be 0
    And the output should contain "\"name\": \"version\""

  Scenario: Schema includes global flags
    When I run "versionator schema"
    Then the exit code should be 0
    And the output should contain "\"globalFlags\":"
    And the output should contain "\"log-format\""

  Scenario: Schema includes subcommands
    When I run "versionator schema"
    Then the exit code should be 0
    And the output should contain "\"subcommands\":"
    And the output should contain "\"name\": \"increment\""

  Scenario: Schema includes template variables
    When I run "versionator schema"
    Then the exit code should be 0
    And the output should contain "\"versionComponents\":"
    And the output should contain "\"Major\""
    And the output should contain "\"vcs\":"
    And the output should contain "\"ShortHash\""

  Scenario: Write schema to file
    When I run "versionator schema --output schema.json"
    Then the exit code should be 0
    And the file "schema.json" should exist
    And the file "schema.json" should contain "\"name\": \"versionator\""
