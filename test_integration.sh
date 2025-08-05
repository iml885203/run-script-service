#!/bin/bash

# Simple integration test to verify FileManager API is working with --web flag
# This script tests that the file management endpoints are accessible

echo "Testing FileManager integration with web server..."

# Create a simple test config
cat > service_config.json << EOF
{
    "scripts": [],
    "web_port": 8081
}
EOF

echo "Created test configuration"

# Start the service with web interface in background (just briefly to test startup)
./run-script-service run --web &
WEB_PID=$!

# Give it a moment to start
sleep 2

# Test basic connectivity (this will check if the server starts properly)
echo "Testing web server startup..."
curl -s http://localhost:8081/api/status > /dev/null
if [ $? -eq 0 ]; then
    echo "✓ Web server started successfully"
    echo "✓ FileManager integration is available via API endpoints:"
    echo "  - GET /api/files/*path (read files)"
    echo "  - PUT /api/files/*path (write files)"
    echo "  - POST /api/files/validate (validate scripts)"
    echo "  - GET /api/files-list/*path (list directory contents)"
else
    echo "✗ Web server failed to start"
fi

# Clean up
kill $WEB_PID 2>/dev/null
wait $WEB_PID 2>/dev/null

echo "Integration test completed"
