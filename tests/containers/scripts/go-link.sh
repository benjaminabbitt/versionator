#!/bin/bash
set -e

echo "=== Go Link Test ==="
echo "Testing versionator link injection"

# Get the ldflags from versionator
# The link command outputs the flags to use
LDFLAGS=$(versionator emit -t '-X main.Version={{MajorMinorPatch}}')

echo "Using ldflags: $LDFLAGS"

# Build with ldflags
echo "Building Go application with version injection..."
go build -ldflags "$LDFLAGS" -o app .

echo "Running application..."
./app

echo ""
echo "=== PASS ==="
