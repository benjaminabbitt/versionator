#!/bin/bash
set -e

echo "=== Python Setuptools Patch Test ==="
echo "Testing versionator emit patch setup.py"

echo "Before patch:"
cat setup.py

# Patch the setup.py
versionator emit patch

echo ""
echo "After patch:"
cat setup.py

# Run Python to read version from setup.py
echo ""
echo "Running Python application..."
python main.py

echo ""
echo "=== PASS ==="
