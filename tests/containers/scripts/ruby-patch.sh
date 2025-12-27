#!/bin/bash
set -e

echo "=== Ruby Patch Test ==="
echo "Testing versionator emit patch testgem.gemspec"

echo "Before patch:"
cat testgem.gemspec

# Patch the gemspec
versionator emit patch testgem.gemspec

echo ""
echo "After patch:"
cat testgem.gemspec

# Run the test
echo ""
echo "Running Ruby application..."
ruby main.rb

echo ""
echo "=== PASS ==="
