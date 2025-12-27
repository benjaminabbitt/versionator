#!/bin/bash
set -e

echo "=== Python Patch Test ==="
echo "Testing versionator emit patch pyproject.toml"

echo "Before patch:"
cat pyproject.toml

# Patch the pyproject.toml
versionator emit patch

echo ""
echo "After patch:"
cat pyproject.toml

# Run Python to read version from pyproject.toml
echo ""
echo "Running Python application..."
python main.py

echo ""
echo "=== PASS ==="
