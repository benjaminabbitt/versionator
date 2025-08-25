#!/bin/bash

# Integration test script for Versionator
# This script tests basic version functionality outside the repository directory

set -e  # Exit on any error

echo "=== Versionator Integration Test ==="

# Check if binary path is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <path-to-versionator-binary>"
    echo "Example: $0 /tmp/test/bin/versionator-linux-amd64"
    exit 1
fi

BINARY_PATH="$1"

# Check if binary exists and is executable
if [ ! -x "$BINARY_PATH" ]; then
    echo "ERROR: Binary not found or not executable: $BINARY_PATH"
    exit 1
fi

echo "Using binary: $BINARY_PATH"

git init .

# Initialize VERSION file without BOM
printf "1.0.0" > VERSION

# Test version reading
echo "=== Testing version reading ==="
VERSION=$($BINARY_PATH version)
echo "Current version: $VERSION"

# Test version incrementing
echo "=== Testing version incrementing ==="
$BINARY_PATH patch increment
NEW_VERSION=$($BINARY_PATH version)
echo "New version: $NEW_VERSION"

# Verify version changed
if [ "$VERSION" = "$NEW_VERSION" ]; then
    echo "ERROR: Version did not increment"
    exit 1
fi

echo "SUCCESS: Version incremented from '$VERSION' to '$NEW_VERSION'"
echo "Integration test passed!"