# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Service Management
```bash
# Install the systemd service
./service_control.sh install

# Service lifecycle commands
./service_control.sh start      # Start the service
./service_control.sh stop       # Stop the service
./service_control.sh restart    # Restart the service
./service_control.sh status     # Check service status
./service_control.sh logs       # View real-time service logs

# Configuration
./service_control.sh set-interval 30m    # Set execution interval (supports s/m/h suffixes)
./service_control.sh show-config         # Display current configuration

# Uninstall the service
./service_control.sh uninstall
```

### Manual Testing
```bash
# Test the main script manually
./run.sh

# Test the service daemon directly
python3 run_script_service.py run

# Make scripts executable if needed
chmod +x run.sh service_control.sh
```

## Architecture

This is a **systemd-based service manager** that executes shell scripts at configurable intervals. The architecture consists of:

1. **Service Daemon** (`run_script_service.py`):
   - Runs continuously as a systemd service
   - Executes `run.sh` at configured intervals (default: 1 hour)
   - Implements automatic log rotation (keeps last 100 lines)
   - Handles graceful shutdown via SIGTERM/SIGINT signals
   - Stores configuration in `service_config.json`

2. **Control Interface** (`service_control.sh`):
   - Provides user-friendly commands for service management
   - Handles systemd integration (install/uninstall/start/stop)
   - Manages configuration updates
   - No external dependencies - uses standard Linux tools

3. **Executable Scripts**:
   - `run.sh` - Main script executed by the service (currently runs Claude CLI operations)
   - Additional scripts: `run-fix-test-push.sh`, `run-migrate-typescript.sh`, `run-refactor.sh`
   - All scripts are executed with captured stdout/stderr logged to `run.log`

4. **Key Design Patterns**:
   - **No external Python dependencies** - uses only standard library
   - **Signal-based graceful shutdown** - properly handles service termination
   - **JSON configuration persistence** - survives service restarts
   - **Structured logging** with timestamps and exit codes
   - **Systemd integration** for reliability and automatic restart on failure

The service runs as user 'logan' and requires sudo for systemd operations. All file paths are absolute to ensure consistent execution regardless of working directory.