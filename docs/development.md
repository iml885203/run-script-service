# Development Guide

This document outlines the **mandatory development workflow and standards** for the run-script-service project.

## 🚨 Mandatory TDD Workflow

**All code changes MUST follow Test-Driven Development (TDD) principles using the Red-Green-Refactor cycle.**

### Prohibited Development Patterns
- ❌ Writing implementation code before tests
- ❌ Large feature commits (>500 lines of changes)
- ❌ Skipping any TDD phase
- ❌ Functional code without corresponding tests

### Required Development Pattern
- ✅ Red-Green-Refactor cycle
- ✅ Small incremental changes (<100 lines per commit)
- ✅ Test-first approach
- ✅ Every commit includes corresponding tests

## 🔄 TDD Cycle Process

### Phase 1: 🔴 Red (Write Failing Test)

```bash
# 1. Create feature branch
git checkout -b feature/descriptive-name

# 2. Write failing test
# Add test case in *_test.go file

# 3. Run tests to ensure failure
make test
# Expected: New test fails, existing tests pass

# 4. Commit failing test
git add .
git commit -m "test: add failing test for [feature description]

Red phase: Test should fail because functionality is not yet implemented
- Add Test[FunctionName] test case
- Define expected behavior and interface"
```

### Phase 2: 🟢 Green (Minimal Implementation)

```bash
# 1. Write minimal code to make test pass
# Only write just enough code to pass the test, no over-engineering

# 2. Run tests to ensure they pass
make test
# Expected: All tests pass

# 3. Commit minimal implementation
git add .
git commit -m "feat: implement minimal [feature description]

Green phase: Minimal implementation to make tests pass
- Implement [FunctionName] basic functionality
- All tests passing"
```

### Phase 3: 🔵 Refactor (Improve Code Quality)

```bash
# 1. Improve code quality
# Enhance design, performance, readability while keeping tests passing

# 2. Continuously run tests
make tdd  # Start file watching mode

# 3. Run full CI checks
make ci
# Expected: Formatting, linting, tests all pass

# 4. Commit refactoring improvements
git add .
git commit -m "refactor: improve [feature description] implementation

Refactor phase: Optimize code quality
- Improve [specific improvements]
- Maintain all tests passing
- Code coverage: X%"
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

## 📋 Commit Checklist

Every commit MUST satisfy the following conditions:

### 🔴 Red Phase Commit
- [ ] Added new test cases
- [ ] New tests fail as expected
- [ ] Existing tests still pass
- [ ] Commit message clearly explains test intent

### 🟢 Green Phase Commit
- [ ] Implemented minimal code to make tests pass
- [ ] All tests pass (`make test`)
- [ ] No over-engineering or extra features
- [ ] Commit message describes implementation

### 🔵 Refactor Phase Commit
- [ ] Code quality improvements made
- [ ] All tests still pass
- [ ] Passes all CI checks (`make ci`)
- [ ] Coverage has not decreased

## 📊 Quality Standards

### Test Coverage
- 🎯 **Target**: >85% coverage
- 🚨 **Minimum**: >80% coverage
- 📈 **Trend**: Coverage must not decrease

### Code Quality
- ✅ Pass `golangci-lint` checks
- ✅ Pass `go vet` checks
- ✅ Conform to `gofmt` formatting
- ✅ Correct import ordering (`goimports`)

### Commit Quality
- 📝 Use [Conventional Commits](https://www.conventionalcommits.org/) format
- 🔍 Atomic commits (single feature change)
- 📏 Small incremental changes (<100 lines)

## 🚨 Violation Handling

### Common Violations
1. **Large feature commits**: PRs with >500 lines of changes
2. **Missing tests**: Functional code without corresponding tests
3. **Skipping Red phase**: Committing passing tests with implementation
4. **Coverage decrease**: New code causing overall coverage to drop

### Enforcement Actions
- 🔙 **Require rework**: Violating PRs must be redone using TDD process
- 📚 **Education**: Provide TDD training and guidance
- 🔒 **Mandatory checks**: CI enforces TDD compliance

## 📖 TDD Example

### Example: Adding Script Path Validation

#### 🔴 Red Phase
```go
// config_test.go
func TestServiceConfig_ValidateScriptPath(t *testing.T) {
    config := &ServiceConfig{
        Scripts: []ScriptConfig{
            {Name: "test", Path: "../invalid/path.sh"},
        },
    }

    err := config.Validate()
    if err == nil {
        t.Error("Expected validation error for invalid script path")
    }

    if !strings.Contains(err.Error(), "invalid script path") {
        t.Errorf("Expected 'invalid script path' error, got: %v", err)
    }
}
```

```bash
make test  # Fails: functionality not implemented
git commit -m "test: add script path validation test

Red phase: Test should fail because ValidateScriptPath not implemented
- Add TestServiceConfig_ValidateScriptPath test
- Check invalid paths should return error"
```

#### 🟢 Green Phase
```go
// config.go
func (c *ServiceConfig) Validate() error {
    for _, script := range c.Scripts {
        if strings.Contains(script.Path, "..") {
            return fmt.Errorf("invalid script path: %s", script.Path)
        }
    }
    return nil
}
```

```bash
make test  # Passes: minimal implementation
git commit -m "feat: implement basic script path validation

Green phase: Minimal implementation checking for '..'
- Implement ServiceConfig.Validate() method
- Check script paths don't contain '..'
- All tests passing"
```

#### 🔵 Refactor Phase
```go
// config.go - improved implementation
func (c *ServiceConfig) Validate() error {
    for _, script := range c.Scripts {
        if err := validateScriptPath(script.Path); err != nil {
            return fmt.Errorf("script %s: %w", script.Name, err)
        }
    }
    return nil
}

func validateScriptPath(path string) error {
    // More comprehensive path validation logic
    if !filepath.IsAbs(path) && strings.Contains(path, "..") {
        return fmt.Errorf("invalid script path: %s", path)
    }
    return nil
}
```

```bash
make ci    # Passes: code quality checks
git commit -m "refactor: improve script path validation

Refactor phase: Extract validateScriptPath function
- Add more comprehensive path validation logic
- Improve error messages with script names
- Maintain all tests passing
- Code coverage: 87%"
```

## Code Standards

### Testing
- All public functions must have tests
- Use table-driven tests for multiple scenarios
- Use mock interfaces for external dependencies
- Aim for >85% code coverage

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
refactor: extract validation logic
```

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
