#!/bin/bash
set -e

echo "=== C# Emit Test ==="
echo "Testing versionator emit csharp"

# Generate version file
versionator emit csharp --output VersionInfo.cs

echo "Generated VersionInfo.cs:"
cat VersionInfo.cs

# Verify Revision is 0 (off by default, like prerelease)
if ! grep -q 'public const int Revision = 0;' VersionInfo.cs; then
    echo "FAIL: Revision should be 0 by default"
    exit 1
fi

# Verify AssemblyVersion uses 4-component format with Revision
if ! grep -q 'public const string AssemblyVersion = "1.2.3.0";' VersionInfo.cs; then
    echo "FAIL: AssemblyVersion should be 1.2.3.0 (4-component format)"
    exit 1
fi

# Build and run
echo ""
echo "Building C# application..."
dotnet build -c Release --nologo -v q

echo ""
echo "Running C# application..."
OUTPUT=$(dotnet run -c Release --no-build --nologo)
echo "$OUTPUT"

# Verify output contains expected values
if ! echo "$OUTPUT" | grep -q 'Revision: 0'; then
    echo "FAIL: Output should show Revision: 0"
    exit 1
fi

echo ""
echo "=== PASS ==="
