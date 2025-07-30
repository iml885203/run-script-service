#!/bin/bash
# Upgrade script for run-script-service
# Updates from manual service file management to dynamic service generation

set -e  # Exit on any error

echo "🔄 Upgrading run-script-service to dynamic service generation..."
echo "================================================"

# Check if we're in the right directory
if [[ ! -f "main.go" || ! -f "service_control.sh" ]]; then
    echo "❌ Error: Please run this script from the run-script-service directory"
    exit 1
fi

# Step 1: Pull latest changes
echo "📥 Pulling latest changes from repository..."
git pull origin master
if [[ $? -ne 0 ]]; then
    echo "❌ Failed to pull latest changes. Please resolve any conflicts first."
    exit 1
fi

# Step 2: Rebuild binary
echo "🔨 Building updated Go binary..."
go build -o run-script-service main.go
if [[ $? -ne 0 ]]; then
    echo "❌ Failed to build binary"
    exit 1
fi

# Step 3: Check service status before uninstalling
echo "🔍 Checking current service status..."
SERVICE_WAS_RUNNING=false
if sudo systemctl is-active --quiet run-script; then
    SERVICE_WAS_RUNNING=true
    echo "ℹ️  Service is currently running - will restart after upgrade"
else
    echo "ℹ️  Service is not running"
fi

# Step 4: Uninstall old service
echo "🗑️  Uninstalling old service configuration..."
./service_control.sh uninstall

# Step 5: Install new service with dynamic generation
echo "⚙️  Installing new service with dynamic configuration..."
./service_control.sh install

# Step 6: Start service if it was running before
if [[ "$SERVICE_WAS_RUNNING" == "true" ]]; then
    echo "▶️  Starting service..."
    ./service_control.sh start
else
    echo "ℹ️  Service not started (wasn't running before upgrade)"
    echo "   Use './service_control.sh start' when ready"
fi

echo ""
echo "✅ Upgrade completed successfully!"
echo "================================================"
echo "🎉 Your service now uses dynamic path generation"
echo "📋 Key improvements:"
echo "   • Service file auto-generated with correct paths"
echo "   • Portable across different directories and users"
echo "   • No manual path configuration needed"
echo ""
echo "📖 Check status: ./service_control.sh status"
echo "📝 View logs:    ./service_control.sh logs"
echo "⚙️  Configuration: ./service_control.sh show-config"
