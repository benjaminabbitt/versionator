#!/bin/bash
set -e

echo "=== Ruby Emit Test ==="
echo "Testing versionator emit ruby"

# Generate version file
versionator emit ruby --output version.rb

echo "Generated version.rb:"
cat version.rb

# Run
echo ""
echo "Running Ruby application..."
ruby main.rb

echo ""
echo "=== PASS ==="
