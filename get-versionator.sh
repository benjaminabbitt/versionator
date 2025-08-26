#!/bin/bash
# Download versionator binary (default for Unix-like systems)

set -e

echo "Creating bin directory..."
mkdir -p bin
echo "Detecting platform..."
echo "For Windows, use the get-versionator-windows script"
echo "Downloading versionator for Linux amd64 (default)..."
curl -L "https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-linux-amd64" -o "bin/versionator"
chmod +x "bin/versionator" 2>/dev/null || echo "Note: chmod not available on this platform"
echo "Successfully downloaded versionator for Linux amd64"
echo "Binary saved as: bin/versionator"