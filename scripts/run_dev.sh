#!/bin/bash

# Development Mode Runner for gomcp
# Runs server with hot reload using air

set -e

echo "Setting up development environment..."

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo "Installing air..."
    go install github.com/air-verse/air@latest
    
    # Add GOPATH/bin to PATH if not already there
    GOBIN=$(go env GOPATH)/bin
    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        export PATH=$PATH:$GOBIN
        echo "Added $GOBIN to PATH"
    fi
fi

# Check if port 8081 is in use
if lsof -Pi :8081 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "Port 8081 is already in use. Stopping existing server..."
    lsof -ti:8081 | xargs kill -9 2>/dev/null || true
    sleep 1
fi

# Clean tmp directory
rm -rf tmp/
mkdir -p tmp/

echo "Starting development server with hot reload..."
echo "Server will run on http://localhost:8081/mcp/sse"
echo "Files will auto-reload on changes"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Run air with Cursor-compatible settings
air
