#!/bin/bash
set -e

echo "=== JavaScript Patch Test ==="
echo "Testing versionator emit patch package.json"

echo "Before patch:"
cat package.json

# Patch the package.json
versionator emit patch

echo ""
echo "After patch:"
cat package.json

# Run Node to read version from package.json
echo ""
echo "Running JavaScript application..."
node main.js

echo ""
echo "=== PASS ==="
