#!/bin/bash
set -e

echo "=== C# Patch Test ==="
echo "Testing versionator emit patch *.csproj"

echo "Before patch:"
cat TestApp.csproj

# Patch the .csproj file
versionator emit patch TestApp.csproj

echo ""
echo "After patch:"
cat TestApp.csproj

# Build and run
echo ""
echo "Building C# application..."
dotnet build -c Release --nologo -v q

echo ""
echo "Running C# application..."
OUTPUT=$(dotnet run -c Release --no-build --nologo)
echo "$OUTPUT"

# Verify output shows Revision: 0 (off by default)
if ! echo "$OUTPUT" | grep -q 'Revision: 0'; then
    echo "FAIL: Output should show Revision: 0"
    exit 1
fi

echo ""
echo "=== PASS ==="
