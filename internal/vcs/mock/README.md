# VCS Mock for Testing

This directory contains a mock implementation of the `VersionControlSystem` interface using [gomock](https://github.com/golang/mock/gomock).

## Overview

The mock VCS allows you to test code that depends on version control operations without requiring an actual VCS repository. This is particularly useful for:

- Unit testing components that interact with VCS
- Testing error scenarios that are difficult to reproduce with real VCS
- Testing in CI/CD environments where VCS setup might be complex
- Isolating tests from external dependencies

## Files

- `mock_vcs.go` - Generated mock implementation of the `VersionControlSystem` interface
- `README.md` - This documentation file

## Usage

### Basic Setup

```go
import (
    "testing"
    "github.com/golang/mock/gomock"
    "versionator/internal/vcs/mock"
)

func TestYourFunction(t *testing.T) {
    // Create a new gomock controller
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Create a mock VCS instance
    mockVCS := mocks.NewMockVersionControlSystem(ctrl)

    // Set up expectations
    mockVCS.EXPECT().Name().Return("mock-git").AnyTimes()
    mockVCS.EXPECT().IsRepository().Return(true).Times(1)

    // Use the mock in your test
    // ... your test code here
}
```

### Common Test Scenarios

#### Testing Repository Detection

```go
// Mock a VCS that detects it's in a repository
mockVCS.EXPECT().IsRepository().Return(true)

// Mock a VCS that's not in a repository
mockVCS.EXPECT().IsRepository().Return(false)
```

#### Testing Repository Root Detection

```go
// Mock successful repository root detection
mockVCS.EXPECT().GetRepositoryRoot().Return("/path/to/repo", nil)

// Mock error when not in a repository
mockVCS.EXPECT().GetRepositoryRoot().Return("", errors.New("not in a repository"))
```

#### Testing Working Directory Status

```go
// Mock clean working directory
mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil)

// Mock dirty working directory
mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, nil)

// Mock error checking working directory
mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, errors.New("failed to check status"))
```

#### Testing VCS Identifier (Hash) Generation

```go
// Mock successful hash generation
mockVCS.EXPECT().GetVCSIdentifier(7).Return("abc1234", nil)

// Mock error in hash generation
mockVCS.EXPECT().GetVCSIdentifier(7).Return("", errors.New("failed to get hash"))
```

#### Testing Tag Operations

```go
// Mock successful tag creation
mockVCS.EXPECT().CreateTag("v1.0.0", "Release version 1.0.0").Return(nil)

// Mock tag creation error
mockVCS.EXPECT().CreateTag("v1.0.0", "Release").Return(errors.New("tag already exists"))

// Mock tag existence check
mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil)
mockVCS.EXPECT().TagExists("v2.0.0").Return(false, nil)
```

### Testing with VCS Registry

You can also test the VCS registry functionality by temporarily replacing the global registry:

```go
func TestWithRegistry(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Save original registry and restore after test
    originalRegistry := vcs.GetRegistry() // Note: This would need to be exposed
    defer func() {
        // Restore original registry
    }()

    // Create mock VCS
    mockVCS := mocks.NewMockVersionControlSystem(ctrl)
    mockVCS.EXPECT().Name().Return("mock").AnyTimes()
    mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()

    // Register mock VCS
    vcs.RegisterVCS(mockVCS)

    // Test code that uses vcs.GetActiveVCS() or vcs.GetVCS("mock")
    activeVCS := vcs.GetActiveVCS()
    // ... assertions
}
```

### Advanced Expectations

#### Using Times() for Call Count Verification

```go
// Expect method to be called exactly once
mockVCS.EXPECT().Name().Return("mock").Times(1)

// Expect method to be called at least once
mockVCS.EXPECT().IsRepository().Return(true).MinTimes(1)

// Expect method to be called any number of times
mockVCS.EXPECT().GetRepositoryRoot().Return("/repo", nil).AnyTimes()
```

#### Using InOrder() for Sequential Calls

```go
gomock.InOrder(
    mockVCS.EXPECT().IsRepository().Return(true),
    mockVCS.EXPECT().GetRepositoryRoot().Return("/repo", nil),
    mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil),
)
```

## Running Tests

To run the VCS tests including the mock tests:

```bash
go test ./internal/vcs -v
```

To run all tests in the project:

```bash
go test ./... -v
```

## Regenerating the Mock

If the `VersionControlSystem` interface changes, you can regenerate the mock using:

```bash
mockgen -source=internal/vcs/vcs.go -destination=internal/vcs/mock/mock_vcs.go -package=mock
```

Or manually update the mock file to match the interface changes.

## Best Practices

1. **Always use `ctrl.Finish()`** - This ensures all expected calls were made
2. **Set realistic expectations** - Mock behavior should match real VCS behavior
3. **Test both success and error scenarios** - Use mocks to test error handling
4. **Use `AnyTimes()` sparingly** - Prefer specific call counts when possible
5. **Clean up after tests** - Restore original state when modifying global registries

## Examples

See `vcs_test.go` for comprehensive examples of how to use the mock VCS in various testing scenarios.