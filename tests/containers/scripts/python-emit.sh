#!/bin/bash
set -e

echo "=== Python Emit Test ==="
echo "Testing versionator emit python"

# Generate version file
versionator emit python --output _version.py

echo "Generated _version.py:"
cat _version.py

# Run Python to import and verify
echo ""
echo "Running Python application..."
python main.py

echo ""
echo "=== PASS ==="
