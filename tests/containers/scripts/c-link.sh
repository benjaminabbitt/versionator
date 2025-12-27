#!/bin/bash
set -e

echo "=== C Link Test ==="
echo "Testing versionator emit build c"

# Get build flags
BUILD_FLAGS=$(versionator emit build c --var VERSION)
echo "Using build flags: $BUILD_FLAGS"

# Compile with version injection
echo ""
echo "Compiling C application with version injection..."
gcc $BUILD_FLAGS -o app main.c

# Run
echo ""
echo "Running C application..."
./app

echo ""
echo "=== PASS ==="
