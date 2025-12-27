#!/bin/bash
set -e

echo "=== C Emit Test ==="
echo "Testing versionator emit c"

# Generate version header
versionator emit c --output version.h

echo "Generated version.h:"
cat version.h

# Compile
echo ""
echo "Compiling C application..."
gcc -o app main.c

# Run
echo ""
echo "Running C application..."
./app

echo ""
echo "=== PASS ==="
