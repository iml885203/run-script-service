#!/bin/bash

SERVICE_NAME="run-script"
SERVICE_FILE="./run-script.service"
GO_BINARY="./run-script-service"

show_help() {
    echo "Run Script Service Control"
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  install           Install the systemd service"
    echo "  uninstall         Uninstall the systemd service"
    echo "  start             Start the service"
    echo "  stop              Stop the service"
    echo "  restart           Restart the service"
    echo "  status            Show service status"
    echo "  logs              Show service logs"
    echo "  set-interval <n>  Set execution interval (in seconds, minutes with 'm', hours with 'h')"
    echo "  show-config       Show current configuration"
    echo ""
    echo "Examples:"
    echo "  $0 set-interval 30        # Run every 30 seconds"
    echo "  $0 set-interval 5m        # Run every 5 minutes"
    echo "  $0 set-interval 2h        # Run every 2 hours"
}

convert_to_seconds() {
    local input="$1"
    local number="${input%[mh]}"
    local unit="${input: -1}"

    if [[ "$input" =~ ^[0-9]+$ ]]; then
        # Pure number, treat as seconds
        echo "$input"
    elif [[ "$unit" == "m" && "$number" =~ ^[0-9]+$ ]]; then
        # Minutes
        echo $((number * 60))
    elif [[ "$unit" == "h" && "$number" =~ ^[0-9]+$ ]]; then
        # Hours
        echo $((number * 3600))
    else
        echo "Invalid format. Use: number (seconds), number+m (minutes), or number+h (hours)" >&2
        return 1
    fi
}

case "$1" in
    install)
        echo "Generating service file..."
        "$GO_BINARY" generate-service
        if [ $? -ne 0 ]; then
            echo "Failed to generate service file"
            exit 1
        fi

        echo "Installing systemd service..."
        sudo cp "$SERVICE_FILE" /etc/systemd/system/
        sudo systemctl daemon-reload
        sudo systemctl enable "$SERVICE_NAME"
        echo "Service installed and enabled"
        ;;

    uninstall)
        echo "Uninstalling systemd service..."
        sudo systemctl stop "$SERVICE_NAME" 2>/dev/null
        sudo systemctl disable "$SERVICE_NAME" 2>/dev/null
        sudo rm -f "/etc/systemd/system/$SERVICE_NAME.service"
        sudo systemctl daemon-reload
        echo "Service uninstalled"
        ;;

    start)
        echo "Starting service..."
        sudo systemctl start "$SERVICE_NAME"
        sleep 1
        sudo systemctl status "$SERVICE_NAME" --no-pager -l
        ;;

    stop)
        echo "Stopping service..."
        sudo systemctl stop "$SERVICE_NAME"
        echo "Service stopped"
        ;;

    restart)
        echo "Restarting service..."
        sudo systemctl restart "$SERVICE_NAME"
        sleep 1
        sudo systemctl status "$SERVICE_NAME" --no-pager -l
        ;;

    status)
        sudo systemctl status "$SERVICE_NAME" --no-pager -l
        ;;

    logs)
        sudo journalctl -u "$SERVICE_NAME" -f
        ;;

    set-interval)
        if [ -z "$2" ]; then
            echo "Please specify interval. Examples: 30, 5m, 2h"
            exit 1
        fi

        seconds=$(convert_to_seconds "$2")
        if [ $? -ne 0 ]; then
            exit 1
        fi

        echo "Setting interval to $seconds seconds..."
        "$GO_BINARY" set-interval "$2"

        # Restart service if it's running
        if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
            echo "Restarting service to apply new interval..."
            sudo systemctl restart "$SERVICE_NAME"
        fi
        ;;

    show-config)
        "$GO_BINARY" show-config
        ;;

    help|--help|-h)
        show_help
        ;;

    *)
        echo "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
