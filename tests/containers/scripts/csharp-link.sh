#!/bin/bash
set -e

echo "=== C# Link Test ==="
echo "Testing versionator emit build csharp"

# Get build flags
BUILD_FLAGS=$(versionator emit build csharp)
echo "Using build flags: $BUILD_FLAGS"

# Build with version injection
echo ""
echo "Building C# application with version injection..."
dotnet build $BUILD_FLAGS --configuration Release --verbosity quiet

# Run
echo ""
echo "Running C# application..."
OUTPUT=$(dotnet run --configuration Release --no-build)
echo "$OUTPUT"

# Verify output shows Revision: 0 (off by default)
if ! echo "$OUTPUT" | grep -q 'Revision: 0'; then
    echo "FAIL: Output should show Revision: 0"
    exit 1
fi

echo ""
echo "=== PASS ==="
