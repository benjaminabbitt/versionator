# Acceptance Tests

This directory contains Cucumber/Gherkin acceptance tests for versionator using [godog](https://github.com/cucumber/godog).

## Structure

```
tests/acceptance/
├── features/              # Gherkin feature files
│   ├── version_management.feature
│   ├── git_tagging.feature
│   ├── version_emission.feature
│   └── integration.feature
├── acceptance_test.go     # Go step definitions
├── Dockerfile             # Docker image for isolated testing
├── docker-compose.yml     # Docker Compose configuration
└── README.md              # This file
```

## Running Tests

### Local (requires versionator in PATH or built)

```bash
# Build versionator first
go build -o versionator .

# Run acceptance tests
cd tests/acceptance
VERSIONATOR_PROJECT_ROOT=$(pwd)/../.. go test -v ./...
```

### Using Docker (recommended for CI)

```bash
# Run from project root
docker-compose -f tests/acceptance/docker-compose.yml up --build

# Or run slow tests
docker-compose -f tests/acceptance/docker-compose.yml run acceptance-tests-slow
```

### Using just

```bash
# Run all acceptance tests
just acceptance-test

# Run acceptance tests in Docker
just acceptance-test-docker
```

## Writing Tests

### Feature Files

Feature files use Gherkin syntax. Example:

```gherkin
Feature: Version Management
  Scenario: Increment major version
    Given a clean git repository
    And a VERSION.json file with version "1.0.0"
    When I run "versionator major"
    And I run "versionator version"
    Then the output should be "2.0.0"
```

### Available Steps

**Background/Setup:**
- `Given a clean git repository`
- `Given versionator is installed`
- `Given a VERSION.json file with version "<version>"`
- `Given a VERSION.json file with prefix "<prefix>" and version "<version>"`
- `Given a committed file "<filename>" with content "<content>"`
- `Given a config file with: <docstring>`

**Actions:**
- `When I run "<command>"`
- `When I commit a file "<filename>" with content "<content>"`
- `When I create <n> commits with message prefix "<prefix>"`

**Assertions:**
- `Then the output should be "<expected>"`
- `Then the output should contain "<substring>"`
- `Then the output should match pattern "<regex>"`
- `Then the exit code should be <code>`
- `Then a git tag "<tag>" should exist`
- `Then the tag "<tag>" should point to HEAD`
- `Then the VERSION.json should have version "<version>"`
- `Then the file "<filename>" should exist`
- `Then the file "<filename>" should contain "<substring>"`

### Slow Tests

Tag scenarios with `@slow` to mark them as slow tests:

```gherkin
@slow
Scenario: Multi-commit release cycle
  ...
```

Slow tests are skipped by default. Run them with:

```bash
go test -v -run TestSlowFeatures ./tests/acceptance/...
```

## Test Isolation

Each scenario runs in an isolated temporary directory with:
- A fresh git repository
- Isolated file system
- No state shared between scenarios

This ensures tests are deterministic and can run in parallel.
