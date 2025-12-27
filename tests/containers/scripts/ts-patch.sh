#!/bin/bash
set -e

echo "=== TypeScript Patch Test ==="
echo "Testing versionator emit patch package.json"

echo "Before patch:"
cat package.json

# Patch the package.json
versionator emit patch

echo ""
echo "After patch:"
cat package.json

# Compile and run
echo ""
echo "Compiling TypeScript..."
npx tsc

echo ""
echo "Running TypeScript application..."
node dist/main.js

echo ""
echo "=== PASS ==="
