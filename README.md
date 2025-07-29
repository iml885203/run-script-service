# Run Script Service

A configurable systemd service that executes a script at regular intervals with automatic logging and log rotation.

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
├── run.sh.example            # Example script template
├── run.sh                    # Your script to be executed (create from example)
├── run_script_service.py     # Main service daemon
├── run-script.service        # Systemd service file
├── service_control.sh        # Control script
├── service_config.json       # Configuration file (auto-generated)
├── run.log                   # Execution log (auto-generated)
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

# Test the service daemon manually
python3 run_script_service.py run
```

### Common Issues

1. **Permission denied**: Ensure scripts are executable with `chmod +x`
2. **Service won't start**: Check the systemd logs with `./service_control.sh logs`
3. **Script not found**: Verify `run.sh` exists and is in the correct directory

## Requirements

- Python 3.6+
- systemd (Linux)
- sudo access for service installation

## License

This project is provided as-is for educational and practical use.