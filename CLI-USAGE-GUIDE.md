# Run Script Service - CLI Usage Guide

## Overview
The run-script-service CLI provides comprehensive management for multiple scripts with scheduling, logging, and web interface capabilities.

## Build and Setup
```bash
# Build the application
go build -o run-script-service main.go

# Make binary executable
chmod +x run-script-service
```

## Core Commands

### Service Management
```bash
# Daemon service control
./run-script-service daemon start          # Start service in background
./run-script-service daemon stop           # Stop background service
./run-script-service daemon status         # Check service status
./run-script-service daemon restart        # Restart service
./run-script-service daemon logs           # View service logs

# Direct execution (foreground)
./run-script-service run                   # Run service in foreground

# Configuration
./run-script-service set-interval <interval>     # Set execution interval
./run-script-service show-config                 # Show current configuration
./run-script-service set-web-port <port>         # Set web server port

# Examples: 30s, 5m, 1h, 3600 (plain seconds)
```

### Script Management
```bash
# Add a new script
./run-script-service add-script --name=<script-name> --path=<script-path> --interval=<interval> [--max-log-lines=<lines>] [--timeout=<seconds>]

# List all scripts
./run-script-service list-scripts

# Enable a script
./run-script-service enable-script <script-name>

# Disable a script
./run-script-service disable-script <script-name>

# Remove a script
./run-script-service remove-script <script-name>

# Run a script once
./run-script-service run-script <script-name>
```

### Log Management
```bash
# View logs (all scripts)
./run-script-service logs --all

# View logs for specific script
./run-script-service logs --script=<script-name>

# View logs with filters
./run-script-service logs --script=<script-name> --limit=<number> --since=<timestamp>

# Clear logs for all scripts
./run-script-service clear-logs --all

# Clear logs for specific script
./run-script-service clear-logs --script=<script-name>
```

## Parameter Details

### add-script Required Parameters
- `--name=<script-name>`: Unique identifier for the script
- `--path=<script-path>`: Path to the executable script file
- `--interval=<interval>`: Execution interval (30s, 5m, 1h, or plain seconds)

### add-script Optional Parameters
- `--max-log-lines=<lines>`: Maximum log lines to keep (default: 100)
- `--timeout=<seconds>`: Script execution timeout (default: 0 = no timeout)

### Interval Format
- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `3600` - Plain number in seconds

## Example Usage

### Basic Script Management
```bash
# Create a test script file
echo '#!/bin/bash\necho "Hello from script 1"\ndate' > test1.sh
chmod +x test1.sh

# Add the script to service
./run-script-service add-script --name=test1 --path=./test1.sh --interval=5m

# List all scripts
./run-script-service list-scripts

# Run the script once
./run-script-service run-script test1

# View its logs
./run-script-service logs --script=test1
```

### Advanced Script Configuration
```bash
# Add script with custom log retention and timeout
./run-script-service add-script --name=long-task --path=./long-script.sh --interval=1h --max-log-lines=500 --timeout=300

# Disable script temporarily
./run-script-service disable-script long-task

# Re-enable when needed
./run-script-service enable-script long-task
```

### Web Interface Setup
```bash
# Set web port
./run-script-service set-web-port 8080

# Start service with web interface
./run-script-service run

# Access web interface at http://localhost:8080
```

## Web API Endpoints

### Script Management API
- `GET /api/scripts` - List all scripts
- `POST /api/scripts` - Add new script
- `PUT /api/scripts/{name}` - Update script
- `DELETE /api/scripts/{name}` - Remove script
- `POST /api/scripts/{name}/run` - Execute script once
- `POST /api/scripts/{name}/enable` - Enable script
- `POST /api/scripts/{name}/disable` - Disable script

### System API
- `GET /api/status` - System status
- `GET /api/logs` - Get logs
- `DELETE /api/logs` - Clear logs

### Real-time Monitoring
- `WebSocket /ws` - Real-time script execution status and system metrics

## File Structure
```
run-script-service-develop/
├── run-script-service          # Main executable
├── service_config.json         # Configuration file
├── run.log                     # Main service log
├── logs/                       # Script-specific logs
│   ├── script1.log
│   └── script2.log
└── web/static/                 # Web interface files
```

## Configuration File Format
The `service_config.json` contains:
```json
{
  "interval": 3600,
  "web_port": 8080,
  "scripts": [
    {
      "name": "test1",
      "path": "./test1.sh",
      "interval": 300,
      "enabled": true,
      "max_log_lines": 100,
      "timeout": 0
    }
  ]
}
```

## Important Notes
- Script paths must point to executable files, not direct commands
- Web interface and API endpoints are enabled by default
- WebSocket connections provide real-time monitoring of script execution
- All interval times are in seconds internally
- Log files are automatically rotated based on max_log_lines setting
- Scripts run in the directory containing the script file
