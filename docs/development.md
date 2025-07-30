# Development Guide

This document outlines the development workflow and standards for the run-script-service project.

## TDD Workflow

This project follows Test-Driven Development (TDD) principles using the Red-Green-Refactor cycle:

### 1. Red Phase
Write a failing test that describes the desired functionality:
```bash
make test  # Should fail
```

### 2. Green Phase
Write the minimal code to make the test pass:
```bash
make test  # Should pass
```

### 3. Refactor Phase
Improve the code while keeping tests passing:
```bash
make test  # Should still pass
```

## Development Commands

### Testing
```bash
# Run all tests
make test

# Run tests with file watching (TDD mode)
make tdd

# Generate coverage report
make coverage
```

### Code Quality
```bash
# Format code
make format

# Run linters
make lint

# Full CI pipeline
make ci
```

### Building
```bash
# Build binary
make build

# Clean build artifacts
make clean
```

## Code Standards

### Testing
- All public functions must have tests
- Use table-driven tests for multiple scenarios
- Use mock interfaces for external dependencies
- Aim for >80% code coverage

### Code Style
- Follow Go conventions (gofmt, goimports)
- Use descriptive variable names
- Keep functions small and focused
- Add comments for exported functions

### Commit Messages
Use conventional commit format:
```
type(scope): description

Examples:
feat: add interval validation
fix: handle graceful shutdown
test: add service integration tests
docs: update development guide
```

### Pre-commit Checklist
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make format`)
- [ ] No linter warnings (`make lint`)
- [ ] Coverage is maintained
- [ ] Documentation is updated

## Project Structure

```
.
├── main.go              # Application entry point
├── main_test.go         # Main function tests
├── service/             # Core service logic
│   ├── config.go        # Configuration management
│   ├── executor.go      # Script execution
│   └── service.go       # Main service implementation
├── mocks/               # Test mocks and interfaces
├── testdata/            # Test data files
└── docs/                # Documentation
```

## Dependencies

### Development Tools
- `golangci-lint`: Code linting
- `goimports`: Import formatting
- `entr`: File watching for TDD mode

### Installation
```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Install entr (Ubuntu/Debian)
sudo apt-get install entr
```

## CI/CD

The project uses GitHub Actions for continuous integration:
- Runs on every push and pull request
- Executes full CI pipeline (`make ci`)
- Generates coverage reports
- Uploads coverage to Codecov

## Troubleshooting

### Common Issues

**Tests failing locally but passing in CI:**
- Check Go version compatibility
- Ensure all dependencies are installed
- Run `make clean && make ci`

**Linter errors:**
- Run `make format` to fix formatting issues
- Check `.golangci.yml` for specific rule configurations
- Some issues may require manual fixes

**Coverage drops:**
- Add tests for new functionality
- Remove dead code
- Check coverage report: `make coverage && open coverage.html`
