# Run Script Service

A high-performance, configurable systemd service built in Go that executes scripts at regular intervals with automatic logging and log rotation.

## Features

- **Configurable Intervals**: Set execution frequency in seconds, minutes, or hours
- **Automatic Logging**: All script execution results are logged to `run.log`
- **Log Rotation**: Automatically keeps only the last 100 lines in the log file
- **Systemd Integration**: Full systemd service support with start/stop/restart capabilities
- **Easy Management**: Simple control script for all operations

## Quick Start

1. **Create your script from the example:**
   ```bash
   cp run.sh.example run.sh
   chmod +x run.sh
   # Edit run.sh with your custom commands
   ```

2. **Install the service:**
   ```bash
   ./service_control.sh install
   ```

3. **Set execution interval (optional):**
   ```bash
   ./service_control.sh set-interval 1h    # Run every hour (default)
   ./service_control.sh set-interval 30m   # Run every 30 minutes
   ./service_control.sh set-interval 120   # Run every 120 seconds
   ```

4. **Start the service:**
   ```bash
   ./service_control.sh start
   ```

## Usage

### Service Management

| Command | Description |
|---------|-------------|
| `./service_control.sh install` | Install and enable the systemd service |
| `./service_control.sh uninstall` | Stop, disable, and remove the service |
| `./service_control.sh start` | Start the service |
| `./service_control.sh stop` | Stop the service |
| `./service_control.sh restart` | Restart the service |
| `./service_control.sh status` | Show service status |
| `./service_control.sh logs` | Show real-time service logs |

### Configuration

| Command | Description |
|---------|-------------|
| `./service_control.sh set-interval <time>` | Set execution interval |
| `./service_control.sh show-config` | Display current configuration |

### Interval Format Examples

- `30` - 30 seconds
- `5m` - 5 minutes
- `2h` - 2 hours
- `3600` - 3600 seconds (1 hour)

## Files Structure

```
run-script-service/
├── main.go                   # Main service daemon (Go)
├── go.mod                    # Go module definition
├── run-script-service        # Compiled binary (auto-generated)
├── run.sh.example            # Example script template
├── run.sh                    # Your script to be executed (create from example)
├── run-script.service        # Systemd service file
├── service_control.sh        # Control script
├── service_config.json       # Configuration file (auto-generated)
├── run.log                   # Execution log (auto-generated)
├── plans/                    # Development plans (see Development section)
└── README.md                 # This file
```

## Log Format

The `run.log` file contains timestamped entries for each script execution:

```
[2024-01-15 14:30:00] Exit code: 0
STDOUT: Script executed successfully
--------------------------------------------------
[2024-01-15 15:30:00] Exit code: 0
STDOUT: Script executed successfully
--------------------------------------------------
```

## Customization

### Modifying the Script

Edit `run.sh` to customize what gets executed:

```bash
#!/bin/bash
# Your custom commands here
echo "$(date): Running my custom task"
# Add your commands below
```

### Configuration File

The service automatically creates `service_config.json` to store settings:

```json
{
  "interval": 3600
}
```

## Troubleshooting

### Check Service Status
```bash
./service_control.sh status
```

### View Recent Logs
```bash
./service_control.sh logs
```

### Manual Test
```bash
# Test the script manually
./run.sh

# Build the binary (if needed)
go build -o run-script-service main.go

# Test the service daemon manually
./run-script-service run
```

### Common Issues

1. **Permission denied**: Ensure scripts are executable with `chmod +x`
2. **Service won't start**: Check the systemd logs with `./service_control.sh logs`
3. **Script not found**: Verify `run.sh` exists and is in the correct directory

## Requirements

- Go 1.21+ (for building from source)
- systemd (Linux)
- sudo access for service installation

## Development

This project follows a structured development approach with detailed plans for each feature enhancement.

### Pre-commit Hooks

To maintain code quality, we use the [pre-commit](https://pre-commit.com/) framework for automatic checks:

```bash
# Install pre-commit (requires Python)
pip install pre-commit

# Install hooks for this repository
make install-precommit

# Optional: run on all files
make setup-precommit
```

The pre-commit hooks will automatically:
- Format code with `go fmt`
- Fix imports with `goimports`
- Run linting with `golangci-lint`
- Execute all tests
- Check for trailing whitespace and other common issues

### Development Commands

```bash
# Code quality
make format              # Format code with go fmt and goimports
make lint               # Run golangci-lint
make test               # Run all tests

# Pre-commit hooks setup
make install-precommit  # Install pre-commit framework hooks
make setup-precommit    # Install and run pre-commit on all files

# Development workflow
make test-watch         # Run tests with file watching (TDD)
make build             # Build binary
make ci                # Full CI pipeline (format + lint + test + build)
make clean             # Clean build artifacts
```

### Development Plans

The `plans/` directory contains detailed implementation plans for upcoming features:

| Plan | Feature | Status |
|------|---------|--------|
| [01-unit-testing.md](plans/01-unit-testing.md) | 單元測試基礎設施 | ✅ Completed |
| [02-tdd-workflow.md](plans/02-tdd-workflow.md) | TDD 開發流程 | ✅ Completed |
| [03-multi-script-support.md](plans/03-multi-script-support.md) | 多腳本支援 | 📋 Planned |
| [04-multi-log-management.md](plans/04-multi-log-management.md) | 多日誌管理 | 📋 Planned |
| [05-web-framework.md](plans/05-web-framework.md) | Web 框架設置 | 📋 Planned |
| [06-web-ui-basic.md](plans/06-web-ui-basic.md) | 基礎 Web UI | 📋 Planned |
| [07-web-editing.md](plans/07-web-editing.md) | Web 編輯功能 | 📋 Planned |

### Development Workflow

Each plan contains:
- **目標**: Clear feature objectives
- **前置需求**: Dependencies on other plans
- **實施步驟**: Step-by-step implementation guide
- **驗收標準**: Acceptance criteria for completion
- **相關檔案**: Files that will be created/modified
- **測試案例**: Test scenarios to validate

### Getting Started with Development

1. Choose a plan from the table above
2. Review the prerequisites
3. Follow the implementation steps using TDD approach
4. Ensure all acceptance criteria are met
5. Update the plan status in this README

### Contributing

When working on features:
- Follow the TDD workflow established in Plan 02
- Update documentation as you implement
- Ensure all tests pass before submitting changes
- Update the plan status table above

## License

This project is provided as-is for educational and practical use.
