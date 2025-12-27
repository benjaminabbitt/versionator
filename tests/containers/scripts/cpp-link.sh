#!/bin/bash
set -e

echo "=== C++ Link Test ==="
echo "Testing versionator emit build cpp"

# Get build flags
BUILD_FLAGS=$(versionator emit build cpp --var VERSION)
echo "Using build flags: $BUILD_FLAGS"

# Compile with version injection
echo ""
echo "Compiling C++ application with version injection..."
g++ $BUILD_FLAGS -o app main.cpp

# Run
echo ""
echo "Running C++ application..."
./app

echo ""
echo "=== PASS ==="
