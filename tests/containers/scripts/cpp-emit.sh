#!/bin/bash
set -e

echo "=== C++ Emit Test ==="
echo "Testing versionator emit cpp"

# Generate version header
versionator emit cpp --output version.hpp

echo "Generated version.hpp:"
cat version.hpp

# Compile
echo ""
echo "Compiling C++ application..."
g++ -o app main.cpp

# Run
echo ""
echo "Running C++ application..."
./app

echo ""
echo "=== PASS ==="
