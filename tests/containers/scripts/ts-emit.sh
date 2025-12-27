#!/bin/bash
set -e

echo "=== TypeScript Emit Test ==="
echo "Testing versionator emit ts"

# Generate version file
versionator emit ts --output version.ts

echo "Generated version.ts:"
cat version.ts

# Install dependencies and build
echo ""
echo "Installing dependencies..."
npm install

echo ""
echo "Building TypeScript application..."
npm run build

echo ""
echo "Running TypeScript application..."
npm run start

echo ""
echo "=== PASS ==="
