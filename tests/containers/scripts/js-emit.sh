#!/bin/bash
set -e

echo "=== JavaScript Emit Test ==="
echo "Testing versionator emit js"

# Generate version file
versionator emit js --output version.js

echo "Generated version.js:"
cat version.js

# Run JavaScript application
echo ""
echo "Running JavaScript application..."
node main.js

echo ""
echo "=== PASS ==="
