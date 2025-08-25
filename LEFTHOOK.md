# Lefthook Git Hooks for Versionator

This project uses [Lefthook](https://github.com/evilmartians/lefthook) to manage git hooks for maintaining code quality and consistency during development.

## What is Lefthook?

Lefthook is a fast and powerful Git hooks manager for Node.js, Ruby, Python, Go, and other languages. It allows you to run linters, tests, and other tools automatically before commits and pushes.

## Installation

### In Dev Container

Lefthook is automatically installed in the dev container. If you're using the provided `.devcontainer`, lefthook will be available immediately.

### Manual Installation

If you're not using the dev container, install lefthook using one of these methods:

#### Using Just (Recommended)
```bash
just lefthook-install
```

#### Manual Installation
```bash
# On Debian/Ubuntu
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.deb.sh' | sudo bash
sudo apt-get update && sudo apt-get install lefthook -y

# On macOS
brew install lefthook

# On other systems
# Download from https://github.com/evilmartians/lefthook/releases
```

## Setup

After installation, set up the git hooks:

```bash
# Install hooks
just lefthook-install
# or
lefthook install
```

## Configuration

The project includes a `lefthook.yml` configuration file with the following hooks:

### Pre-commit Hooks
- **fmt**: Automatically formats Go code using `go fmt`
- **imports**: Organizes imports using `goimports`
- **vet**: Runs static analysis using `go vet`
- **lint**: Runs `golangci-lint` on changed files

### Pre-push Hooks
- **test**: Runs all tests using `go test ./...`
- **test-race**: Runs tests with race detection
- **mod-tidy**: Ensures `go.mod` and `go.sum` are tidy

### Commit Message Hook
- **format**: Validates commit messages follow conventional commits format

## Usage

### Automatic Execution

Once installed, hooks will run automatically:
- **Pre-commit hooks** run before each commit
- **Pre-push hooks** run before each push
- **Commit-msg hooks** validate your commit messages

### Manual Execution

You can run hooks manually using just commands:

```bash
# Check lefthook status
just lefthook-status

# Run specific hook manually
just lefthook-run pre-commit
just lefthook-run pre-push

# Update/reinstall hooks
just lefthook-update

# Uninstall hooks
just lefthook-uninstall
```

Or use lefthook directly:

```bash
# Run all pre-commit hooks
lefthook run pre-commit

# Run specific command
lefthook run pre-commit fmt

# Skip hooks for a single commit (use sparingly)
git commit --no-verify -m "emergency fix"
```

## Hook Details

### Pre-commit: Code Formatting and Quality

1. **go fmt**: Ensures consistent Go code formatting
2. **goimports**: Automatically adds/removes imports and formats them
3. **go vet**: Catches common Go programming errors
4. **golangci-lint**: Comprehensive linting with multiple analyzers

### Pre-push: Testing and Validation

1. **go test**: Runs all tests to ensure functionality
2. **go test -race**: Detects race conditions in concurrent code
3. **go mod tidy**: Ensures module dependencies are clean

### Commit Message: Conventional Commits

Enforces conventional commit format:
```
type(scope): description

Types: feat, fix, docs, style, refactor, test, chore, ci, build, perf
Examples:
- feat(auth): add user authentication
- fix(api): resolve null pointer exception
- docs(readme): update installation instructions
```

## Customization

To modify the hooks, edit `lefthook.yml`:

```yaml
pre-commit:
  commands:
    your-custom-hook:
      glob: "*.go"
      run: your-command
```

After making changes, update the hooks:
```bash
just lefthook-update
```

## Troubleshooting

### Common Issues

1. **Hooks not running**: Ensure you've run `lefthook install`
2. **Permission denied**: Check that lefthook binary has execute permissions
3. **golangci-lint not found**: Ensure golangci-lint is installed (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

### Debugging

```bash
# Check lefthook status
lefthook version
lefthook install --force

# View hook logs
cat .lefthook/output.log

# Test specific hook
lefthook run pre-commit --verbose
```

### Skipping Hooks

Sometimes you may need to skip hooks:

```bash
# Skip all hooks for one commit
git commit --no-verify -m "emergency fix"

# Skip specific hooks (set in lefthook.yml)
LEFTHOOK_EXCLUDE=lint git commit -m "work in progress"
```

## Integration with Development Workflow

1. **Initial setup**: Run `just dev-setup` which includes lefthook installation
2. **Daily development**: Hooks run automatically, ensuring code quality
3. **CI/CD**: The same checks run in your CI pipeline for consistency
4. **Team collaboration**: All team members have the same code quality standards

## Best Practices

1. **Keep hooks fast**: Long-running hooks slow down development
2. **Fix hook failures**: Don't skip hooks unless absolutely necessary
3. **Update regularly**: Keep lefthook and linters up to date
4. **Test hooks**: Verify hook behavior after configuration changes
5. **Document exceptions**: If you must skip hooks, document why

## Resources

- [Lefthook GitHub Repository](https://github.com/evilmartians/lefthook)
- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [golangci-lint Documentation](https://golangci-lint.run/)