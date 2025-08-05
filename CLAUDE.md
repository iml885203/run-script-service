# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Service Management
```bash
# Service lifecycle commands
./run-script-service daemon start      # Start the service in background
./run-script-service daemon stop       # Stop the service
./run-script-service daemon status     # Check service status
./run-script-service daemon restart    # Restart the service
./run-script-service daemon logs       # View service logs

# Configuration
./run-script-service set-interval 30m    # Set execution interval (supports s/m/h suffixes)
./run-script-service show-config         # Display current configuration
```

### Manual Testing
```bash
# Test the main script manually
./run.sh

# Test the service daemon directly
./run-script-service run

# Build the Go binary (if needed)
go build -o run-script-service main.go

# Make scripts executable if needed
chmod +x run.sh run-script-service
```

### Development Commands (TDD Workflow)
```bash
# TDD Cycle - MUST follow Red-Green-Refactor cycle
make tdd              # Start file watching mode, auto-run tests
make test             # Run all tests
make coverage         # Generate test coverage report

# Code Quality Checks
make format           # Format code (gofmt + goimports)
make lint             # Run golangci-lint checks
make ci               # Complete CI pipeline (format + lint + test + build)

# Build & Clean
make build            # Build Go binary
make clean            # Clean artifacts
```

## Development Workflow

### 🚨 Mandatory TDD Process

**All code changes MUST follow Test-Driven Development (TDD) cycle**:

1. **🔴 Red Phase**: Write failing test first
   ```bash
   # Write test, ensure it fails
   make test  # New test should fail
   git commit -m "test: add failing test for [feature]"
   ```

2. **🟢 Green Phase**: Write minimal implementation to pass tests
   ```bash
   # Implement minimal functionality
   make test  # All tests should pass
   git commit -m "feat: implement minimal [feature]"
   ```

3. **🔵 Refactor Phase**: Refactor and improve code
   ```bash
   # Improve code quality while keeping tests passing
   make ci    # Complete checks
   git commit -m "refactor: improve [feature] implementation"
   ```

### 📋 Development Checklist

Every commit must satisfy:
- [ ] Follow Red-Green-Refactor cycle
- [ ] All tests pass (`make test`)
- [ ] Code coverage >80%
- [ ] Pass linting checks (`make lint`)
- [ ] Small incremental changes (<100 lines)
- [ ] Use Conventional Commits format

### 📚 Related Documentation
- **[docs/development.md](docs/development.md)** - Complete TDD workflow and development guide
- **[Makefile](Makefile)** - Development command definitions

## Architecture

This is a **Go-based service manager** that executes shell scripts at configurable intervals. The architecture consists of:

1. **Service Daemon** (`main.go` compiled to `run-script-service`):
   - Runs continuously as a background process
   - Executes scripts at configured intervals (default: 1 hour)
   - Implements automatic log rotation (keeps last 100 lines)
   - Handles graceful shutdown via SIGTERM/SIGINT signals
   - Stores configuration in `service_config.json`
   - Includes built-in web interface for monitoring and control

2. **Integrated CLI** (built into `run-script-service`):
   - Provides daemon management commands (start/stop/status/restart/logs)
   - Handles background process management with PID files
   - Manages configuration updates
   - No external dependencies - single Go binary

3. **Executable Scripts**:
   - `run.sh` - Main script executed by the service (currently runs Claude CLI operations)
   - Additional scripts: `run-fix-test-push.sh`, `run-migrate-typescript.sh`, `run-refactor.sh`
   - All scripts are executed with captured stdout/stderr logged to `run.log`

4. **Key Design Patterns**:
   - **No external dependencies** - uses only Go standard library
   - **Signal-based graceful shutdown** - properly handles service termination
   - **JSON configuration persistence** - survives service restarts
   - **Structured logging** with timestamps and exit codes
   - **Built-in web interface** - real-time monitoring and control via HTTP
   - **Cross-platform compatibility** - single binary deployment

The service runs as a background process and can be managed via the integrated CLI commands. PID files are used for reliable process management. All file paths are absolute to ensure consistent execution regardless of working directory.

## New Usage Pattern (No External Scripts)

```bash
# Start service in background
./run-script-service daemon start

# Check status
./run-script-service daemon status

# View service logs
./run-script-service daemon logs

# Stop service
./run-script-service daemon stop

# Restart service
./run-script-service daemon restart
```
