#!/bin/bash

set -e

echo "=== AssetHub Setup Script ==="

# Check Go version
GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
if [ -z "$GO_VERSION" ]; then
    echo "Error: Go is not installed"
    exit 1
fi
echo "Go version: $GO_VERSION"

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Copy config if not exists
if [ ! -f configs/config.yaml ]; then
    if [ -f configs/config.example.yaml ]; then
        cp configs/config.example.yaml configs/config.yaml
        echo "Created configs/config.yaml from example"
    fi
fi

# Copy .env if not exists
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "Created .env from example"
    fi
fi

echo ""
echo "=== Setup Complete ==="
echo "Next steps:"
echo "  1. Edit configs/config.yaml with your settings"
echo "  2. Ensure PostgreSQL and Redis are running"
echo "  3. Run: make run"
