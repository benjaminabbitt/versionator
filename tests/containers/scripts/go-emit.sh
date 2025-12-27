#!/bin/bash
set -e

echo "=== Go Emit Test ==="
echo "Testing versionator emit go"

# Generate version file
mkdir -p version
versionator emit go --output version/version.go

echo "Generated version/version.go:"
cat version/version.go

# Build and run
echo ""
echo "Building Go application..."
go build -o app .

echo "Running application..."
./app

echo ""
echo "=== PASS ==="
