# Go Development Guidelines

If `CLAUDE.project.md` exists, it supplements/overrides these guidelines.

---

## TDD
- Red → Green → Refactor
- Test naming: `Test<Action>_<Condition>_<Result>` (PascalCase)
- Verify test fails before implementing
- Unit tests: `*_test.go` co-located
- Integration tests: `tests/integration/` or build tags
- Acceptance tests: `tests/acceptance/features/*.feature` (godog)
- Mark slow tests: `//go:build slow`
- Build and run actual binaries for acceptance tests

## Tooling
- Test: `testing` (stdlib)
- Acceptance: godog (Gherkin)
- Lint: golangci-lint
- Format: gofmt/goimports
- Logging: zap
- Tasks: just
- Git hooks: lefthook

```justfile
TOP := `git rev-parse --show-toplevel`

test:
    go test {{TOP}}/...

lint:
    golangci-lint run {{TOP}}/...

fmt:
    gofmt -w {{TOP}}
    goimports -w {{TOP}}
```

```yaml
# lefthook.yml
pre-commit:
  parallel: true
  commands:
    lint:
      run: golangci-lint run
    format:
      run: gofmt -l . | grep -q . && exit 1 || exit 0
    test:
      run: go test ./...
```

## Structure
```
project/
├── cmd/app/main.go
├── internal/
│   ├── logmsg/messages.go
│   ├── errmsg/messages.go
│   └── module/
│       ├── module.go
│       └── module_test.go
├── tests/
│   ├── integration/
│   └── acceptance/features/
├── .devcontainer/
├── lefthook.yml
├── justfile
└── go.mod
```

## Error Constants
```go
// errmsg/messages.go
package errmsg

const DivideByZero = "cannot divide by zero"

// usage: return 0, errors.New(errmsg.DivideByZero)
// test: if err.Error() != errmsg.DivideByZero { t.Errorf(...) }
```

## Log Messages
```go
// logmsg/messages.go
package logmsg

const UserCreated = "user_created"

// usage: logger.Info(logmsg.UserCreated, zap.String("username", u))
```

## IoC Pattern
```go
// Testing constructor - accepts interfaces
func NewUserService(repo UserRepository, logger *zap.Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}

// Default factory - exclude from coverage
func NewUserServiceDefault(db *Database) *UserService { // nolint:unused
    logger, _ := zap.NewProduction()
    return NewUserService(NewSQLUserRepository(db), logger)
}
```

- Define interfaces for dependencies
- Testing constructor accepts all deps
- Default factory excluded from coverage (tested via integration)

## Error Handling
- Return `error` as last value
- Wrap: `fmt.Errorf("context: %w", err)`
- Check immediately, fail fast
- Specific error types
- Log with context, not just message
- Document expected errors

## Concurrency
- Channels for communication
- `sync.Mutex` for shared state
- `sync.WaitGroup` for goroutine coordination
- `context.Context` for cancellation
- Document lock ordering
- Minimal lock hold times

## Code Quality
- Fix all golangci-lint warnings
- Low coupling, high cohesion
- Small, focused functions
- Composition over inheritance
- Markers: `TODO`, `FIXME`, `NOTE`, `HACK`
- Coverage target: >80%
- No file creation unless necessary
- No progress docs in code

## Review Perspectives
- Domain: business logic correct?
- Concurrency: races? sync? deadlocks?
- Security: validation? injection? TOCTOU?
- Readability: clear? documented?
- Architecture: testable? separation?

## Pre-commit
- golangci-lint, gofmt, go test
- Fix all issues before commit
- `--no-verify` only for WIP feature branches

## Security
- No secrets in code (use env vars)
- Validate external input
- Parameterized queries
- Keep deps updated
- Least privilege

## Performance
- Profile before optimizing
- Document requirements in tests
- Use appropriate data structures
- State algorithmic complexity
- Consider allocation patterns

## Workarounds
When stuck, present options:
1. Fix properly (estimate effort)
2. Workaround (document trade-offs)
3. Disable test (document why)
4. Alternative approach

Document decision and reasoning.

## Communication
- State limitations immediately
- Ask when ambiguous
- Lead with key info
- Test before complete
- Direct, no sycophancy

## Questions to Ask
New feature: acceptance criteria? performance reqs? error cases? security? logging? dependencies? testing?
Design: dependencies? interfaces? context to log? error conditions?
Problems: fix vs workaround? pros/cons? tech debt impact?

## MCP Servers
- mcp-tasks: task tracking (`CLAUDE.todo.md`)
- Check availability before use
