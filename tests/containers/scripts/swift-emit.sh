#!/bin/bash
set -e

echo "=== Swift Emit Test ==="
echo "Testing versionator emit swift"

# Generate version file
versionator emit swift --output version.swift

echo "Generated version.swift:"
cat version.swift

# Compile (include version.swift first, then main.swift)
echo ""
echo "Compiling Swift application..."
swiftc -o app version.swift main.swift

# Run
echo ""
echo "Running Swift application..."
./app

echo ""
echo "=== PASS ==="
